package migration

import (
	"encoding/binary"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/storage/db"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func TestMigration_Migrate(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "store", uuid.New().String())
	store, err := db.NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(dir)
	}()

	key := []byte{byte(storage.KeyPrefixVersion)}
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, 1)
	if err := store.Put(key, buf[:n]); err != nil {
		t.Fatal(err)
	}

	blk := mock.StateBlockWithoutWork()
	k, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock, blk.GetHash())
	v, _ := blk.Serialize()
	if err := store.Put(k, v); err != nil {
		t.Fatal(err)
	}
	m1 := MigrationV1ToV15{}
	if err := m1.Migrate(store); err != nil {
		t.Fatal(err)
	}

	acc1 := mock.AccountMeta(mock.Address())
	acc1.Tokens[0].Type = config.ChainToken()
	acc2 := mock.AccountMeta(mock.Address())
	acc2.Tokens[0].Type = config.ChainToken()

	acc1K, _ := storage.GetKeyOfParts(storage.KeyPrefixAccount, acc1.Address)
	acc2K, _ := storage.GetKeyOfParts(storage.KeyPrefixAccount, acc2.Address)
	acc1V, _ := acc1.Serialize()
	acc2V, _ := acc2.Serialize()
	if err := store.Put(acc1K, acc1V); err != nil {
		t.Fatal(err)
	}
	if err := store.Put(acc2K, acc2V); err != nil {
		t.Fatal(err)
	}

	frontier := &types.Frontier{
		OpenBlock:   mock.Hash(),
		HeaderBlock: mock.Hash(),
	}
	frontierK, err := storage.GetKeyOfParts(storage.KeyPrefixFrontier, frontier.HeaderBlock)
	if err != nil {
		t.Fatal(err)
	}
	frontierV := frontier.OpenBlock[:]
	if err := store.Put(frontierK, frontierV); err != nil {
		t.Fatal(err)
	}
	migrations := []Migration{MigrationV15ToV16{}}
	if err := Upgrade(migrations, store); err != nil {
		t.Fatal(err)
	}
}

func TestMigration_MigrateV15ToV16(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "store", uuid.New().String())
	store, err := db.NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(dir)
	}()

	key := []byte{byte(storage.KeyPrefixVersion)}
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, 15)
	if err := store.Put(key, buf[:n]); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		pendingKey := &types.PendingKey{
			Address: mock.Address(),
			Hash:    mock.Hash(),
		}
		pendingInfo := &types.PendingInfo{
			Source: mock.Address(),
			Type:   mock.Hash(),
			Amount: types.Balance{Int: big.NewInt(100)},
		}
		kBytes, err := pendingKey.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}
		pk := make([]byte, 0)
		pk = append(pk, byte(storage.KeyPrefixPending))
		pk = append(pk, kBytes...)
		iBytes, err := pendingInfo.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Put(pk, iBytes); err != nil {
			t.Fatal(err)
		}
	}
	migrations := []Migration{MigrationV15ToV16{}}
	if err := Upgrade(migrations, store); err != nil {
		t.Fatal(err)
	}
}

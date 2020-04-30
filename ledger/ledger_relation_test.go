package ledger

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

// StructA (store to badger)  -->  StructB (store to relation)

type StructA struct {
	Hash    types.Hash    `json:"hash"`
	Address types.Address `json:"address"`
	Typ     int           `json:"typ"`
	Time    int64         `json:"time"`
}

func (a *StructA) Serialize() ([]byte, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (a *StructA) Deserialize(text []byte) error {
	return json.Unmarshal(text, a)
}

func (a *StructA) ConvertToSchema() ([]types.Schema, error) {
	return []types.Schema{
		&StructB{
			Hash:    a.Hash.String(),
			Address: a.Address.String(),
		},
	}, nil
}

type StructB struct {
	Hash    string `db:"hash" typ:"char(64)" key:""`
	Address string `db:"address" typ:"char(64)"`
}

func (b *StructB) IdentityID() string {
	return "structb"
}

func (b *StructB) DeleteKey() string {
	panic("implement me")
}

func TestLedger_Relation(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	if err := l.RegisterInterface(new(StructA), []types.Schema{new(StructB)}); err != nil {
		t.Fatal(err)
	}

	sa := &StructA{
		Hash:    mock.Hash(),
		Address: mock.Address(),
		Typ:     1,
		Time:    time.Now().Unix(),
	}

	key, _ := types.HashBytes(sa.Hash[:], sa.Address[:])
	if err := l.cache.Put(key[:], sa); err != nil {
		t.Fatal(err)
	}

	l.Flush() // dump StructA to badger and StructB to relation

	var s []StructB
	sql := fmt.Sprintf("select * from StructB")
	err := l.SelectRelation(&s, sql)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 1 {
		t.Fatal()
	}
	if s[0].Hash != sa.Hash.String() {
		t.Fatal()
	}

	sb := StructB{}
	sql = fmt.Sprintf("select * from StructB")
	err = l.GetRelation(&sb, sql)
	if err != nil {
		t.Fatal(err)
	}
	if sb.Hash != sa.Hash.String() {
		t.Fatal()
	}
}

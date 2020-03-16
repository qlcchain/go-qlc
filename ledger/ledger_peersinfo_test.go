package ledger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
)

func setupPeersInfoTestCase(t *testing.T) (func(t *testing.T), *Ledger) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	l := NewLedger(cm.ConfigFile)

	return func(t *testing.T) {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l
}

func generatePeersInfo() *types.PeerInfo {
	peerID := random.RandomHexString(46)
	return &types.PeerInfo{
		PeerID:  peerID,
		Address: "/ip4/192.168.80.1/tcp/9001",
		Version: "v1.3.1",
		Rtt:     1,
	}
}

func TestLedger_AddPeerInfo(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	pi := generatePeersInfo()
	err := l.AddPeerInfo(pi)
	if err != nil {
		t.Fatal(err)
	}
	err = l.AddPeerInfo(pi)
	if err == nil {
		t.Fatal(err)
	}
}

func TestLedger_GetPeerInfo(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	pi := generatePeersInfo()
	err := l.AddPeerInfo(pi)
	if err != nil {
		t.Fatal(err)
	}
	pi2, err := l.GetPeerInfo(pi.PeerID)
	if err != nil {
		t.Fatal(err)
	}
	if pi.PeerID != pi2.PeerID {
		t.Fatal("peerID mismatch")
	}
	if pi.Address != pi2.Address {
		t.Fatal("Address mismatch")
	}
	if pi.Version != pi2.Version {
		t.Fatal("Version mismatch")
	}
	if pi.Rtt != pi2.Rtt {
		t.Fatal("Rtt mismatch")
	}

	// key not found
	if _, err := l.GetPeerInfo(random.RandomHexString(46)); err != ErrPovHeaderNotFound {
		t.Fatal()
	}

	// deserialize error
	key := random.RandomHexString(46)
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPeerInfo, []byte(key))
	if err != nil {
		t.Fatal()
	}
	d1 := make([]byte, 0)
	_ = random.Bytes(d1)
	if err := l.store.Put(k, d1); err != nil {
		t.Fatal(err)
	}
	if r, err := l.GetPeerInfo(key); err == nil {
		t.Fatal(err, r)
	}

	if err := l.GetPeersInfo(func(info *types.PeerInfo) error {
		return nil
	}); err == nil {
		t.Fatal(err)
	}
}

func TestLedger_GetPeersInfo(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	pi := generatePeersInfo()
	err := l.AddPeerInfo(pi)
	if err != nil {
		t.Fatal(err)
	}
	pi2 := generatePeersInfo()
	err = l.AddPeerInfo(pi2)
	if err != nil {
		t.Fatal(err)
	}

	pis := make([]*types.PeerInfo, 0)
	err = l.GetPeersInfo(func(pi *types.PeerInfo) error {
		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pis) != 2 {
		t.Fatal("GetPeersInfo err")
	}
}

func TestLedger_UpdatePeerInfo(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	pi := generatePeersInfo()
	err := l.UpdatePeerInfo(pi)
	if err == nil {
		t.Fatal(err)
	}
	err = l.AddPeerInfo(pi)
	if err != nil {
		t.Fatal(err)
	}
	pi.Rtt = 2
	err = l.UpdatePeerInfo(pi)
	if err != nil {
		t.Fatal(err)
	}
	pi2, err := l.GetPeerInfo(pi.PeerID)
	if err != nil {
		t.Fatal(err)
	}
	pis := make([]*types.PeerInfo, 0)
	err = l.GetPeersInfo(func(pi *types.PeerInfo) error {
		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pis) != 1 {
		t.Fatal("UpdatePeerInfo err")
	}
	if pi2.Rtt != 2 {
		t.Fatal("rtt error")
	}
}

func TestLedger_AddOrUpdatePeerInfo(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	pi := generatePeersInfo()
	err := l.AddPeerInfo(pi)
	if err != nil {
		t.Fatal(err)
	}
	pi.Rtt = 2
	err = l.AddOrUpdatePeerInfo(pi)
	if err != nil {
		t.Fatal(err)
	}
	pi2, err := l.GetPeerInfo(pi.PeerID)
	if err != nil {
		t.Fatal(err)
	}
	pis := make([]*types.PeerInfo, 0)
	err = l.GetPeersInfo(func(pi *types.PeerInfo) error {
		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pis) != 1 {
		t.Fatal("AddOrUpdatePeerInfo err")
	}
	if pi2.Rtt != 2 {
		t.Fatal("rtt error")
	}
}

func TestLedger_CountPeersInfo(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	pi := generatePeersInfo()
	err := l.AddPeerInfo(pi)
	if err != nil {
		t.Fatal(err)
	}
	pi2 := generatePeersInfo()
	err = l.AddPeerInfo(pi2)
	if err != nil {
		t.Fatal(err)
	}
	pis, err := l.CountPeersInfo()
	if err != nil {
		t.Fatal(err)
	}
	if pis != 2 {
		t.Fatal("PeerInfo Count err")
	}
}

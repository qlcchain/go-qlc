package ledger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func setupPovTestCase(t *testing.T) (func(t *testing.T), *Ledger) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	l := NewLedger(cm.ConfigFile)

	return func(t *testing.T) {
		//err := l.Store.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l
}

func generatePovBlock(prevBlock *types.PovBlock) (*types.PovBlock, *types.PovTD) {
	return mock.GeneratePovBlock(prevBlock, 0)
}

func TestLedger_AddPovBlock(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	block, td := generatePovBlock(nil)
	err := l.AddPovBlock(block, td)
	if err != nil {
		t.Fatal(err)
	}

	retBlk, err := l.GetPovBlockByHeightAndHash(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	if retBlk.GetHash() != block.GetHash() {
		t.Fatalf("block hash not equal, %s != %s", retBlk.GetHash(), block.GetHash())
	}

	retHdr, err := l.GetPovHeader(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	if retHdr.GetHash() != block.GetHash() {
		t.Fatalf("header hash not equal, %s != %s", retBlk.GetHash(), block.GetHash())
	}

	retBody, err := l.GetPovBody(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	if len(retBody.Txs) != len(block.GetAllTxs()) {
		t.Fatalf("body txs not equal, %d != %d", len(retBody.Txs), len(block.GetAllTxs()))
	}

	retTD, err := l.GetPovTD(block.GetHash(), block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	if retTD.Chain.Cmp(&td.Chain) != 0 {
		t.Fatalf("td not equal, %s != %s", retTD.Chain.String(), td.Chain.String())
	}
}

func TestLedger_DeletePovBlock(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	block, td := generatePovBlock(nil)
	err := l.AddPovBlock(block, td)
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovBlock(block)
	if err != nil {
		t.Fatal(err)
	}

	retBlk, err := l.GetPovBlockByHeightAndHash(block.GetHeight(), block.GetHash())
	if retBlk != nil {
		t.Fatal("block exist after delete")
	}

	retTD, err := l.GetPovTD(block.GetHash(), block.GetHeight())
	if retTD != nil {
		t.Fatal("td exist after delete")
	}
}

func TestLedger_LatestPovBlock(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	block, td := generatePovBlock(nil)
	err := l.AddPovBlock(block, td)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovBestHash(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovLatestHeight(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	latestBlk, err := l.GetLatestPovBlock()
	if err != nil {
		t.Fatal(err)
	}

	if latestBlk.GetHash() != block.GetHash() {
		t.Fatalf("best hash not equal, %s != %s", latestBlk.GetHash(), block.GetHash())
	}
}

func TestLedger_AllPovBlocks(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	var blocks []*types.PovBlock

	block1, td1 := generatePovBlock(nil)
	err := l.AddPovBlock(block1, td1)
	if err != nil {
		t.Fatal(err)
	}
	blocks = append(blocks, block1)

	block2, td2 := generatePovBlock(block1)
	err = l.AddPovBlock(block2, td2)
	if err != nil {
		t.Fatal(err)
	}
	blocks = append(blocks, block2)

	block3, td3 := generatePovBlock(block2)
	err = l.AddPovBlock(block3, td3)
	if err != nil {
		t.Fatal(err)
	}
	blocks = append(blocks, block3)

	var retBlocks []*types.PovBlock
	err = l.GetAllPovBlocks(func(retBlk *types.PovBlock) error {
		exist := false
		for _, expectBlk := range blocks {
			if retBlk.GetHash() == expectBlk.GetHash() {
				exist = true
				break
			}
		}
		if !exist {
			t.Fatalf("block %s is not expect", retBlk.GetHash())
		}

		retBlocks = append(retBlocks, retBlk)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(retBlocks) != len(blocks) {
		t.Fatalf("block length not equal, %d != %d", len(retBlocks), len(blocks))
	}
}

func TestLedger_PovTxLookup(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	block, _ := generatePovBlock(nil)
	err := l.AddPovBestHash(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovLatestHeight(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	txH0, _ := types.NewHash("0000000000000000000000000000000000000000000000000000000000000001")
	txL0 := &types.PovTxLookup{BlockHash: block.GetHash(), BlockHeight: block.GetHeight(), TxIndex: 0}
	err = l.AddPovTxLookup(txH0, txL0)
	if err != nil {
		t.Fatal(err)
	}

	txH1, _ := types.NewHash("0000000000000000000000000000000000000000000000000000000000000002")
	txL1 := &types.PovTxLookup{BlockHash: block.GetHash(), BlockHeight: block.GetHeight(), TxIndex: 1}
	err = l.AddPovTxLookup(txH1, txL1)
	if err != nil {
		t.Fatal(err)
	}

	txH2, _ := types.NewHash("0000000000000000000000000000000000000000000000000000000000000003")
	txL2 := &types.PovTxLookup{BlockHash: block.GetHash(), BlockHeight: block.GetHeight(), TxIndex: 2}
	err = l.AddPovTxLookup(txH2, txL2)
	if err != nil {
		t.Fatal(err)
	}

	retTxL0, err := l.GetPovTxLookup(txH0)
	if retTxL0 == nil {
		t.Fatalf("tx %s not exist", txH0)
	}

	retTxL1, err := l.GetPovTxLookup(txH0)
	if retTxL1 == nil {
		t.Fatalf("tx %s not exist", txH1)
	}

	retTxL2, err := l.GetPovTxLookup(txH0)
	if retTxL2 == nil {
		t.Fatalf("tx %s not exist", txH2)
	}

	err = l.DeletePovTxLookup(txH0)
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovTxLookup(txH1)
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovTxLookup(txH2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_PovMinerStats(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	dayStat0 := types.NewPovMinerDayStat()
	dayStat0.DayIndex = 0
	dayStat0.MinerStats["miner0"] = &types.PovMinerStatItem{FirstHeight: 0, LastHeight: 99, BlockNum: 0, RewardAmount: types.ZeroBalance}
	dayStat0.MinerNum = uint32(len(dayStat0.MinerStats))
	err := l.AddPovMinerStat(dayStat0)
	if err != nil {
		t.Fatal(err)
	}

	dayStat2 := types.NewPovMinerDayStat()
	dayStat2.DayIndex = 1
	dayStat2.MinerStats["miner2"] = &types.PovMinerStatItem{FirstHeight: 200, LastHeight: 299, BlockNum: 20, RewardAmount: types.NewBalance(20)}
	dayStat2.MinerNum = uint32(len(dayStat0.MinerStats))
	err = l.AddPovMinerStat(dayStat2)
	if err != nil {
		t.Fatal(err)
	}

	dayStat1 := types.NewPovMinerDayStat()
	dayStat1.DayIndex = 1
	dayStat1.MinerStats["miner1"] = &types.PovMinerStatItem{FirstHeight: 100, LastHeight: 199, BlockNum: 10, RewardAmount: types.NewBalance(10)}
	dayStat1.MinerNum = uint32(len(dayStat0.MinerStats))
	err = l.AddPovMinerStat(dayStat1)
	if err != nil {
		t.Fatal(err)
	}

	latestMinerStat, err := l.GetLatestPovMinerStat()
	if err != nil {
		t.Fatal(err)
	}
	if latestMinerStat.DayIndex != dayStat2.DayIndex {
		t.Fatalf("latest day index not equal, %d != %d", latestMinerStat.DayIndex, dayStat2.DayIndex)
	}

	err = l.DeletePovMinerStat(0)
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovMinerStat(1)
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovMinerStat(2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_PovTxLookupBatch(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	batchAdd := l.Store.NewWriteBatch()

	block, _ := generatePovBlock(nil)
	err := l.AddPovBestHash(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovLatestHeight(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	txH0, _ := types.NewHash("0000000000000000000000000000000000000000000000000000000000000001")
	txL0 := &types.PovTxLookup{BlockHash: block.GetHash(), BlockHeight: block.GetHeight(), TxIndex: 0}
	err = l.AddPovTxLookupInBatch(txH0, txL0, batchAdd)
	if err != nil {
		t.Fatal(err)
	}

	txH1, _ := types.NewHash("0000000000000000000000000000000000000000000000000000000000000002")
	txL1 := &types.PovTxLookup{BlockHash: block.GetHash(), BlockHeight: block.GetHeight(), TxIndex: 1}
	err = l.AddPovTxLookupInBatch(txH1, txL1, batchAdd)
	if err != nil {
		t.Fatal(err)
	}

	txH2, _ := types.NewHash("0000000000000000000000000000000000000000000000000000000000000003")
	txL2 := &types.PovTxLookup{BlockHash: block.GetHash(), BlockHeight: block.GetHeight(), TxIndex: 2}
	err = l.AddPovTxLookupInBatch(txH2, txL2, batchAdd)
	if err != nil {
		t.Fatal(err)
	}

	err = batchAdd.Flush()
	if err != nil {
		t.Fatal(err)
	}

	retTxL0, err := l.GetPovTxLookup(txH0)
	if retTxL0 == nil {
		t.Fatalf("tx %s not exist", txH0)
	}

	retTxL1, err := l.GetPovTxLookup(txH0)
	if retTxL1 == nil {
		t.Fatalf("tx %s not exist", txH1)
	}

	retTxL2, err := l.GetPovTxLookup(txH0)
	if retTxL2 == nil {
		t.Fatalf("tx %s not exist", txH2)
	}

	batchDel := l.Store.NewWriteBatch()
	err = l.DeletePovTxLookupInBatch(txH0, batchDel)
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovTxLookupInBatch(txH1, batchDel)
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovTxLookupInBatch(txH2, batchDel)
	if err != nil {
		t.Fatal(err)
	}
	err = batchDel.Flush()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_PovTxlScanCursor(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	err := l.SetPovTxlScanCursor(34127856)
	if err != nil {
		t.Fatal(err)
	}

	height, err := l.GetPovTxlScanCursor()
	if err != nil {
		t.Fatal(err)
	}
	if height != 34127856 {
		t.Fatal(err)
	}
}

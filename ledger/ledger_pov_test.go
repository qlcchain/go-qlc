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

func TestLedger_AddPovHeader(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	block, td := generatePovBlock(nil)

	err := l.AddPovHeader(block.GetHeader())
	if err != nil {
		t.Fatal(err)
	}
	exist := l.HasPovHeader(block.GetHeight(), block.GetHash())
	if !exist {
		t.Fatal(err)
	}

	err = l.AddPovBody(block.GetHeight(), block.GetHash(), block.GetBody())
	if err != nil {
		t.Fatal(err)
	}
	exist = l.HasPovBody(block.GetHeight(), block.GetHash())
	if !exist {
		t.Fatal(err)
	}

	err = l.AddPovHeight(block.GetHash(), block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
	exist = l.HasPovHeight(block.GetHash())
	if !exist {
		t.Fatal(err)
	}

	err = l.AddPovBestHash(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovTD(block.GetHash(), block.GetHeight(), td)
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovTD(block.GetHash(), block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovBestHash(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovHeight(block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovBody(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovHeader(block.GetHeight(), block.GetHash())
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddPovBlock(t *testing.T) {
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
	err = l.SetPovLatestHeight(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	exist := l.HasPovBlock(block.GetHeight(), block.GetHash())
	if !exist {
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
		t.Fatalf("header hash not equal, %s != %s", retHdr.GetHash(), block.GetHash())
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

	retHdr, err = l.GetPovHeaderByHeight(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
	if retHdr.GetHash() != block.GetHash() {
		t.Fatalf("header hash not equal, %s != %s", retHdr.GetHash(), block.GetHash())
	}

	retHdr, err = l.GetPovHeaderByHash(block.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if retHdr.GetHash() != block.GetHash() {
		t.Fatalf("header hash not equal, %s != %s", retHdr.GetHash(), block.GetHash())
	}

	retBlk, err = l.GetPovBlockByHeight(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
	if retBlk.GetHash() != block.GetHash() {
		t.Fatalf("block hash not equal, %s != %s", retBlk.GetHash(), block.GetHash())
	}

	retBlk, err = l.GetPovBlockByHash(block.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if retBlk.GetHash() != block.GetHash() {
		t.Fatalf("block hash not equal, %s != %s", retBlk.GetHash(), block.GetHash())
	}

	err = l.DropAllPovBlocks()
	if err != nil {
		t.Fatal(err)
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

	err = l.SetPovLatestHeight(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	latestHdr, err := l.GetLatestPovHeader()
	if err != nil {
		t.Fatal(err)
	}

	if latestHdr.GetHash() != block.GetHash() {
		t.Fatalf("latest header hash not equal, %s != %s", latestHdr.GetHash(), block.GetHash())
	}

	latestBlk, err := l.GetLatestPovBlock()
	if err != nil {
		t.Fatal(err)
	}

	if latestBlk.GetHash() != block.GetHash() {
		t.Fatalf("latest block hash not equal, %s != %s", latestBlk.GetHash(), block.GetHash())
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
	err = l.AddPovBestHash(block1.GetHeight(), block1.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	err = l.SetPovLatestHeight(block1.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
	blocks = append(blocks, block1)

	block2, td2 := generatePovBlock(block1)
	err = l.AddPovBlock(block2, td2)
	if err != nil {
		t.Fatal(err)
	}
	err = l.AddPovBestHash(block2.GetHeight(), block2.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	err = l.SetPovLatestHeight(block2.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
	blocks = append(blocks, block2)

	block3, td3 := generatePovBlock(block2)
	err = l.AddPovBlock(block3, td3)
	if err != nil {
		t.Fatal(err)
	}
	err = l.AddPovBestHash(block3.GetHeight(), block3.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	err = l.SetPovLatestHeight(block3.GetHeight())
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

	var retBestBlks []*types.PovBlock
	err = l.GetAllPovBestBlocks(func(retBlk *types.PovBlock) error {
		retBestBlks = append(retBestBlks, retBlk)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(retBestBlks) != len(blocks) {
		t.Fatalf("best block length not equal, %d != %d", len(retBestBlks), len(blocks))
	}

	var retHeaders []*types.PovHeader
	err = l.GetAllPovHeaders(func(retHdr *types.PovHeader) error {
		retHeaders = append(retHeaders, retHdr)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(retHeaders) != len(blocks) {
		t.Fatalf("header length not equal, %d != %d", len(retHeaders), len(blocks))
	}

	var retBestHdrs []*types.PovHeader
	err = l.GetAllPovBestHeaders(func(retHdr *types.PovHeader) error {
		retBestHdrs = append(retBestHdrs, retHdr)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(retBestHdrs) != len(blocks) {
		t.Fatalf("best header length not equal, %d != %d", len(retBestHdrs), len(blocks))
	}

	var retBestHashes []types.Hash
	err = l.GetAllPovBestHashes(func(height uint64, hash types.Hash) error {
		retBestHashes = append(retBestHashes, hash)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(retBestHashes) != len(blocks) {
		t.Fatalf("best hash length not equal, %d != %d", len(retBestHashes), len(blocks))
	}

	retHdrs1, err := l.BatchGetPovHeadersByHeightAsc(block1.GetHeight(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(retHdrs1) != 3 {
		t.Fatal("len(retHdrs1) != 3", len(retHdrs1))
	}

	retHdrs2, err := l.BatchGetPovHeadersByHeightDesc(block3.GetHeight(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(retHdrs2) != 3 {
		t.Fatal("len(retHdrs2) != 3", len(retHdrs2))
	}

	blkCnt, err := l.CountPovBlocks()
	if int(blkCnt) != len(blocks) {
		t.Fatal("CountPovBlocks", blkCnt, len(retHdrs2))
	}
	bhCnt, err := l.CountPovBestHashs()
	if int(bhCnt) != len(blocks) {
		t.Fatal("CountPovBestHashs", bhCnt, len(retHdrs2))
	}
	err = l.DropAllPovBlocks()
	if err != nil {
		t.Fatal(err)
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

	err = l.SetPovLatestHeight(block.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	var allTxls []*types.PovTxLookup

	txH0, _ := types.NewHash("0000000000000000000000000000000000000000000000000000000000000001")
	txL0 := &types.PovTxLookup{BlockHash: block.GetHash(), BlockHeight: block.GetHeight(), TxIndex: 0}
	err = l.AddPovTxLookup(txH0, txL0)
	if err != nil {
		t.Fatal(err)
	}
	allTxls = append(allTxls, txL0)

	txH1, _ := types.NewHash("0000000000000000000000000000000000000000000000000000000000000002")
	txL1 := &types.PovTxLookup{BlockHash: block.GetHash(), BlockHeight: block.GetHeight(), TxIndex: 1}
	err = l.AddPovTxLookup(txH1, txL1)
	if err != nil {
		t.Fatal(err)
	}
	allTxls = append(allTxls, txL1)

	txH2, _ := types.NewHash("0000000000000000000000000000000000000000000000000000000000000003")
	txL2 := &types.PovTxLookup{BlockHash: block.GetHash(), BlockHeight: block.GetHeight(), TxIndex: 2}
	err = l.AddPovTxLookup(txH2, txL2)
	if err != nil {
		t.Fatal(err)
	}
	allTxls = append(allTxls, txL2)

	txCnt, err := l.CountPovTxs()
	if int(txCnt) != len(allTxls) {
		t.Fatal("CountPovTxs", txCnt, len(allTxls))
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

	var allDss []*types.PovMinerDayStat
	dayStat0 := types.NewPovMinerDayStat()
	dayStat0.DayIndex = 0
	dayStat0.MinerStats["miner0"] = &types.PovMinerStatItem{FirstHeight: 0, LastHeight: 99, BlockNum: 0, RewardAmount: types.ZeroBalance}
	dayStat0.MinerNum = uint32(len(dayStat0.MinerStats))
	err := l.AddPovMinerStat(dayStat0)
	if err != nil {
		t.Fatal(err)
	}
	allDss = append(allDss, dayStat0)

	dayStat2 := types.NewPovMinerDayStat()
	dayStat2.DayIndex = 2
	dayStat2.MinerStats["miner2"] = &types.PovMinerStatItem{FirstHeight: 200, LastHeight: 299, BlockNum: 20, RewardAmount: types.NewBalance(20)}
	dayStat2.MinerNum = uint32(len(dayStat0.MinerStats))
	err = l.AddPovMinerStat(dayStat2)
	if err != nil {
		t.Fatal(err)
	}
	allDss = append(allDss, dayStat2)

	dayStat1 := types.NewPovMinerDayStat()
	dayStat1.DayIndex = 1
	dayStat1.MinerStats["miner1"] = &types.PovMinerStatItem{FirstHeight: 100, LastHeight: 199, BlockNum: 10, RewardAmount: types.NewBalance(10)}
	dayStat1.MinerNum = uint32(len(dayStat0.MinerStats))
	err = l.AddPovMinerStat(dayStat1)
	if err != nil {
		t.Fatal(err)
	}
	allDss = append(allDss, dayStat1)

	latestMinerStat, err := l.GetLatestPovMinerStat()
	if err != nil {
		t.Fatal(err)
	}
	if latestMinerStat.DayIndex != dayStat2.DayIndex {
		t.Fatalf("latest day index not equal, %d != %d", latestMinerStat.DayIndex, dayStat2.DayIndex)
	}

	for _, ds := range allDss {
		exist := l.HasPovMinerStat(ds.DayIndex)
		if !exist {
			t.Fatalf("day index not exist, %d", ds.DayIndex)
		}

		_, err := l.GetPovMinerStat(ds.DayIndex)
		if err != nil {
			t.Fatal(err)
		}
	}

	var retDss []*types.PovMinerDayStat
	err = l.GetAllPovMinerStats(func(retDs *types.PovMinerDayStat) error {
		retDss = append(retDss, retDs)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(retDss) != len(allDss) {
		t.Fatalf("GetAllPovMinerStats %d != %d", len(retDss), len(allDss))
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

	err = l.SetPovLatestHeight(block.GetHeight())
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
	exist := l.HasPovTxLookup(txH0)
	if !exist {
		t.Fatalf("tx %s not exist", txH0)
	}

	retTxL1, err := l.GetPovTxLookup(txH1)
	if retTxL1 == nil {
		t.Fatalf("tx %s not exist", txH1)
	}
	exist = l.HasPovTxLookup(txH1)
	if !exist {
		t.Fatalf("tx %s not exist", txH1)
	}

	retTxL2, err := l.GetPovTxLookup(txH2)
	if retTxL2 == nil {
		t.Fatalf("tx %s not exist", txH2)
	}
	exist = l.HasPovTxLookup(txH2)
	if !exist {
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

func TestLedger_PovDiffStats(t *testing.T) {
	teardownTestCase, l := setupPovTestCase(t)
	defer teardownTestCase(t)

	var allDss []*types.PovDiffDayStat

	dayStat0 := types.NewPovDiffDayStat()
	dayStat0.DayIndex = 0
	dayStat0.AvgDiffRatio = 10
	err := l.AddPovDiffStat(dayStat0)
	if err != nil {
		t.Fatal(err)
	}
	allDss = append(allDss, dayStat0)

	dayStat2 := types.NewPovDiffDayStat()
	dayStat2.DayIndex = 2
	dayStat2.AvgDiffRatio = 200
	err = l.AddPovDiffStat(dayStat2)
	if err != nil {
		t.Fatal(err)
	}
	allDss = append(allDss, dayStat2)

	dayStat1 := types.NewPovDiffDayStat()
	dayStat1.DayIndex = 1
	dayStat1.AvgDiffRatio = 100
	err = l.AddPovDiffStat(dayStat1)
	if err != nil {
		t.Fatal(err)
	}
	allDss = append(allDss, dayStat1)

	for _, ds := range allDss {
		_, err := l.GetPovDiffStat(ds.DayIndex)
		if err != nil {
			t.Fatal(err)
		}
	}

	var retDss []*types.PovDiffDayStat
	err = l.GetAllPovDiffStats(func(retDs *types.PovDiffDayStat) error {
		retDss = append(retDss, retDs)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(retDss) != len(allDss) {
		t.Fatalf("GetAllPovDiffStats %d != %d", len(retDss), len(allDss))
	}

	latestDayStat, err := l.GetLatestPovDiffStat()
	if err != nil {
		t.Fatal(err)
	}
	if latestDayStat.DayIndex != dayStat2.DayIndex {
		t.Fatalf("latest day index not equal, %d != %d", latestDayStat.DayIndex, dayStat2.DayIndex)
	}

	err = l.DeletePovDiffStat(2)
	if err != nil {
		t.Fatal(err)
	}

	latestDayStat, err = l.GetLatestPovDiffStat()
	if err != nil {
		t.Fatal(err)
	}
	if latestDayStat.DayIndex != dayStat1.DayIndex {
		t.Fatalf("latest day index not equal, %d != %d", latestDayStat.DayIndex, dayStat1.DayIndex)
	}

	err = l.DeletePovDiffStat(1)
	if err != nil {
		t.Fatal(err)
	}

	err = l.DeletePovDiffStat(0)
	if err != nil {
		t.Fatal(err)
	}

	latestDayStat, err = l.GetLatestPovDiffStat()
	if err == nil {
		t.Fatal(err)
	}
}

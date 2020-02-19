package contract

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"math/big"
	"testing"
)

func TestMinerReward_GetLastRewardHeight(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	h, err := m.GetLastRewardHeight(ctx, account)
	if h != 0 {
		t.Fatal()
	}

	data, _ := cabi.MinerABI.PackVariable(cabi.VariableNameMinerReward, uint64(1440), uint64(100), common.TimeNow().Unix(), big.NewInt(200))
	err = ctx.SetStorage(types.MinerAddress.Bytes(), account[:], data)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SaveStorage()
	if err != nil {
		t.Fatal(err)
	}

	h, err = m.GetLastRewardHeight(ctx, account)
	if h != 1440 || err != nil {
		t.Fatal(h, err)
	}
}

func TestMinerReward_GetRewardHistory(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	mi, err := m.GetRewardHistory(ctx, account)
	if mi != nil || err == nil {
		t.Fatal()
	}

	timeStamp := common.TimeNow().Unix()
	data, _ := cabi.MinerABI.PackVariable(cabi.VariableNameMinerReward, uint64(1440), uint64(100), timeStamp, big.NewInt(200))
	err = ctx.SetStorage(types.MinerAddress.Bytes(), account[:], data)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SaveStorage()
	if err != nil {
		t.Fatal(err)
	}

	mi, err = m.GetRewardHistory(ctx, account)
	if mi == nil || mi.EndHeight != uint64(1440) || mi.RewardBlocks != uint64(100) || mi.RewardAmount.Cmp(big.NewInt(200)) != 0 || mi.Timestamp != timeStamp {
		t.Fatal()
	}
}

func TestMinerReward_GetNodeRewardHeight(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	h, err := m.GetNodeRewardHeight(ctx)
	if h != 0 {
		t.Fatal()
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	pb.Header.BasHdr.Height = common.PovMinerRewardHeightStart - 1
	err = addLatestPovBlock(pb, td, l)
	if err != nil {
		t.Fatal(err)
	}

	h, err = m.GetNodeRewardHeight(ctx)
	if h != 0 {
		t.Fatal(err)
	}

	pb, td = mock.GeneratePovBlock(pb, 0)
	pb.Header.BasHdr.Height = common.PovMinerRewardHeightStart + common.PovMinerRewardHeightGapToLatest + 1
	err = addLatestPovBlock(pb, td, l)
	if err != nil {
		t.Fatal(err)
	}

	h, err = m.GetNodeRewardHeight(ctx)
	if h != 0 {
		t.Fatal(err)
	}

	pb, td = mock.GeneratePovBlock(pb, 0)
	pb.Header.BasHdr.Height = common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay) + common.PovMinerRewardHeightGapToLatest + 1
	err = addLatestPovBlock(pb, td, l)
	if err != nil {
		t.Fatal(err)
	}

	h, err = m.GetNodeRewardHeight(ctx)
	t.Log(h)
	if h != common.PovMinerRewardHeightStart+uint64(common.POVChainBlocksPerDay)-1 {
		t.Fatal(err)
	}
}

func TestMinerReward_GetAvailRewardInfo(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	lastHeight := uint64(0)
	nodeHeight := common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay)
	mi, err := m.GetAvailRewardInfo(ctx, account, nodeHeight, lastHeight)
	if mi != nil {
		t.Fatal(err)
	}

	ds := types.NewPovMinerDayStat()
	it := types.NewPovMinerStatItem()
	it.FirstHeight = 1440
	it.LastHeight = 2880
	it.BlockNum = 240
	it.RewardAmount = types.NewBalance(100)
	it.IsMiner = true
	ds.DayIndex = uint32(common.PovMinerRewardHeightStart / uint64(common.POVChainBlocksPerDay))
	ds.MinerStats[account.String()] = it
	err = l.AddPovMinerStat(ds)
	if err != nil {
		t.Fatal(err)
	}

	mi, err = m.GetAvailRewardInfo(ctx, account, nodeHeight, lastHeight)
	if err != nil || mi.RewardBlocks != uint64(it.BlockNum) || mi.RewardAmount.Cmp(it.RewardAmount.Int) != 0 || mi.EndHeight != it.LastHeight-1 {
		t.Fatal(err)
	}
}

func TestMinerReward_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlock()
	blk.Address = types.ZeroAddress
	blk.Token = types.ZeroHash

	_, _, err := m.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	account := types.ZeroAddress
	beneficial := mock.Address()
	startHeight := common.PovMinerRewardHeightStart
	endHeight := common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay) - 1
	rewardBlocks := uint64(240)
	rewardAmount := types.NewBalance(100)
	blk.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, account, beneficial, startHeight, endHeight, rewardBlocks, rewardAmount.Int)
	_, _, err = m.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	account = mock.Address()
	blk.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, account, beneficial, startHeight, endHeight, rewardBlocks, rewardAmount.Int)
	_, _, err = m.ProcessSend(ctx, blk)
	if err != ErrAccountInvalid {
		t.Fatal(err)
	}

	blk.Address = account
	_, _, err = m.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = common.ChainToken()
	_, _, err = m.ProcessSend(ctx, blk)
	if err != ErrAccountNotExist {
		t.Fatal(err)
	}

	am := mock.AccountMeta(blk.Address)
	err = l.AddAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = m.ProcessSend(ctx, blk)
	if err != ErrGetNodeHeight {
		t.Fatal(err)
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	err = addLatestPovBlock(pb, td, l)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = m.ProcessSend(ctx, blk)
	if err != ErrEndHeightInvalid {
		t.Fatal(err)
	}

	pb, td = mock.GeneratePovBlock(pb, 0)
	pb.Header.BasHdr.Height = common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay) + common.PovMinerRewardHeightGapToLatest
	err = addLatestPovBlock(pb, td, l)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = m.ProcessSend(ctx, blk)
	if err != ErrCalcAmount {
		t.Fatal(err)
	}

	ds := types.NewPovMinerDayStat()
	it := types.NewPovMinerStatItem()
	it.FirstHeight = 1440
	it.LastHeight = 2880
	it.BlockNum = 120
	it.RewardAmount = types.NewBalance(10)
	ds.DayIndex = uint32(common.PovMinerRewardHeightStart / uint64(common.POVChainBlocksPerDay))
	ds.MinerStats[account.String()] = it
	err = l.AddPovMinerStat(ds)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = m.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	it.BlockNum = 240
	it.RewardAmount = types.NewBalance(100)
	err = l.AddPovMinerStat(ds)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = m.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	timestamp := common.TimeNow().Unix()
	data, err := cabi.MinerABI.PackVariable(cabi.VariableNameMinerReward, endHeight, rewardBlocks, timestamp, rewardAmount.Int)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetStorage(types.MinerAddress.Bytes(), account[:], data)
	ctx.SaveStorage()
	_, _, err = m.ProcessSend(ctx, blk)
	if err != ErrClaimRepeat {
		t.Fatal(err)
	}
}

func TestMinerReward_SetStorage(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlock()
	blk.Address = mock.Address()
	blk.Timestamp = common.TimeNow().Unix()
	amount := big.NewInt(100)
	err := m.SetStorage(ctx, 2879, amount, 240, blk)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SaveStorage()

	mi, err := m.GetRewardHistory(ctx, blk.Address)
	if err != nil || mi.EndHeight != 2879 || mi.RewardBlocks != 240 || mi.RewardAmount.Cmp(amount) != 0 {
		t.Fatal()
	}
}

func TestMinerReward_DoReceive(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	sendBlk := mock.StateBlock()
	recvBlk := mock.StateBlock()
	blks, err := m.DoReceive(ctx, recvBlk, sendBlk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	account := types.ZeroAddress
	beneficial := mock.Address()
	startHeight := common.PovMinerRewardHeightStart
	endHeight := common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay) - 1
	rewardBlocks := uint64(240)
	rewardAmount := types.NewBalance(100)
	sendBlk.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, account, beneficial, startHeight, endHeight, rewardBlocks, rewardAmount.Int)
	blks, err = m.DoReceive(ctx, recvBlk, sendBlk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	account = mock.Address()
	sendBlk.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, account, beneficial, startHeight, endHeight, rewardBlocks, rewardAmount.Int)
	blks, err = m.DoReceive(ctx, recvBlk, sendBlk)
	if err != ErrAccountInvalid {
		t.Fatal(err)
	}

	sendBlk.Address = account
	blks, err = m.DoReceive(ctx, recvBlk, sendBlk)
	if err != nil {
		t.Fatal(err)
	}

	retBlock := blks[0].Block
	if retBlock.Balance.Compare(rewardAmount) != types.BalanceCompEqual || retBlock.Address != beneficial || retBlock.Previous != types.ZeroHash {
		t.Fatal()
	}

	am := mock.AccountMeta(beneficial)
	err = l.AddAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}

	blks, err = m.DoReceive(ctx, recvBlk, sendBlk)
	if err != nil {
		t.Fatal(err)
	}

	retBlock = blks[0].Block
	if retBlock.Balance.Compare(rewardAmount) != types.BalanceCompEqual || retBlock.Address != beneficial ||
		retBlock.Previous != types.ZeroHash || retBlock.Representative != am.Tokens[0].Representative {
		t.Fatal()
	}

	am = mock.AccountMeta(beneficial)
	am.Tokens[0].Type = common.GasToken()
	err = l.UpdateAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}

	blks, err = m.DoReceive(ctx, recvBlk, sendBlk)
	if err != nil {
		t.Fatal(err)
	}

	retBlock = blks[0].Block
	if retBlock.Balance.Sub(am.Tokens[0].Balance).Compare(rewardAmount) != types.BalanceCompEqual || retBlock.Address != beneficial ||
		retBlock.Previous != am.Tokens[0].Header || retBlock.Representative != am.Tokens[0].Representative {
		t.Fatal()
	}

	am = mock.AccountMeta(beneficial)
	am.Tokens = []*types.TokenMeta{}
	err = l.UpdateAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}

	blks, err = m.DoReceive(ctx, recvBlk, sendBlk)
	if err != nil {
		t.Fatal(err)
	}

	retBlock = blks[0].Block
	if retBlock.Representative != sendBlk.Representative {
		t.Fatal()
	}
}

func TestMinerReward_DoGap(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlock()
	gap, _, _ := m.DoGap(ctx, blk)
	if gap != common.ContractNoGap {
		t.Fatal(gap)
	}

	account := mock.Address()
	beneficial := mock.Address()
	startHeight := common.PovMinerRewardHeightStart
	endHeight := common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay) - 1
	rewardBlocks := uint64(240)
	rewardAmount := types.NewBalance(100)
	blk.Data, _ = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, account, beneficial, startHeight, endHeight, rewardBlocks, rewardAmount.Int)
	gap, height, _ := m.DoGap(ctx, blk)
	if gap != common.ContractRewardGapPov || height.(uint64) != endHeight+common.PovMinerRewardHeightGapToLatest {
		t.Fatal(gap)
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	err := addLatestPovBlock(pb, td, l)
	if err != nil {
		t.Fatal(err)
	}
	gap, height, _ = m.DoGap(ctx, blk)
	if gap != common.ContractRewardGapPov || height.(uint64) != endHeight+common.PovMinerRewardHeightGapToLatest {
		t.Fatal(gap)
	}

	pb, td = mock.GeneratePovBlock(pb, 0)
	pb.Header.BasHdr.Height = common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay) + common.PovMinerRewardHeightGapToLatest
	err = addLatestPovBlock(pb, td, l)
	if err != nil {
		t.Fatal(err)
	}
	gap, height, _ = m.DoGap(ctx, blk)
	if gap != common.ContractNoGap {
		t.Fatal(gap)
	}
}

func TestMinerReward_checkParamExistInOldRewardInfos(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlock()
	account := mock.Address()

	param := &cabi.MinerRewardParam{
		Coinbase:     account,
		Beneficial:   mock.Address(),
		StartHeight:  1440,
		EndHeight:    2879,
		RewardBlocks: 240,
		RewardAmount: big.NewInt(100),
	}
	err := m.checkParamExistInOldRewardInfos(ctx, param)
	if err != nil {
		t.Fatal()
	}

	blk.Address = account
	blk.Timestamp = common.TimeNow().Unix()
	amount := big.NewInt(100)
	err = m.SetStorage(ctx, 2879, amount, 240, blk)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SaveStorage()

	err = m.checkParamExistInOldRewardInfos(ctx, param)
	if err == nil {
		t.Fatal()
	}
}

func TestMinerReward_calcRewardBlocksByDayStats(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)
	account := mock.Address()

	startHeight := common.PovMinerRewardHeightStart + 1
	endHeight := uint64(1)
	_, _, err := m.calcRewardBlocksByDayStats(ctx, account, startHeight, endHeight)
	if err == nil {
		t.Fatal()
	}

	startHeight = common.PovMinerRewardHeightStart
	endHeight = uint64(1)
	_, _, err = m.calcRewardBlocksByDayStats(ctx, account, startHeight, endHeight)
	if err == nil {
		t.Fatal()
	}

	ds := types.NewPovMinerDayStat()
	it := types.NewPovMinerStatItem()
	it.FirstHeight = 1440
	it.LastHeight = 2880
	it.BlockNum = 120
	it.RewardAmount = types.NewBalance(10)
	ds.DayIndex = uint32(common.PovMinerRewardHeightStart / uint64(common.POVChainBlocksPerDay))
	ds.MinerStats[mock.Address().String()] = it
	err = l.AddPovMinerStat(ds)
	if err != nil {
		t.Fatal(err)
	}

	startHeight = common.PovMinerRewardHeightStart
	endHeight = startHeight + uint64(common.POVChainBlocksPerDay) - 1
	rewardBlocks, rewardAmount, err := m.calcRewardBlocksByDayStats(ctx, account, startHeight, endHeight)
	if err != nil || rewardAmount.Compare(types.NewBalance(0)) != types.BalanceCompEqual || rewardBlocks != 0 {
		t.Fatal(err)
	}

	ds = types.NewPovMinerDayStat()
	it = types.NewPovMinerStatItem()
	it.FirstHeight = 1440
	it.LastHeight = 2880
	it.BlockNum = 120
	it.RewardAmount = types.NewBalance(10)
	ds.DayIndex = uint32(common.PovMinerRewardHeightStart / uint64(common.POVChainBlocksPerDay))
	ds.MinerStats[account.String()] = it
	err = l.AddPovMinerStat(ds)
	if err != nil {
		t.Fatal(err)
	}

	rewardBlocks, rewardAmount, err = m.calcRewardBlocksByDayStats(ctx, account, startHeight, endHeight)
	if err != nil || rewardAmount.Compare(it.RewardAmount) != types.BalanceCompEqual || rewardBlocks != 120 {
		t.Fatal()
	}
}

func TestMinerReward_GetTargetReceiver(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	m := new(MinerReward)
	ctx := vmstore.NewVMContext(l)

	blk := mock.StateBlock()
	account := mock.Address()
	beneficial := account
	startHeight := common.PovMinerRewardHeightStart
	endHeight := common.PovMinerRewardHeightStart + uint64(common.POVChainBlocksPerDay) - 1
	rewardBlocks := uint64(240)
	rewardAmount := types.NewBalance(100)
	blk.Data, _ = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, account, beneficial, startHeight, endHeight, rewardBlocks, rewardAmount.Int)
	receiver := m.GetTargetReceiver(ctx, blk)
	if receiver != beneficial {
		t.Fatal()
	}
}

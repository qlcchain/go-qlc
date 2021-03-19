package api

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type QGasSwapAPI struct {
	l                ledger.Store
	pledgeContract   *contract.QGasPledge
	withdrawContract *contract.QGasWithdraw
	cc               *chainctx.ChainContext
}

func NewQGasSwapAPI(l ledger.Store, cfgFile string) *QGasSwapAPI {
	return &QGasSwapAPI{
		l:                l,
		pledgeContract:   &contract.QGasPledge{},
		withdrawContract: &contract.QGasWithdraw{},
		cc:               chainctx.NewChainContext(cfgFile),
	}
}

type QGasPledgeParams struct {
	PledgeAddress types.Address `json:"pledgeAddress"`
	Amount        types.Balance `json:"amount"`
	ToAddress     types.Address `json:"toAddress"`
}

func (q *QGasSwapAPI) GetPledgeSendBlock(param *QGasPledgeParams) (*types.StateBlock, error) {
	if !q.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if param == nil {
		return nil, ErrParameterNil
	}

	am, err := q.l.GetAccountMeta(param.PledgeAddress)
	if am == nil {
		return nil, fmt.Errorf("invalid user account:%s, %s", param.PledgeAddress.String(), err)
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have any chain token", param.PledgeAddress.String())
	}
	if tm.Balance.Compare(param.Amount) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have no enough balance, want pledge %s, but only %s", param.PledgeAddress.String(), param.Amount.String(), tm.Balance.String())
	}
	povHeader, err := q.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	ctx := vmstore.NewVMContext(q.l, &contractaddress.QGasSwapAddress)

	pledgeParam := &abi.QGasPledgeParam{
		PledgeAddress: param.PledgeAddress,
		Amount:        param.Amount.Int,
		ToAddress:     param.ToAddress,
	}
	data, err := pledgeParam.ToABI()
	if err != nil {
		return nil, fmt.Errorf("error pledge param abi: %s", err)
	}

	block := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.PledgeAddress,
		Balance:        tm.Balance.Sub(param.Amount),
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Previous:       tm.Header,
		Link:           types.Hash(contractaddress.QGasSwapAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}
	if _, _, err := q.pledgeContract.ProcessSend(ctx, block); err != nil {
		return nil, err
	} else {
		h := vmstore.TrieHash(ctx)
		if h != nil {
			block.Extra = h
		}
	}
	return block, nil
}

func (q *QGasSwapAPI) GetPledgeRewardBlock(sendHash types.Hash) (*types.StateBlock, error) {
	if sendHash.IsZero() {
		return nil, ErrParameterNil
	}
	if !q.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	sendBlk, err := q.l.GetStateBlock(sendHash)
	if err != nil {
		return nil, err
	}

	povHeader, err := q.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	rewardBlk := &types.StateBlock{
		Type:      types.ContractReward,
		Token:     config.GasToken(),
		Link:      sendHash,
		PoVHeight: povHeader.GetHeight(),
		Timestamp: common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(q.l, &contractaddress.RepAddress)
	blocks, err := q.pledgeContract.DoReceive(vmContext, rewardBlk, sendBlk)
	if err != nil {
		return nil, fmt.Errorf("qgas pledge DoReceive error: %s", err)
	}
	if len(blocks) > 0 {
		return rewardBlk, nil
	}

	return nil, errors.New("can not generate reward recv block")
}

type QGasWithdrawParams struct {
	WithdrawAddress types.Address `json:"withdrawAddress"`
	Amount          types.Balance `json:"amount"`
	FromAddress     types.Address `json:"fromAddress"`
	LinkHash        types.Hash    `json:"linkHash"`
}

func (q *QGasSwapAPI) GetWithdrawSendBlock(param *QGasWithdrawParams) (*types.StateBlock, error) {
	if !q.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if param == nil {
		return nil, ErrParameterNil
	}

	am, err := q.l.GetAccountMeta(param.FromAddress)
	if am == nil {
		return nil, fmt.Errorf("invalid user account:%s, %s", param.FromAddress.String(), err)
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have any chain token", param.FromAddress.String())
	}
	if tm.Balance.Compare(param.Amount) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have no enough balance, want pledge %s, but only %s", param.FromAddress.String(), param.Amount.String(), tm.Balance.String())
	}
	povHeader, err := q.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	withdrawParam := &abi.QGasWithdrawParam{
		WithdrawAddress: param.WithdrawAddress,
		Amount:          param.Amount.Int,
		FromAddress:     param.FromAddress,
		LinkHash:        param.LinkHash,
	}
	data, err := withdrawParam.ToABI()
	if err != nil {
		return nil, fmt.Errorf("error pledge param abi: %s", err)
	}

	block := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.FromAddress,
		Balance:        tm.Balance.Sub(param.Amount),
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Previous:       tm.Header,
		Link:           types.Hash(contractaddress.QGasSwapAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}
	ctx := vmstore.NewVMContext(q.l, &contractaddress.QGasSwapAddress)
	if _, _, err := q.withdrawContract.ProcessSend(ctx, block); err != nil {
		return nil, err
	} else {
		h := vmstore.TrieHash(ctx)
		if h != nil {
			block.Extra = h
		}
	}
	return block, nil
}

func (q *QGasSwapAPI) GetWithdrawRewardBlock(sendHash types.Hash) (*types.StateBlock, error) {
	if sendHash.IsZero() {
		return nil, ErrParameterNil
	}
	if !q.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	sendBlk, err := q.l.GetStateBlock(sendHash)
	if err != nil {
		return nil, err
	}

	povHeader, err := q.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	rewardBlk := &types.StateBlock{
		Type:      types.ContractReward,
		Token:     config.GasToken(),
		Link:      sendHash,
		PoVHeight: povHeader.GetHeight(),
		Timestamp: common.TimeNow().Unix(),
	}

	ctx := vmstore.NewVMContext(q.l, &contractaddress.QGasSwapAddress)
	blocks, err := q.withdrawContract.DoReceive(ctx, rewardBlk, sendBlk)
	if err != nil {
		return nil, fmt.Errorf("qgas pledge DoReceive error: %s", err)
	}
	if len(blocks) > 0 {
		return rewardBlk, nil
	}

	return nil, errors.New("can not generate reward recv block")
}

type QGasSwapInfo struct {
	SwapType    string
	FromAddress types.Address
	Amount      types.Balance
	ToAddress   types.Address
	SendHash    types.Hash
	RewardHash  types.Hash
	LinkHash    types.Hash
	Time        string
}

func (q *QGasSwapAPI) GetAllSwapInfos(count int, offset int, isPledge *bool) ([]*QGasSwapInfo, error) {
	c, o, err := checkOffset(count, &offset)
	if err != nil {
		return nil, err
	}
	var infos []*abi.QGasSwapInfo
	if isPledge == nil {
		infos, err = abi.GetQGasSwapInfos(q.l, types.ZeroAddress, abi.QGasAll)
	} else {
		if *isPledge {
			infos, err = abi.GetQGasSwapInfos(q.l, types.ZeroAddress, abi.QGasPledge)
		} else {
			infos, err = abi.GetQGasSwapInfos(q.l, types.ZeroAddress, abi.QGasWithdraw)
		}
	}
	if err != nil {
		return nil, err
	}
	pList := make([]*QGasSwapInfo, 0)
	for _, info := range infos {
		pList = append(pList, q.toSwapInfo(info))
	}
	return infosByOffset(pList, c, o), nil
}

func (q *QGasSwapAPI) GetSwapInfosByAddress(addr types.Address, count int, offset int, isPledge *bool) ([]*QGasSwapInfo, error) {
	c, o, err := checkOffset(count, &offset)
	if err != nil {
		return nil, err
	}
	var infos []*abi.QGasSwapInfo
	if isPledge == nil {
		infos, err = abi.GetQGasSwapInfos(q.l, addr, abi.QGasAll)
	} else {
		if *isPledge {
			infos, err = abi.GetQGasSwapInfos(q.l, addr, abi.QGasPledge)
		} else {
			infos, err = abi.GetQGasSwapInfos(q.l, addr, abi.QGasWithdraw)
		}
	}
	if err != nil {
		return nil, err
	}
	pList := make([]*QGasSwapInfo, 0)
	for _, info := range infos {
		pList = append(pList, q.toSwapInfo(info))
	}
	return infosByOffset(pList, c, o), nil
}

func (q *QGasSwapAPI) toSwapInfo(info *abi.QGasSwapInfo) *QGasSwapInfo {
	pInfo := new(QGasSwapInfo)
	pInfo.SwapType = abi.QGasSwapTypeToString(info.SwapType)
	pInfo.Amount = types.Balance{Int: info.Amount}
	pInfo.FromAddress = info.FromAddress
	pInfo.ToAddress = info.ToAddress
	pInfo.SendHash = info.SendHash
	pInfo.LinkHash = info.LinkHash
	if h, err := q.l.GetBlockLink(pInfo.SendHash); err == nil {
		pInfo.RewardHash = h
	}
	pInfo.Time = time.Unix(info.Time, 0).Format(time.RFC3339)
	return pInfo
}

func (q *QGasSwapAPI) GetSwapAmount() (map[string]*big.Int, error) {
	return abi.GetQGasSwapAmount(q.l, types.ZeroAddress)
}

func (q *QGasSwapAPI) GetSwapAmountByAddress(addr types.Address) (map[string]*big.Int, error) {
	return abi.GetQGasSwapAmount(q.l, addr)
}

func infosByOffset(pList []*QGasSwapInfo, count, offset int) []*QGasSwapInfo {
	sort.Slice(pList, func(i, j int) bool { return pList[i].Time < pList[j].Time })
	if len(pList) > offset {
		if len(pList) >= offset+count {
			return pList[offset : count+offset]
		}
		return pList[offset:]
	} else {
		return make([]*QGasSwapInfo, 0)
	}
}

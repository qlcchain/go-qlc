package api

import (
	"fmt"
	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type DoDBillingApi struct {
	logger *zap.SugaredLogger
	l      ledger.Store
	cc     *chainctx.ChainContext
	ctx    *vmstore.VMContext
	sa     *contract.DoDSetAccount
	ss     *contract.DoDSetService
	us     *contract.DoDUpdateUsage
}

func NewDoDBillingApi(cfgFile string, l ledger.Store) *DoDBillingApi {
	api := &DoDBillingApi{
		l:      l,
		logger: log.NewLogger("dod billing"),
		cc:     chainctx.NewChainContext(cfgFile),
		ctx:    vmstore.NewVMContext(l),
		sa:     &contract.DoDSetAccount{},
		ss:     &contract.DoDSetService{},
		us:     &contract.DoDUpdateUsage{},
	}
	return api
}

func (d *DoDBillingApi) GetSetAccountBlock(param *abi.DoDAccount) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !d.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	povHeader, err := d.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := d.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Account)
	}

	data, _ := abi.DoDBillingABI.PackMethod(abi.MethodNameDoDSetAccount, param.AccountName, param.AccountInfo, param.AccountType, param.UUID)

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Account,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(contractaddress.DoDBillingAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	return send, nil
}

// need to split the params according to different services
type DoDConnection struct {
	Account         types.Address `json:"account"`
	AccountName     string        `json:"accountName"`
	ConnectionID    string        `json:"connectionID"`
	ServiceType     string        `json:"serviceType"`
	ChargeType      string        `json:"chargeType"`
	PaidRule        string        `json:"paidRule"`
	BuyMode         string        `json:"buyMode"`
	Location        string        `json:"location"`
	StartTime       uint64        `json:"startTime"`
	EndTime         uint64        `json:"endTime"`
	Price           float64       `json:"price"`
	Unit            string        `json:"unit"`
	Currency        string        `json:"currency"`
	Balance         float64       `json:"balance"`
	TempStartTime   uint64        `json:"tempStartTime"`
	TempEndTime     uint64        `json:"tempEndTime"`
	TempPrice       float64       `json:"tempPrice"`
	TempBandwidth   string        `json:"tempBandwidth"`
	Bandwidth       string        `json:"bandwidth"`
	Quota           float64       `json:"quota"`
	UsageLimitation float64       `json:"usageLimitation"`
	MinBandwidth    string        `json:"minBandwidth"`
	ExpireTime      uint64        `json:"expireTime"`
}

func (d *DoDBillingApi) GetSetServiceBlock(param *DoDConnection) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !d.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	ct := abi.DoDString2ChargeType(param.ChargeType)
	pr := abi.DoDString2PaidRule(param.PaidRule)
	bm := abi.DoDString2BuyMode(param.BuyMode)

	povHeader, err := d.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := d.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Account)
	}

	data, _ := abi.DoDBillingABI.PackMethod(abi.MethodNameDoDSetService,
		param.ConnectionID,
		param.ServiceType,
		ct,
		pr,
		bm,
		param.Location,
		param.Price,
		param.Unit,
		param.Currency)

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Account,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(contractaddress.DoDBillingAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	return send, nil
}

package api

import (
	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type KYCApi struct {
	logger *zap.SugaredLogger
	l      ledger.Store
	ca     *ContractApi
	cc     *chainctx.ChainContext
	ctx    *vmstore.VMContext
	a      *contract.KYCAdminHandOver
	s      *contract.KYCStatusUpdate
	t      *contract.KYCTradeAddressUpdate
}

func NewKYCApi(cfgFile string, l ledger.Store) *KYCApi {
	api := &KYCApi{
		l:      l,
		logger: log.NewLogger("api kyc status"),
		cc:     chainctx.NewChainContext(cfgFile),
		ctx:    vmstore.NewVMContext(l, &contractaddress.KYCAddress),
		a:      &contract.KYCAdminHandOver{},
		s:      &contract.KYCStatusUpdate{},
		t:      &contract.KYCTradeAddressUpdate{},
	}
	api.ca = NewContractApi(api.cc, api.l)
	return api
}

type KYCAdminUpdateParam struct {
	Admin     types.Address `json:"admin"`
	Successor types.Address `json:"successor"`
	Comment   string        `json:"comment"`
}

type KYCAdminUser struct {
	Account types.Address `json:"admin"`
	Comment string        `json:"comment"`
}

type KYCUpdateStatusParam struct {
	Admin        types.Address `json:"admin"`
	ChainAddress types.Address `json:"chainAddress"`
	Status       string        `json:"status"`
}

type KYCStatusInfo struct {
	ChainAddress types.Address `json:"chainAddress"`
	Status       string        `json:"status"`
}

type KYCUpdateTradeAddressParam struct {
	Admin        types.Address `json:"admin"`
	ChainAddress types.Address `json:"chainAddress"`
	Action       string        `json:"action"`
	TradeAddress string        `json:"tradeAddress"`
	Comment      string        `json:"comment"`
}

type KYCTradeAddress struct {
	Address string `json:"address"`
	Comment string `json:"comment"`
}

type KYCTradeAddressPack struct {
	ChainAddress types.Address      `json:"chainAddress"`
	TradeAddress []*KYCTradeAddress `json:"tradeAddress"`
}

func (p *KYCApi) GetAdminHandoverBlock(param *KYCAdminUpdateParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := abi.KYCStatusABI.PackMethod(abi.MethodNameKYCAdminHandOver, param.Successor, param.Comment)
	if err != nil {
		return nil, err
	}

	pa := &ContractSendBlockPara{
		Address:   param.Admin,
		TokenName: "QLC",
		To:        contractaddress.KYCAddress,
		Amount:    types.NewBalance(0),
		Data:      data,
	}

	return p.ca.GenerateSendBlock(pa)
}

func (p *KYCApi) GetAdmin() (*KYCAdminUser, error) {
	a, err := abi.KYCGetAdmin(p.ctx)
	if err != nil {
		return nil, err
	}

	au := &KYCAdminUser{
		Account: a[0].Account,
		Comment: a[0].Comment,
	}
	return au, nil
}

func (p *KYCApi) GetUpdateStatusBlock(param *KYCUpdateStatusParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := abi.KYCStatusABI.PackMethod(abi.MethodNameKYCStatusUpdate, param.ChainAddress, param.Status)
	if err != nil {
		return nil, err
	}

	pa := &ContractSendBlockPara{
		Address:   param.Admin,
		TokenName: "QLC",
		To:        contractaddress.KYCAddress,
		Amount:    types.NewBalance(0),
		Data:      data,
	}

	return p.ca.GenerateSendBlock(pa)
}

func (p *KYCApi) GetStatusCount() int {
	ss, err := abi.KYCGetAllStatus(p.ctx)
	if err != nil {
		return 0
	} else {
		return len(ss)
	}
}

func (p *KYCApi) GetStatus(count int, offset int) ([]*KYCStatusInfo, error) {
	ss, err := abi.KYCGetAllStatus(p.ctx)
	if err != nil {
		return nil, err
	}

	sis := make([]*KYCStatusInfo, 0)
	for i, n := range ss {
		if i < offset {
			continue
		}

		if i >= offset+count {
			break
		}

		si := &KYCStatusInfo{
			ChainAddress: n.ChainAddress,
			Status:       n.Status,
		}
		sis = append(sis, si)
	}

	return sis, nil
}

func (p *KYCApi) GetStatusByChainAddress(ca types.Address) (*KYCStatusInfo, error) {
	ss, err := abi.KYCGetStatusByChainAddress(p.ctx, ca)
	if err != nil {
		return nil, err
	}

	si := &KYCStatusInfo{
		ChainAddress: ss.ChainAddress,
		Status:       ss.Status,
	}

	return si, nil
}

func (p *KYCApi) GetStatusByTradeAddress(ta string) (*KYCStatusInfo, error) {
	s, err := abi.KYCGetStatusByTradeAddress(p.ctx, ta)
	if err != nil {
		return nil, err
	}

	si := &KYCStatusInfo{
		ChainAddress: s.ChainAddress,
		Status:       s.Status,
	}

	return si, nil
}

func (p *KYCApi) GetUpdateTradeAddressBlock(param *KYCUpdateTradeAddressParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	action, err := abi.KYCTradeAddressActionFromString(param.Action)
	if err != nil {
		return nil, err
	}

	data, err := abi.KYCStatusABI.PackMethod(abi.MethodNameKYCTradeAddressUpdate, param.ChainAddress,
		action, param.TradeAddress, param.Comment)
	if err != nil {
		return nil, err
	}

	pa := &ContractSendBlockPara{
		Address:   param.Admin,
		TokenName: "QLC",
		To:        contractaddress.KYCAddress,
		Amount:    types.NewBalance(0),
		Data:      data,
	}

	return p.ca.GenerateSendBlock(pa)
}

func (p *KYCApi) GetTradeAddress(address types.Address) (*KYCTradeAddressPack, error) {
	kas, err := abi.KYCGetTradeAddress(p.ctx, address)
	if err != nil {
		return nil, err
	}

	ktap := new(KYCTradeAddressPack)
	ktap.ChainAddress = address
	ktap.TradeAddress = make([]*KYCTradeAddress, 0)

	for _, ka := range kas {
		kta := &KYCTradeAddress{
			Address: ka.TradeAddress,
			Comment: ka.Comment,
		}
		ktap.TradeAddress = append(ktap.TradeAddress, kta)
	}

	return ktap, nil
}

package api

import (
	"fmt"

	"go.uber.org/zap"

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
)

type PtmKeyApi struct {
	logger *zap.SugaredLogger
	l      ledger.Store
	cc     *chainctx.ChainContext
	ctx    *vmstore.VMContext
	pu     *contract.PtmKeyUpdate
	pdb    *contract.PtmKeyDelete
}

func NewPtmKeyApi(cfgFile string, l ledger.Store) *PtmKeyApi {
	api := &PtmKeyApi{
		l:      l,
		logger: log.NewLogger("api ptmkey"),
		cc:     chainctx.NewChainContext(cfgFile),
		ctx:    vmstore.NewVMContext(l, &contractaddress.PtmKeyKVAddress),
		pu:     &contract.PtmKeyUpdate{},
		pdb:    &contract.PtmKeyDelete{},
	}
	return api
}

type PtmKeyUpdateParam struct {
	Account types.Address `json:"account"`
	Btype   string        `json:"btype"`
	Pubkey  string        `json:"pubkey"`
}

func (p *PtmKeyApi) GetPtmKeyUpdateBlock(param *PtmKeyUpdateParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	ac, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := ac.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Account)
	}
	pt := common.PtmKeyBtypeFromString(param.Btype)
	if err := abi.PtmKeyInfoCheck(p.ctx, pt, []byte(param.Pubkey)); err != nil {
		return nil, err
	}
	//fmt.Printf("GetPtmKeyUpdateBlock:pt(%d),Pubkey(%s)\n", pt, param.Pubkey)
	data, err := abi.PtmKeyABI.PackMethod(abi.MethodNamePtmKeyUpdate, pt, param.Pubkey)
	if err != nil {
		fmt.Printf("GetPtmKeyUpdateBlock:PackMethod err(%s)\n", err)
	}
	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Account,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           types.ToBalance(ac.CoinVote),
		Network:        types.ToBalance(ac.CoinNetwork),
		Oracle:         types.ToBalance(ac.CoinOracle),
		Storage:        types.ToBalance(ac.CoinStorage),
		Link:           types.Hash(contractaddress.PtmKeyKVAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}
	// fmt.Printf("GetPtmKeyUpdateBlock:data(%s)\n", data)
	// fmt.Printf("GetPtmKeyUpdateBlock:send(%s)\n", send)
	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PtmKeyKVAddress)
	_, _, err = p.pu.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = h
	}
	return send, nil
}

type PtmKeyDeleteParam struct {
	Account types.Address `json:"account"`
	Btype   string        `json:"btype"`
}

func (p *PtmKeyApi) GetPtmKeyDeleteBlock(param *PtmKeyDeleteParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	ac, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, fmt.Errorf("GetAccountMeta: %s", err)
	}

	tm := ac.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Account)
	}
	pt := common.PtmKeyBtypeFromString(param.Btype)
	data, err := abi.PtmKeyABI.PackMethod(abi.MethodNamePtmKeyDelete, pt)
	if err != nil {
		fmt.Printf("GetPtmKeyDeleteBlock:PackMethod err(%s)\n", err)
	}
	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Account,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           types.ToBalance(ac.CoinVote),
		Network:        types.ToBalance(ac.CoinNetwork),
		Oracle:         types.ToBalance(ac.CoinOracle),
		Storage:        types.ToBalance(ac.CoinStorage),
		Link:           types.Hash(contractaddress.PtmKeyKVAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}
	//fmt.Printf("GetPtmKeyDeleteBlock:send(%s)\n", send)
	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PtmKeyKVAddress)
	_, _, err = p.pdb.ProcessSend(vmContext, send)
	if err != nil {
		return nil, fmt.Errorf("ProcessSend: %s", err)
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = h
	}
	return send, nil
}

func (p *PtmKeyApi) GetPtmKeyByAccount(account types.Address) ([]*PtmKeyUpdateParam, error) {
	pk := make([]*PtmKeyUpdateParam, 0)

	pkstorage, err := abi.GetPtmKeyByAccount(p.ctx, account)
	if err != nil {
		return nil, err
	}
	for _, v := range pkstorage {
		vr := &PtmKeyUpdateParam{
			Account: v.Account,
			Btype:   common.PtmKeyBtypeToString(v.Btype),
			Pubkey:  string(v.Pubkey[:]),
		}
		pk = append(pk, vr)
	}
	return pk, nil
}

func (p *PtmKeyApi) GetPtmKeyByAccountAndBtype(account types.Address, Btype string) ([]*PtmKeyUpdateParam, error) {
	pk := make([]*PtmKeyUpdateParam, 0)
	byte := common.PtmKeyBtypeFromString(Btype)
	pkstorage, err := abi.GetPtmKeyByAccountAndBtype(p.ctx, account, byte)
	if err != nil {
		return nil, err
	}
	for _, v := range pkstorage {
		vr := &PtmKeyUpdateParam{
			Account: v.Account,
			Btype:   Btype,
			Pubkey:  string(v.Pubkey[:]),
		}
		pk = append(pk, vr)
	}
	return pk, nil
}

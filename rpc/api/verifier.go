package api

import (
	"fmt"
	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type VerifierApi struct {
	logger *zap.SugaredLogger
	l      *ledger.Ledger
	cc     *chainctx.ChainContext
	vr     *contract.VerifierRegister
	vu     *contract.VerifierUnregister
	ctx    *vmstore.VMContext
}

type VerifierRegInfo struct {
	Account types.Address `json:"account"`
	VType   string        `json:"type"`
	VInfo   string        `json:"id"`
}

type VerifierUnRegInfo struct {
	Account types.Address `json:"account"`
	VType   string        `json:"type"`
}

func NewVerifierApi(cfgFile string, l *ledger.Ledger) *VerifierApi {
	api := &VerifierApi{
		l:      l,
		logger: log.NewLogger("api_verifier"),
		cc:     chainctx.NewChainContext(cfgFile),
		vr:     &contract.VerifierRegister{},
		vu:     &contract.VerifierUnregister{},
		ctx:    vmstore.NewVMContext(l),
	}
	return api
}

func (v *VerifierApi) GetRegisterBlock(param *VerifierRegInfo) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !v.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	vt := types.OracleStringToType(param.VType)
	if err := cabi.VerifierRegInfoCheck(v.ctx, param.Account, vt, param.VInfo); err != nil {
		return nil, err
	}

	am, err := v.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Account)
	}

	minPledgeAmount := types.Balance{Int: types.MinVerifierPledgeAmount}
	if am.CoinOracle.Compare(minPledgeAmount) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have not enough oracle pledge %s, expect %s", param.Account, am.CoinOracle, minPledgeAmount)
	}

	data, err := cabi.VerifierABI.PackMethod(cabi.MethodNameVerifierRegister, param.Account, vt, param.VInfo)
	if err != nil {
		return nil, err
	}

	povHeader, err := v.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

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
		Link:           types.Hash(types.VerifierAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(v.l)
	err = v.vr.SetStorage(vmContext, param.Account, vt, param.VInfo)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (v *VerifierApi) GetUnregisterBlock(param *VerifierUnRegInfo) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !v.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	vt := types.OracleStringToType(param.VType)
	if err := cabi.VerifierUnRegInfoCheck(v.ctx, param.Account, vt); err != nil {
		return nil, err
	}

	data, err := cabi.VerifierABI.PackMethod(cabi.MethodNameVerifierUnregister, param.Account, vt)
	if err != nil {
		return nil, err
	}

	am, err := v.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.Account)
	}

	povHeader, err := v.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

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
		Link:           types.Hash(types.VerifierAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(v.l)
	err = v.vu.SetStorage(vmContext, param.Account, vt)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (v *VerifierApi) GetAllVerifiers() ([]*VerifierRegInfo, error) {
	vrs := make([]*VerifierRegInfo, 0)
	
	rawVr, err := cabi.GetAllVerifiers(v.ctx)
	if err != nil {
		return nil, err
	}
	
	for _, v := range rawVr {
		vr := &VerifierRegInfo{
			Account: v.Account,
			VType:   types.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
		}
		vrs = append(vrs, vr)
	}

	return vrs, nil
}

func (v *VerifierApi) GetVerifiersByType(vType string) ([]*VerifierRegInfo, error) {
	vrs := make([]*VerifierRegInfo, 0)

	vt := types.OracleStringToType(vType)
	if vt == types.OracleTypeInvalid {
		return nil, fmt.Errorf("verifier type err")
	}

	rawVr, err := cabi.GetVerifiersByType(v.ctx, vt)
	if err != nil {
		return nil, err
	}

	for _, v := range rawVr {
		vr := &VerifierRegInfo{
			Account: v.Account,
			VType:   types.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
		}
		vrs = append(vrs, vr)
	}

	return vrs, nil
}

func (v *VerifierApi) GetVerifiersByAccount(account string) ([]*VerifierRegInfo, error) {
	vrs := make([]*VerifierRegInfo, 0)

	addr, err := types.HexToAddress(account)
	if err != nil {
		return nil, fmt.Errorf("account format err(%s)", err)
	}

	rawVr, err := cabi.GetVerifiersByAccount(v.ctx, addr)
	if err != nil {
		return nil, err
	}

	for _, v := range rawVr {
		vr := &VerifierRegInfo{
			Account: v.Account,
			VType:   types.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
		}
		vrs = append(vrs, vr)
	}

	return vrs, nil
}

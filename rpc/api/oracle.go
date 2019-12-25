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

type OracleApi struct {
	logger *zap.SugaredLogger
	l      *ledger.Ledger
	cc     *chainctx.ChainContext
	or     *contract.Oracle
	ctx    *vmstore.VMContext
}

func NewOracleApi(cfgFile string, l *ledger.Ledger) *OracleApi {
	api := &OracleApi{
		l:      l,
		logger: log.NewLogger("api_oracle"),
		cc:     chainctx.NewChainContext(cfgFile),
		or:     &contract.Oracle{},
		ctx:    vmstore.NewVMContext(l),
	}
	return api
}

type OracleParam struct {
	Account types.Address `json:"account"`
	OType   string        `json:"type"`
	OID     string        `json:"id"`
	PubKey  string        `json:"pk"`
	Code    string        `json:"code"`
	Hash    string        `json:"hash"`
}

func (o *OracleApi) GetOracleBlock(param *OracleParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !o.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	id, err := types.Sha256HashData([]byte(param.OID))
	if err != nil {
		return nil, err
	}

	hash, err := types.NewHash(param.Hash)
	if err != nil {
		return nil, err
	}

	pk := types.NewHexBytesFromHex(param.PubKey)
	ot := types.OracleStringToType(param.OType)

	err = cabi.OracleInfoCheck(o.ctx, param.Account, ot, id, pk, param.Code, hash)
	if err != nil {
		return nil, err
	}

	am, err := o.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have gas token", param.Account)
	}

	oracleCost := types.Balance{Int: types.OracleCost}
	if tm.Balance.Compare(oracleCost) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have not enough qgas %s, expect %s", param.Account, tm.Balance, oracleCost)
	}

	data, err := cabi.OracleABI.PackMethod(cabi.MethodNameOracle, param.Account, ot, id, pk, param.Code, hash)
	if err != nil {
		return nil, err
	}

	povHeader, err := o.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Account,
		Balance:        tm.Balance.Sub(oracleCost),
		Previous:       tm.Header,
		Link:           types.Hash(types.OracleAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(o.l)
	err = o.or.SetStorage(vmContext, param.Account, ot, id, pk, param.Code, hash)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (o *OracleApi) GetOracleInfosByType(oType string) ([]*OracleParam, error) {
	ot := types.OracleStringToType(oType)
	oi := make([]*OracleParam, 0)
	var infos []*cabi.OracleInfo

	if oType == "" {
		infos = cabi.GetAllOracleInfo(o.ctx)
	} else {
		infos = cabi.GetOracleInfoByType(o.ctx, ot)
	}

	if infos != nil {
		for _, i := range infos {
			or := &OracleParam{
				Account: i.Account,
				OType:   types.OracleTypeToString(i.OType),
				OID:     i.OID.String(),
				PubKey:  types.NewHexBytesFromData(i.PubKey).String(),
				Code:    i.Code,
				Hash:    i.Hash.String(),
			}
			oi = append(oi, or)
		}
	}

	return oi, nil
}

func (o *OracleApi) GetOracleInfosByAccountAndType(account types.Address, oType string) ([]*OracleParam, error) {
	ot := types.OracleStringToType(oType)
	oi := make([]*OracleParam, 0)
	var infos []*cabi.OracleInfo

	if oType == "" {
		infos = cabi.GetOracleInfoByAccount(o.ctx, account)
	} else {
		infos = cabi.GetOracleInfoByAccountAndType(o.ctx, account, ot)
	}

	if infos != nil {
		for _, i := range infos {
			or := &OracleParam{
				Account: i.Account,
				OType:   types.OracleTypeToString(i.OType),
				OID:     i.OID.String(),
				PubKey:  types.NewHexBytesFromData(i.PubKey).String(),
				Code:    i.Code,
				Hash:    i.Hash.String(),
			}
			oi = append(oi, or)
		}
	}

	return oi, nil
}

func (o *OracleApi) GetOracleInfosByHash(hash string) ([]*OracleParam, error) {
	oi := make([]*OracleParam, 0)
	h, err := types.NewHash(hash)
	if err != nil {
		return nil, err
	}

	infos := cabi.GetOracleInfoByHash(o.ctx, h)
	if infos != nil {
		for _, i := range infos {
			or := &OracleParam{
				Account: i.Account,
				OType:   types.OracleTypeToString(i.OType),
				OID:     i.OID.String(),
				PubKey:  types.NewHexBytesFromData(i.PubKey).String(),
				Code:    i.Code,
				Hash:    i.Hash.String(),
			}
			oi = append(oi, or)
		}
	}

	return oi, nil
}

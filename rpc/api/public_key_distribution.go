package api

import (
	"fmt"
	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
	"time"
)

type PublicKeyDistributionApi struct {
	logger *zap.SugaredLogger
	l      *ledger.Ledger
	cc     *chainctx.ChainContext
	vr     *contract.VerifierRegister
	vu     *contract.VerifierUnregister
	pu     *contract.Publish
	up     *contract.UnPublish
	or     *contract.Oracle
	ctx    *vmstore.VMContext
}

func NewPublicKeyDistributionApi(cfgFile string, l *ledger.Ledger) *PublicKeyDistributionApi {
	api := &PublicKeyDistributionApi{
		l:      l,
		logger: log.NewLogger("api_verifier"),
		cc:     chainctx.NewChainContext(cfgFile),
		vr:     &contract.VerifierRegister{},
		vu:     &contract.VerifierUnregister{},
		pu:     &contract.Publish{},
		up:     &contract.UnPublish{},
		or:     &contract.Oracle{},
		ctx:    vmstore.NewVMContext(l),
	}
	return api
}

type VerifierRegParam struct {
	Account types.Address `json:"account"`
	VType   string        `json:"type"`
	VInfo   string        `json:"id"`
}

type VerifierUnRegParam struct {
	Account types.Address `json:"account"`
	VType   string        `json:"type"`
}

func (p *PublicKeyDistributionApi) GetVerifierRegisterBlock(param *VerifierRegParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	vt := common.OracleStringToType(param.VType)
	if err := abi.VerifierRegInfoCheck(p.ctx, param.Account, vt, param.VInfo); err != nil {
		return nil, err
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Account)
	}

	if am.CoinOracle.Compare(common.MinVerifierPledgeAmount) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have not enough oracle pledge %s, expect %s", param.Account, am.CoinOracle, common.MinVerifierPledgeAmount)
	}

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierRegister, vt, param.VInfo)
	if err != nil {
		return nil, err
	}

	povHeader, err := p.l.GetLatestPovHeader()
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
		Link:           types.Hash(types.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.vr.SetStorage(vmContext, param.Account, vt, param.VInfo)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (p *PublicKeyDistributionApi) GetVerifierUnregisterBlock(param *VerifierUnRegParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	vt := common.OracleStringToType(param.VType)

	vs, _ := abi.GetVerifierInfoByAccountAndType(p.ctx, param.Account, vt)
	if vs == nil {
		return nil, fmt.Errorf("there is no valid verifier to unregister(%s-%s)", param.Account, param.VType)
	}

	if err := abi.VerifierUnRegInfoCheck(p.ctx, param.Account, vt); err != nil {
		return nil, err
	}

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierUnregister, vt)
	if err != nil {
		return nil, err
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.Account)
	}

	povHeader, err := p.l.GetLatestPovHeader()
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
		Link:           types.Hash(types.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.vu.SetStorage(vmContext, param.Account, vt, vs.VInfo)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (p *PublicKeyDistributionApi) GetAllVerifiers() ([]*VerifierRegParam, error) {
	vrs := make([]*VerifierRegParam, 0)

	rawVr, err := abi.GetAllVerifiers(p.ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range rawVr {
		vr := &VerifierRegParam{
			Account: v.Account,
			VType:   common.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
		}
		vrs = append(vrs, vr)
	}

	return vrs, nil
}

func (p *PublicKeyDistributionApi) GetVerifiersByType(vType string) ([]*VerifierRegParam, error) {
	vrs := make([]*VerifierRegParam, 0)

	vt := common.OracleStringToType(vType)
	if vt == common.OracleTypeInvalid {
		return nil, fmt.Errorf("verifier type err")
	}

	rawVr, err := abi.GetVerifiersByType(p.ctx, vt)
	if err != nil {
		return nil, err
	}

	for _, v := range rawVr {
		vr := &VerifierRegParam{
			Account: v.Account,
			VType:   common.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
		}
		vrs = append(vrs, vr)
	}

	return vrs, nil
}

func (p *PublicKeyDistributionApi) GetVerifiersByAccount(account string) ([]*VerifierRegParam, error) {
	vrs := make([]*VerifierRegParam, 0)

	addr, err := types.HexToAddress(account)
	if err != nil {
		return nil, fmt.Errorf("account format err(%s)", err)
	}

	rawVr, err := abi.GetVerifiersByAccount(p.ctx, addr)
	if err != nil {
		return nil, err
	}

	for _, v := range rawVr {
		vr := &VerifierRegParam{
			Account: v.Account,
			VType:   common.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
		}
		vrs = append(vrs, vr)
	}

	return vrs, nil
}

type PublishParam struct {
	Account   types.Address   `json:"account"`
	PType     string          `json:"type"`
	PID       string          `json:"id"`
	PubKey    string          `json:"pubKey"`
	Fee       types.Balance   `json:"fee"`
	Verifiers []types.Address `json:"verifiers"`
	Codes     []types.Hash    `json:"codes"`
	Hash      string          `json:"hash"`
}

type UnPublishParam struct {
	Account types.Address `json:"account"`
	PType   string        `json:"type"`
	PID     string        `json:"id"`
	PubKey  string        `json:"pubKey"`
	Hash    string        `json:"hash"`
}

type VerifierContent struct {
	Account types.Address `json:"account"`
	PubKey  string        `json:"pubKey"`
	Code    string        `json:"code"`
	Hash    types.Hash    `json:"hash"`
}

type PublishRet struct {
	Block     *types.StateBlock           `json:"block"`
	Verifiers map[string]*VerifierContent `json:"verifiers"`
}

func (p *PublicKeyDistributionApi) GetPublishBlock(param *PublishParam) (*PublishRet, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	id, err := types.Sha256HashData([]byte(param.PID))
	if err != nil {
		return nil, fmt.Errorf("id err(%s)", param.PID)
	}

	pk := types.NewHexBytesFromHex(param.PubKey)
	pt := common.OracleStringToType(param.PType)
	if err := abi.PublishInfoCheck(p.ctx, param.Account, pt, id, pk, param.Fee); err != nil {
		return nil, err
	}

	if len(param.Verifiers) < 1 || len(param.Verifiers) > 5 {
		return nil, fmt.Errorf("verifier num err")
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have gas token", param.Account)
	}

	if tm.Balance.Compare(param.Fee) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("fee(%s), account(%s) has not enough qgas", tm.Balance, param.Fee)
	}

	vcs := make(map[string]*VerifierContent)
	verifiers := make([]types.Address, 0)
	codesHash := make([]types.Hash, 0)
	seedBase := time.Now().UnixNano()
	for i, addr := range param.Verifiers {
		vs, err := abi.GetVerifierInfoByAccountAndType(p.ctx, addr, common.OracleTypeEmail)
		if err != nil {
			return nil, err
		}

		code := util.RandomFixedStringWithSeed(common.RandomCodeLen, seedBase+int64(i))

		hashRawCode := append([]byte(param.PubKey), []byte(code)...)
		hashCode, err := types.Sha256HashData(hashRawCode)
		if err != nil {
			return nil, err
		}

		verifiers = append(verifiers, addr)
		codesHash = append(codesHash, hashCode)

		vc := &VerifierContent{
			Account: addr,
			PubKey:  param.PubKey,
			Code:    code,
			Hash:    tm.Header,
		}

		vcs[vs.VInfo] = vc
	}

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, pk[:], verifiers, codesHash, param.Fee.Int)
	if err != nil {
		return nil, err
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Account,
		Balance:        tm.Balance.Sub(param.Fee),
		Previous:       tm.Header,
		Link:           types.Hash(types.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.pu.SetStorage(vmContext, param.Account, pt, id, pk, verifiers, codesHash, param.Fee, tm.Header)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	ret := &PublishRet{
		Block:     send,
		Verifiers: vcs,
	}

	return ret, nil
}

func (p *PublicKeyDistributionApi) GetUnPublishBlock(param *UnPublishParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	id, err := types.Sha256HashData([]byte(param.PID))
	if err != nil {
		return nil, fmt.Errorf("id err(%s)", param.PID)
	}

	pk := types.NewHexBytesFromHex(param.PubKey)
	hash, err := types.NewHash(param.Hash)
	if err != nil {
		return nil, fmt.Errorf("hash(%s) err(%s)", param.Hash, err)
	}

	pt := common.OracleStringToType(param.PType)
	if err := abi.UnPublishInfoCheck(p.ctx, param.Account, pt, id, pk, hash); err != nil {
		return nil, err
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have gas token", param.Account)
	}

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDUnPublish, pt, id, pk, hash)
	if err != nil {
		return nil, err
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Account,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Link:           types.Hash(types.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.up.SetStorage(vmContext, pt, id, pk, hash)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (p *PublicKeyDistributionApi) GetPubKeyByTypeAndID(pType string, pID string) ([]*PublishParam, error) {
	pubs := make([]*PublishParam, 0)

	pt := common.OracleStringToType(pType)
	id, err := types.Sha256HashData([]byte(pID))
	if err != nil {
		return nil, fmt.Errorf("get id hash err")
	}

	infos := abi.GetPublishInfoByTypeAndId(p.ctx, pt, id)
	if infos != nil {
		for _, i := range infos {
			p := &PublishParam{
				Account:   i.Account,
				PType:     common.OracleTypeToString(i.PType),
				PID:       i.PID.String(),
				PubKey:    types.NewHexBytesFromData(i.PubKey).String(),
				Fee:       types.Balance{Int: i.Fee},
				Verifiers: i.Verifiers,
				Codes:     i.Codes,
				Hash:      i.Hash.String(),
			}
			pubs = append(pubs, p)
		}
	}

	return pubs, nil
}

func (p *PublicKeyDistributionApi) GetPublishInfosByType(pType string) ([]*PublishParam, error) {
	pt := common.OracleStringToType(pType)
	pubs := make([]*PublishParam, 0)
	var infos []*abi.PublishInfo

	if pType == "" {
		infos = abi.GetAllPublishInfo(p.ctx)
	} else {
		infos = abi.GetPublishInfoByType(p.ctx, pt)
	}

	if infos != nil {
		for _, i := range infos {
			p := &PublishParam{
				Account:   i.Account,
				PType:     common.OracleTypeToString(i.PType),
				PID:       i.PID.String(),
				PubKey:    types.NewHexBytesFromData(i.PubKey).String(),
				Fee:       types.Balance{Int: i.Fee},
				Verifiers: i.Verifiers,
				Codes:     i.Codes,
				Hash:      i.Hash.String(),
			}
			pubs = append(pubs, p)
		}
	}

	return pubs, nil
}

func (p *PublicKeyDistributionApi) GetPublishInfosByAccountAndType(account types.Address, pType string) ([]*PublishParam, error) {
	pt := common.OracleStringToType(pType)
	pubs := make([]*PublishParam, 0)
	var infos []*abi.PublishInfo

	if pType == "" {
		infos = abi.GetPublishInfoByAccount(p.ctx, account)
	} else {
		infos = abi.GetPublishInfoByAccountAndType(p.ctx, account, pt)
	}

	if infos != nil {
		for _, i := range infos {
			p := &PublishParam{
				Account:   i.Account,
				PType:     common.OracleTypeToString(i.PType),
				PID:       i.PID.String(),
				PubKey:    types.NewHexBytesFromData(i.PubKey).String(),
				Fee:       types.Balance{Int: i.Fee},
				Verifiers: i.Verifiers,
				Codes:     i.Codes,
				Hash:      i.Hash.String(),
			}
			pubs = append(pubs, p)
		}
	}

	return pubs, nil
}

type OracleParam struct {
	Account types.Address `json:"account"`
	OType   string        `json:"type"`
	OID     string        `json:"id"`
	PubKey  string        `json:"pk"`
	Code    string        `json:"code"`
	Hash    string        `json:"hash"`
}

func (p *PublicKeyDistributionApi) GetOracleBlock(param *OracleParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
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
	ot := common.OracleStringToType(param.OType)

	if abi.CheckOracleInfoExist(p.ctx, param.Account, ot, id, pk, hash) {
		return nil, fmt.Errorf("(%s) oracle info for hash(%s) exist", param.Account, hash)
	}

	err = abi.OracleInfoCheck(p.ctx, param.Account, ot, id, pk, param.Code, hash)
	if err != nil {
		return nil, err
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have gas token", param.Account)
	}

	if tm.Balance.Compare(common.OracleCost) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have not enough qgas %s, expect %s", param.Account, tm.Balance, common.OracleCost)
	}

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, param.Code, hash)
	if err != nil {
		return nil, err
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Account,
		Balance:        tm.Balance.Sub(common.OracleCost),
		Previous:       tm.Header,
		Link:           types.Hash(types.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.or.SetStorage(vmContext, param.Account, ot, id, pk, param.Code, hash)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (p *PublicKeyDistributionApi) GetOracleInfosByType(oType string) ([]*OracleParam, error) {
	ot := common.OracleStringToType(oType)
	oi := make([]*OracleParam, 0)
	var infos []*abi.OracleInfo

	if oType == "" {
		infos = abi.GetAllOracleInfo(p.ctx)
	} else {
		infos = abi.GetOracleInfoByType(p.ctx, ot)
	}

	if infos != nil {
		for _, i := range infos {
			or := &OracleParam{
				Account: i.Account,
				OType:   common.OracleTypeToString(i.OType),
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

func (p *PublicKeyDistributionApi) GetOracleInfosByTypeAndID(oType string, id string) ([]*OracleParam, error) {
	ot := common.OracleStringToType(oType)
	oi := make([]*OracleParam, 0)

	idHash, err := types.Sha256HashData([]byte(id))
	if err != nil {
		return nil, err
	}

	infos := abi.GetOracleInfoByTypeAndID(p.ctx, ot, idHash)
	if infos != nil {
		for _, i := range infos {
			or := &OracleParam{
				Account: i.Account,
				OType:   common.OracleTypeToString(i.OType),
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

func (p *PublicKeyDistributionApi) GetOracleInfosByAccountAndType(account types.Address, oType string) ([]*OracleParam, error) {
	ot := common.OracleStringToType(oType)
	oi := make([]*OracleParam, 0)
	var infos []*abi.OracleInfo

	if oType == "" {
		infos = abi.GetOracleInfoByAccount(p.ctx, account)
	} else {
		infos = abi.GetOracleInfoByAccountAndType(p.ctx, account, ot)
	}

	if infos != nil {
		for _, i := range infos {
			or := &OracleParam{
				Account: i.Account,
				OType:   common.OracleTypeToString(i.OType),
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

func (p *PublicKeyDistributionApi) GetOracleInfosByHash(hash string) ([]*OracleParam, error) {
	oi := make([]*OracleParam, 0)
	h, err := types.NewHash(hash)
	if err != nil {
		return nil, err
	}

	infos := abi.GetOracleInfoByHash(p.ctx, h)
	if infos != nil {
		for _, i := range infos {
			or := &OracleParam{
				Account: i.Account,
				OType:   common.OracleTypeToString(i.OType),
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

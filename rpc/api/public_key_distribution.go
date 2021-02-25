package api

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"time"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/contract/dpki"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type PublicKeyDistributionApi struct {
	logger *zap.SugaredLogger
	l      ledger.Store
	cc     *chainctx.ChainContext
	vr     *contract.VerifierRegister
	vu     *contract.VerifierUnregister
	pu     *contract.Publish
	up     *contract.UnPublish
	or     *contract.Oracle
	reward *contract.PKDReward
	ctx    *vmstore.VMContext
}

func NewPublicKeyDistributionApi(cfgFile string, l ledger.Store) *PublicKeyDistributionApi {
	api := &PublicKeyDistributionApi{
		l:      l,
		logger: log.NewLogger("api_verifier"),
		cc:     chainctx.NewChainContext(cfgFile),
		vr:     &contract.VerifierRegister{},
		vu:     &contract.VerifierUnregister{},
		pu:     &contract.Publish{},
		up:     &contract.UnPublish{},
		or:     &contract.Oracle{},
		ctx:    vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress),
	}
	return api
}

type VerifierRegParam struct {
	Account types.Address `json:"account"`
	VType   string        `json:"type"`
	VInfo   string        `json:"id"`
	VKey    string        `json:"key"`
}

type VerifierUnRegParam struct {
	Account types.Address `json:"account"`
	VType   string        `json:"type"`
}

type PublishInfoState struct {
	*PublishParam

	State *types.PovPublishState `json:"state"`
}

func (p *PublicKeyDistributionApi) GetVerifierRegisterBlock(param *VerifierRegParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	vk := types.NewHexBytesFromHex(param.VKey)
	if vk == nil {
		return nil, ErrInvalidParam
	}

	vt := common.OracleStringToType(param.VType)
	if err := abi.VerifierRegInfoCheck(p.ctx, param.Account, vt, param.VInfo, vk); err != nil {
		return nil, err
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Account)
	}

	if am.CoinOracle.Compare(common.MinVerifierPledgeAmount) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have not enough oracle pledge %s, expect %s", param.Account, am.CoinOracle, common.MinVerifierPledgeAmount)
	}

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierRegister, vt, param.VInfo, vk)
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
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Link:           types.Hash(contractaddress.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PubKeyDistributionAddress)
	_, _, err = p.vr.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = h
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

	tm := am.Token(config.ChainToken())
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
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Link:           types.Hash(contractaddress.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PubKeyDistributionAddress)
	_, _, err = p.vu.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = h
	}

	return send, nil
}

func (p *PublicKeyDistributionApi) GetAllVerifiers() ([]*VerifierRegParam, error) {
	vrs := make([]*VerifierRegParam, 0)

	rawVr, err := abi.GetAllVerifiers(p.l)
	if err != nil {
		return nil, err
	}

	for _, v := range rawVr {
		vr := &VerifierRegParam{
			Account: v.Account,
			VType:   common.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
			VKey:    types.NewHexBytesFromData(v.VKey).String(),
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

	rawVr, err := abi.GetVerifiersByType(p.l, vt)
	if err != nil {
		return nil, err
	}

	for _, v := range rawVr {
		vr := &VerifierRegParam{
			Account: v.Account,
			VType:   common.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
			VKey:    types.NewHexBytesFromData(v.VKey).String(),
		}
		vrs = append(vrs, vr)
	}

	return vrs, nil
}

func (p *PublicKeyDistributionApi) GetActiveVerifiers(vType string) ([]*VerifierRegParam, error) {
	vrs := make([]*VerifierRegParam, 0)

	vt := common.OracleStringToType(vType)
	if vt == common.OracleTypeInvalid {
		return nil, fmt.Errorf("verifier type err")
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	rawVr, err := abi.GetVerifiersByType(p.l, vt)
	if err != nil {
		return nil, err
	}

	pb, err := p.l.GetLatestPovBlock()
	if err != nil {
		return nil, err
	}

	csDB := p.getCSDB(false, pb.GetHeight())
	if csDB == nil {
		return nil, errors.New("failed to get contract state db")
	}

	for _, v := range rawVr {
		var ah uint64
		vs, _ := dpki.PovGetVerifierState(csDB, v.Account[:])
		if vs != nil {
			ah, _ = vs.ActiveHeight[vType]
		}

		if pb.GetHeight()-ah > uint64(common.POVChainBlocksPerDay) {
			continue
		}

		vr := &VerifierRegParam{
			Account: v.Account,
			VType:   common.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
			VKey:    types.NewHexBytesFromData(v.VKey).String(),
		}

		vrs = append(vrs, vr)
	}

	// randomize the verifiers that we will get
	if len(vrs) > common.VerifierMaxNum {
		vrr := make([]*VerifierRegParam, 0)
		rnd := rand.Perm(len(vrs))

		for i := 0; i < common.VerifierMaxNum; i++ {
			vrr = append(vrr, vrs[rnd[i]])
		}

		return vrr, nil
	} else {
		return vrs, nil
	}
}

func (p *PublicKeyDistributionApi) GetVerifiersByAccount(account types.Address) ([]*VerifierRegParam, error) {
	vrs := make([]*VerifierRegParam, 0)

	rawVr, err := abi.GetVerifiersByAccount(p.l, account)
	if err != nil {
		return nil, err
	}

	for _, v := range rawVr {
		vr := &VerifierRegParam{
			Account: v.Account,
			VType:   common.OracleTypeToString(v.VType),
			VInfo:   v.VInfo,
			VKey:    types.NewHexBytesFromData(v.VKey).String(),
		}
		vrs = append(vrs, vr)
	}

	return vrs, nil
}

func (p *PublicKeyDistributionApi) GetVerifierStateByBlockHeight(height uint64, address types.Address) (*types.PovVerifierState, error) {
	csDB := p.getCSDB(false, height)
	if csDB == nil {
		return nil, errors.New("failed to get contract state db")
	}
	vsRawKey := address.Bytes()

	vs, err := dpki.PovGetVerifierState(csDB, vsRawKey)
	if err != nil {
		return nil, err
	}

	return vs, nil
}

type PKDVerifierStateList struct {
	VerifierNum  int                                       `json:"verifierNum"`
	AllVerifiers map[types.Address]*types.PovVerifierState `json:"allVerifiers"`
}

func (p *PublicKeyDistributionApi) GetAllVerifierStatesByBlockHeight(height uint64) (*PKDVerifierStateList, error) {
	csDB := p.getCSDB(false, height)
	if csDB == nil {
		return nil, errors.New("failed to get contract state db")
	}

	rspData := new(PKDVerifierStateList)
	rspData.AllVerifiers = make(map[types.Address]*types.PovVerifierState)

	itor := csDB.NewCurTireIterator(statedb.PovCreateContractLocalStateKey(dpki.PovContractStatePrefixPKDVS, nil))

	for key, val, ok := itor.Next(); ok; key, val, ok = itor.Next() {
		verAddr, err := types.BytesToAddress(key[2:])
		if err != nil {
			return nil, fmt.Errorf("deserialize verifier state key err %s", err)
		}

		ps := types.NewPovVerifierState()
		err = ps.Deserialize(val)
		if err != nil {
			return nil, fmt.Errorf("deserialize verifier state value err %s", err)
		}

		rspData.AllVerifiers[verAddr] = ps
	}
	rspData.VerifierNum = len(rspData.AllVerifiers)

	return rspData, nil
}

type PublishParam struct {
	Account   types.Address   `json:"account"`
	PType     string          `json:"type"`
	PID       string          `json:"id"`
	PubKey    string          `json:"pubKey"`
	KeyType   string          `json:"keyType"`
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
	KeyType string        `json:"keyType"`
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
	kt := common.PublicKeyTypeFromString(param.KeyType)
	if err := abi.PublishInfoCheck(p.ctx, param.Account, pt, id, kt, pk, param.Fee); err != nil {
		return nil, err
	}

	if len(param.Verifiers) < 3 || len(param.Verifiers) > 5 {
		return nil, fmt.Errorf("verifier num err")
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.GasToken())
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

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, kt, pk[:], verifiers, codesHash, param.Fee.Int)
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
		Link:           types.Hash(contractaddress.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PubKeyDistributionAddress)
	_, _, err = p.pu.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = h
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

	kt := common.PublicKeyTypeFromString(param.KeyType)
	pk := types.NewHexBytesFromHex(param.PubKey)
	hash, err := types.NewHash(param.Hash)
	if err != nil {
		return nil, fmt.Errorf("hash(%s) err(%s)", param.Hash, err)
	}

	pt := common.OracleStringToType(param.PType)
	if err := abi.UnPublishInfoCheck(p.ctx, param.Account, pt, id, kt, pk, hash); err != nil {
		return nil, err
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have gas token", param.Account)
	}

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDUnPublish, pt, id, kt, pk, hash)
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
		Link:           types.Hash(contractaddress.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PubKeyDistributionAddress)
	_, _, err = p.up.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = h
	}

	return send, nil
}

func (p *PublicKeyDistributionApi) getCSDB(isLatest bool, height uint64) *statedb.PovContractStateDB {
	var povHdr *types.PovHeader

	if isLatest {
		povHdr, _ = p.l.GetLatestPovHeader()
	} else {
		povHdr, _ = p.l.GetPovHeaderByHeight(height)
	}

	if povHdr != nil {
		gsdb := statedb.NewPovGlobalStateDB(p.l.DBStore(), povHdr.GetStateHash())
		csdb, _ := gsdb.LookupContractStateDB(contractaddress.PubKeyDistributionAddress)
		return csdb
	}

	return nil
}

func (p *PublicKeyDistributionApi) sortPublishInfo(pubs []*PublishInfoState) {
	sort.Slice(pubs, func(i, j int) bool {
		psi := pubs[i].State
		psj := pubs[j].State

		if psi == nil {
			psi = new(types.PovPublishState)
		}

		if psj == nil {
			psj = new(types.PovPublishState)
		}

		if psi.VerifiedStatus == types.PovPublishStatusVerified {
			if psj.VerifiedStatus != types.PovPublishStatusVerified {
				return true
			} else {
				return psi.VerifiedHeight > psj.VerifiedHeight
			}
		} else {
			return psi.PublishHeight < psj.PublishHeight
		}
	})
}

func (p *PublicKeyDistributionApi) fillPublishInfoState(csdb *statedb.PovContractStateDB, pubInfo *abi.PublishInfo) *PublishInfoState {
	kh := common.PublicKeyWithTypeHash(pubInfo.KeyType, pubInfo.PubKey)
	pubInfoKey := &abi.PublishInfoKey{
		PType:  pubInfo.PType,
		PID:    pubInfo.PID,
		PubKey: kh,
		Hash:   pubInfo.Hash,
	}
	psRawKey := pubInfoKey.ToRawKey()

	pis := &PublishInfoState{}
	pis.PublishParam = &PublishParam{
		Account:   pubInfo.Account,
		PType:     common.OracleTypeToString(pubInfo.PType),
		PID:       pubInfo.PID.String(),
		KeyType:   common.PublicKeyTypeToString(pubInfo.KeyType),
		PubKey:    types.NewHexBytesFromData(pubInfo.PubKey).String(),
		Fee:       types.Balance{Int: pubInfo.Fee},
		Verifiers: pubInfo.Verifiers,
		Codes:     pubInfo.Codes,
		Hash:      pubInfo.Hash.String(),
	}
	if csdb != nil {
		pis.State, _ = dpki.PovGetPublishState(csdb, psRawKey)
	}

	return pis
}

func (p *PublicKeyDistributionApi) GetPubKeyByTypeAndID(pType string, pID string) ([]*PublishInfoState, error) {
	pubs := make([]*PublishInfoState, 0)

	pt := common.OracleStringToType(pType)
	id, err := types.Sha256HashData([]byte(pID))
	if err != nil {
		return nil, fmt.Errorf("get id hash err")
	}

	csDB := p.getCSDB(true, 0)

	infos := abi.GetPublishInfoByTypeAndId(p.l, pt, id)
	if infos != nil {
		for _, i := range infos {
			pis := p.fillPublishInfoState(csDB, i)
			pubs = append(pubs, pis)
		}
	}

	p.sortPublishInfo(pubs)
	return pubs, nil
}

func (p *PublicKeyDistributionApi) GetRecommendPubKey(pType string, pID string) (*PublishInfoState, error) {
	pis, err := p.GetPubKeyByTypeAndID(pType, pID)
	if err != nil {
		return nil, err
	}

	if len(pis) > 0 {
		return pis[0], nil
	}

	return nil, nil
}

func (p *PublicKeyDistributionApi) GetPublishInfosByType(pType string) ([]*PublishInfoState, error) {
	pt := common.OracleStringToType(pType)
	pubs := make([]*PublishInfoState, 0)
	var infos []*abi.PublishInfo

	if pType == "" {
		infos = abi.GetAllPublishInfo(p.l)
	} else {
		infos = abi.GetPublishInfoByType(p.l, pt)
	}

	if infos != nil {
		csDB := p.getCSDB(true, 0)

		for _, i := range infos {
			pis := p.fillPublishInfoState(csDB, i)
			pubs = append(pubs, pis)
		}
	}

	p.sortPublishInfo(pubs)
	return pubs, nil
}

func (p *PublicKeyDistributionApi) GetPublishInfosByAccountAndType(account types.Address, pType string) ([]*PublishInfoState, error) {
	pt := common.OracleStringToType(pType)
	pubs := make([]*PublishInfoState, 0)
	var infos []*abi.PublishInfo

	if pType == "" {
		infos = abi.GetPublishInfoByAccount(p.l, account)
	} else {
		infos = abi.GetPublishInfoByAccountAndType(p.l, account, pt)
	}

	if infos != nil {
		csDB := p.getCSDB(true, 0)

		for _, i := range infos {
			pis := p.fillPublishInfoState(csDB, i)
			pubs = append(pubs, pis)
		}
	}

	p.sortPublishInfo(pubs)
	return pubs, nil
}

type OracleParam struct {
	Account types.Address `json:"account"`
	OType   string        `json:"type"`
	OID     string        `json:"id"`
	KeyType string        `json:"keyType"`
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

	kt := common.PublicKeyTypeFromString(param.KeyType)
	pk := types.NewHexBytesFromHex(param.PubKey)
	ot := common.OracleStringToType(param.OType)

	if abi.CheckOracleInfoExist(p.ctx, param.Account, ot, id, kt, pk, hash) {
		return nil, fmt.Errorf("(%s) oracle info for hash(%s) exist", param.Account, hash)
	}

	err = abi.OracleInfoCheck(p.ctx, param.Account, ot, id, kt, pk, param.Code, hash)
	if err != nil {
		return nil, err
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have gas token", param.Account)
	}

	if tm.Balance.Compare(common.OracleCost) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have not enough qgas %s, expect %s", param.Account, tm.Balance, common.OracleCost)
	}

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, param.Code, hash)
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
		Link:           types.Hash(contractaddress.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PubKeyDistributionAddress)
	_, _, err = p.or.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = h
	}

	return send, nil
}

func (p *PublicKeyDistributionApi) oracleParamConvert(i *abi.OracleInfo) *OracleParam {
	return &OracleParam{
		Account: i.Account,
		OType:   common.OracleTypeToString(i.OType),
		OID:     i.OID.String(),
		KeyType: common.PublicKeyTypeToString(i.KeyType),
		PubKey:  types.NewHexBytesFromData(i.PubKey).String(),
		Code:    i.Code,
		Hash:    i.Hash.String(),
	}
}

func (p *PublicKeyDistributionApi) GetOracleInfosByType(oType string) ([]*OracleParam, error) {
	ot := common.OracleStringToType(oType)
	oi := make([]*OracleParam, 0)
	var infos []*abi.OracleInfo

	if oType == "" {
		infos = abi.GetAllOracleInfo(p.l)
	} else {
		infos = abi.GetOracleInfoByType(p.l, ot)
	}

	if infos != nil {
		for _, i := range infos {
			or := p.oracleParamConvert(i)
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

	infos := abi.GetOracleInfoByTypeAndID(p.l, ot, idHash)
	if infos != nil {
		for _, i := range infos {
			or := p.oracleParamConvert(i)
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
		infos = abi.GetOracleInfoByAccount(p.l, account)
	} else {
		infos = abi.GetOracleInfoByAccountAndType(p.l, account, ot)
	}

	if infos != nil {
		for _, i := range infos {
			or := p.oracleParamConvert(i)
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

	infos := abi.GetOracleInfoByHash(p.l, h)
	if infos != nil {
		for _, i := range infos {
			or := p.oracleParamConvert(i)
			oi = append(oi, or)
		}
	}

	return oi, nil
}

type PKDRewardParam struct {
	Account      types.Address `json:"account"`
	Beneficial   types.Address `json:"beneficial"`
	EndHeight    uint64        `json:"endHeight"`
	RewardAmount *big.Int      `json:"rewardAmount"`
}

func (p *PublicKeyDistributionApi) PackRewardData(param *PKDRewardParam) ([]byte, error) {
	return abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDReward,
		param.Account, param.Beneficial, param.EndHeight, param.RewardAmount)
}

func (p *PublicKeyDistributionApi) UnpackRewardData(data []byte) (*PKDRewardParam, error) {
	abiParam := new(dpki.PKDRewardParam)
	err := abi.PublicKeyDistributionABI.UnpackMethod(abiParam, abi.MethodNamePKDReward, data)
	if err != nil {
		return nil, err
	}
	apiParam := new(PKDRewardParam)
	apiParam.Account = abiParam.Account
	apiParam.Beneficial = abiParam.Beneficial
	apiParam.EndHeight = abiParam.EndHeight
	apiParam.RewardAmount = abiParam.RewardAmount
	return apiParam, nil
}

func (p *PublicKeyDistributionApi) GetRewardSendBlock(param *PKDRewardParam) (*types.StateBlock, error) {
	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if param.Account.IsZero() {
		return nil, errors.New("invalid reward param account")
	}

	if param.Beneficial.IsZero() {
		return nil, errors.New("invalid reward param beneficial")
	}

	if param.RewardAmount == nil || param.RewardAmount.Sign() <= 0 {
		return nil, errors.New("invalid reward param rewardAmount")
	}

	am, err := p.l.GetAccountMeta(param.Account)
	if am == nil {
		return nil, fmt.Errorf("account not exist, %s", err)
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("account does not have gas token")
	}

	data, err := p.PackRewardData(param)
	if err != nil {
		return nil, err
	}

	latestPovHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:    types.ContractSend,
		Address: param.Account,

		Token:          tm.Type,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Representative: tm.Representative,

		//Vote:    types.ZeroBalance,
		//Network: types.ZeroBalance,
		//Oracle:  types.ZeroBalance,
		//Storage: types.ZeroBalance,

		Link:      types.Hash(contractaddress.PubKeyDistributionAddress),
		Data:      data,
		Timestamp: common.TimeNow().Unix(),

		PoVHeight: latestPovHeader.GetHeight(),
	}

	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PubKeyDistributionAddress)
	_, _, err = p.reward.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = h
	}

	return send, nil
}

func (p *PublicKeyDistributionApi) GetRewardRecvBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if input.GetType() != types.ContractSend {
		return nil, errors.New("input block type is not contract send")
	}
	if input.GetLink() != contractaddress.PubKeyDistributionAddress.ToHash() {
		return nil, errors.New("input address is not contract PublicKeyDistribution")
	}

	reward := &types.StateBlock{}

	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PubKeyDistributionAddress)
	blocks, err := p.reward.DoReceive(vmContext, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		return reward, nil
	}

	return nil, errors.New("can not generate reward recv block")
}

func (p *PublicKeyDistributionApi) GetRewardRecvBlockBySendHash(sendHash types.Hash) (*types.StateBlock, error) {
	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	input, err := p.l.GetStateBlock(sendHash)
	if err != nil {
		return nil, err
	}

	return p.GetRewardRecvBlock(input)
}

type PKDHistoryRewardInfo struct {
	LastEndHeight  uint64        `json:"lastEndHeight"`
	LastBeneficial types.Address `json:"lastBeneficial"`
	LastRewardTime int64         `json:"lastRewardTime"`
	RewardAmount   types.Balance `json:"rewardAmount"`
}

func (p *PublicKeyDistributionApi) GetRewardHistory(account types.Address) (*PKDHistoryRewardInfo, error) {
	history := new(PKDHistoryRewardInfo)
	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PubKeyDistributionAddress)
	info, err := p.reward.GetRewardInfo(vmContext, account)
	if err != nil {
		return nil, err
	}

	history.LastEndHeight = info.EndHeight
	history.LastBeneficial = info.Beneficial
	history.LastRewardTime = info.Timestamp
	history.RewardAmount = types.Balance{Int: info.RewardAmount}

	return history, nil
}

type PKDAvailRewardInfo struct {
	LastEndHeight     uint64        `json:"lastEndHeight"`
	LatestBlockHeight uint64        `json:"latestBlockHeight"`
	NodeRewardHeight  uint64        `json:"nodeRewardHeight"`
	AvailEndHeight    uint64        `json:"availEndHeight"`
	AvailRewardAmount types.Balance `json:"availRewardAmount"`
	NeedCallReward    bool          `json:"needCallReward"`
}

func (p *PublicKeyDistributionApi) GetAvailRewardInfo(account types.Address) (*PKDAvailRewardInfo, error) {
	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	rsp := new(PKDAvailRewardInfo)

	latestPovHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}
	rsp.LatestBlockHeight = latestPovHeader.GetHeight()

	vmContext := vmstore.NewVMContext(p.l, &contractaddress.PubKeyDistributionAddress)

	lastRwdInfo, _ := p.GetRewardHistory(account)
	if lastRwdInfo != nil {
		rsp.LastEndHeight = lastRwdInfo.LastEndHeight
	}

	rsp.NodeRewardHeight, err = abi.PovGetNodeRewardHeightByDay(vmContext)
	if err != nil {
		return nil, err
	}

	if rsp.LastEndHeight >= rsp.NodeRewardHeight {
		return rsp, err
	}

	rsp.AvailEndHeight = rsp.NodeRewardHeight

	var lastVs *types.PovVerifierState
	if lastRwdInfo != nil {
		lastVs, err = p.reward.GetVerifierState(vmContext, rsp.LastEndHeight, account)
		if err != nil {
			return nil, err
		}
	} else {
		lastVs = types.NewPovVerifierState()
	}

	curVs, err := p.reward.GetVerifierState(vmContext, rsp.NodeRewardHeight, account)
	if err != nil {
		return nil, err
	}

	availRwdAmount := types.NewBigNumFromInt(0).Sub(curVs.TotalReward, lastVs.TotalReward)
	rsp.AvailRewardAmount = types.NewBalanceFromBigInt(availRwdAmount.ToBigInt())

	if rsp.AvailEndHeight <= rsp.NodeRewardHeight && rsp.AvailRewardAmount.Int64() > 0 {
		rsp.NeedCallReward = true
	}

	return rsp, nil
}

func (p *PublicKeyDistributionApi) GetVerifierHeartBlock(account types.Address, vTypes []string) (*types.StateBlock, error) {
	if len(vTypes) == 0 {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	vts := make([]uint32, 0)
	for _, v := range vTypes {
		vt := common.OracleStringToType(v)
		if vt == common.OracleTypeInvalid {
			return nil, ErrVerifierType
		}

		err := abi.VerifierPledgeCheck(p.ctx, account)
		if err != nil {
			return nil, contract.ErrNotEnoughPledge
		}

		_, err = abi.GetVerifierInfoByAccountAndType(p.ctx, account, vt)
		if err != nil {
			return nil, contract.ErrGetVerifier
		}

		vts = append(vts, vt)
	}

	am, err := p.l.GetAccountMeta(account)
	if err != nil {
		return nil, contract.ErrAccountNotExist
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, ErrNoGas
	}

	if tm.Balance.Compare(common.OracleCost) == types.BalanceCompSmaller {
		return nil, contract.ErrNotEnoughFee
	}

	data, err := abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierHeart, vts)
	if err != nil {
		return nil, contract.ErrPackMethod
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, ErrGetPovHeader
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        account,
		Balance:        tm.Balance.Sub(common.OracleCost),
		Previous:       tm.Header,
		Link:           types.Hash(contractaddress.PubKeyDistributionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	return send, nil
}

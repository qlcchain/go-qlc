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
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
	"time"
)

type PublisherApi struct {
	logger *zap.SugaredLogger
	l      *ledger.Ledger
	cc     *chainctx.ChainContext
	pub    *contract.Publish
	up     *contract.UnPublish
	ctx    *vmstore.VMContext
}

func NewPublisherApi(cfgFile string, l *ledger.Ledger) *PublisherApi {
	api := &PublisherApi{
		l:      l,
		logger: log.NewLogger("api_oracle"),
		cc:     chainctx.NewChainContext(cfgFile),
		pub:    &contract.Publish{},
		up:     &contract.UnPublish{},
		ctx:    vmstore.NewVMContext(l),
	}
	return api
}

type PublishParam struct {
	Account   types.Address   `json:"account"`
	PType     string          `json:"type"`
	PID       string          `json:"id"`
	PubKey    string          `json:"pubKey"`
	Fee       types.Balance   `json:"fee"`
	Verifiers []types.Address `json:"verifiers"`
	Codes     []types.Hash    `json:"codes"`
}

type UnPublishParam struct {
	Account types.Address `json:"account"`
	PType   string        `json:"type"`
	PID     string        `json:"id"`
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

func (p *PublisherApi) GetPublishBlock(param *PublishParam) (*PublishRet, error) {
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
	pt := types.OracleStringToType(param.PType)
	if err := cabi.PublishInfoCheck(p.ctx, param.Account, pt, id, pk, param.Fee); err != nil {
		return nil, err
	}

	if len(param.Verifiers) < 1 || len(param.Verifiers) > 5 {
		return nil, fmt.Errorf("verifier num err")
	}

	vcs := make(map[string]*VerifierContent)
	verifiers := make([]types.Address, 0)
	codesHash := make([]types.Hash, 0)
	seedBase := time.Now().UnixNano()
	for i, addr := range param.Verifiers {
		email, err := cabi.GetVerifierInfoByAccountAndType(p.ctx, addr, types.OracleTypeEmail)
		if err != nil {
			return nil, err
		}

		code := util.RandomFixedStringWithSeed(types.RandomCodeLen, seedBase+int64(i))

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
		}

		vcs[email] = vc
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

	data, err := cabi.PublisherABI.PackMethod(cabi.MethodNamePublish, param.Account, pt, id, pk[:], verifiers, codesHash, param.Fee.Int)
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
		Link:           types.Hash(types.PublisherAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.pub.SetStorage(vmContext, param.Account, pt, id, pk, verifiers, codesHash, param.Fee)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	sendHash := send.GetHash()
	for _, vc := range vcs {
		vc.Hash = sendHash
	}

	ret := &PublishRet{
		Block:     send,
		Verifiers: vcs,
	}

	return ret, nil
}

func (p *PublisherApi) GetUnPublishBlock(param *UnPublishParam) (*types.StateBlock, error) {
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

	pt := types.OracleStringToType(param.PType)
	if err := cabi.UnPublishInfoCheck(p.ctx, param.Account, pt, id); err != nil {
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

	data, err := cabi.PublisherABI.PackMethod(cabi.MethodNameUnPublish, param.Account, pt, id)
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
		Link:           types.Hash(types.PublisherAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.up.SetStorage(vmContext, param.Account, pt, id)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (p *PublisherApi) GetPubKeyByTypeAndID(pType string, pID string) ([]*PublishParam, error) {
	pubs := make([]*PublishParam, 0)

	pt := types.OracleStringToType(pType)
	id, err := types.Sha256HashData([]byte(pID))
	if err != nil {
		return nil, fmt.Errorf("get id hash err")
	}

	infos := cabi.GetPublishInfoByTypeAndId(p.ctx, pt, id)
	if infos != nil {
		for _, i := range infos {
			p := &PublishParam{
				Account:   i.Account,
				PType:     types.OracleTypeToString(i.PType),
				PID:       i.PID.String(),
				PubKey:    types.NewHexBytesFromData(i.PubKey).String(),
				Fee:       types.Balance{Int: i.Fee},
				Verifiers: i.Verifiers,
				Codes:     i.Codes,
			}
			pubs = append(pubs, p)
		}
	}

	return pubs, nil
}

func (p *PublisherApi) GetPublishInfosByType(pType string) ([]*PublishParam, error) {
	pt := types.OracleStringToType(pType)
	pubs := make([]*PublishParam, 0)
	var infos []*cabi.PublishInfo

	if pType == "" {
		infos = cabi.GetAllPublishInfo(p.ctx)
	} else {
		infos = cabi.GetPublishInfoByType(p.ctx, pt)
	}

	if infos != nil {
		for _, i := range infos {
			p := &PublishParam{
				Account:   i.Account,
				PType:     types.OracleTypeToString(i.PType),
				PID:       i.PID.String(),
				PubKey:    types.NewHexBytesFromData(i.PubKey).String(),
				Fee:       types.Balance{Int: i.Fee},
				Verifiers: i.Verifiers,
				Codes:     i.Codes,
			}
			pubs = append(pubs, p)
		}
	}

	return pubs, nil
}

func (p *PublisherApi) GetPublishInfosByAccountAndType(account types.Address, pType string) ([]*PublishParam, error) {
	pt := types.OracleStringToType(pType)
	pubs := make([]*PublishParam, 0)
	var infos []*cabi.PublishInfo

	if pType == "" {
		infos = cabi.GetPublishInfoByAccount(p.ctx, account)
	} else {
		infos = cabi.GetPublishInfoByAccountAndType(p.ctx, account, pt)
	}

	if infos != nil {
		for _, i := range infos {
			p := &PublishParam{
				Account:   i.Account,
				PType:     types.OracleTypeToString(i.PType),
				PID:       i.PID.String(),
				PubKey:    types.NewHexBytesFromData(i.PubKey).String(),
				Fee:       types.Balance{Int: i.Fee},
				Verifiers: i.Verifiers,
				Codes:     i.Codes,
			}
			pubs = append(pubs, p)
		}
	}

	return pubs, nil
}

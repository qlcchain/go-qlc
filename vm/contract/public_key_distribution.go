package contract

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/qlcchain/go-qlc/vm/contract/dpki"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type VerifierRegister struct {
	BaseContract
}

func (vr *VerifierRegister) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != common.ChainToken() {
		return nil, nil, fmt.Errorf("not qlc chain")
	}

	reg := new(abi.VerifierRegInfo)
	err := abi.PublicKeyDistributionABI.UnpackMethod(reg, abi.MethodNamePKDVerifierRegister, block.Data)
	if err != nil {
		return nil, nil, err
	}

	err = abi.VerifierPledgeCheck(ctx, block.Address)
	if err != nil {
		return nil, nil, err
	}

	err = abi.VerifierRegInfoCheck(ctx, block.Address, reg.VType, reg.VInfo)
	if err != nil {
		return nil, nil, err
	}

	block.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierRegister, reg.VType, reg.VInfo)
	if err != nil {
		return nil, nil, err
	}

	err = vr.SetStorage(ctx, block.Address, reg.VType, reg.VInfo)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (vr *VerifierRegister) SetStorage(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) error {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDVerifierInfo, vInfo, true)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PKDStorageTypeVerifier)
	key = append(key, util.BE_Uint32ToBytes(vType)...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
	if err != nil {
		return err
	}

	return nil
}

func (vr *VerifierRegister) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.NewBalance(0), nil
}

func (vr *VerifierRegister) GetRefundData() []byte {
	return []byte{1}
}

func (vr *VerifierRegister) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	return common.ContractNoGap, nil, nil
}

func (vr *VerifierRegister) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

type VerifierUnregister struct {
	BaseContract
}

func (vu *VerifierUnregister) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != common.ChainToken() {
		return nil, nil, fmt.Errorf("not qlc chain")
	}

	reg := new(abi.VerifierRegInfo)
	err := abi.PublicKeyDistributionABI.UnpackMethod(reg, abi.MethodNamePKDVerifierUnregister, block.Data)
	if err != nil {
		return nil, nil, err
	}

	err = abi.VerifierUnRegInfoCheck(ctx, block.Address, reg.VType)
	if err != nil {
		return nil, nil, err
	}

	vs, _ := abi.GetVerifierInfoByAccountAndType(ctx, block.Address, reg.VType)
	if vs == nil {
		return nil, nil, fmt.Errorf("there is no valid verifier to unregister(%s-%s)", block.Address, common.OracleTypeToString(reg.VType))
	}

	block.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierUnregister, reg.VType)
	if err != nil {
		return nil, nil, err
	}

	err = vu.SetStorage(ctx, block.Address, reg.VType, vs.VInfo)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (vu *VerifierUnregister) SetStorage(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) error {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDVerifierInfo, vInfo, false)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PKDStorageTypeVerifier)
	key = append(key, util.BE_Uint32ToBytes(vType)...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
	if err != nil {
		return err
	}

	return nil
}

func (vu *VerifierUnregister) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.NewBalance(0), nil
}

func (vu *VerifierUnregister) GetRefundData() []byte {
	return []byte{1}
}

func (vu *VerifierUnregister) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	return common.ContractNoGap, nil, nil
}

func (vu *VerifierUnregister) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

type Publish struct {
	BaseContract
}

func (p *Publish) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != common.GasToken() {
		return nil, nil, fmt.Errorf("not gas chain")
	}

	info := new(abi.PublishInfo)
	err := abi.PublicKeyDistributionABI.UnpackMethod(info, abi.MethodNamePKDPublish, block.GetData())
	if err != nil {
		return nil, nil, err
	}

	fee := types.Balance{Int: info.Fee}
	err = abi.PublishInfoCheck(ctx, block.Address, info.PType, info.PID, info.PubKey, fee)
	if err != nil {
		return nil, nil, err
	}

	amount, err := ctx.CalculateAmount(block)
	if err != nil {
		return nil, nil, err
	}

	if amount.Compare(fee) != types.BalanceCompEqual {
		return nil, nil, fmt.Errorf("balance mismatch(data:%s--amount:%s)", fee, amount)
	}

	block.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, info.PType, info.PID,
		info.PubKey, info.Verifiers, info.Codes, info.Fee)
	if err != nil {
		return nil, nil, err
	}

	err = p.SetStorage(ctx, block.Address, info.PType, info.PID, info.PubKey, info.Verifiers, info.Codes, fee, block.Previous)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (p *Publish) SetStorage(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash, pk []byte,
	vs []types.Address, cs []types.Hash, fee types.Balance, hash types.Hash) error {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDPublishInfo, account, vs, cs, fee.Int, true)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PKDStorageTypePublisher)
	key = append(key, util.BE_Uint32ToBytes(pt)...)
	key = append(key, id[:]...)
	key = append(key, pk...)
	key = append(key, hash[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
	if err != nil {
		return err
	}

	return nil
}

func (p *Publish) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.NewBalance(0), nil
}

func (p *Publish) GetRefundData() []byte {
	return []byte{1}
}

func (p *Publish) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	return common.ContractNoGap, nil, nil
}

func (p *Publish) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

type UnPublish struct {
	BaseContract
}

func (up *UnPublish) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != common.GasToken() {
		return nil, nil, fmt.Errorf("not gas chain")
	}

	info := new(abi.UnPublishInfo)
	err := abi.PublicKeyDistributionABI.UnpackMethod(info, abi.MethodNamePKDUnPublish, block.GetData())
	if err != nil {
		return nil, nil, err
	}

	err = abi.UnPublishInfoCheck(ctx, block.Address, info.PType, info.PID, info.PubKey, info.Hash)
	if err != nil {
		return nil, nil, err
	}

	block.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDUnPublish, info.PType, info.PID, info.PubKey, info.Hash)
	if err != nil {
		return nil, nil, err
	}

	err = up.SetStorage(ctx, info.PType, info.PID, info.PubKey, info.Hash)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (up *UnPublish) SetStorage(ctx *vmstore.VMContext, pt uint32, id types.Hash, pk []byte, hash types.Hash) error {
	var key []byte
	key = append(key, abi.PKDStorageTypePublisher)
	key = append(key, util.BE_Uint32ToBytes(pt)...)
	key = append(key, id[:]...)
	key = append(key, pk...)
	key = append(key, hash[:]...)
	dataOld, err := ctx.GetStorage(types.PubKeyDistributionAddress[:], key)
	if err != nil {
		return err
	}

	var info abi.PubKeyInfo
	err = abi.PublicKeyDistributionABI.UnpackVariable(&info, abi.VariableNamePKDPublishInfo, dataOld)
	if err != nil {
		return nil
	}

	dataNew, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDPublishInfo, info.Account, info.Verifiers, info.Codes, info.Fee, false)
	if err != nil {
		return err
	}

	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, dataNew)
	if err != nil {
		return err
	}

	return nil
}

func (up *UnPublish) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.NewBalance(0), nil
}

func (up *UnPublish) GetRefundData() []byte {
	return []byte{1}
}

func (up *UnPublish) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	return common.ContractNoGap, nil, nil
}

func (up *UnPublish) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

type Oracle struct {
	BaseContract
}

func (o *Oracle) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != common.GasToken() {
		return nil, nil, fmt.Errorf("not gas chain")
	}

	info := new(abi.OracleInfo)
	err := abi.PublicKeyDistributionABI.UnpackMethod(info, abi.MethodNamePKDOracle, block.GetData())
	if err != nil {
		return nil, nil, err
	}

	// check verifier if the block is not synced
	if !block.IsFromSync() {
		err = abi.VerifierPledgeCheck(ctx, block.GetAddress())
		if err != nil {
			return nil, nil, err
		}

		_, err = abi.GetVerifierInfoByAccountAndType(ctx, block.GetAddress(), info.OType)
		if err != nil {
			return nil, nil, err
		}
	}

	err = abi.OracleInfoCheck(ctx, block.Address, info.OType, info.OID, info.PubKey, info.Code, info.Hash)
	if err != nil {
		return nil, nil, err
	}

	amount, err := ctx.CalculateAmount(block)
	if err != nil {
		return nil, nil, err
	}

	if amount.Compare(common.OracleCost) != types.BalanceCompEqual {
		return nil, nil, fmt.Errorf("balance(exp:%s-%s) wrong", common.OracleCost, amount)
	}

	block.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, info.OType, info.OID, info.PubKey, info.Code, info.Hash)
	if err != nil {
		return nil, nil, err
	}

	err = o.SetStorage(ctx, block.Address, info.OType, info.OID, info.PubKey, info.Code, info.Hash)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (o *Oracle) SetStorage(ctx *vmstore.VMContext, account types.Address, ot uint32, id types.Hash, pk []byte, code string, hash types.Hash) error {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDOracleInfo, code)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PKDStorageTypeOracle)
	key = append(key, util.BE_Uint32ToBytes(ot)...)
	key = append(key, id[:]...)
	key = append(key, pk...)
	key = append(key, hash[:]...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
	if err != nil {
		return err
	}

	return nil
}

func (o *Oracle) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.NewBalance(0), nil
}

func (o *Oracle) GetRefundData() []byte {
	return []byte{1}
}

func (o *Oracle) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	info := new(abi.OracleInfo)
	err := abi.PublicKeyDistributionABI.UnpackMethod(info, abi.MethodNamePKDOracle, block.GetData())
	if err != nil {
		return common.ContractNoGap, nil, err
	}

	pi := abi.GetPublishInfo(ctx, info.OType, info.OID, info.PubKey, info.Hash)
	if pi == nil {
		return common.ContractDPKIGapPublish, nil, nil
	}

	return common.ContractNoGap, nil, nil
}

func (o *Oracle) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

func (o *Oracle) DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error {
	oraInfo := new(abi.OracleInfo)
	err := abi.PublicKeyDistributionABI.UnpackMethod(oraInfo, abi.MethodNamePKDOracle, block.GetData())
	if err != nil {
		return err
	}

	pubInfoKey := &abi.PublishInfoKey{
		PType:  oraInfo.OType,
		PID:    oraInfo.OID,
		PubKey: oraInfo.PubKey,
		Hash:   oraInfo.Hash,
	}
	psRawKey := pubInfoKey.ToRawKey()

	ps, _ := dpki.PovGetPublishState(csdb, psRawKey)
	if ps == nil {
		ps = types.NewPovPublishState()

		pubInfo := abi.GetPublishInfoByKey(ctx, oraInfo.OType, oraInfo.OID, oraInfo.PubKey, oraInfo.Hash)
		if pubInfo == nil {
			return errors.New("publish info not exist")
		}

		ps.BonusFee = types.NewBigNumFromBigInt(pubInfo.Fee)
	}

	if ps.VerifiedStatus == types.PovPublishStatusVerified {
		if povHeight > ps.VerifiedHeight+common.OracleExpirePovHeight {
			return nil
		}
		if len(ps.OracleAccounts) >= common.OracleVerifyMaxAccount {
			return nil
		}
	}

	for _, oa := range ps.OracleAccounts {
		if oa == block.Address {
			return nil
		}
	}
	ps.OracleAccounts = append(ps.OracleAccounts, block.Address)

	vsChangeAddrs := make([]types.Address, 0)

	if ps.VerifiedStatus == types.PovPublishStatusInit {
		if len(ps.OracleAccounts) >= common.OracleVerifyMinAccount {
			ps.VerifiedStatus = types.PovPublishStatusVerified
			ps.VerifiedHeight = povHeight

			vsChangeAddrs = ps.OracleAccounts
		}
	} else {
		vsChangeAddrs = append(vsChangeAddrs, block.Address)
	}

	err = dpki.PovSetPublishState(csdb, psRawKey, ps)
	if err != nil {
		return err
	}

	// update verifier state
	divBonusFee := new(types.BigNum).Div(ps.BonusFee, types.NewBigNumFromInt(5))
	for _, vsAddr := range vsChangeAddrs {
		var vsRawKey []byte
		vsRawKey = append(vsRawKey, vsAddr.Bytes()...)
		vsVal, _ := dpki.PovGetVerifierState(csdb, vsRawKey)
		if vsVal == nil {
			vsVal = types.NewPovVerifierState()
		}

		vsVal.TotalVerify += 1
		vsVal.TotalReward = new(types.BigNum).Add(vsVal.TotalReward, divBonusFee)

		err = dpki.PovSetVerifierState(csdb, vsRawKey, vsVal)
		if err != nil {
			return err
		}
	}

	return nil
}

type PKDReward struct {
	BaseContract
}

func (r *PKDReward) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(dpki.PKDRewardParam)
	err := abi.PublicKeyDistributionABI.UnpackMethod(param, abi.MethodNamePKDReward, block.Data)
	if err != nil {
		return nil, nil, err
	}

	if param.Account != block.Address {
		return nil, nil, errors.New("account is not verifier")
	}

	if block.Token != common.GasToken() {
		return nil, nil, errors.New("token is not gas token")
	}

	// check account exist
	am, _ := ctx.GetAccountMeta(param.Account)
	if am == nil {
		return nil, nil, errors.New("verifier account not exist")
	}

	nodeRewardHeight, err := abi.GetNodeRewardHeight(ctx)
	if err != nil {
		return nil, nil, err
	}

	if param.EndHeight > nodeRewardHeight {
		return nil, nil, fmt.Errorf("end height %d greater than node height %d", param.EndHeight, nodeRewardHeight)
	}

	oldInfo, err := r.GetRewardInfo(ctx, param.Account)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, nil, err
	}
	if oldInfo == nil {
		oldInfo = new(dpki.PKDRewardInfo)
		oldInfo.RewardAmount = big.NewInt(0)
	}

	if param.EndHeight <= oldInfo.EndHeight {
		return nil, nil, fmt.Errorf("param height %d lesser than last height %d", param.EndHeight, oldInfo.EndHeight)
	}

	lastVs, err := r.GetVerifierState(ctx, param.EndHeight, param.Account)
	if err != nil {
		return nil, nil, err
	}

	curVs, err := r.GetVerifierState(ctx, param.EndHeight, param.Account)
	if err != nil {
		return nil, nil, err
	}

	calcRewardAmount := types.NewBigNumFromInt(0).Sub(curVs.TotalReward, lastVs.TotalReward)
	if calcRewardAmount.CmpBigInt(param.RewardAmount) != 0 {
		return nil, nil, fmt.Errorf("calc reward %s not equal param reward %s", calcRewardAmount, param.RewardAmount)
	}

	newInfo := new(dpki.PKDRewardInfo)
	newInfo.Beneficial = param.Beneficial
	newInfo.EndHeight = param.EndHeight
	newInfo.RewardAmount = new(big.Int).Add(param.RewardAmount, oldInfo.RewardAmount)
	newInfo.Timestamp = block.Timestamp

	err = r.SetRewardInfo(ctx, param.Account, newInfo)
	if err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: param.Beneficial,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: types.Balance{Int: param.RewardAmount},
			Type:   common.GasToken(),
		}, nil
}

func (r *PKDReward) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(dpki.PKDRewardParam)

	err := abi.PublicKeyDistributionABI.UnpackMethod(param, abi.MethodNamePKDReward, input.Data)
	if err != nil {
		return nil, err
	}

	if param.Account != input.Address {
		return nil, errors.New("input account is not verifier")
	}

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = common.GasToken()
	block.Link = input.GetHash()
	block.PoVHeight = input.PoVHeight
	block.Timestamp = common.TimeNow().Unix()

	// pledge fields only for QLC token
	block.Vote = types.NewBalance(0)
	block.Oracle = types.NewBalance(0)
	block.Storage = types.NewBalance(0)
	block.Network = types.NewBalance(0)

	amBnf, _ := ctx.GetAccountMeta(param.Beneficial)
	if amBnf != nil {
		tmBnf := amBnf.Token(common.GasToken())
		if tmBnf != nil {
			block.Balance = tmBnf.Balance.Add(types.Balance{Int: param.RewardAmount})
			block.Representative = tmBnf.Representative
			block.Previous = tmBnf.Header
		} else {
			block.Balance = types.Balance{Int: param.RewardAmount}
			if len(amBnf.Tokens) > 0 {
				block.Representative = amBnf.Tokens[0].Representative
			} else {
				block.Representative = input.Representative
			}
			block.Previous = types.ZeroHash
		}
	} else {
		block.Balance = types.Balance{Int: param.RewardAmount}
		block.Representative = input.Representative
		block.Previous = types.ZeroHash
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: param.Beneficial,
			BlockType: types.ContractReward,
			Amount:    types.Balance{Int: param.RewardAmount},
			Token:     common.GasToken(),
			Data:      []byte{},
		},
	}, nil
}

func (r *PKDReward) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	param := new(dpki.PKDRewardParam)
	err := abi.PublicKeyDistributionABI.UnpackMethod(param, abi.MethodNamePKDReward, block.Data)
	if err != nil {
		return common.ContractNoGap, nil, err
	}

	needHeight := param.EndHeight + common.PovMinerRewardHeightGapToLatest

	latestBlock, err := ctx.GetLatestPovBlock()
	if err != nil || latestBlock == nil {
		return common.ContractRewardGapPov, needHeight, nil
	}

	nodeHeight := latestBlock.GetHeight()
	if nodeHeight < needHeight {
		return common.ContractRewardGapPov, needHeight, nil
	}

	return common.ContractNoGap, nil, err
}

func (r *PKDReward) SetRewardInfo(ctx *vmstore.VMContext, address types.Address, rwdInfo *dpki.PKDRewardInfo) error {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDRewardInfo,
		rwdInfo.Beneficial, rwdInfo.EndHeight, rwdInfo.RewardAmount, rwdInfo.Timestamp)
	if err != nil {
		return err
	}

	var rwdInfoKey []byte
	rwdInfoKey = append(rwdInfoKey, abi.PKDStorageTypeReward)
	rwdInfoKey = append(rwdInfoKey, address.Bytes()...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress.Bytes(), rwdInfoKey, data)
	if err != nil {
		return errors.New("save contract data err")
	}

	return nil
}

func (r *PKDReward) GetRewardInfo(ctx *vmstore.VMContext, address types.Address) (*dpki.PKDRewardInfo, error) {
	var rwdInfoKey []byte
	rwdInfoKey = append(rwdInfoKey, abi.PKDStorageTypeReward)
	rwdInfoKey = append(rwdInfoKey, address.Bytes()...)

	valBytes, err := ctx.GetStorage(types.PubKeyDistributionAddress[:], rwdInfoKey)
	if err != nil {
		return nil, err
	}

	rwdInfo := new(dpki.PKDRewardInfo)
	rwdInfo.RewardAmount = big.NewInt(0)

	err = abi.PublicKeyDistributionABI.UnpackVariable(&rwdInfo, abi.VariableNamePKDRewardInfo, valBytes)
	if err != nil {
		return nil, err
	}

	return rwdInfo, nil
}

func (r *PKDReward) GetVerifierState(ctx *vmstore.VMContext, povHeight uint64, address types.Address) (*types.PovVerifierState, error) {
	povHdr, err := ctx.GetPovHeaderByHeight(povHeight)
	if err != nil {
		return nil, err
	}

	gsdb := statedb.NewPovGlobalStateDB(ctx.Ledger.DBStore(), povHdr.GetStateHash())
	csdb, err := gsdb.LookupContractStateDB(types.PubKeyDistributionAddress)
	if err != nil {
		return nil, err
	}

	vsRawKey := address.Bytes()
	return dpki.PovGetVerifierState(csdb, vsRawKey)
}

package contract

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	PovStatePrefixPKDPS = byte(1) // publish state
	PovStatePrefixPKDVS = byte(2) // verifier state
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

	ps, _ := o.GetPublishState(csdb, psRawKey)
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

	err = o.SetPublishState(csdb, psRawKey, ps)
	if err != nil {
		return err
	}

	// update verifier state
	divBonusFee := new(types.BigNum).Div(ps.BonusFee, types.NewBigNumFromInt(5))
	for _, vsAddr := range vsChangeAddrs {
		var vsRawKey []byte
		vsRawKey = append(vsRawKey, vsAddr.Bytes()...)
		vsVal, _ := o.GetVerifierState(csdb, vsRawKey)
		if vsVal == nil {
			vsVal = types.NewPovVerifierState()
		}

		vsVal.TotalVerify += 1
		vsVal.TotalReward = new(types.BigNum).Add(vsVal.TotalReward, divBonusFee)

		err = o.SetVerifierState(csdb, vsRawKey, vsVal)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Oracle) SetPublishState(csdb *statedb.PovContractStateDB, rawKey []byte, ps *types.PovPublishState) error {
	trieKey := types.PovCreateStateKey(PovStatePrefixPKDPS, rawKey)

	val, err := ps.Serialize()
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, val)
}

func (o *Oracle) GetPublishState(csdb *statedb.PovContractStateDB, rawKey []byte) (*types.PovPublishState, error) {
	trieKey := types.PovCreateStateKey(PovStatePrefixPKDPS, rawKey)

	valBytes, err := csdb.GetValue(trieKey)
	if err != nil {
		return nil, err
	}
	if len(valBytes) == 0 {
		return nil, errors.New("trie get value return empty")
	}

	ps := types.NewPovPublishState()
	err = ps.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize publish state err %s", err)
	}

	return ps, nil
}

func (o *Oracle) SetVerifierState(csdb *statedb.PovContractStateDB, rawKey []byte, ps *types.PovVerifierState) error {
	trieKey := types.PovCreateStateKey(PovStatePrefixPKDPS, rawKey)

	val, err := ps.Serialize()
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, val)
}

func (o *Oracle) GetVerifierState(csdb *statedb.PovContractStateDB, rawKey []byte) (*types.PovVerifierState, error) {
	trieKey := types.PovCreateStateKey(PovStatePrefixPKDVS, rawKey)

	valBytes, err := csdb.GetValue(trieKey)
	if err != nil {
		return nil, err
	}
	if len(valBytes) == 0 {
		return nil, errors.New("trie get value return empty")
	}

	vs := types.NewPovVerifierState()
	err = vs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize verifier state err %s", err)
	}

	return vs, nil
}

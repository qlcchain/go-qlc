package contract

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type VerifierRegister struct {
	WithSignNoPending
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

func (vr *VerifierRegister) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}

func (vr *VerifierRegister) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

type VerifierUnregister struct {
	WithSignNoPending
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
		return nil, nil, fmt.Errorf("there is no valid verifier to unregister(%s-%s)", block.Address, types.OracleTypeToString(reg.VType))
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

func (vu *VerifierUnregister) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}

func (vu *VerifierUnregister) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

type Publish struct {
	WithSignNoPending
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

	err = p.SetStorage(ctx, block.Address, info.PType, info.PID, info.PubKey, info.Verifiers, info.Codes, fee)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (p *Publish) SetStorage(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash, pk []byte,
	vs []types.Address, cs []types.Hash, fee types.Balance) error {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDPublishInfo, vs, cs, fee.Int, true)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PKDStorageTypePublisher)
	key = append(key, util.BE_Uint32ToBytes(pt)...)
	key = append(key, id[:]...)
	key = append(key, pk...)
	key = append(key, account[:]...)
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

func (p *Publish) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}

func (p *Publish) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

type UnPublish struct {
	WithSignNoPending
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

	err = abi.UnPublishInfoCheck(ctx, block.Address, info.PType, info.PID)
	if err != nil {
		return nil, nil, err
	}

	block.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDUnPublish, info.PType, info.PID)
	if err != nil {
		return nil, nil, err
	}

	err = up.SetStorage(ctx, block.Address, info.PType, info.PID)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (up *UnPublish) SetStorage(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash) error {
	pubInfo := abi.GetPublishInfo(ctx, account, pt, id)
	if len(pubInfo) == 0 {
		return fmt.Errorf("get publish info err(%s-%s-%s)", account, types.OracleTypeToString(pt), id)
	}

	info := pubInfo[0]
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDPublishInfo, info.Verifiers, info.Codes, info.Fee, false)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PKDStorageTypePublisher)
	key = append(key, util.BE_Uint32ToBytes(pt)...)
	key = append(key, id[:]...)
	key = append(key, info.PubKey...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
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

func (up *UnPublish) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}

func (up *UnPublish) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

type Oracle struct {
	WithSignNoPending
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
	if !block.IsSync() {
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

	if amount.Compare(types.OracleCost) != types.BalanceCompEqual {
		return nil, nil, fmt.Errorf("balance(exp:%s-%s) wrong", types.OracleCost, amount)
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
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDOracleInfo, code, hash)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PKDStorageTypeOracle)
	key = append(key, util.BE_Uint32ToBytes(ot)...)
	key = append(key, id[:]...)
	key = append(key, pk...)
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

func (o *Oracle) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}

func (o *Oracle) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

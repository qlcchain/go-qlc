package contract

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type Publish struct {
	WithSignNoPending
}

func (p *Publish) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	info := new(cabi.PublishInfo)
	err := cabi.PublisherABI.UnpackMethod(info, cabi.MethodNamePublish, block.GetData())
	if err != nil {
		return nil, nil, err
	}

	fee := types.Balance{Int: info.Fee}
	err = cabi.PublishInfoCheck(ctx, info.Account, info.PType, info.PID, info.PubKey, fee)
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

	block.Data, err = cabi.PublisherABI.PackMethod(cabi.MethodNamePublish, info.Account, info.PType, info.PID,
		info.PubKey, info.Verifiers, info.Codes, info.Fee)
	if err != nil {
		return nil, nil, err
	}

	err = p.SetStorage(ctx, info.Account, info.PType, info.PID, info.PubKey, info.Verifiers, info.Codes, fee)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (p *Publish) SetStorage(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash, pk []byte,
	vs []types.Address, cs []types.Hash, fee types.Balance) error {
	data, err := cabi.PublisherABI.PackVariable(cabi.VariableNamePublishInfo, vs, cs, fee.Int)
	if err != nil {
		return err
	}

	key := append(util.BE_Uint32ToBytes(pt), id[:]...)
	key = append(key, pk...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.PublisherAddress[:], key, data)
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
	info := new(cabi.UnPublishInfo)
	err := cabi.PublisherABI.UnpackMethod(info, cabi.MethodNameUnPublish, block.GetData())
	if err != nil {
		return nil, nil, err
	}

	err = cabi.UnPublishInfoCheck(ctx, info.Account, info.PType, info.PID)
	if err != nil {
		return nil, nil, err
	}

	block.Data, err = cabi.PublisherABI.PackMethod(cabi.MethodNameUnPublish, info.Account, info.PType, info.PID)
	if err != nil {
		return nil, nil, err
	}

	err = up.SetStorage(ctx, info.Account, info.PType, info.PID)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (up *UnPublish) SetStorage(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash) error {
	pk := cabi.GetPublishKeyByAccountAndTypeAndID(ctx, account, pt, id)
	if pk == nil {
		return fmt.Errorf("get publish key err")
	}

	key := append(util.BE_Uint32ToBytes(pt), id[:]...)
	key = append(key, pk...)
	key = append(key, account[:]...)
	ctx.DelStorage(types.PublisherAddress[:], key)
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

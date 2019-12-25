package contract

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type Oracle struct {
	WithSignNoPending
}

func (o *Oracle) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	info := new(cabi.OracleInfo)
	err := cabi.OracleABI.UnpackMethod(info, cabi.MethodNameOracle, block.GetData())
	if err != nil {
		return nil, nil, err
	}

	err = cabi.OracleInfoCheck(ctx, info.Account, info.OType, info.OID, info.PubKey, info.Code, info.Hash)
	if err != nil {
		return nil, nil, err
	}

	amount, err := ctx.CalculateAmount(block)
	if err != nil {
		return nil, nil, err
	}

	fee := types.Balance{Int: types.OracleCost}
	if amount.Compare(fee) != types.BalanceCompEqual {
		return nil, nil, fmt.Errorf("balance(exp:%s-%s) wrong", fee, amount)
	}

	block.Data, err = cabi.OracleABI.PackMethod(cabi.MethodNameOracle, info.Account, info.OType, info.OID, info.PubKey, info.Code, info.Hash)
	if err != nil {
		return nil, nil, err
	}

	err = o.SetStorage(ctx, info.Account, info.OType, info.OID, info.PubKey, info.Code, info.Hash)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}

func (o *Oracle) SetStorage(ctx *vmstore.VMContext, account types.Address, ot uint32, id types.Hash, pk []byte, code string, hash types.Hash) error {
	data, err := cabi.OracleABI.PackVariable(cabi.VariableNameOracleInfo, code, hash)
	if err != nil {
		return err
	}

	key := append(util.BE_Uint32ToBytes(ot), id[:]...)
	key = append(key, pk...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.OracleAddress[:], key, data)
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

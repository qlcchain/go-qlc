package contract

import (
	_ "errors"
	"fmt"
	_ "net"
	_ "strconv"
	_ "strings"

	_ "github.com/qlcchain/go-qlc/common/topic"

	"github.com/qlcchain/go-qlc/common"
	_ "github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type PtmKeyUpdate struct {
	BaseContract
}

func (pku *PtmKeyUpdate) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	pki := new(abi.PtmKeyInfo)
	err := abi.PtmKeyABI.UnpackMethod(pki, abi.MethodNamePtmKeyUpdate, block.GetData())
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}
	err = abi.PtmKeyInfoCheck(ctx, pki.Btype, []byte(pki.Pubkey))
	if err != nil {
		return nil, nil, ErrCheckParam
	}
	//fmt.Printf("ProcessSend:get pki Btype(%d) Pubkey(%s)\n", pki.Btype, pki.Pubkey)
	err = pku.SetStorage(ctx, block.Address, pki.Btype, []byte(pki.Pubkey))
	if err != nil {
		return nil, nil, ErrSetStorage
	}

	return nil, nil, nil
}

func (pku *PtmKeyUpdate) SetStorage(ctx *vmstore.VMContext, account types.Address, vBtype uint16, vKey []byte) error {
	data, err := abi.PtmKeyABI.PackVariable(abi.VariableNamePtmKeyStorageVar, string(vKey[:]), true)
	if err != nil {
		//fmt.Printf("SetStorage:PackVariable err(%s)\n", err)
		return err
	}
	if vBtype != common.PtmKeyVBtypeDefault {
		return ErrCheckParam
	}
	if len(vKey) < 44 {
		return ErrCheckParam
	}
	var key []byte
	key = append(key, account[:]...)
	key = append(key, util.BE_Uint16ToBytes(vBtype)...)
	//fmt.Printf("SetStorage:get key(%s) data(%s)\n", string(key[:]), data)
	err = ctx.SetStorage(contractaddress.PtmKeyKVAddress[:], key, data)
	if err != nil {
		//fmt.Printf("SetStorage:err(%s)\n", err)
		return err
	}

	return nil
}

type PtmKeyDelete struct {
	BaseContract
}

func (pkd *PtmKeyDelete) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}
	pki := new(abi.PtmKeyDeleteInfo)
	err := abi.PtmKeyABI.UnpackMethod(pki, abi.MethodNamePtmKeyDelete, block.GetData())
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	pks, _ := abi.GetPtmKeyByAccountAndBtype(ctx, block.Address, pki.Btype)
	if pks != nil {
		for _, pk := range pks {
			err = pkd.SetStorage(ctx, block.Address, pk.Btype, []byte(pk.Pubkey))
			if err != nil {
				return nil, nil, ErrSetStorage
			}
		}
	} else {
		return nil, nil, fmt.Errorf("PtmKeyDelete:get nil")
	}

	return nil, nil, nil
}

func (pkd *PtmKeyDelete) SetStorage(ctx *vmstore.VMContext, account types.Address, vBtype uint16, vKey []byte) error {
	data, err := abi.PtmKeyABI.PackVariable(abi.VariableNamePtmKeyStorageVar, string(vKey), false)
	if err != nil {
		return err
	}
	if vBtype != common.PtmKeyVBtypeDefault {
		return ErrCheckParam
	}
	if len(vKey) < 44 {
		return ErrCheckParam
	}
	var key []byte
	key = append(key, account[:]...)
	key = append(key, util.BE_Uint16ToBytes(vBtype)...)
	err = ctx.SetStorage(contractaddress.PtmKeyKVAddress[:], key, data)
	if err != nil {
		return err
	}

	return nil
}

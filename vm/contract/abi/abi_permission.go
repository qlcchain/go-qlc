package abi

import (
	"errors"
	"strings"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonPermission = `
	[
		{"type":"function","name":"PermissionAdminUpdate","inputs":[
			{"name":"addr","type":"address"},
			{"name":"comment","type":"string"}
		]},
		{"type":"function","name":"PermissionNodeAdd","inputs":[
			{"name":"kind","type":"uint8"},
			{"name":"node","type":"string"},
			{"name":"comment","type":"string"}
		]},
		{"type":"function","name":"PermissionNodeUpdate","inputs":[
			{"name":"index","type":"uint32"},
			{"name":"kind","type":"uint8"},
			{"name":"node","type":"string"},
			{"name":"comment","type":"string"}
		]},
		{"type":"function","name":"PermissionNodeRemove","inputs":[
			{"name":"index","type":"uint32"}
		]}
	]`

	MethodNamePermissionAdminUpdate = "PermissionAdminUpdate"
	MethodNamePermissionNodeAdd     = "PermissionNodeAdd"
	MethodNamePermissionNodeUpdate  = "PermissionNodeUpdate"
	MethodNamePermissionNodeRemove  = "PermissionNodeRemove"
)

var (
	PermissionABI, _ = abi.JSONToABIContract(strings.NewReader(JsonPermission))
)

func PermissionInit(ctx *vmstore.VMContext) error {
	_, err := GetPermissionAdmin(ctx)
	if err != nil {
		if err == vmstore.ErrStorageNotFound {
			adm := &AdminAccount{
				Addr:    cfg.GenesisAddress(),
				Comment: "Initial Admin",
				Status:  PermissionAdminStatusActive,
			}

			data, err := adm.MarshalMsg(nil)
			if err != nil {
				return errors.New("permission init err")
			}

			var key []byte
			key = append(key, PermissionDataAdmin)
			err = ctx.SetStorage(contractaddress.PermissionAddress[:], key, data)
			if err != nil {
				return errors.New("permission init err")
			}

			err = ctx.SaveStorage()
			if err != nil {
				return errors.New("permission init err")
			}
		} else {
			return err
		}
	}

	return nil
}

func GetPermissionAdmin(ctx *vmstore.VMContext) (*AdminAccount, error) {
	var key []byte
	key = append(key, PermissionDataAdmin)
	data, err := ctx.GetStorage(contractaddress.PermissionAddress[:], key)
	if err != nil {
		return nil, err
	}

	admin := new(AdminAccount)
	_, err = admin.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func GetPermissionNodeIndex(ctx *vmstore.VMContext) uint32 {
	var key []byte
	key = append(key, PermissionDataNodeIndex)
	data, err := ctx.GetStorage(contractaddress.PermissionAddress[:], key)
	if err != nil {
		return 0
	}

	return util.BE_BytesToUint32(data)
}

func GetPermissionNode(ctx *vmstore.VMContext, index uint32) (*PermNode, error) {
	var key []byte
	key = append(key, PermissionDataNode)
	key = append(key, util.BE_Uint32ToBytes(index)...)
	data, err := ctx.GetStorage(contractaddress.PermissionAddress[:], key)
	if err != nil {
		return nil, err
	}

	pn := new(PermNode)
	_, err = pn.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	pn.Index = index
	return pn, nil
}

func GetAllPermissionNodes(ctx *vmstore.VMContext) ([]*PermNode, error) {
	nodes := make([]*PermNode, 0)

	itKey := append(contractaddress.PermissionAddress[:], PermissionDataNode)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		// prefix(1) + addr(32) + type(1) + index(4)
		index := util.BE_BytesToUint32(key[34:])

		pn := new(PermNode)
		_, err := pn.UnmarshalMsg(value)
		if err != nil {
			return err
		}

		if pn.Valid {
			pn.Index = index
			nodes = append(nodes, pn)
		}

		return nil
	})

	return nodes, err
}

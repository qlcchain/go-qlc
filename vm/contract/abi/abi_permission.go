package abi

import (
	"errors"
	"strings"

	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"

	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"

	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonPermission = `[
		{"type":"function","name":"PermissionAdminHandOver","inputs":[
			{"name":"account","type":"address"},
			{"name":"comment","type":"string"}
		]},
		{"type":"function","name":"PermissionNodeUpdate","inputs":[
			{"name":"nodeId","type":"string"},
			{"name":"nodeUrl","type":"string"},
			{"name":"comment","type":"string"}
		]}
	]`

	MethodNamePermissionAdminHandOver = "PermissionAdminHandOver"
	MethodNamePermissionNodeUpdate    = "PermissionNodeUpdate"
)

var (
	PermissionABI, _ = abi.JSONToABIContract(strings.NewReader(JsonPermission))
)

func PermissionIsAdmin(ctx *vmstore.VMContext, addr types.Address) bool {
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return false
	}
	// if there is no admin, genesis is admin
	if addr == cfg.GenesisAddress() {
		itor := csdb.NewCurTireIterator(statedb.PovCreateContractLocalStateKey(PermissionDataAdmin, nil))

		for _, val, ok := itor.Next(); ok; _, val, ok = itor.Next() {
			admin := new(AdminAccount)
			_, err := admin.UnmarshalMsg(val)
			if err != nil {
				continue
			}

			if admin.Valid {
				return false
			}
		}

		return true
	} else {
		trieKey := statedb.PovCreateContractLocalStateKey(PermissionDataAdmin, addr.Bytes())

		valBytes, err := csdb.GetValue(trieKey)
		if err != nil || len(valBytes) == 0 {
			return false
		}

		admin := new(AdminAccount)
		_, err = admin.UnmarshalMsg(valBytes)
		if err != nil {
			return false
		}

		return admin.Valid
	}
}

func PermissionGetAdmin(ctx *vmstore.VMContext) ([]*AdminAccount, error) {
	admins := make([]*AdminAccount, 0)
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return nil, err
	}
	itor := csdb.NewCurTireIterator(statedb.PovCreateContractLocalStateKey(PermissionDataAdmin, nil))

	for key, val, ok := itor.Next(); ok; key, val, ok = itor.Next() {
		admin := new(AdminAccount)
		_, err := admin.UnmarshalMsg(val)
		if err != nil {
			return nil, err
		}

		addr, err := types.BytesToAddress(key[2:])
		if err != nil {
			return nil, err
		}

		if admin.Valid {
			admin.Account = addr
			admins = append(admins, admin)
		}
	}

	if len(admins) == 0 {
		genesisAdmin := &AdminAccount{
			Account: cfg.GenesisAddress(),
			Comment: "Initial admin",
		}
		admins = append(admins, genesisAdmin)
	}

	return admins, nil
}

func PermissionUpdateNode(csdb *statedb.PovContractStateDB, id, url, comment string) error {
	trieKey := statedb.PovCreateContractLocalStateKey(PermissionDataNode, []byte(id))

	node := new(PermNode)
	node.NodeUrl = url
	node.Comment = comment
	node.Valid = true
	data, err := node.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, data)
}

func PermissionGetNode(ctx *vmstore.VMContext, id string) (*PermNode, error) {
	trieKey := statedb.PovCreateContractLocalStateKey(PermissionDataNode, []byte(id))
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return nil, err
	}
	valBytes, err := csdb.GetValue(trieKey)
	if err != nil || len(valBytes) == 0 {
		return nil, errors.New("get node err")
	}

	pn := new(PermNode)
	_, err = pn.UnmarshalMsg(valBytes)
	if err != nil {
		return nil, err
	}

	if pn.Valid {
		pn.NodeId = id
		return pn, nil
	} else {
		return nil, errors.New("node is removed")
	}
}

func PermissionGetAllNodes(store ledger.Store) ([]*PermNode, error) {
	nodes := make([]*PermNode, 0)
	ctx := vmstore.NewVMContext(store, &contractaddress.PermissionAddress)
	csdb, err := ctx.PoVContractState()
	if err != nil {
		return nil, err
	}
	itor := csdb.NewCurTireIterator(statedb.PovCreateContractLocalStateKey(PermissionDataNode, nil))

	for key, val, ok := itor.Next(); ok; key, val, ok = itor.Next() {
		pn := new(PermNode)
		_, err := pn.UnmarshalMsg(val)
		if err != nil {
			return nil, err
		}

		if pn.Valid {
			pn.NodeId = string(key[2:])
			nodes = append(nodes, pn)
		}
	}

	return nodes, nil
}

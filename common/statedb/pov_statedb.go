package statedb

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/trie"
)

type PovGlobalStateDB struct {
	treeRoot *trie.Trie

	asCache map[types.Address]*types.PovAccountState
	asDirty map[types.Address]struct{}

	rsCache map[types.Address]*types.PovRepState
	rsDirty map[types.Address]struct{}

	csCache map[types.Address]*types.PovContractState
	csDirty map[types.Address]struct{}
}

type PovContractStateDB struct {
	treeRoot *trie.Trie
	kvCache  map[string][]byte
	kvDirty  map[string]struct{}
}

func NewPovGlobalStateDB(t *trie.Trie) *PovGlobalStateDB {
	return &PovGlobalStateDB{
		treeRoot: t,

		asCache: make(map[types.Address]*types.PovAccountState),
		asDirty: make(map[types.Address]struct{}),

		rsCache: make(map[types.Address]*types.PovRepState),
		rsDirty: make(map[types.Address]struct{}),

		csCache: make(map[types.Address]*types.PovContractState),
		csDirty: make(map[types.Address]struct{}),
	}
}

func (gsdb *PovGlobalStateDB) SetAccountState(address types.Address, as *types.PovAccountState) error {
	as.Account = address
	gsdb.asCache[address] = as
	gsdb.asDirty[address] = struct{}{}
	return nil
}

func (gsdb *PovGlobalStateDB) GetAccountState(address types.Address) (*types.PovAccountState, error) {
	if as := gsdb.asCache[address]; as != nil {
		return as, nil
	}

	keyBytes := types.PovCreateAccountStateKey(address)
	valBytes := gsdb.treeRoot.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("trie get value return empty")
	}

	as := types.NewPovAccountState()
	err := as.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize account state err %s", err)
	}
	gsdb.asCache[address] = as

	return as, nil
}

func (gsdb *PovGlobalStateDB) SetRepState(address types.Address, rs *types.PovRepState) error {
	rs.Account = address
	gsdb.rsCache[address] = rs
	gsdb.rsDirty[address] = struct{}{}
	return nil
}

func (gsdb *PovGlobalStateDB) GetRepState(address types.Address) (*types.PovRepState, error) {
	if rs := gsdb.rsCache[address]; rs != nil {
		return rs, nil
	}

	keyBytes := types.PovCreateRepStateKey(address)
	valBytes := gsdb.treeRoot.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("trie get value return empty")
	}

	rs := types.NewPovRepState()
	err := rs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize rep state err %s", err)
	}
	gsdb.rsCache[address] = rs

	return rs, nil
}

func (gsdb *PovGlobalStateDB) SetContractState(address types.Address, cs *types.PovContractState) error {
	gsdb.csCache[address] = cs
	gsdb.csDirty[address] = struct{}{}
	return nil
}

func (gsdb *PovGlobalStateDB) GetContractState(address types.Address) (*types.PovContractState, error) {
	if cs := gsdb.csCache[address]; cs != nil {
		return cs, nil
	}

	keyBytes := types.PovCreateStateKey(types.PovStatePrefixCCS, address.Bytes())
	valBytes := gsdb.treeRoot.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("trie get value return empty")
	}

	cs := types.NewPovContractState()
	err := cs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize contract state err %s", err)
	}

	gsdb.csCache[address] = cs
	return cs, nil
}

func (gsdb *PovGlobalStateDB) CommitToTrie() error {
	for address := range gsdb.asDirty {
		as := gsdb.asCache[address]
		if as == nil {
			continue
		}

		valBytes, err := as.Serialize()
		if err != nil {
			return fmt.Errorf("serialize new account state err %s", err)
		}
		if len(valBytes) == 0 {
			return errors.New("serialize new account state got empty value")
		}

		keyBytes := types.PovCreateAccountStateKey(address)
		gsdb.treeRoot.SetValue(keyBytes, valBytes)
	}

	for address := range gsdb.rsDirty {
		rs := gsdb.rsCache[address]
		if rs == nil {
			continue
		}

		valBytes, err := rs.Serialize()
		if err != nil {
			return fmt.Errorf("serialize new rep state err %s", err)
		}
		if len(valBytes) == 0 {
			return errors.New("serialize new rep state got empty value")
		}

		keyBytes := types.PovCreateRepStateKey(address)
		gsdb.treeRoot.SetValue(keyBytes, valBytes)
	}

	for address := range gsdb.csDirty {
		cs := gsdb.csCache[address]
		if cs == nil {
			continue
		}

		valBytes, err := cs.Serialize()
		if err != nil {
			return fmt.Errorf("serialize new contract state err %s", err)
		}
		if len(valBytes) == 0 {
			return errors.New("serialize new contract state got empty value")
		}

		keyBytes := types.PovCreateStateKey(types.PovStatePrefixCCS, address.Bytes())
		gsdb.treeRoot.SetValue(keyBytes, valBytes)
	}

	return nil
}

func NewPovContractStateDB(t *trie.Trie) *PovContractStateDB {
	return &PovContractStateDB{
		treeRoot: t,
		kvCache:  make(map[string][]byte),
		kvDirty:  make(map[string]struct{}),
	}
}

func (csdb *PovContractStateDB) SetValue(key []byte, val []byte) error {
	csdb.kvCache[string(key)] = val
	csdb.kvDirty[string(key)] = struct{}{}
	return nil
}

func (csdb *PovContractStateDB) GetValue(key []byte) ([]byte, error) {
	if val, ok := csdb.kvCache[string(key)]; ok {
		return val, nil
	}

	val := csdb.treeRoot.GetValue(key)
	csdb.kvCache[string(key)] = val

	return val, nil
}

func (csdb *PovContractStateDB) CommitToTrie() error {
	for key := range csdb.kvDirty {
		val := csdb.kvCache[key]
		if val == nil {
			continue
		}

		csdb.treeRoot.SetValue([]byte(key), csdb.kvCache[key])
	}

	return nil
}

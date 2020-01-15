package statedb

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/trie"
)

type PovStateDB struct {
	treeRoot   *trie.Trie
	asCache    map[types.Address]*types.PovAccountState
	rsCache    map[types.Address]*types.PovRepState
	pkdPsCache map[string]*types.PovPublishState
	pkdVsCache map[string]*types.PovVerifierState
}

func NewPovStateDB(t *trie.Trie) *PovStateDB {
	return &PovStateDB{
		treeRoot: t,
		asCache:  make(map[types.Address]*types.PovAccountState),
		rsCache:  make(map[types.Address]*types.PovRepState),

		pkdPsCache: make(map[string]*types.PovPublishState),
	}
}

func (sdb *PovStateDB) SetAccountState(address types.Address, as *types.PovAccountState) error {
	as.Account = address
	sdb.asCache[address] = as
	return nil
}

func (sdb *PovStateDB) GetAccountState(address types.Address) (*types.PovAccountState, error) {
	if as := sdb.asCache[address]; as != nil {
		return as, nil
	}

	keyBytes := types.PovCreateAccountStateKey(address)
	valBytes := sdb.treeRoot.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("trie get value return empty")
	}

	as := types.NewPovAccountState()
	err := as.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize account state err %s", err)
	}
	sdb.asCache[address] = as

	return as, nil
}

func (sdb *PovStateDB) SetRepState(address types.Address, rs *types.PovRepState) error {
	rs.Account = address
	sdb.rsCache[address] = rs
	return nil
}

func (sdb *PovStateDB) GetRepState(address types.Address) (*types.PovRepState, error) {
	if rs := sdb.rsCache[address]; rs != nil {
		return rs, nil
	}

	keyBytes := types.PovCreateRepStateKey(address)
	valBytes := sdb.treeRoot.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("trie get value return empty")
	}

	rs := types.NewPovRepState()
	err := rs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize rep state err %s", err)
	}
	sdb.rsCache[address] = rs

	return rs, nil
}

func (sdb *PovStateDB) SetPublishState(key []byte, ps *types.PovPublishState) error {
	sdb.pkdPsCache[string(key)] = ps
	return nil
}

func (sdb *PovStateDB) GetPublishState(key []byte) (*types.PovPublishState, error) {
	keyStr := string(key)
	if ps := sdb.pkdPsCache[keyStr]; ps != nil {
		return ps, nil
	}

	keyBytes := types.PovCreateStateKey(types.PovStatePrefixPKDPS, key)
	valBytes := sdb.treeRoot.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("trie get value return empty")
	}

	ps := types.NewPovPublishState()
	err := ps.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize publish state err %s", err)
	}
	sdb.pkdPsCache[keyStr] = ps

	return ps, nil
}

func (sdb *PovStateDB) SetVerifierState(key []byte, ps *types.PovVerifierState) error {
	sdb.pkdVsCache[string(key)] = ps
	return nil
}

func (sdb *PovStateDB) GetVerifierState(key []byte) (*types.PovVerifierState, error) {
	keyStr := string(key)
	if ps := sdb.pkdVsCache[keyStr]; ps != nil {
		return ps, nil
	}

	keyBytes := types.PovCreateStateKey(types.PovStatePrefixPKDVS, key)
	valBytes := sdb.treeRoot.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("trie get value return empty")
	}

	vs := types.NewPovVerifierState()
	err := vs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize verifier state err %s", err)
	}
	sdb.pkdVsCache[keyStr] = vs

	return vs, nil
}

func (sdb *PovStateDB) CommitToTrie() error {
	for address, as := range sdb.asCache {
		valBytes, err := as.Serialize()
		if err != nil {
			return fmt.Errorf("serialize new account state err %s", err)
		}
		if len(valBytes) == 0 {
			return errors.New("serialize new account state got empty value")
		}

		keyBytes := types.PovCreateAccountStateKey(address)
		sdb.treeRoot.SetValue(keyBytes, valBytes)
	}

	for address, rs := range sdb.rsCache {
		valBytes, err := rs.Serialize()
		if err != nil {
			return fmt.Errorf("serialize new rep state err %s", err)
		}
		if len(valBytes) == 0 {
			return errors.New("serialize new rep state got empty value")
		}

		keyBytes := types.PovCreateRepStateKey(address)
		sdb.treeRoot.SetValue(keyBytes, valBytes)
	}

	for keyStr, ps := range sdb.pkdPsCache {
		valBytes, err := ps.Serialize()
		if err != nil {
			return fmt.Errorf("serialize new publish state err %s", err)
		}
		if len(valBytes) == 0 {
			return errors.New("serialize new publish state got empty value")
		}

		keyBytes := types.PovCreateStateKey(types.PovStatePrefixPKDPS, []byte(keyStr))
		sdb.treeRoot.SetValue(keyBytes, valBytes)
	}

	for keyStr, ps := range sdb.pkdVsCache {
		valBytes, err := ps.Serialize()
		if err != nil {
			return fmt.Errorf("serialize new verifier state err %s", err)
		}
		if len(valBytes) == 0 {
			return errors.New("serialize new verifier state got empty value")
		}

		keyBytes := types.PovCreateStateKey(types.PovStatePrefixPKDVS, []byte(keyStr))
		sdb.treeRoot.SetValue(keyBytes, valBytes)
	}

	return nil
}

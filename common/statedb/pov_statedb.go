package statedb

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
)

const (
	PovGlobalStatePrefixAcc = byte(1)
	PovGlobalStatePrefixRep = byte(2)
	PovGlobalStatePrefixCS  = byte(201) // Contract State

	PovStatusOffline = 0
	PovStatusOnline  = 1
)

// PovCreateGlobalStateKey used for global trie key only
// prefix MUST be UNIQUE in global namespace
func PovCreateGlobalStateKey(prefix byte, rawKey []byte) []byte {
	var key []byte
	key = append(key, storage.KeyPrefixTriePovState)
	key = append(key, prefix)
	if rawKey != nil {
		key = append(key, rawKey...)
	}
	return key
}

func PovCreateAccountStateKey(address types.Address) []byte {
	addrBytes := address.Bytes()
	return PovCreateGlobalStateKey(PovGlobalStatePrefixAcc, addrBytes)
}

func PovCreateRepStateKey(address types.Address) []byte {
	addrBytes := address.Bytes()
	return PovCreateGlobalStateKey(PovGlobalStatePrefixRep, addrBytes)
}

func PovCreateContractStateKey(address types.Address) []byte {
	addrBytes := address.Bytes()
	return PovCreateGlobalStateKey(PovGlobalStatePrefixCS, addrBytes)
}

func PovStateKeyToAddress(key []byte) (types.Address, error) {
	return types.BytesToAddress(key[2:])
}

// PovCreateContractLocalStateKey used for contract trie tree key only
// prefix MUST be UNIQUE in contract namespace, not in global namespace
func PovCreateContractLocalStateKey(prefix byte, rawKey []byte) []byte {
	var key []byte
	key = append(key, storage.KeyPrefixTriePovState)
	key = append(key, prefix)
	if rawKey != nil {
		key = append(key, rawKey...)
	}
	return key
}

type PovGlobalStateDB struct {
	db       storage.Store
	logger   *zap.SugaredLogger
	prevTrie *trie.Trie
	curTrie  *trie.Trie

	kvCache map[string][]byte
	kvDirty map[string]struct{}

	asCache map[types.Address]*types.PovAccountState
	asDirty map[types.Address]struct{}

	rsCache map[types.Address]*types.PovRepState
	rsDirty map[types.Address]struct{}

	allCSDBs map[types.Address]*PovContractStateDB
}

type PovContractStateDB struct {
	CS       *types.PovContractState
	dirty    bool // means fields in cs have been updated
	db       storage.Store
	prevTrie *trie.Trie
	curTrie  *trie.Trie
	kvCache  map[string][]byte
	kvDirty  map[string]struct{}
}

func NewPovGlobalStateDB(db storage.Store, prevStateHash types.Hash) *PovGlobalStateDB {
	gsdb := &PovGlobalStateDB{
		db:     db,
		logger: log.NewLogger("pov_statedb"),

		kvCache: make(map[string][]byte),
		kvDirty: make(map[string]struct{}),

		asCache: make(map[types.Address]*types.PovAccountState),
		asDirty: make(map[types.Address]struct{}),

		rsCache: make(map[types.Address]*types.PovRepState),
		rsDirty: make(map[types.Address]struct{}),

		allCSDBs: make(map[types.Address]*PovContractStateDB),
	}

	if prevStateHash.IsZero() {
		gsdb.curTrie = trie.NewTrie(db, nil, GetStateDBTriePool())
	} else {
		gsdb.prevTrie = trie.NewTrie(db, &prevStateHash, GetStateDBTriePool())
		if gsdb.prevTrie != nil {
			gsdb.curTrie = gsdb.prevTrie.Clone()
		} else {
			gsdb.curTrie = trie.NewTrie(db, nil, GetStateDBTriePool())
		}
	}

	return gsdb
}

func (gsdb *PovGlobalStateDB) GetPrevHash() types.Hash {
	if gsdb.prevTrie != nil {
		rootHashPtr := gsdb.prevTrie.Hash()
		if rootHashPtr != nil {
			return *rootHashPtr
		}
	}
	return types.ZeroHash
}

func (gsdb *PovGlobalStateDB) GetCurHash() types.Hash {
	if gsdb.curTrie != nil {
		rootHashPtr := gsdb.curTrie.Hash()
		if rootHashPtr != nil {
			return *rootHashPtr
		}
	}
	return types.ZeroHash
}

func (gsdb *PovGlobalStateDB) GetPrevTrie() *trie.Trie {
	return gsdb.prevTrie
}

func (gsdb *PovGlobalStateDB) GetCurTrie() *trie.Trie {
	return gsdb.curTrie
}

func (gsdb *PovGlobalStateDB) SetValue(key []byte, val []byte) error {
	gsdb.kvCache[string(key)] = val
	gsdb.kvDirty[string(key)] = struct{}{}
	return nil
}

func (gsdb *PovGlobalStateDB) GetValue(key []byte) ([]byte, error) {
	if val, ok := gsdb.kvCache[string(key)]; ok {
		return val, nil
	}

	val := gsdb.curTrie.GetValue(key)
	if len(val) == 0 {
		return nil, errors.New("key not exist in trie")
	}
	gsdb.kvCache[string(key)] = val

	return val, nil
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

	keyBytes := PovCreateAccountStateKey(address)
	valBytes := gsdb.curTrie.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("key not exist in trie")
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

	keyBytes := PovCreateRepStateKey(address)
	valBytes := gsdb.curTrie.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("key not exist in trie")
	}

	rs := types.NewPovRepState()
	err := rs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize rep state err %s", err)
	}
	gsdb.rsCache[address] = rs

	return rs, nil
}

func (gsdb *PovGlobalStateDB) GetContractState(address types.Address) (*types.PovContractState, error) {
	keyBytes := PovCreateContractStateKey(address)
	valBytes := gsdb.curTrie.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil, errors.New("key not exist in trie")
	}

	cs := types.NewPovContractState()
	err := cs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize contract state err %s", err)
	}

	return cs, nil
}

func (gsdb *PovGlobalStateDB) LookupContractStateDB(address types.Address) (*PovContractStateDB, error) {
	if cs := gsdb.allCSDBs[address]; cs != nil {
		return cs, nil
	}

	cs := types.NewPovContractState()

	keyBytes := PovCreateContractStateKey(address)
	valBytes := gsdb.curTrie.GetValue(keyBytes)
	if len(valBytes) > 0 {
		err := cs.Deserialize(valBytes)
		if err != nil {
			return nil, fmt.Errorf("deserialize contract state err %s", err)
		}
	}

	csdb := NewPovContractStateDB(gsdb.db, cs)
	gsdb.allCSDBs[address] = csdb
	return csdb, nil
}

func (gsdb *PovGlobalStateDB) SetContractValue(address types.Address, key []byte, val []byte) error {
	csdb, err := gsdb.LookupContractStateDB(address)
	if err != nil {
		return err
	}
	err = csdb.SetValue(key, val)
	return err
}

func (gsdb *PovGlobalStateDB) GetContractValue(address types.Address, key []byte) ([]byte, error) {
	csdb, err := gsdb.LookupContractStateDB(address)
	if err != nil {
		return nil, err
	}
	val, err := csdb.GetValue(key)
	return val, err
}

func (gsdb *PovGlobalStateDB) CommitToTrie() error {
	// global key value
	if len(gsdb.kvDirty) > 0 {
		for key := range gsdb.kvDirty {
			val := gsdb.kvCache[key]
			if val == nil {
				continue
			}

			gsdb.curTrie.SetValue([]byte(key), gsdb.kvCache[key])
		}
		gsdb.kvDirty = make(map[string]struct{})
	}

	// global account state
	if len(gsdb.asDirty) > 0 {
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

			keyBytes := PovCreateAccountStateKey(address)
			gsdb.curTrie.SetValue(keyBytes, valBytes)
		}
		gsdb.asDirty = make(map[types.Address]struct{})
	}

	// global rep state
	if len(gsdb.rsDirty) > 0 {
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

			keyBytes := PovCreateRepStateKey(address)
			gsdb.curTrie.SetValue(keyBytes, valBytes)
		}
		gsdb.rsDirty = make(map[types.Address]struct{})
	}

	// all contracts state
	for address, csdb := range gsdb.allCSDBs {
		dirty, err := csdb.CommitToTrie()
		if err != nil {
			return err
		}
		if !dirty {
			continue
		}

		valBytes, err := csdb.CS.Serialize()
		if err != nil {
			return fmt.Errorf("serialize new contract state err %s", err)
		}
		if len(valBytes) == 0 {
			return errors.New("serialize new contract state got empty value")
		}

		keyBytes := PovCreateContractStateKey(address)
		gsdb.curTrie.SetValue(keyBytes, valBytes)
	}

	return nil
}

func (gsdb *PovGlobalStateDB) CommitToDB(batch storage.Batch) error {
	// contract state
	for _, csdb := range gsdb.allCSDBs {
		dbErr := csdb.CommitToDB(batch)
		if dbErr != nil {
			return dbErr
		}
	}

	// global state
	saveCallback, dbErr := gsdb.curTrie.SaveInTxn(batch)
	if dbErr != nil {
		return dbErr
	}
	if saveCallback != nil {
		saveCallback()
	}

	return nil
}

func (gsdb *PovGlobalStateDB) NewCurTireIterator(prefix []byte) *trie.Iterator {
	return gsdb.curTrie.NewIterator(prefix)
}

func (gsdb *PovGlobalStateDB) NewPrevTireIterator(prefix []byte) *trie.Iterator {
	return gsdb.prevTrie.NewIterator(prefix)
}

func NewPovContractStateDB(db storage.Store, cs *types.PovContractState) *PovContractStateDB {
	var prevTrie *trie.Trie
	var currentTrie *trie.Trie

	prevStateHash := cs.StateHash
	if prevStateHash.IsZero() {
		currentTrie = trie.NewTrie(db, nil, GetStateDBTriePool())
	} else {
		prevTrie = trie.NewTrie(db, &prevStateHash, GetStateDBTriePool())
		if prevTrie != nil {
			currentTrie = prevTrie.Clone()
		} else {
			currentTrie = trie.NewTrie(db, nil, GetStateDBTriePool())
		}
	}

	return &PovContractStateDB{
		db:       db,
		CS:       cs,
		prevTrie: prevTrie,
		curTrie:  currentTrie,
		kvCache:  make(map[string][]byte),
		kvDirty:  make(map[string]struct{}),
	}
}

func (csdb *PovContractStateDB) GetPrevHash() types.Hash {
	if csdb.prevTrie != nil {
		rootHashPtr := csdb.prevTrie.Hash()
		if rootHashPtr != nil {
			return *rootHashPtr
		}
	}
	return types.ZeroHash
}

func (csdb *PovContractStateDB) GetCurHash() types.Hash {
	if csdb.curTrie != nil {
		rootHashPtr := csdb.curTrie.Hash()
		if rootHashPtr != nil {
			return *rootHashPtr
		}
	}
	return types.ZeroHash
}

func (csdb *PovContractStateDB) GetPrevTrie() *trie.Trie {
	return csdb.prevTrie
}

func (csdb *PovContractStateDB) GetCurTrie() *trie.Trie {
	return csdb.curTrie
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

	val := csdb.curTrie.GetValue(key)
	if len(val) == 0 {
		return nil, errors.New("key not exist in trie")
	}
	csdb.kvCache[string(key)] = val

	return val, nil
}

func (csdb *PovContractStateDB) CommitToTrie() (bool, error) {
	if len(csdb.kvDirty) > 0 {
		for key := range csdb.kvDirty {
			val := csdb.kvCache[key]
			if val == nil {
				continue
			}

			csdb.curTrie.SetValue([]byte(key), csdb.kvCache[key])
		}
		csdb.kvDirty = make(map[string]struct{})

		csdb.CS.StateHash = *csdb.curTrie.Hash()

		csdb.dirty = true
	}

	return csdb.dirty, nil
}

func (csdb *PovContractStateDB) CommitToDB(batch storage.Batch) error {
	saveCallback, dbErr := csdb.curTrie.SaveInTxn(batch)
	if dbErr != nil {
		return dbErr
	}
	if saveCallback != nil {
		saveCallback()
	}
	return nil
}

func (csdb *PovContractStateDB) NewCurTireIterator(prefix []byte) *trie.Iterator {
	return csdb.curTrie.NewIterator(prefix)
}

func (csdb *PovContractStateDB) NewPrevTireIterator(prefix []byte) *trie.Iterator {
	return csdb.prevTrie.NewIterator(prefix)
}

func GetStateDBTriePool() *trie.NodePool {
	return trie.GetGlobalTriePool()
}

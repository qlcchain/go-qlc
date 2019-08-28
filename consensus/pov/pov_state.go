package pov

import (
	"errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/trie"
)

func (bc *PovBlockChain) TrieDb() db.Store {
	return bc.getLedger().DBStore()
}

func (bc *PovBlockChain) GetStateTrie(stateHash *types.Hash) *trie.Trie {
	if stateHash != nil {
		v, _ := bc.trieCache.Get(*stateHash)
		if v != nil {
			return v.(*trie.Trie)
		}
	}
	t := trie.NewTrie(bc.TrieDb(), stateHash, bc.trieNodePool)
	if stateHash != nil {
		_ = bc.trieCache.Set(*stateHash, t)
	}
	return t
}

func (bc *PovBlockChain) NewStateTrie() *trie.Trie {
	return trie.NewTrie(bc.TrieDb(), nil, bc.trieNodePool)
}

func (bc *PovBlockChain) GenStateTrie(prevStateHash types.Hash, txs []*types.PovTransaction) (*trie.Trie, error) {
	var currentTrie *trie.Trie
	prevTrie := bc.GetStateTrie(&prevStateHash)
	if prevTrie != nil {
		currentTrie = prevTrie.Clone()
	} else {
		currentTrie = bc.NewStateTrie()
	}
	if currentTrie == nil {
		return nil, errors.New("failed to make current trie")
	}

	for _, tx := range txs {
		err := bc.ApplyTransaction(currentTrie, tx.Block)
		if err != nil {
			return nil, err
		}
	}

	return currentTrie, nil
}

func (bc *PovBlockChain) ApplyTransaction(trie *trie.Trie, stateBlock *types.StateBlock) error {
	oldAs := bc.GetAccountState(trie, stateBlock.Address)
	var newAs *types.PovAccountState
	if oldAs != nil {
		newAs = oldAs.Clone()
	} else {
		newAs = types.NewPovAccountState()
	}

	bc.updateAccountState(trie, stateBlock, oldAs, newAs)

	bc.updateRepState(trie, stateBlock, oldAs, newAs)

	return nil
}

func (bc *PovBlockChain) updateAccountState(trie *trie.Trie, block *types.StateBlock, oldAs *types.PovAccountState, newAs *types.PovAccountState) {
	hash := block.GetHash()
	rep := block.GetRepresentative()
	token := block.GetToken()
	balance := block.GetBalance()

	tsNew := &types.PovTokenState{
		Type:           token,
		Hash:           hash,
		Representative: rep,
		Balance:        balance,
	}

	if oldAs != nil {
		if block.GetToken() == common.ChainToken() {
			newAs.Balance = balance
			newAs.Oracle = block.GetOracle()
			newAs.Network = block.GetNetwork()
			newAs.Vote = block.GetVote()
			newAs.Storage = block.GetStorage()
		}

		tsNewExist := newAs.GetTokenState(block.GetToken())
		if tsNewExist != nil {
			tsNewExist.Representative = rep
			tsNewExist.Balance = balance
			tsNewExist.Hash = hash
		} else {
			newAs.TokenStates = append(newAs.TokenStates, tsNew)
		}
	} else {
		newAs.TokenStates = []*types.PovTokenState{tsNew}

		if block.GetToken() == common.ChainToken() {
			newAs.Balance = balance
			newAs.Oracle = block.GetOracle()
			newAs.Network = block.GetNetwork()
			newAs.Vote = block.GetVote()
			newAs.Storage = block.GetStorage()
		}
	}

	bc.SetAccountState(trie, block.Address, newAs)
}

func (bc *PovBlockChain) updateRepState(trie *trie.Trie, block *types.StateBlock, oldBlkAs *types.PovAccountState, newBlkAs *types.PovAccountState) {
	if block.GetToken() != common.ChainToken() {
		return
	}

	// change balance should modify one account's repState
	// change representative should modify two account's repState

	var oldBlkTs *types.PovTokenState
	var newBlkTs *types.PovTokenState

	if oldBlkAs != nil {
		oldBlkTs = oldBlkAs.GetTokenState(block.GetToken())
	}
	if newBlkAs != nil {
		newBlkTs = newBlkAs.GetTokenState(block.GetToken())
	}

	if oldBlkTs != nil && !oldBlkTs.Representative.IsZero() {
		var lastRepOldRs *types.PovRepState
		var lastRepNewRs *types.PovRepState

		lastRepOldRs = bc.GetRepState(trie, oldBlkTs.Representative)
		if lastRepOldRs != nil {
			lastRepNewRs = lastRepOldRs.Clone()
		} else {
			lastRepNewRs = types.NewPovRepState()
		}

		// old(last) representative minus old account balance
		if oldBlkAs != nil && lastRepOldRs != nil && lastRepNewRs != nil {
			lastRepNewRs.Balance = lastRepOldRs.Balance.Sub(oldBlkAs.Balance)
			lastRepNewRs.Vote = lastRepOldRs.Vote.Sub(oldBlkAs.Vote)
			lastRepNewRs.Network = lastRepOldRs.Network.Sub(oldBlkAs.Network)
			lastRepNewRs.Oracle = lastRepOldRs.Oracle.Sub(oldBlkAs.Oracle)
			lastRepNewRs.Storage = lastRepOldRs.Storage.Sub(oldBlkAs.Storage)
			lastRepNewRs.Total = lastRepOldRs.Total.Sub(oldBlkAs.TotalBalance())
		}

		bc.SetRepState(trie, oldBlkTs.Representative, lastRepNewRs)
	}

	if newBlkTs != nil && !newBlkTs.Representative.IsZero() {
		var currRepOldRs *types.PovRepState
		var currRepNewRs *types.PovRepState

		currRepOldRs = bc.GetRepState(trie, newBlkTs.Representative)
		if currRepOldRs != nil {
			currRepNewRs = currRepOldRs.Clone()
		} else {
			currRepNewRs = types.NewPovRepState()
		}

		// new(current) representative plus new account balance
		currRepNewRs.Balance = currRepNewRs.Balance.Add(block.Balance)
		currRepNewRs.Vote = currRepNewRs.Vote.Add(block.Vote)
		currRepNewRs.Network = currRepNewRs.Network.Add(block.Network)
		currRepNewRs.Oracle = currRepNewRs.Oracle.Add(block.Oracle)
		currRepNewRs.Storage = currRepNewRs.Storage.Add(block.Storage)
		currRepNewRs.Total = currRepNewRs.Total.Add(block.TotalBalance())

		bc.SetRepState(trie, newBlkTs.Representative, currRepNewRs)
	}
}

func (bc *PovBlockChain) GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState {
	keyBytes := types.PovCreateAccountStateKey(address)
	valBytes := trie.GetValue(keyBytes)
	if len(valBytes) <= 0 {
		return nil
	}

	as := new(types.PovAccountState)
	err := as.Deserialize(valBytes)
	if err != nil {
		bc.logger.Errorf("deserialize old account state err %s", err)
		return nil
	}

	//bc.logger.Debugf("get account %s state %s", address, as)

	return as
}

func (bc *PovBlockChain) SetAccountState(trie *trie.Trie, address types.Address, as *types.PovAccountState) {
	//bc.logger.Debugf("set account %s state %s", address, as)

	valBytes, err := as.Serialize()
	if err != nil {
		bc.logger.Errorf("serialize new account state err %s", err)
		return
	}

	keyBytes := types.PovCreateAccountStateKey(address)
	trie.SetValue(keyBytes, valBytes)
	return
}

func (bc *PovBlockChain) GetRepState(trie *trie.Trie, address types.Address) *types.PovRepState {
	keyBytes := types.PovCreateRepStateKey(address)
	valBytes := trie.GetValue(keyBytes)
	if len(valBytes) <= 0 {
		return nil
	}

	rs := new(types.PovRepState)
	err := rs.Deserialize(valBytes)
	if err != nil {
		bc.logger.Errorf("deserialize old rep state err %s", err)
		return nil
	}

	//bc.logger.Debugf("get rep %s state %s", address, as)

	return rs
}

func (bc *PovBlockChain) SetRepState(trie *trie.Trie, address types.Address, rs *types.PovRepState) {
	//bc.logger.Debugf("set rep %s state %s", address, as)

	valBytes, err := rs.Serialize()
	if err != nil {
		bc.logger.Errorf("serialize new rep state err %s", err)
		return
	}

	keyBytes := types.PovCreateRepStateKey(address)
	trie.SetValue(keyBytes, valBytes)
	return
}

func (bc *PovBlockChain) GetAllRepStates(trie *trie.Trie) []*types.PovRepState {
	var allRss []*types.PovRepState

	it := trie.NewIterator([]byte{types.PovStatePrefixRep})

	_, valBytes, ok := it.Next()
	for ok {
		if len(valBytes) > 0 {
			rs := new(types.PovRepState)
			err := rs.Deserialize(valBytes)
			if err != nil {
				bc.logger.Errorf("deserialize old rep state err %s", err)
				return nil
			}

			allRss = append(allRss, rs)
		}

		_, valBytes, ok = it.Next()
	}

	//bc.logger.Debugf("get all rep state %d", len(allRss))

	return allRss
}

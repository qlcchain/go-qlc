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

	bc.updateRepresentativeState(trie, stateBlock, oldAs, newAs)

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

func (bc *PovBlockChain) updateRepresentativeState(trie *trie.Trie, block *types.StateBlock, oldBlkAs *types.PovAccountState, newBlkAs *types.PovAccountState) {
	if block.GetToken() != common.ChainToken() {
		return
	}

	// change balance should modify one account's repState
	// change representative should modify two account's repState

	var oldBlkTs *types.PovTokenState
	if oldBlkAs != nil {
		oldBlkTs = oldBlkAs.GetTokenState(block.GetToken())
	}
	if oldBlkTs != nil && !oldBlkTs.Representative.IsZero() {
		var lastRepOldAs *types.PovAccountState
		var lastRepOldRs *types.PovRepState

		var lastRepNewAs *types.PovAccountState
		var lastRepNewRs *types.PovRepState

		lastRepOldAs = bc.GetAccountState(trie, oldBlkTs.Representative)
		if lastRepOldAs != nil {
			lastRepOldRs = lastRepOldAs.RepState

			lastRepNewAs = lastRepOldAs.Clone()
			lastRepNewRs = lastRepNewAs.RepState
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

		bc.SetAccountState(trie, oldBlkTs.Representative, lastRepNewAs)
	}

	newBlkTs := newBlkAs.GetTokenState(block.GetToken())
	if newBlkTs != nil && !newBlkTs.Representative.IsZero() {
		var currRepOldAs *types.PovAccountState

		var currRepNewAs *types.PovAccountState
		var currRepNewRs *types.PovRepState

		currRepOldAs = bc.GetAccountState(trie, newBlkTs.Representative)
		if currRepOldAs != nil {
			currRepNewAs = currRepOldAs.Clone()
		} else {
			currRepNewAs = types.NewPovAccountState()
		}

		if currRepNewAs.RepState == nil {
			currRepNewAs.RepState = types.NewPovRepState()
		}
		currRepNewRs = currRepNewAs.RepState

		// new(current) representative plus new account balance
		currRepNewRs.Balance = currRepNewRs.Balance.Add(block.Balance)
		currRepNewRs.Vote = currRepNewRs.Vote.Add(block.Vote)
		currRepNewRs.Network = currRepNewRs.Network.Add(block.Network)
		currRepNewRs.Oracle = currRepNewRs.Oracle.Add(block.Oracle)
		currRepNewRs.Storage = currRepNewRs.Storage.Add(block.Storage)
		currRepNewRs.Total = currRepNewRs.Total.Add(block.TotalBalance())

		bc.SetAccountState(trie, newBlkTs.Representative, currRepNewAs)
	}
}

func (bc *PovBlockChain) GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState {
	stateBytes := trie.GetValue(address.Bytes())
	if len(stateBytes) <= 0 {
		return nil
	}

	as := new(types.PovAccountState)
	err := as.Deserialize(stateBytes)
	if err != nil {
		bc.logger.Errorf("deserialize old account state err %s", err)
		return nil
	}

	//bc.logger.Debugf("get account %s state %s", address, as)

	return as
}

func (bc *PovBlockChain) SetAccountState(trie *trie.Trie, address types.Address, as *types.PovAccountState) {
	//bc.logger.Debugf("set account %s state %s", address, as)

	newStateBytes, err := as.Serialize()
	if err != nil {
		bc.logger.Errorf("serialize new account state err %s", err)
		return
	}

	trie.SetValue(address.Bytes(), newStateBytes)
	return
}

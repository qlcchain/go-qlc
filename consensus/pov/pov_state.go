package pov

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/trie"
)

func (bc *PovBlockChain) TrieDb() db.Store {
	return bc.getLedger().DBStore()
}

func (bc *PovBlockChain) GetStateTrie(stateHash *types.Hash) *trie.Trie {
	t := trie.NewTrie(bc.TrieDb(), stateHash, bc.trieNodePool)
	return t
}

func (bc *PovBlockChain) NewStateTrie() *trie.Trie {
	return trie.NewTrie(bc.TrieDb(), nil, bc.trieNodePool)
}

func (bc *PovBlockChain) GenStateTrie(height uint64, prevStateHash types.Hash,
	txs []*types.PovTransaction) (*trie.Trie, error) {
	var currentTrie *trie.Trie
	prevTrie := bc.GetStateTrie(&prevStateHash)
	if prevTrie == nil {
		return nil, fmt.Errorf("failed to get prev trie %s", prevStateHash)
	}
	currentTrie = prevTrie.Clone()
	if currentTrie == nil {
		return nil, errors.New("failed to make current trie by clone prev trie")
	}

	for _, tx := range txs {
		err := bc.ApplyTransaction(height, currentTrie, tx.Block)
		if err != nil {
			bc.logger.Errorf("failed to apply tx %s", tx.Hash)
			return nil, err
		}
	}

	return currentTrie, nil
}

func (bc *PovBlockChain) ApplyTransaction(height uint64, trie *trie.Trie, stateBlock *types.StateBlock) error {
	var err error

	oldAs := bc.GetAccountState(trie, stateBlock.GetAddress())
	var newAs *types.PovAccountState
	if oldAs != nil {
		newAs = oldAs.Clone()
	} else {
		newAs = types.NewPovAccountState()
	}

	err = bc.updateAccountState(trie, stateBlock, oldAs, newAs)
	if err != nil {
		return err
	}

	if stateBlock.GetType() != types.Online {
		err = bc.updateRepState(trie, stateBlock, oldAs, newAs)
		if err != nil {
			return err
		}
	}

	if stateBlock.GetType() == types.Online {
		err = bc.updateRepOnline(height, trie, stateBlock, oldAs, newAs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bc *PovBlockChain) updateAccountState(trie *trie.Trie, block *types.StateBlock,
	oldAs *types.PovAccountState, newAs *types.PovAccountState) error {
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

	err := bc.SetAccountState(trie, block.GetAddress(), newAs)
	if err != nil {
		return err
	}

	return nil
}

func (bc *PovBlockChain) updateRepState(trie *trie.Trie, block *types.StateBlock,
	oldBlkAs *types.PovAccountState, newBlkAs *types.PovAccountState) error {
	if block.GetToken() != common.ChainToken() {
		return nil
	}

	// change balance should modify one account's repState
	// change representative should modify two account's repState
	var err error

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

			err = bc.SetRepState(trie, oldBlkTs.Representative, lastRepNewRs)
			if err != nil {
				return err
			}
		}
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

		err = bc.SetRepState(trie, newBlkTs.Representative, currRepNewRs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bc *PovBlockChain) updateRepOnline(height uint64, trie *trie.Trie, block *types.StateBlock,
	oldBlkAs *types.PovAccountState, newBlkAs *types.PovAccountState) error {
	var newRs *types.PovRepState

	oldRs := bc.GetRepState(trie, block.GetAddress())
	if oldRs != nil {
		newRs = oldRs.Clone()
	} else {
		newRs = types.NewPovRepState()
	}

	newRs.Status = types.PovStatusOnline
	newRs.Height = height

	err := bc.SetRepState(trie, block.GetAddress(), newRs)
	if err != nil {
		return err
	}

	return nil
}

func (bc *PovBlockChain) GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState {
	keyBytes := types.PovCreateAccountStateKey(address)
	valBytes := trie.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil
	}

	as := types.NewPovAccountState()
	err := as.Deserialize(valBytes)
	if err != nil {
		bc.logger.Errorf("deserialize old account state err %s", err)
		return nil
	}

	return as
}

func (bc *PovBlockChain) SetAccountState(trie *trie.Trie, address types.Address, as *types.PovAccountState) error {
	as.Account = address

	valBytes, err := as.Serialize()
	if err != nil {
		bc.logger.Errorf("serialize new account state err %s", err)
		return err
	}
	if len(valBytes) == 0 {
		return errors.New("serialize new account state got empty value")
	}

	keyBytes := types.PovCreateAccountStateKey(address)
	trie.SetValue(keyBytes, valBytes)
	return nil
}

func (bc *PovBlockChain) GetRepState(trie *trie.Trie, address types.Address) *types.PovRepState {
	keyBytes := types.PovCreateRepStateKey(address)
	valBytes := trie.GetValue(keyBytes)
	if len(valBytes) == 0 {
		return nil
	}

	rs := types.NewPovRepState()
	err := rs.Deserialize(valBytes)
	if err != nil {
		bc.logger.Errorf("deserialize old rep state err %s", err)
		return nil
	}

	return rs
}

func (bc *PovBlockChain) SetRepState(trie *trie.Trie, address types.Address, rs *types.PovRepState) error {
	rs.Account = address

	valBytes, err := rs.Serialize()
	if err != nil {
		bc.logger.Errorf("serialize new rep state err %s", err)
		return err
	}
	if len(valBytes) == 0 {
		return errors.New("serialize new rep state got empty value")
	}

	keyBytes := types.PovCreateRepStateKey(address)
	trie.SetValue(keyBytes, valBytes)
	return nil
}

func (bc *PovBlockChain) GetAllOnlineRepStates(header *types.PovHeader) []*types.PovRepState {
	var allRss []*types.PovRepState
	supply := common.GenesisBlock().Balance
	minVoteWeight, _ := supply.Div(common.DposVoteDivisor)

	stateHash := header.GetStateHash()
	stateTrie := bc.GetStateTrie(&stateHash)
	if stateTrie == nil {
		return nil
	}

	repPrefix := types.PovCreateStatePrefix(types.PovStatePrefixRep)
	it := stateTrie.NewIterator(repPrefix)
	if it == nil {
		return nil
	}

	key, valBytes, ok := it.Next()
	for ; ok; key, valBytes, ok = it.Next() {
		if len(valBytes) == 0 {
			continue
		}

		rs := types.NewPovRepState()
		err := rs.Deserialize(valBytes)
		if err != nil {
			bc.logger.Errorf("deserialize old rep state, key %s err %s", hex.EncodeToString(key), err)
			return nil
		}

		if rs.Status != types.PovStatusOnline {
			continue
		}

		if rs.Height > header.GetHeight() || rs.Height < (header.GetHeight()+1-common.DPosOnlinePeriod) {
			continue
		}

		if rs.CalcTotal().Compare(minVoteWeight) == types.BalanceCompSmaller {
			continue
		}

		allRss = append(allRss, rs)
	}

	return allRss
}

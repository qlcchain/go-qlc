package pov

import (
	"encoding/hex"
	"fmt"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func (bc *PovBlockChain) TrieDb() storage.Store {
	return bc.getLedger().DBStore()
}

func (bc *PovBlockChain) TransitStateDB(height uint64, txs []*types.PovTransaction, gsdb *statedb.PovGlobalStateDB) error {
	for _, tx := range txs {
		err := bc.ApplyTransaction(height, gsdb, tx)
		if err != nil {
			bc.logger.Errorf("failed to apply tx %s, err %s", tx.Hash, err)
			return err
		}
	}

	err := gsdb.CommitToTrie()
	if err != nil {
		bc.logger.Errorf("failed to commit trie for block %d", height)
		return err
	}

	return nil
}

func (bc *PovBlockChain) ApplyTransaction(height uint64, sdb *statedb.PovGlobalStateDB, tx *types.PovTransaction) error {
	var err error

	oldAs, _ := sdb.GetAccountState(tx.Block.GetAddress())
	var newAs *types.PovAccountState
	if oldAs != nil {
		newAs = oldAs.Clone()
	} else {
		newAs = types.NewPovAccountState()
	}

	err = bc.updateAccountState(sdb, tx, oldAs, newAs)
	if err != nil {
		return err
	}

	if tx.Block.GetType() != types.Online {
		err = bc.updateRepState(sdb, tx, oldAs, newAs)
		if err != nil {
			return err
		}
	}

	if tx.Block.GetType() == types.Online {
		err = bc.updateRepOnline(height, sdb, tx)
		if err != nil {
			return err
		}
	}

	err = bc.updateContractState(height, sdb, tx)
	if err != nil {
		return err
	}

	return nil
}

func (bc *PovBlockChain) updateAccountState(sdb *statedb.PovGlobalStateDB, tx *types.PovTransaction,
	oldAs *types.PovAccountState, newAs *types.PovAccountState) error {
	block := tx.Block
	hash := tx.GetHash()

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
		if block.GetToken() == config.ChainToken() {
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

		if block.GetToken() == config.ChainToken() {
			newAs.Balance = balance
			newAs.Oracle = block.GetOracle()
			newAs.Network = block.GetNetwork()
			newAs.Vote = block.GetVote()
			newAs.Storage = block.GetStorage()
		}
	}

	err := sdb.SetAccountState(block.GetAddress(), newAs)
	if err != nil {
		return err
	}

	return nil
}

func (bc *PovBlockChain) updateRepState(sdb *statedb.PovGlobalStateDB, tx *types.PovTransaction,
	oldBlkAs *types.PovAccountState, newBlkAs *types.PovAccountState) error {
	block := tx.Block

	if block.GetToken() != config.ChainToken() {
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

		lastRepOldRs, _ = sdb.GetRepState(oldBlkTs.Representative)
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

			err = sdb.SetRepState(oldBlkTs.Representative, lastRepNewRs)
			if err != nil {
				return err
			}
		}
	}

	if newBlkTs != nil && !newBlkTs.Representative.IsZero() {
		var currRepOldRs *types.PovRepState
		var currRepNewRs *types.PovRepState

		currRepOldRs, _ = sdb.GetRepState(newBlkTs.Representative)
		if currRepOldRs != nil {
			currRepNewRs = currRepOldRs.Clone()
		} else {
			currRepNewRs = types.NewPovRepState()
		}

		// new(current) representative plus new account balance
		currRepNewRs.Balance = currRepNewRs.Balance.Add(block.Balance)
		currRepNewRs.Vote = currRepNewRs.Vote.Add(block.GetVote())
		currRepNewRs.Network = currRepNewRs.Network.Add(block.GetNetwork())
		currRepNewRs.Oracle = currRepNewRs.Oracle.Add(block.GetOracle())
		currRepNewRs.Storage = currRepNewRs.Storage.Add(block.GetStorage())
		currRepNewRs.Total = currRepNewRs.Total.Add(block.TotalBalance())

		err = sdb.SetRepState(newBlkTs.Representative, currRepNewRs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bc *PovBlockChain) updateRepOnline(height uint64, sdb *statedb.PovGlobalStateDB, tx *types.PovTransaction) error {
	var newRs *types.PovRepState

	block := tx.Block

	oldRs, _ := sdb.GetRepState(block.GetAddress())
	if oldRs != nil {
		newRs = oldRs.Clone()
	} else {
		newRs = types.NewPovRepState()
	}

	newRs.Status = statedb.PovStatusOnline
	newRs.Height = height

	err := sdb.SetRepState(block.GetAddress(), newRs)
	if err != nil {
		return err
	}

	return nil
}

func (bc *PovBlockChain) updateContractState(height uint64, gsdb *statedb.PovGlobalStateDB, tx *types.PovTransaction) error {
	var sendBlk *types.StateBlock
	var methodSig []byte
	var err error

	txBlock := tx.Block

	if config.IsGenesisBlock(txBlock) {
		bc.logger.Infof("no need to update contract state for genesis block %s", tx.Hash)
		return nil
	}

	ca := types.Address{}
	if txBlock.GetType() == types.ContractSend {
		// check private tx
		if txBlock.IsPrivate() && !txBlock.IsRecipient() {
			return nil
		}

		ca = types.Address(txBlock.GetLink())
		methodSig = txBlock.GetPayload()
	} else if txBlock.GetType() == types.ContractReward {
		// check private tx
		if txBlock.IsPrivate() && !txBlock.IsRecipient() {
			return nil
		}

		sendBlk, err = bc.ledger.GetStateBlockConfirmed(txBlock.GetLink())
		if err != nil {
			bc.logger.Errorf("failed to get chain contract send block err %s", err)
			return err
		}

		// check private tx
		if sendBlk.IsPrivate() && !sendBlk.IsRecipient() {
			return nil
		}

		ca = types.Address(sendBlk.GetLink())
		methodSig = sendBlk.GetPayload()
	} else {
		return nil
	}

	cf, ok, err := contract.GetChainContract(ca, methodSig)
	if err != nil {
		bc.logger.Errorf("failed to get chain contract err %s", err)
		return err
	}
	if !ok {
		err := fmt.Errorf("chain contract %s method not exist", ca)
		bc.logger.Errorf("failed to get chain contract err %s", err)
		return err
	}

	if !cf.GetDescribe().WithPovState() {
		return nil
	}

	csdb, err := gsdb.LookupContractStateDB(ca)
	if err != nil {
		return err
	}

	vmCtx := vmstore.NewVMContextWithBlock(bc.ledger, txBlock)
	if vmCtx == nil {
		return fmt.Errorf("update contract state: can not get vm context, %s", err)
	}
	if txBlock.GetType() == types.ContractSend {
		err = cf.DoSendOnPov(vmCtx, csdb, height, txBlock)
		if err != nil {
			return err
		}
	} else if txBlock.GetType() == types.ContractReward {
		err = cf.DoReceiveOnPov(vmCtx, csdb, height, txBlock, sendBlk)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bc *PovBlockChain) GetAllOnlineRepStates(header *types.PovHeader) []*types.PovRepState {
	var allRss []*types.PovRepState
	supply := config.GenesisBlock().Balance
	minVoteWeight, _ := supply.Div(common.DposVoteDivisor)

	gsdb := statedb.NewPovGlobalStateDB(bc.TrieDb(), header.GetStateHash())
	stateTrie := gsdb.GetPrevTrie()
	if stateTrie == nil {
		return nil
	}

	repPrefix := statedb.PovCreateGlobalStateKey(statedb.PovGlobalStatePrefixRep, nil)
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

		if rs.Status != statedb.PovStatusOnline {
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

package process

import (
	"fmt"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

func (lv *LedgerVerifier) BlockCacheCheck(block *types.StateBlock) (ProcessResult, error) {
	lv.logger.Infof("check cache block, %s(%s)", block.GetHash(), block.GetType().String())
	if c, ok := lv.cacheBlockCheck[block.Type]; ok {
		r, err := c.Check(lv, block)
		if err != nil {
			lv.logger.Errorf("error:%s, block:%s", err.Error(), block.GetHash().String())
		}
		if r != Progress {
			lv.logger.Infof("check cache result:%s,(%s, %s)", r.String(), block.GetHash().String(), block.GetType().String())
		}
		return r, err
	} else {
		return Other, fmt.Errorf("unsupport block type %s", block.Type.String())
	}
}

func newCacheBlockCheck() map[types.BlockType]blockCheck {
	r := make(map[types.BlockType]blockCheck)
	r[types.Open] = &cacheOpenBlockCheck{}
	r[types.Send] = &cacheSendBlockCheck{}
	r[types.Receive] = &cacheReceiveBlockCheck{}
	r[types.Change] = &cacheChangeBlockCheck{}
	r[types.Online] = &cacheChangeBlockCheck{}
	r[types.ContractSend] = &cacheContractSendBlockCheck{}
	r[types.ContractReward] = &cacheContractReceiveBlockCheck{}
	return r
}

type cacheSendBlockCheck struct {
	cacheBlockBaseInfoCheck
	cacheBlockForkCheck
	cacheBlockBalanceCheck
}

func (c *cacheSendBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	return c.balance(lv, block)
}

type cacheContractSendBlockCheck struct {
	cacheBlockBaseInfoCheck
	cacheBlockForkCheck
	cacheBlockBalanceCheck
	cacheBlockContractCheck
}

func (c *cacheContractSendBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//ignore chain genesis block
	if config.IsGenesisBlock(block) {
		return Progress, nil
	}
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.balance(lv, block); r != Progress || err != nil {
		return r, err
	}
	return c.contract(lv, block)
}

type cacheReceiveBlockCheck struct {
	cacheBlockBaseInfoCheck
	cacheBlockForkCheck
	cacheBlockSourceCheck
	cacheBlockPendingCheck
}

func (c *cacheReceiveBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.source(lv, block); r != Progress || err != nil {
		return r, err
	}
	return c.pending(lv, block)
}

type cacheContractReceiveBlockCheck struct {
	cacheBlockBaseInfoCheck
	cacheBlockForkCheck
	cacheBlockPendingCheck
	cacheBlockSourceCheck
	cacheBlockContractCheck
}

func (c *cacheContractReceiveBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//ignore chain genesis block
	if config.IsGenesisBlock(block) {
		return Progress, nil
	}
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.source(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.pending(lv, block); r != Progress || err != nil {
		return r, err
	}
	return c.contract(lv, block)
}

type cacheOpenBlockCheck struct {
	cacheBlockBaseInfoCheck
	cacheBlockForkCheck
	cacheBlockSourceCheck
	cacheBlockPendingCheck
}

func (c *cacheOpenBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//check previous
	if !block.Previous.IsZero() {
		return Other, fmt.Errorf("open block previous is not zero")
	}

	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.source(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	return c.pending(lv, block)
}

type cacheChangeBlockCheck struct {
	cacheBlockBaseInfoCheck
	cacheBlockForkCheck
	cacheBlockBalanceCheck
}

func (c *cacheChangeBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	// check link
	if !block.Link.IsZero() {
		return Other, fmt.Errorf("invalid link hash")
	}
	// check chain token
	if block.Token != config.ChainToken() {
		return Other, fmt.Errorf("invalid token Id")
	}
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	return c.balance(lv, block)
}

func (lv *LedgerVerifier) BlockCacheProcess(block *types.StateBlock) error {
	lv.logger.Infof("block cache process: %s(%s) ", block.GetHash().String(), block.GetType().String())
	am, err := lv.l.GetAccountMeta(block.GetAddress())
	if err != nil && err != ledger.ErrAccountNotFound {
		return fmt.Errorf("get account meta cache error: %s", err)
	}
	batch := lv.l.DBStore().Batch(true)
	if err := lv.l.AddBlockCache(block, batch); err != nil {
		return fmt.Errorf("update block cache error: %s", err)
	}
	if err := lv.updateAccountMetaCache(block, am, batch); err != nil {
		return fmt.Errorf("update account meta cache error: %s", err)
	}
	return lv.l.DBStore().PutBatch(batch)
}

func (lv *LedgerVerifier) updateAccountMetaCache(block *types.StateBlock, am *types.AccountMeta, batch storage.Batch) error {
	hash := block.GetHash()
	rep := block.GetRepresentative()
	address := block.GetAddress()
	token := block.GetToken()
	balance := block.GetBalance()

	tmNew := &types.TokenMeta{
		Type:           token,
		Header:         hash,
		Representative: rep,
		OpenBlock:      hash,
		Balance:        balance,
		BlockCount:     1,
		BelongTo:       address,
		Modified:       common.TimeNow().UTC().Unix(),
	}

	if am != nil {
		tm := am.Token(block.GetToken())
		if block.GetToken() == config.ChainToken() {
			am.CoinBalance = balance
			am.CoinOracle = block.GetOracle()
			am.CoinNetwork = block.GetNetwork()
			am.CoinVote = block.GetVote()
			am.CoinStorage = block.GetStorage()
		}
		if tm != nil {
			tm.Header = hash
			tm.Representative = rep
			tm.Balance = balance
			tm.BlockCount = tm.BlockCount + 1
			tm.Modified = common.TimeNow().UTC().Unix()
		} else {
			am.Tokens = append(am.Tokens, tmNew)
		}
		return lv.l.AddOrUpdateAccountMetaCache(am, batch)
	} else {
		account := types.AccountMeta{
			Address: address,
			Tokens:  []*types.TokenMeta{tmNew},
		}

		if block.GetToken() == config.ChainToken() {
			account.CoinBalance = balance
			account.CoinOracle = block.GetOracle()
			account.CoinNetwork = block.GetNetwork()
			account.CoinVote = block.GetVote()
			account.CoinStorage = block.GetStorage()
		}
		return lv.l.AddAccountMetaCache(&account, batch)
	}
}

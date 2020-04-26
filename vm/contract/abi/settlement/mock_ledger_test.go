/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

var (
	account1 = mock.Account1
	account2 = mock.Account2
)

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	t.Parallel()
	dir := filepath.Join(cfg.QlcTestDataDir(), "settlement", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := cfg.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	l := ledger.NewLedger(cm.ConfigFile)

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}

	//verifier := process.NewLedgerVerifier(l)

	for i := range blocks {
		block := blocks[i]

		//if err := verifier.BlockProcess(block); err != nil {
		//	t.Fatal(err)
		//}
		if err := updateBlock(l, block); err != nil {
			t.Fatal(err)
		}
	}

	return func(t *testing.T) {
		//err := l.DBStore.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l
}

func updateBlock(l *ledger.Ledger, block *types.StateBlock) error {
	return l.Cache().BatchUpdate(func(c *ledger.Cache) error {
		err := l.UpdateStateBlock(block, c)
		if err != nil {
			return err
		}
		am, err := l.GetAccountMetaConfirmed(block.GetAddress(), c)
		if err != nil && err != ledger.ErrAccountNotFound {
			return fmt.Errorf("get account meta error: %s", err)
		}
		tm, err := l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken())
		if err != nil && err != ledger.ErrAccountNotFound && err != ledger.ErrTokenNotFound {
			return fmt.Errorf("get token meta error: %s", err)
		}
		err = updateFrontier(l, block, tm, c)
		if err != nil {
			return err
		}
		err = updateAccountMeta(l, block, am, c)
		if err != nil {
			return err
		}
		return nil
	})
}

func updateFrontier(l *ledger.Ledger, block *types.StateBlock, tm *types.TokenMeta, c *ledger.Cache) error {
	hash := block.GetHash()
	frontier := &types.Frontier{
		HeaderBlock: hash,
	}
	if tm != nil {
		if frontier, err := l.GetFrontier(tm.Header, c); err == nil {
			if err := l.DeleteFrontier(frontier.HeaderBlock, c); err != nil {
				return err
			}
		} else {
			return err
		}
		frontier.OpenBlock = tm.OpenBlock
	} else {
		frontier.OpenBlock = hash
	}
	if err := l.AddFrontier(frontier, c); err != nil {
		return err
	}
	return nil
}

func updateAccountMeta(l *ledger.Ledger, block *types.StateBlock, am *types.AccountMeta, c *ledger.Cache) error {
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
		if block.GetToken() == cfg.ChainToken() {
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
		if err := l.UpdateAccountMeta(am, c); err != nil {
			return err
		}
	} else {
		account := types.AccountMeta{
			Address: address,
			Tokens:  []*types.TokenMeta{tmNew},
		}

		if block.GetToken() == cfg.ChainToken() {
			account.CoinBalance = balance
			account.CoinOracle = block.GetOracle()
			account.CoinNetwork = block.GetNetwork()
			account.CoinVote = block.GetVote()
			account.CoinStorage = block.GetStorage()
		}
		if err := l.AddAccountMeta(&account, c); err != nil {
			return err
		}
	}
	return nil
}

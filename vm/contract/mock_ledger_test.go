/*
 * Copyright (c) 2020. QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"encoding/hex"
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
)

var (
	// qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq
	priv1, _ = hex.DecodeString("7098c089e66bd66476e3b88df8699bcd4dacdd5e1e5b41b3c598a8a36d851184d992a03b7326b7041f689ae727292d761b329a960f3e4335e0a7dcf2c43c4bcf")
	// qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te
	priv2, _ = hex.DecodeString("31ee4e16826569dc631b969e71bd4c46d5c0df0daeca6933f46586f36f49537cd929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393")
	account1 = types.NewAccount(priv1)
	account2 = types.NewAccount(priv2)
)

func setupLedgerForTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	dir := filepath.Join(cfg.QlcTestDataDir(), "settlement", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := cfg.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	l := ledger.NewLedger(cm.ConfigFile)
	//ctx := vmstore.NewVMContext(l)
	//verifier := process.NewLedgerVerifier(l)
	//
	//for _, v := range common.genesisInfos {
	//	mb := v.GenesisMintageBlock
	//	gb := v.GenesisBlock
	//	err := ctx.SetStorage(types.MintageAddress[:], v.GenesisBlock.Token[:], v.GenesisBlock.Data)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	if b, err := l.HasStateBlock(mb.GetHash()); !b && err == nil {
	//		if err := l.AddStateBlock(&mb); err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//	if b, err := l.HasStateBlock(gb.GetHash()); !b && err == nil {
	//		if err := verifier.BlockProcess(&gb); err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//}
	//_ = ctx.SaveStorage()

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}

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

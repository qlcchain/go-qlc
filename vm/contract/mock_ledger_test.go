/*
 * Copyright (c) 2019 QLC Chain Team
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

const (
	accountBlocks = `[
	{
		"type": "Open",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "100000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "c0d330096ec4ab6ccf5481e06cc54e74b14f534e99e38df486f47d1123cbd1ae",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "747648bafd344347582876662641c4b8ffbf20a85ba01dc559ff930435bc5bad",
		"povHeight": 0,
		"timestamp": 1580997079,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "000000000122e972",
		"signature": "5460905ad2096d1822dc086e8fe375409f9fc87f0e8288ca215a399eb2fee6c6c5fc94b53a18f62fe6d124f869cbac1737b762c9a8f7654d1b7ecacc480f010a"
	},
	{
		"type": "Send",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "40000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "cad0cad8a98813787dc11ba2413afca574f2d62e222fd2644cc33c7d70124d90",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "f82eae0fa0f56a53e9d217140eaa33219c7cb910439501f333383f4d6147618c",
		"povHeight": 0,
		"timestamp": 1580997083,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000002aa56ad",
		"signature": "dd0af652dc5acca94547b5e130a38a2728531235d224e6031f13d9958221fcd4852bee6684fa1b29c5170f71f7f301b64eda8d10208ecc12b2862b34ea049a0e"
	},
	{
		"type": "Open",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te",
		"balance": "60000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "b05f7c462867df6f24b810c0b28b50d709667feb7d870a2b1db23bb3fa491249",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "eb9c1dcccaba3937d8745c364dadb1ca056cfa9540184277ad6fe8af66f81358",
		"povHeight": 0,
		"timestamp": 1580997093,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "00000000002389ad",
		"signature": "5d35efd693b85ccf4a01e4f132aa0a248b328b10024af2a22e65474038a4aea3decb6c414c5563b326f509f4cf5eac852c81317a96a8c349b965849c31e5580d"
	}
]`
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
	if err := json.Unmarshal([]byte(accountBlocks), &blocks); err != nil {
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

// +build !testnet

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
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var (
	priv, _   = hex.DecodeString("9783e5073f3a32895bc3ebb1f266b028cfbcea2dd35db57e1243a327058b4e1729eb31893dbde6f18a979c537885175032eec84e88ece0b8c341aab9f48223ed")
	acc       = types.NewAccount(priv)
	addr      = acc.Address()
	blks_data = `[
  {
    "type": "Open",
    "token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
    "address": "qlc_1chd886muhh8y87bh94mh44jgn3kxu66x49ew4we8ifcq9ta6azftarn4a47",
    "balance": "100000000000",
    "vote": "0",
    "network": "0",
    "storage": "0",
    "oracle": "0",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "link": "c0d330096ec4ab6ccf5481e06cc54e74b14f534e99e38df486f47d1123cbd1ae",
    "sender": "MTU4MTExMTAwMDA=",
    "receiver": "MTU4MDAwMDExMTE=",
    "message": "270c0fd4be6ccfe0e730a6e4733ae4ad81551f1df0548c3498c6eda1fb54fabc",
    "povHeight": 0,
    "timestamp": 1569317230,
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "representative": "qlc_1chd886muhh8y87bh94mh44jgn3kxu66x49ew4we8ifcq9ta6azftarn4a47",
    "work": "00000000007e1b53",
    "signature": "74dbf93011109c50be76fcc515c37cce7476928703cf8983ad42886580dc95521922250a1083b1d65cc8043086520be95ed1372ea07e185fb429d3267da90f05"
  },
  {
    "type": "Send",
    "token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
    "address": "qlc_1chd886muhh8y87bh94mh44jgn3kxu66x49ew4we8ifcq9ta6azftarn4a47",
    "balance": "40000000000",
    "vote": "0",
    "network": "0",
    "storage": "0",
    "oracle": "0",
    "previous": "ef133fc0202af329cf3d968c69af26b923a1059d70d9a3c2141e6e32040f8d54",
    "link": "c696653ec78f433585f4a30af47d702432bc4956b5f004f419e958ae99a82f02",
    "sender": "MTU4MTExMTAwMDA=",
    "receiver": "MTU4MDAwMDExMTE=",
    "message": "1ec85e93daeb09bff0638c5b580812f2c6eb7e8006a9bdf06f37e42d284c96a5",
    "povHeight": 0,
    "timestamp": 1569317233,
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "representative": "qlc_1chd886muhh8y87bh94mh44jgn3kxu66x49ew4we8ifcq9ta6azftarn4a47",
    "work": "000000000003f998",
    "signature": "375d17bf1b3124e41515ff8ce64c08a092cf49a9c478076e93f22d7b3176a703222adc759e55bdbed172a5bf79ea1c3e7b3bdbf5a8e8a16d5212327742aaea05"
  },
  {
    "type": "Open",
    "token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
    "address": "qlc_3jnpenzeh5t58p4zbarcyjyq1b3kqj6offhi1mt3mtcrotetidr4gqpqecn4",
    "balance": "60000000000",
    "vote": "0",
    "network": "0",
    "storage": "0",
    "oracle": "0",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "link": "9a8a816d8ea407030c95295ef05c47c125961804fc748a6060c03327a848dffe",
    "sender": "MTU4MTExMTAwMDA=",
    "receiver": "MTU4MDAwMDExMTE=",
    "message": "d2c2cf85f44bccfac4b202f4cf354c2717f4ab047ef20569b235fdb28c3962ad",
    "povHeight": 0,
    "timestamp": 1569317233,
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "representative": "qlc_1chd886muhh8y87bh94mh44jgn3kxu66x49ew4we8ifcq9ta6azftarn4a47",
    "work": "00000000009dce75",
    "signature": "311ae25027c2fc584b4fd6e28a7008e3a4f6ca722c1130f5c698f0252df1d8c1b867946e0c917b8642923bf433151b91c7ea0c993517b0dcc5b042ac08c9bd0b"
  },
  {
    "type": "Change",
    "token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
    "address": "qlc_3jnpenzeh5t58p4zbarcyjyq1b3kqj6offhi1mt3mtcrotetidr4gqpqecn4",
    "balance": "60000000000",
    "vote": "0",
    "network": "0",
    "storage": "0",
    "oracle": "0",
    "previous": "94e6abbce31090f699e0e7b7c798ccd6e89f71ee3b15040f69b5532f873d953f",
    "link": "0000000000000000000000000000000000000000000000000000000000000000",
    "sender": "MTU4MTExMTAwMDA=",
    "receiver": "MTU4MDAwMDExMTE=",
    "message": "dd039923d32a0f11d8be68f7aef924835f9a3d3636214f1235a6507248069482",
    "povHeight": 0,
    "timestamp": 1569317237,
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "representative": "qlc_3jnpenzeh5t58p4zbarcyjyq1b3kqj6offhi1mt3mtcrotetidr4gqpqecn4",
    "work": "00000000000261b1",
    "signature": "44992faf6e640c79eccbd5e24d15d79d32289ef04a19d43f85e90eaaa28db6b7673a2c417b8bd30e7a01bc6c430cd16f8235243988c653014a92961bcb218009"
  },
  {
    "type": "Send",
    "token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
    "address": "qlc_3jnpenzeh5t58p4zbarcyjyq1b3kqj6offhi1mt3mtcrotetidr4gqpqecn4",
    "balance": "30000000000",
    "vote": "0",
    "network": "0",
    "storage": "0",
    "oracle": "0",
    "previous": "35226153d01564d2d92be96e8941b871bc96c79064228fdfd6c409daac3989f4",
    "link": "29eb31893dbde6f18a979c537885175032eec84e88ece0b8c341aab9f48223ed",
    "sender": "MTU4MTExMTAwMDA=",
    "receiver": "MTU4MDAwMDExMTE=",
    "message": "c1a4d1897db706667934d0987f320d54bef582ef85136c5fbf6bb5aa80c5d73f",
    "povHeight": 0,
    "timestamp": 1569317237,
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "representative": "qlc_3jnpenzeh5t58p4zbarcyjyq1b3kqj6offhi1mt3mtcrotetidr4gqpqecn4",
    "work": "000000000014c332",
    "signature": "057127965ec7a09e77bd47d984b3d616d8c15948a26f5410711d99ef811edbe47b90c2cc1b3d5696312b829d81923c1b7617c13eb0cf748710fffcdf9dbcff03"
  },
  {
    "type": "Receive",
    "token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
    "address": "qlc_1chd886muhh8y87bh94mh44jgn3kxu66x49ew4we8ifcq9ta6azftarn4a47",
    "balance": "70000000000",
    "vote": "0",
    "network": "0",
    "storage": "0",
    "oracle": "0",
    "previous": "9a8a816d8ea407030c95295ef05c47c125961804fc748a6060c03327a848dffe",
    "link": "60b614c24efc71b2ef5c522f28cffc58f3a7e381902a545154f852ca1c141705",
    "sender": "MTU4MTExMTAwMDA=",
    "receiver": "MTU4MDAwMDExMTE=",
    "message": "4c2b43748d8bfb2cf6575c737f2dc7504ad7d7ae6221e03f5056730561c3c0d5",
    "povHeight": 0,
    "timestamp": 1569317237,
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "representative": "qlc_1chd886muhh8y87bh94mh44jgn3kxu66x49ew4we8ifcq9ta6azftarn4a47",
    "work": "000000000062070c",
    "signature": "cc502d7b8007cedf2017c554e20a4eb34982e6564580efe2b118b9b48438e85e9622412677288528aa7e8f751a4386e3455866e66bb03a8f2a39a60aa3e3410d"
  }
]
`
)

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	dir := filepath.Join(cfg.QlcTestDataDir(), "destroy", uuid.New().String())

	_ = os.RemoveAll(filepath.Join(cfg.QlcTestDataDir(), "destroy"))
	l := ledger.NewLedger(dir)
	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(blks_data), &blocks); err != nil {
		t.Fatal(err)
	}
	for _, block := range blocks[0:2] {
		err := l.BatchUpdate(func(txn db.StoreTxn) error {
			err := l.AddStateBlock(block, txn)
			if err != nil {
				return err
			}
			am, err := l.GetAccountMetaConfirmed(block.GetAddress(), txn)
			if err != nil && err != ledger.ErrAccountNotFound {
				return fmt.Errorf("get account meta error: %s", err)
			}
			tm, err := l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken(), txn)
			if err != nil && err != ledger.ErrAccountNotFound && err != ledger.ErrTokenNotFound {
				return fmt.Errorf("get token meta error: %s", err)
			}
			err = updateFrontier(l, block, tm, txn)
			if err != nil {
				return err
			}
			err = updateAccountMeta(l, block, am, txn)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	return func(t *testing.T) {
		//err := l.Store.Erase()
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

func updateFrontier(l *ledger.Ledger, block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	hash := block.GetHash()
	frontier := &types.Frontier{
		HeaderBlock: hash,
	}
	if tm != nil {
		if frontier, err := l.GetFrontier(tm.Header, txn); err == nil {
			if err := l.DeleteFrontier(frontier.HeaderBlock, txn); err != nil {
				return err
			}
		} else {
			return err
		}
		frontier.OpenBlock = tm.OpenBlock
	} else {
		frontier.OpenBlock = hash
	}
	if err := l.AddFrontier(frontier, txn); err != nil {
		return err
	}
	return nil
}

func updateAccountMeta(l *ledger.Ledger, block *types.StateBlock, am *types.AccountMeta, txn db.StoreTxn) error {
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
		if block.GetToken() == common.ChainToken() {
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
		if err := l.UpdateAccountMeta(am, txn); err != nil {
			return err
		}
	} else {
		account := types.AccountMeta{
			Address: address,
			Tokens:  []*types.TokenMeta{tmNew},
		}

		if block.GetToken() == common.ChainToken() {
			account.CoinBalance = balance
			account.CoinOracle = block.GetOracle()
			account.CoinNetwork = block.GetNetwork()
			account.CoinVote = block.GetVote()
			account.CoinStorage = block.GetStorage()
		}
		if err := l.AddAccountMeta(&account, txn); err != nil {
			return err
		}
	}
	return nil
}

func TestDestroyContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	tm, err := l.GetTokenMeta(acc.Address(), common.GasToken())
	if err != nil {
		t.Fatal(err)
	}

	param := &abi.DestroyParam{
		Owner:    acc.Address(),
		Token:    common.GasToken(),
		Previous: tm.Header,
		Amount:   big.NewInt(100),
	}

	param.Sign, err = param.Signature(acc)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(util.ToIndentString(param))

	if b, err := param.Verify(); err != nil {
		t.Fatal(err)
	} else if !b {
		t.Fatal("invalid sign")
	}

	vmContext := vmstore.NewVMContext(l)
	b := &BlackHole{}
	sendBlock, err := abi.PackSendBlock(vmContext, param)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(sendBlock))
	key, info, err := b.ProcessSend(vmContext, sendBlock)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(util.ToIndentString(key), util.ToIndentString(info))
	}
	recv := &types.StateBlock{
		Timestamp: common.TimeNow().Unix(),
	}
	blocks, err := b.DoReceive(vmContext, recv, sendBlock)
	if err != nil {
		t.Fatal(err)
	} else {
		if len(blocks) > 0 {
			t.Log(util.ToIndentString(blocks[0].Block))
		}
	}

	err = vmContext.SaveStorage()
	if err != nil {
		t.Fatal(err)
	}

	if infos, err := abi.GetDestroyInfoDetail(vmContext, &addr); err == nil {
		for idx, info := range infos {
			t.Log(idx, util.ToIndentString(info))
		}
	} else {
		t.Fatal(err)
	}
}

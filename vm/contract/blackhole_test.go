// +build !testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"math/big"
	"testing"

	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestDestroyContract(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	addr := account1.Address()

	tm, err := l.GetTokenMeta(addr, cfg.GasToken())
	if err != nil {
		t.Fatal(err, cfg.GasToken())
	}

	param := &abi.DestroyParam{
		Owner:    addr,
		Token:    cfg.GasToken(),
		Previous: tm.Header,
		Amount:   big.NewInt(100),
	}

	param.Sign, err = param.Signature(account1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(util.ToIndentString(param))

	if b, err := param.Verify(); err != nil {
		t.Fatal(err)
	} else if !b {
		t.Fatal("invalid sign")
	}

	vmContext := vmstore.NewVMContext(l, &contractaddress.BlackHoleAddress)
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
		if _, _, err = b.ProcessSend(vmContext, sendBlock); err != nil {
			t.Fatal(err)
		}
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

	err = l.SaveStorage(vmstore.ToCache(vmContext))
	if err != nil {
		t.Fatal(err)
	}

	if infos, err := abi.GetDestroyInfoDetail(l, &addr); err == nil {
		for idx, info := range infos {
			t.Log(idx, util.ToIndentString(info))
		}
	} else {
		t.Fatal(err)
	}
}

// +build testnet

package api

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
)

func TestContractApi_Privacy(t *testing.T) {
	tearDown, md := setupTestCaseContractApi(t)
	defer tearDown(t)

	md.eb.Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(10 * time.Millisecond)

	api := NewContractApi(md.cc, md.l)

	paraList := []string{
		hex.EncodeToString(util.RandomFixedBytes(32)),
		hex.EncodeToString(util.RandomFixedBytes(32)),
	}
	data, err := api.PackChainContractData(contractaddress.PrivacyDemoKVAddress, "PrivacyDemoKVSet", paraList)
	if err != nil {
		t.Fatal(err)
	}

	sendPara := &ContractSendBlockPara{
		Address:   config.GenesisAddress(),
		To:        contractaddress.PrivacyDemoKVAddress,
		TokenName: "QLC",
		Amount:    types.NewBalance(0),
		Data:      data,

		PrivateFrom: util.RandomFixedString(32),
		PrivateFor:  []string{util.RandomFixedString(32)},
	}
	sendBlk, err := api.GenerateSendBlock(sendPara)
	if err != nil {
		t.Fatal(err)
	}
	if sendBlk == nil {
		t.Fatal("recvBlk is nil")
	}

	md.l.AddStateBlock(sendBlk)

	recvPara := &ContractRewardBlockPara{
		SendHash: sendBlk.GetHash(),
	}
	_, err = api.GenerateRewardBlock(recvPara)
	if err == nil {
		t.Fatal("DemoKV not support recv")
	}
}

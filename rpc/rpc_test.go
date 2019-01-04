package rpc

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/test/mock"
)

func TestRPC_RPC(t *testing.T) {
	//cfg := RPCConfig{
	//	RPCEnabled: true,
	//	WSEnabled:  true,
	//	IPCEnabled: true,
	//}
	//r := NewRPC(cfg)
	//r.StartRPC()
	//defer r.stopInProcess()
	//defer r.stopHTTP()
	//logger.Info("rpc started")

	client, err := Dial("http://127.0.0.1:29735")
	//client, err := Dial("ws://127.0.0.1:29734")
	//client, err := Dial("29736")
	if err != nil {
		logger.Info(err)
	}

	var resp string
	err = client.Call(&resp, "qlcclassic_testAdd", 1100, "me")
	if err != nil {
		logger.Info(err)
	}
	logger.Info(resp)

	var resp2 interface{}
	err = client.Call(&resp2, "qlcclassic_testGetBlock")
	if err != nil {
		logger.Info(err)
	}
	logger.Info(resp2)

	var resp3 types.StateBlock
	err = client.Call(&resp3, "qlcclassic_testGetBlock")
	if err != nil {
		logger.Info(err)
	}
	logger.Info(resp3)

	var resp6 interface{}
	err = client.Call(&resp6, "qlcclassic_testAddBlock", mock.StateBlock())
	if err != nil {
		logger.Info(err)
	}
	logger.Info(resp6)

	var resp4 interface{}
	address := mock.Address()

	err = client.Call(&resp4, "qlcclassic_testAccounts", []string{address.String(), address.String()})
	if err != nil {
		logger.Info(err)
	}
	logger.Info(resp4)

	//l := ledger.NewLedger(filepath.Join(config.QlcTestDataDir(), "ledger"))
	//block := mock.StateBlock()
	//if err := l.AddBlock(block); err != nil {
	//	logger.Warn(err)
	//}
	//var resp types.StateBlock
	//err = client.Call(&resp, "ledger_getBlockByHash", block.GetHash())
	//if err != nil {
	//	logger.Info(err)
	//}
	//logger.Info(resp)
	//logger.Info(resp.GetHash())
}

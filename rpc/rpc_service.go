package rpc

import (
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
)

type RPCService struct {
	rpc *RPC
}

func NewRPCService(cfg *config.Config, dpos *consensus.DposService) *RPCService {
	return &RPCService{NewRPC(cfg, dpos)}
}

func (rs *RPCService) Init() error {
	logger.Info("rpc service init")
	/*
		addr1, _ := types.HexToAddress("qlc_3nihnp4a5zf5iq9pz54twp1dmksxnouc4i5k4y6f8gbnkc41p1b5ewm3inpw")
		addr2, _ := types.HexToAddress("qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic")
		ac1 := mock.AccountMeta(addr1)
		err := rs.rpc.ledger.AddOrUpdateAccountMeta(ac1)
		ac2 := mock.AccountMeta(addr2)
		err = rs.rpc.ledger.AddOrUpdateAccountMeta(ac2)

		b1 := mock.StateBlock()
		if err := rs.rpc.ledger.AddBlock(b1); err != nil {
			logger.Fatal(err)
		}
		b2 := mock.StateBlock()
		if rs.rpc.ledger.AddBlock(b2); err != nil {
			logger.Fatal(err)
		}
		fmt.Println(b1.GetHash(), b2.GetHash())
		b3 := mock.StateBlock()
		rs.rpc.ledger.AddBlock(b3)

		//accounts pending
		pendingkey := types.PendingKey{
			Address: addr1,
			Hash:    b2.GetHash(),
		}
		sb2 := b2.(*types.StateBlock)
		pendinginfo := types.PendingInfo{
			Source: addr2,
			Type:   sb2.GetToken(),
			Amount: types.StringToBalance("1998888"),
		}
		rs.rpc.ledger.AddPending(pendingkey, &pendinginfo)

		pendingkey2 := types.PendingKey{
			Address: addr2,
			Hash:    b3.GetHash(),
		}
		sb3 := b3.(*types.StateBlock)
		pendinginfo2 := types.PendingInfo{
			Source: addr1,
			Type:   sb3.GetToken(),
			Amount: types.StringToBalance("8888"),
		}
		rs.rpc.ledger.AddPending(pendingkey2, &pendinginfo2)

		blocks, err := mock.BlockChain()
		if err != nil {
			logger.Fatal(err)
		}
		for i, b := range blocks {
			fmt.Println(i, " ", b.GetAddress())
			if err := rs.rpc.ledger.BlockProcess(b); err != nil {
				logger.Fatal(err)
			}
		}

		if err != nil {
			return err
		}
	*/
	return nil
}

func (rs *RPCService) Start() error {
	err := rs.rpc.StartRPC()
	if err != nil {
		return err
	}
	return nil
}

func (rs *RPCService) Stop() error {
	logger.Info("rpc stopping...")
	rs.rpc.StopRPC()
	return nil
}

func (rs *RPCService) Status() int32 {
	panic("implement me")
}

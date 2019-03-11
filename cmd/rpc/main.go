/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/ledger/process"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/test/mock"
)

var logger = log.NewLogger("main")

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	switch os.Args[1] {

	case "rpc":
		dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
		initData(dir)
		cm := config.NewCfgManager(dir)
		cfg, err := cm.Load()
		if cfg.RPC.Enable == false {
			return
		}

		dp := &consensus.DposService{}
		rs := rpc.NewRPCService(cfg, dp)
		err = rs.Init()
		if err != nil {
			logger.Fatal(err)
		}
		err = rs.Start()
		if err != nil {
			logger.Fatal(err)
		}
		defer func() {
			rs.Stop()
			os.RemoveAll(dir)
		}()
		logger.Info("rpc started")
		s := <-c
		fmt.Println("Got signal: ", s)
	}
}

func initData(p string) {
	dir := filepath.Join(p, "ledger")
	ledger := ledger.NewLedger(dir)
	defer ledger.Close()
	verifier := process.NewLedgerVerifier(ledger)

	// accountsFrontiers / accountInfo
	var am1 types.AccountMeta
	addr1, _ := types.HexToAddress("qlc_3nihnp4a5zf5iq9pz54twp1dmksxnouc4i5k4y6f8gbnkc41p1b5ewm3inpw")
	am1.Address = addr1
	t1 := mock.TokenMeta(addr1)
	t1.Type = common.QLCChainToken
	// delegators
	t1.Representative, _ = types.HexToAddress("qlc_3pu4ggyg36nienoa9s9x95a615m1natqcqe7bcrn3t3ckq1srnnkh8q5xst5")
	t1.Balance = types.Balance{Int: big.NewInt(int64(110000000000000000))}
	// accountBlocksCount
	t1.BlockCount = 12
	am1.Tokens = append(am1.Tokens, t1)
	t2 := mock.TokenMeta(addr1)
	t2.Type = mock.SmartContractBlock().GetHash()
	am1.Tokens = append(am1.Tokens, t2)
	ledger.AddAccountMeta(&am1)

	var am2 types.AccountMeta
	addr2, _ := types.HexToAddress("qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic")
	am2.Address = addr2
	t3 := mock.TokenMeta(addr2)
	t3.Type = common.QLCChainToken
	t3.Balance = types.Balance{Int: big.NewInt(int64(2000000000000000))}
	t3.Representative, _ = types.HexToAddress("qlc_3pu4ggyg36nienoa9s9x95a615m1natqcqe7bcrn3t3ckq1srnnkh8q5xst5")

	am2.Tokens = append(am2.Tokens, t3)
	ledger.AddAccountMeta(&am2)
	fmt.Println("am1", am1, *am1.Tokens[0], *am1.Tokens[1])
	fmt.Println("am2", am2, *am2.Tokens[0])

	// accountHistoryTopn
	blocks, _ := mock.BlockChain()
	verifier.BlockProcess(blocks[0])
	for i, b := range blocks[1:] {
		fmt.Println(i)
		verifier.Process(b)
	}
	fmt.Println("accountHistoryTopn, ", blocks[0].GetAddress())

	//accountbalance    accountpending
	pendingkey := types.PendingKey{
		Address: addr1,
		Hash:    blocks[0].GetHash(),
	}
	pendinginfo := types.PendingInfo{
		Source: addr2,
		Type:   blocks[0].GetToken(),
		Amount: types.StringToBalance("2100000000000000000"),
	}
	ledger.AddPending(pendingkey, &pendinginfo)

	pendingkey2 := types.PendingKey{
		Address: addr1,
		Hash:    blocks[1].GetHash(),
	}
	pendinginfo2 := types.PendingInfo{
		Source: addr2,
		Type:   blocks[1].GetToken(),
		Amount: types.StringToBalance("80000000000000"),
	}
	ledger.AddPending(pendingkey2, &pendinginfo2)

	//blockAccount
	sb := types.StateBlock{
		Type:    types.Send,
		Address: addr1,
		Token:   common.QLCChainToken,
	}
	ledger.AddStateBlock(&sb)

	//accountVotingWeight
	ledger.AddRepresentation(addr1, types.Balance{Int: big.NewInt(int64(120000000000000))})
	ledger.AddRepresentation(addr2, types.Balance{Int: big.NewInt(int64(100000000000))})

	// unchecked
	ledger.AddUncheckedBlock(mock.Hash(), mock.StateBlock(), types.UncheckedKindLink, types.UnSynchronized)
	ledger.AddUncheckedBlock(mock.Hash(), mock.StateBlock(), types.UncheckedKindPrevious, types.UnSynchronized)

	for i := 0; i < 10; i++ {
		s := mock.SmartContractBlock()
		ledger.AddSmartContractBlock(*s)
	}

	// change block
	addr5, _ := types.HexToAddress("qlc_3c6ezoskbkgajq8f89ntcu75fdpcsokscgp9q5cdadndg1ju85fief7rrt11")

	sb3 := types.StateBlock{
		Type:    types.Change,
		Address: addr5,
		Token:   common.QLCChainToken,
	}
	ledger.AddStateBlock(&sb3)
	fmt.Println("hash,", sb.GetHash())

	// generate block
	var am5 types.AccountMeta
	am5.Address = addr5
	t5 := mock.TokenMeta(addr5)
	t5.Type = common.QLCChainToken
	t5.Header = sb3.GetHash()
	t5.Balance = types.Balance{Int: big.NewInt(int64(120000000000000))}
	t5.Representative, _ = types.HexToAddress("qlc_3pu4ggyg36nienoa9s9x95a615m1natqcqe7bcrn3t3ckq1srnnkh8q5xst5")
	am5.Tokens = append(am5.Tokens, t5)
	ledger.AddAccountMeta(&am5)
	ledger.AddRepresentation(t5.Representative, types.Balance{Int: big.NewInt(int64(200000000000))})

	//wallet
	// seed : 3197189ef9ef28f2496a24a18f740820915bf3fd7076a46513301c52b3d3b59d
	// address : qlc_3p1mnf5w3opm6sf4f9m7faeamks6cdeemx7p63tp4c9z456emzhhb1n9srco
	// public :  d813a347c0d6d3265a269e656a1889cb2452d8c9f4b620756128ff10c8c9fdef
	addr6, _ := types.HexToAddress("qlc_3p1mnf5w3opm6sf4f9m7faeamks6cdeemx7p63tp4c9z456emzhhb1n9srco")
	var am6 types.AccountMeta
	am6.Address = addr6
	t6 := mock.TokenMeta(addr6)
	t6.Type = common.QLCChainToken
	t6.Balance = types.Balance{Int: big.NewInt(int64(600000000000000))}
	am6.Tokens = append(am6.Tokens, t6)

	t7 := mock.TokenMeta(addr6)
	t7.Type = common.QLCChainToken
	t7.Balance = types.Balance{Int: big.NewInt(int64(90000000000000))}
	am6.Tokens = append(am6.Tokens, t7)

	t8 := mock.TokenMeta(addr6)
	t8.Type = mock.SmartContractBlock().GetHash()
	t8.Balance = types.Balance{Int: big.NewInt(int64(10000000000))}
	am6.Tokens = append(am6.Tokens, t8)
	ledger.AddAccountMeta(&am6)

	// sender or receiver
	p1 := &types.StateBlock{
		Address:  mock.Address(),
		Token:    common.QLCChainToken,
		Sender:   "1801111111",
		Receiver: "",
	}
	p2 := &types.StateBlock{
		Address:  mock.Address(),
		Token:    common.QLCChainToken,
		Sender:   "1801111111",
		Receiver: "18000000000",
	}
	p3 := &types.StateBlock{
		Address:  mock.Address(),
		Token:    common.QLCChainToken,
		Sender:   "1801111111",
		Receiver: "",
	}
	if err := ledger.AddStateBlock(p1); err != nil {
		fmt.Errorf("err block, %s", p1)
	}
	if err := ledger.AddStateBlock(p2); err != nil {
		fmt.Errorf("err block, %s", p2)
	}
	if err := ledger.AddStateBlock(p3); err != nil {
		fmt.Errorf("err block, %s", p3)
	}
}

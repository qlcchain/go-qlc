/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/qlcchain/go-qlc/cmd/util"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/abiosoft/ishell"

	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
)

func addPledgeRecvPendCmdByShell(parentCmd *ishell.Cmd) {
	priKey := util.Flag{
		Name:  "priKey",
		Must:  true,
		Usage: "account private hex string",
	}
	hash := util.Flag{
		Name:  "hash",
		Must:  true,
		Usage: "send block hash hex string",
	}
	method := util.Flag{
		Name:  "method",
		Must:  true,
		Usage: "method type, etc pledge/withdraw",
	}

	s := &ishell.Cmd{
		Name: "recvPending",
		Help: "pledge receive pending",
		Func: func(c *ishell.Context) {
			args := []util.Flag{priKey, hash, method}
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			priKeyP := util.StringVar(c.Args, priKey)
			hashP := util.StringVar(c.Args, hash)
			methodP := util.StringVar(c.Args, method)

			if err := pledgeRecvPendAction(priKeyP, hashP, methodP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(s)
}

func pledgeRecvPendAction(priKeyP, hashP, method string) error {
	pBytes, err := hex.DecodeString(priKeyP)
	if err != nil {
		return err
	}
	account := types.NewAccount(pBytes)

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	reward := types.StateBlock{}

	if method == "pledge" {
		err = client.Call(&reward, "pledge_getPledgeRewardBlockBySendHash", hashP)
	} else if method == "withdraw" {
		err = client.Call(&reward, "pledge_getWithdrawRewardBlockBySendHash", hashP)
	} else {
		return errors.New("invalid method value")
	}
	if err != nil {
		return err
	}

	rewardHash := reward.GetHash()
	reward.Signature = account.Sign(rewardHash)
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	fmt.Printf("RewardBlock:\n%s\n", cutil.ToIndentString(reward))
	fmt.Println("address", reward.Address, "rewardHash", rewardHash)

	isRecvOK := false
	for try := 0; try < 3; try++ {
		err = client.Call(nil, "ledger_process", &reward)
		if err != nil {
			fmt.Printf("reward block, try %d err %s\n", try, err)
			time.Sleep(1 * time.Second)
			continue
		}

		isRecvOK = true
		break
	}

	if !isRecvOK {
		fmt.Printf("failed to recv reward\n")
		return nil
	}

	fmt.Printf("success to recv reward, account vote %s\n", reward.GetVote())
	return nil
}

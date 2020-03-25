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

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPledgeWithdrawCmdByShell(parentCmd *ishell.Cmd) {
	beneficialAccount := util.Flag{
		Name:  "beneficialAccount",
		Must:  true,
		Usage: "beneficial account private hex string",
	}
	pledgeAccount := util.Flag{
		Name:  "pledgeAccount",
		Must:  true,
		Usage: "pledge account private hex string",
	}
	amount := util.Flag{
		Name:  "amount",
		Must:  true,
		Usage: "pledge amount",
	}
	pType := util.Flag{
		Name:  "pType",
		Must:  true,
		Usage: "pledge type",
	}
	nep5TxId := util.Flag{
		Name:  "nep5TxId",
		Must:  true,
		Usage: "NEP5 TX ID",
	}
	args := []util.Flag{beneficialAccount, pledgeAccount, amount, pType, nep5TxId}
	s := &ishell.Cmd{
		Name:                "withdraw",
		Help:                "withdraw token",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			beneficialAccountP := util.StringVar(c.Args, beneficialAccount)
			pledgeAccountP := util.StringVar(c.Args, pledgeAccount)
			amountP := util.StringVar(c.Args, amount)
			pTypeP := util.StringVar(c.Args, pType)
			nep5TxIdP := util.StringVar(c.Args, nep5TxId)

			if err := withdrawPledgeAction(beneficialAccountP, pledgeAccountP, amountP, pTypeP, nep5TxIdP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(s)
}

func addPledgeWithdrawCmdByCobra(parentCmd *cobra.Command) {
	var beneficialAccountP string
	var pledgeAccountP string
	var amountP string
	var pTypeP string
	var nep5TxIdP string

	var cmd = &cobra.Command{
		Use:   "withdraw",
		Short: "withdraw token",
		Run: func(cmd *cobra.Command, args []string) {
			err := withdrawPledgeAction(beneficialAccountP, pledgeAccountP, amountP, pTypeP, nep5TxIdP)
			if err != nil {
				cmd.Println(err)
			}
		},
	}
	cmd.Flags().StringVar(&pledgeAccountP, "pAccount", "", "pledge account private hex string")
	cmd.Flags().StringVar(&beneficialAccountP, "bAccount", "", "beneficial account private hex string")
	cmd.Flags().StringVar(&amountP, "amount", "", "pledge amount")
	cmd.Flags().StringVar(&pTypeP, "pledgeType", "", "pledge type")
	cmd.Flags().StringVar(&nep5TxIdP, "nep5TxId", "", "NEP5 TX ID")
	parentCmd.AddCommand(cmd)
}

func withdrawPledgeAction(beneficialAccount, pledgeAccount, amount, pType, NEP5TxId string) error {
	pBytes, err := hex.DecodeString(pledgeAccount)
	if err != nil {
		return err
	}
	p := types.NewAccount(pBytes)

	bBytes, err := hex.DecodeString(beneficialAccount)
	if err != nil {
		return err
	}
	b := types.NewAccount(bBytes)

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	am := types.StringToBalance(amount)

	withdrawPledgeParam := api.WithdrawPledgeParam{
		Beneficial: b.Address(), Amount: am, PType: pType, NEP5TxId: NEP5TxId}

	send := types.StateBlock{}
	err = client.Call(&send, "pledge_getWithdrawPledgeBlock", &withdrawPledgeParam)
	if err != nil {
		return err
	}
	sendHash := send.GetHash()
	send.Signature = b.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	fmt.Printf("sendHash:%s\n", sendHash)
	sendOk := false
	for try := 0; try < 3; try++ {
		err = client.Call(nil, "ledger_process", &send)
		if err != nil {
			fmt.Printf("send block, try %d err %s\n", try, err)
			time.Sleep(1 * time.Second)
			continue
		}

		sendOk = true
		break
	}
	if !sendOk {
		return errors.New("failed process send block")
	}

	time.Sleep(3 * time.Second)

	reward := types.StateBlock{}
	err = client.Call(&reward, "pledge_getWithdrawRewardBlockBySendHash", &sendHash)
	if err != nil {
		return err
	}

	rewardHash := reward.GetHash()
	reward.Signature = p.Sign(rewardHash)
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	fmt.Printf("rewardHash:%s\n", rewardHash)
	recvOk := false
	for try := 0; try < 3; try++ {
		err = client.Call(nil, "ledger_process", &reward)
		if err != nil {
			fmt.Printf("reward block, try %d err %s\n", try, err)
			time.Sleep(1 * time.Second)
			continue
		}

		recvOk = true
		break
	}
	if !recvOk {
		return errors.New("failed process recv block")
	}

	return nil
}

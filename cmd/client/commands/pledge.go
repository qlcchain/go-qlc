/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/rpc/api"
	"github.com/spf13/cobra"
)

func pledge() {
	var beneficialAccountP string
	var pledgeAccountP string
	var amountP string
	var pTypeP string

	if interactive {
		beneficialAccount := Flag{
			Name:  "beneficialAccount",
			Must:  true,
			Usage: "beneficial account private hex string",
		}
		pledgeAccount := Flag{
			Name:  "pledgeAccount",
			Must:  true,
			Usage: "pledge account private hex string",
		}
		amount := Flag{
			Name:  "amount",
			Must:  true,
			Usage: "pledge amount",
		}
		pType := Flag{
			Name:  "pType",
			Must:  true,
			Usage: "pledge type",
		}

		s := &ishell.Cmd{
			Name: "pledge",
			Help: "pledge token",
			Func: func(c *ishell.Context) {
				args := []Flag{beneficialAccount, pledgeAccount, amount, pType}
				if HelpText(c, args) {
					return
				}
				err := CheckArgs(c, args)
				if err != nil {
					Warn(err)
					return
				}

				beneficialAccountP = StringVar(c.Args, beneficialAccount)
				pledgeAccountP = StringVar(c.Args, pledgeAccount)
				amountP = StringVar(c.Args, amount)
				pTypeP = StringVar(c.Args, pType)
				if err != nil {
					Warn(err)
					return
				}

				fmt.Println(beneficialAccountP, pledgeAccountP, amountP, pTypeP)
				if err := pledgeAction(beneficialAccountP, pledgeAccountP, amountP, pTypeP); err != nil {
					Warn(err)
					return
				}
			},
		}
		shell.AddCmd(s)
	} else {
		var accountCmd = &cobra.Command{
			Use:   "pledge",
			Short: "pledge token",
			Run: func(cmd *cobra.Command, args []string) {
				err := pledgeAction(beneficialAccountP, pledgeAccountP, amountP, pTypeP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		accountCmd.Flags().StringVar(&pledgeAccountP, "pAccount", "", "pledge account private hex string")
		accountCmd.Flags().StringVar(&beneficialAccountP, "bAccount", "", "beneficial account private hex string")
		accountCmd.Flags().StringVar(&amountP, "amount", "", "pledge amount")
		accountCmd.Flags().StringVar(&pTypeP, "pledgeType", "", "pledge type")
		rootCmd.AddCommand(accountCmd)
	}
}

func pledgeAction(beneficialAccount, pledgeAccount, amount, pType string) error {
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

	pledgeParam := api.PledgeParam{
		Beneficial: b.Address(), PledgeAddress: p.Address(), Amount: am,
		PType: pType,
	}

	send := types.StateBlock{}
	err = client.Call(&send, "pledge_getPledgeBlock", &pledgeParam)
	if err != nil {
		return err
	}
	send.Timestamp = time.Now().UTC().Unix()
	sendHash := send.GetHash()
	send.Signature = p.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	err = client.Call(nil, "ledger_process", &send)
	if err != nil {
		return err
	}

	reward := types.StateBlock{}
	err = client.Call(&reward, "pledge_getPledgeRewardBlock", &send)

	if err != nil {
		return err
	}
	reward.Timestamp = time.Now().UTC().Unix()
	reward.Signature = b.Sign(reward.GetHash())
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	err = client.Call(nil, "ledger_process", &reward)
	if err != nil {
		return err
	}
	return nil
}

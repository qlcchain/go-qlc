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

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/rpc/api"
	"github.com/spf13/cobra"
)

func withdrawMintage() {
	var accountP string
	var tokenIdP string

	if interactive {
		account := util.Flag{
			Name:  "account",
			Must:  true,
			Usage: "account private hex string",
		}
		tokenId := util.Flag{
			Name:  "tokenId",
			Must:  true,
			Usage: "token id hash hex string",
		}

		s := &ishell.Cmd{
			Name: "withdrawMine",
			Help: "withdraw mine token",
			Func: func(c *ishell.Context) {
				args := []util.Flag{account, tokenId}
				if util.HelpText(c, args) {
					return
				}
				err := util.CheckArgs(c, args)
				if err != nil {
					util.Warn(err)
					return
				}

				accountP = util.StringVar(c.Args, account)
				tokenIdP = util.StringVar(c.Args, tokenId)

				fmt.Println(accountP, tokenIdP)
				if err := withdrawMintageAction(accountP, tokenIdP); err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(s)
	} else {
		var accountCmd = &cobra.Command{
			Use:   "withdrawMine",
			Short: "withdraw mine token",
			Run: func(cmd *cobra.Command, args []string) {
				err := withdrawMintageAction(accountP, tokenIdP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		accountCmd.Flags().StringVar(&accountP, "account", "", "account private hex string")
		accountCmd.Flags().StringVar(&tokenIdP, "tokenId", "", "token id hash hex string")
		rootCmd.AddCommand(accountCmd)
	}
}

func withdrawMintageAction(account, tokenId string) error {
	bytes, err := hex.DecodeString(account)
	if err != nil {
		return err
	}
	a := types.NewAccount(bytes)

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	id, err := types.NewHash(tokenId)
	if err != nil {
		return err
	}

	withdrawMintageParam := api.WithdrawParams{
		SelfAddr: a.Address(), TokenId: id}

	send := types.StateBlock{}
	err = client.Call(&send, "mintage_getWithdrawMintageBlock", &withdrawMintageParam)
	if err != nil {
		return err
	}

	sendHash := send.GetHash()
	send.Signature = a.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	reward := types.StateBlock{}
	err = client.Call(&reward, "mintage_getWithdrawRewardBlock", &send)

	if err != nil {
		return err
	}

	//reward.Address = a.Address()
	//reward.Representative = a.Address()
	//reward.Link = sendHash
	reward.Signature = a.Sign(reward.GetHash())
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	err = client.Call(nil, "ledger_process", &send)
	if err != nil {
		return err
	}

	err = client.Call(nil, "ledger_process", &reward)
	if err != nil {
		return err
	}
	return nil
}

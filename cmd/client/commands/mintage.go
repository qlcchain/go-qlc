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

func mintage() {
	var accountP string
	var preHashP string
	var tokenNameP string
	var tokenSymbolP string
	var totalSupplyP string
	var decimalsP int

	if interactive {
		account := util.Flag{
			Name:  "account",
			Must:  true,
			Usage: "account private hex string",
		}
		preHash := util.Flag{
			Name:  "preHash",
			Must:  true,
			Usage: "account previous hash hex string",
		}
		tokenName := util.Flag{
			Name:  "tokenName",
			Must:  true,
			Usage: "token name",
		}
		tokenSymbol := util.Flag{
			Name:  "tokenSymbol",
			Must:  true,
			Usage: "token symbol",
		}
		totalSupply := util.Flag{
			Name:  "totalSupply",
			Must:  true,
			Usage: "token total supply",
		}
		decimals := util.Flag{
			Name:  "decimals",
			Must:  true,
			Usage: "token decimals",
		}

		s := &ishell.Cmd{
			Name: "mine",
			Help: "mine token",
			Func: func(c *ishell.Context) {
				args := []util.Flag{account, preHash, tokenName, tokenSymbol, totalSupply, decimals}
				if util.HelpText(c, args) {
					return
				}
				err := util.CheckArgs(c, args)
				if err != nil {
					util.Warn(err)
					return
				}

				accountP = util.StringVar(c.Args, account)
				preHashP = util.StringVar(c.Args, preHash)
				tokenNameP = util.StringVar(c.Args, tokenName)
				tokenSymbolP = util.StringVar(c.Args, tokenSymbol)
				totalSupplyP = util.StringVar(c.Args, totalSupply)
				decimalsP, err = util.IntVar(c.Args, decimals)
				if err != nil {
					util.Warn(err)
					return
				}

				fmt.Println(accountP, preHashP, tokenNameP, tokenSymbolP, totalSupplyP, decimalsP)
				if err := mintageAction(accountP, preHashP, tokenNameP, tokenSymbolP, totalSupplyP, decimalsP); err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(s)
	} else {
		var accountCmd = &cobra.Command{
			Use:   "mine",
			Short: "mine token",
			Run: func(cmd *cobra.Command, args []string) {
				err := mintageAction(accountP, preHashP, tokenNameP, tokenSymbolP, totalSupplyP, decimalsP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		accountCmd.Flags().StringVar(&accountP, "account", "", "account private hex string")
		accountCmd.Flags().StringVar(&preHashP, "preHash", "", "account previous hash hex string")
		accountCmd.Flags().StringVar(&tokenNameP, "tokenName", "", "token name")
		accountCmd.Flags().StringVar(&tokenSymbolP, "tokenSymbol", "", "token symbol")
		accountCmd.Flags().StringVar(&totalSupplyP, "totalSupply", "", "token total supply")
		accountCmd.Flags().IntVar(&decimalsP, "decimals", 8, "token decimals")
		rootCmd.AddCommand(accountCmd)
	}
}

func mintageAction(account, preHash, tokenName, tokenSymbol, totalSupply string, decimals int) error {
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

	previous, err := types.NewHash(preHash)
	if err != nil {
		return err
	}

	d := uint8(decimals)
	NEP5tTxId := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	mintageParam := api.MintageParams{
		SelfAddr: a.Address(), PrevHash: previous, TokenName: tokenName,
		TotalSupply: totalSupply, TokenSymbol: tokenSymbol, Decimals: d, Beneficial: a.Address(),
		NEP5TxId: NEP5tTxId,
	}

	send := types.StateBlock{}
	err = client.Call(&send, "mintage_getMintageBlock", &mintageParam)
	if err != nil {
		return err
	}

	sendHash := send.GetHash()
	send.Signature = a.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	err = client.Call(nil, "ledger_process", &send)
	if err != nil {
		return err
	}

	reward := types.StateBlock{}
	err = client.Call(&reward, "mintage_getRewardBlock", &send)

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

	err = client.Call(nil, "ledger_process", &reward)
	if err != nil {
		return err
	}
	return nil
}

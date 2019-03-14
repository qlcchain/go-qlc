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
	var pledgeAmountP int
	if interactive {
		account := Flag{
			Name:  "account",
			Must:  true,
			Usage: "account private hex string",
		}
		preHash := Flag{
			Name:  "preHash",
			Must:  true,
			Usage: "account previous hash hex string",
		}
		tokenName := Flag{
			Name:  "tokenName",
			Must:  true,
			Usage: "token name",
		}
		tokenSymbol := Flag{
			Name:  "tokenSymbol",
			Must:  true,
			Usage: "token symbol",
		}
		totalSupply := Flag{
			Name:  "totalSupply",
			Must:  true,
			Usage: "token total supply",
		}
		decimals := Flag{
			Name:  "decimals",
			Must:  true,
			Usage: "token decimals",
		}
		pledgeAmount := Flag{
			Name:  "pledgeAmount",
			Must:  true,
			Usage: "token decimals",
		}

		s := &ishell.Cmd{
			Name: "mine",
			Help: "mine token",
			Func: func(c *ishell.Context) {
				args := []Flag{account, preHash, tokenName, tokenSymbol, totalSupply, decimals, pledgeAmount}
				if HelpText(c, args) {
					return
				}
				err := CheckArgs(c, args)
				if err != nil {
					Warn(err)
					return
				}

				accountP = StringVar(c.Args, account)
				preHashP = StringVar(c.Args, preHash)
				tokenNameP = StringVar(c.Args, tokenName)
				tokenSymbolP = StringVar(c.Args, tokenSymbol)
				totalSupplyP = StringVar(c.Args, totalSupply)
				decimalsP, err = IntVar(c.Args, decimals)
				if err != nil {
					Warn(err)
					return
				}
				pledgeAmountP, err = IntVar(c.Args, pledgeAmount)
				if err != nil {
					Warn(err)
					return
				}
				fmt.Println(accountP, preHashP, tokenNameP, tokenSymbolP, totalSupplyP, decimalsP, pledgeAmountP)
				if err := mintageAction(accountP, preHashP, tokenNameP, tokenSymbolP, totalSupplyP, decimalsP, pledgeAmountP); err != nil {
					Warn(err)
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
				err := mintageAction(accountP, preHashP, tokenNameP, tokenSymbolP, totalSupplyP, decimalsP, pledgeAmountP)
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
		accountCmd.Flags().IntVar(&pledgeAmountP, "pledgeAmount", 100, "pledge Amount")
		rootCmd.AddCommand(accountCmd)
	}
}

func mintageAction(account, preHash, tokenName, tokenSymbol, totalSupply string, decimals int, pledgeAmount int) error {
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

	mintageParam := api.MintageParams{
		SelfAddr: a.Address(), PrevHash: previous, TokenName: tokenName,
		TotalSupply: totalSupply, TokenSymbol: tokenSymbol, Decimals: d,
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

	reward.Address = a.Address()
	reward.Representative = a.Address()
	reward.Link = sendHash
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

func buildBatchElem(method string, result interface{}, args ...interface{}) rpc.BatchElem {
	var batchArgs []interface{}
	batchArgs = append(batchArgs, args...)
	var err error
	return rpc.BatchElem{Method: method, Args: batchArgs, Result: result, Error: err}
}

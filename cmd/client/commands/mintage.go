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
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/rpc/api"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
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
	var data []byte
	previous, err := types.NewHash(preHash)
	if err != nil {
		return err
	}
	d := uint8(decimals)

	mintageParam := api.MintageParams{
		SelfAddr: a.Address(), PrevHash: previous, TokenName: tokenName,
		TotalSupply: totalSupply, TokenSymbol: tokenSymbol, Decimals: d,
	}
	err = client.Call(&data, "mintage_getMintageData", &mintageParam)
	if err != nil {
		return err
	}

	//generate send contract, genesis block and broadcast to network
	token, err := cabi.ParseTokenInfo(data)
	if err != nil {
		return err
	}

	var am api.APIAccount
	err = client.Call(&am, "ledger_accountInfo", a.Address())
	if err != nil {
		return err
	}

	tm := new(api.APITokenMeta)
	for _, t := range am.Tokens {
		if t.TokenMeta.Type == common.QLCChainToken {
			tm = t
		}
	}
	if tm.TokenMeta.Type.IsZero() {
		return fmt.Errorf("account does not have chain token")
	}
	send := types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        a.Address(),
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Link:           types.Hash(types.MintageAddress),
		Representative: tm.Representative,
		Data:           data,
	}
	send.Signature = a.Sign(send.GetHash())
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()
	err = client.Call(nil, "ledger_process", &send)
	if err != nil {
		return err
	}

	reward := types.StateBlock{
		Type:           types.ContractReward,
		Token:          token.TokenId,
		Address:        a.Address(),
		Balance:        types.Balance{token.TotalSupply},
		Previous:       types.ZeroHash,
		Link:           send.GetHash(),
		Representative: a.Address(),
	}
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

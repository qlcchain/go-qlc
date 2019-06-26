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

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func generateTestLedger() {
	var fromAccountP string
	var repCountsP int
	var toAccountsP []string
	if interactive {
		from := util.Flag{
			Name:  "from",
			Must:  true,
			Usage: "send account private key",
			Value: "",
		}
		repCounts := util.Flag{
			Name:  "repCounts",
			Must:  true,
			Usage: "rep counts",
			Value: "",
		}
		to := util.Flag{
			Name:  "to",
			Must:  false,
			Usage: "receive accounts",
			Value: []string{},
		}
		c := &ishell.Cmd{
			Name: "generateTestLedger",
			Help: "generate test ledger",
			Func: func(c *ishell.Context) {
				args := []util.Flag{from, repCounts, to}
				if util.HelpText(c, args) {
					return
				}
				if err := util.CheckArgs(c, args); err != nil {
					util.Warn(err)
					return
				}
				fromAccountP = util.StringVar(c.Args, from)
				repCountsP, _ = util.IntVar(c.Args, repCounts)
				toAccountsP = util.StringSliceVar(c.Args, to)
				if len(toAccountsP) != 0 {
					if (repCountsP - 1) != len(toAccountsP) {
						util.Info("account count is not match rep count,pls check")
						return
					}
				}
				err := generateLedger(fromAccountP, repCountsP, toAccountsP)
				if err != nil {
					util.Info(err)
					return
				}

				util.Info("generate test ledger success")
			},
		}
		shell.AddCmd(c)
	} else {
		var generateTestLedgerCmd = &cobra.Command{
			Use:   "generateTestLedger",
			Short: "generate test ledger",
			Run: func(cmd *cobra.Command, args []string) {
				err := generateLedger(fromAccountP, repCountsP, toAccountsP)
				if err != nil {
					cmd.Println(err)
					return
				}
				cmd.Println("generate test ledger success")
			},
		}
		generateTestLedgerCmd.Flags().StringVar(&fromAccountP, "from", "", "send account private key")
		generateTestLedgerCmd.Flags().IntVar(&repCountsP, "repCounts", 1, "rep counts")
		generateTestLedgerCmd.Flags().StringSliceVar(&toAccountsP, "to", []string{}, "receive accounts")
		rootCmd.AddCommand(generateTestLedgerCmd)
	}
}

func generateLedger(fromAccountP string, repCountsP int, toAccountsP []string) error {
	bytes, err := hex.DecodeString(fromAccountP)
	if err != nil {
		return err
	}
	fromAccount := types.NewAccount(bytes)
	err = sendReceiveAndChangeAction(repCountsP, fromAccount, toAccountsP)
	if err != nil {
		return err
	}
	return nil
}

func sendReceiveAndChangeAction(repCountsP int, from *types.Account, toAccountsP []string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()
	var accountInfo *api.APIAccount
	addr := from.Address()
	err = client.Call(&accountInfo, "ledger_accountInfo", addr)
	if err != nil {
		return err
	}
	var totalBalance *types.Balance
	if accountInfo != nil {
		totalBalance = accountInfo.CoinBalance
		amount, err := totalBalance.Div(int64(repCountsP))
		if err != nil {
			return err
		}
		fmt.Printf("totalBalance is [%s],amount is [%s]\n", totalBalance, amount)
		for i := 0; i < (repCountsP - 1); i++ {
			var para api.APISendBlockPara
			var to *types.Account
			if len(toAccountsP) == 0 {
				seed, err := types.NewSeed()
				if err == nil {
					to, err = seed.Account(0)
					if err == nil {
						fmt.Println("Seed:", seed.String())
						fmt.Println("Address:", to.Address())
						fmt.Println("Private:", hex.EncodeToString(to.PrivateKey()))
						para = api.APISendBlockPara{
							From:      from.Address(),
							TokenName: "QLC",
							To:        to.Address(),
							Amount:    amount,
						}
					} else {
						return err
					}
				} else {
					return err
				}
			} else {
				bytes, err := hex.DecodeString(toAccountsP[i])
				if err != nil {
					return err
				}
				to = types.NewAccount(bytes)
				para = api.APISendBlockPara{
					From:      from.Address(),
					TokenName: "QLC",
					To:        to.Address(),
					Amount:    amount,
				}
			}
			var sendBlock types.StateBlock
			err = client.Call(&sendBlock, "ledger_generateSendBlock", &para, hex.EncodeToString(from.PrivateKey()))
			if err != nil {
				return err
			}
			var h types.Hash
			err = client.Call(&h, "ledger_process", &sendBlock)
			if err != nil {
				return err
			}
			var receiveBlock types.StateBlock
			err = client.Call(&receiveBlock, "ledger_generateReceiveBlock", &sendBlock, hex.EncodeToString(to.PrivateKey()))
			if err != nil {
				return err
			}

			var h1 types.Hash
			err = client.Call(&h1, "ledger_process", &receiveBlock)
			if err != nil {
				return err
			}
			var changeBlock types.StateBlock
			err = client.Call(&changeBlock, "ledger_generateChangeBlock", to.Address(), to.Address(), hex.EncodeToString(to.PrivateKey()))
			if err != nil {
				return err
			}
			var h2 types.Hash
			err = client.Call(&h2, "ledger_process", &changeBlock)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

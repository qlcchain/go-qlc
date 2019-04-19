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

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/rpc/api"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

var (
	testGenesisAddress = "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4"
)

func generateTestLedger() {
	var fromAccountP string
	var repCountsP int
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
		c := &ishell.Cmd{
			Name: "generateTestLedger",
			Help: "generate test ledger",
			Func: func(c *ishell.Context) {
				args := []util.Flag{from, repCounts}
				if util.HelpText(c, args) {
					return
				}
				if err := util.CheckArgs(c, args); err != nil {
					util.Warn(err)
					return
				}
				fromAccountP = util.StringVar(c.Args, from)
				repCountsP, _ = util.IntVar(c.Args, repCounts)
				err := generateLedger(fromAccountP, repCountsP)
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
				err := generateLedger(fromAccountP, repCountsP)
				if err != nil {
					cmd.Println(err)
					return
				}
				cmd.Println("generate test ledger success")
			},
		}
		generateTestLedgerCmd.Flags().StringVar(&fromAccountP, "from", "", "send account private key")
		generateTestLedgerCmd.Flags().IntVar(&repCountsP, "repCounts", 1, "rep counts")
		rootCmd.AddCommand(generateTestLedgerCmd)
	}
}

func generateLedger(fromAccountP string, repCountsP int) error {
	bytes, err := hex.DecodeString(fromAccountP)
	if err != nil {
		return err
	}
	fromAccount := types.NewAccount(bytes)
	err = sendReceiveAndChangeAction(repCountsP, fromAccount)
	if err != nil {
		return err
	}
	return nil
}

func sendReceiveAndChangeAction(repCountsP int, from *types.Account) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()
	var accountInfo *api.APIAccount
	err = client.Call(&accountInfo, "ledger_accountInfo", testGenesisAddress)
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
			seed, err := types.NewSeed()
			if err == nil {
				if to, err := seed.Account(0); err == nil {
					fmt.Println("Seed:", seed.String())
					fmt.Println("Address:", to.Address())
					fmt.Println("Private:", hex.EncodeToString(to.PrivateKey()))
					para := api.APISendBlockPara{
						From:      from.Address(),
						TokenName: "QLC",
						To:        to.Address(),
						Amount:    amount,
					}
					var sendBlock types.StateBlock
					err = client.Call(&sendBlock, "ledger_generateSendBlock", para, hex.EncodeToString(from.PrivateKey()))
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
					//from := sendBlock.Address
					//s := fmt.Sprintf("Receive QLC from %s to %s （hash: %s）", from.String(), to.String(), receiveBlock.GetHash())
					//if interactive {
					//	Info(s)
					//} else {
					//	fmt.Println(s)
					//}

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
		}
	}
	return nil
}

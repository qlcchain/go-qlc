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
	"math/rand"
	"time"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/rpc/api"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func generateTestPair() {
	var fromAccountP string
	var toAccountsP []string
	var tpsP int
	var txCountP int

	if interactive {
		from := util.Flag{
			Name:  "from",
			Must:  true,
			Usage: "send account private key",
			Value: "",
		}
		toAccounts := util.Flag{
			Name:  "toAccounts",
			Must:  true,
			Usage: "to account private key",
			Value: "",
		}
		tps := util.Flag{
			Name:  "tps",
			Must:  true,
			Usage: "tx per sec",
			Value: 1,
		}
		txCount := util.Flag{
			Name:  "txCount",
			Must:  true,
			Usage: "tx count",
			Value: 1,
		}
		c := &ishell.Cmd{
			Name: "generateTestPair",
			Help: "generate test pair send txs",
			Func: func(c *ishell.Context) {
				args := []util.Flag{from, toAccounts, tps, txCount}
				if util.HelpText(c, args) {
					return
				}
				if err := util.CheckArgs(c, args); err != nil {
					util.Warn(err)
					return
				}
				fromAccountP = util.StringVar(c.Args, from)
				toAccountsP = util.StringSliceVar(c.Args, toAccounts)
				tpsP, _ = util.IntVar(c.Args, tps)
				txCountP, _ = util.IntVar(c.Args, txCount)
				err := randSendTxs(fromAccountP, toAccountsP, tpsP, txCountP)
				if err != nil {
					util.Info(err)
					return
				}
				util.Info("generate test send pair txs success")
			},
		}
		shell.AddCmd(c)
	} else {
		var randSendCmd = &cobra.Command{
			Use:   "generateTestPair",
			Short: "generate test pair send txs",
			Run: func(cmd *cobra.Command, args []string) {
				err := randSendTxs(fromAccountP, toAccountsP, tpsP, txCountP)
				if err != nil {
					cmd.Println(err)
					return
				}
				cmd.Println("generate test pair send txs success")
			},
		}
		randSendCmd.Flags().StringVar(&fromAccountP, "from", "", "send account private key")
		randSendCmd.Flags().StringSliceVar(&toAccountsP, "toAccounts", nil, "to account private key")
		randSendCmd.Flags().IntVar(&tpsP, "tps", 1, "tx per sec")
		randSendCmd.Flags().IntVar(&txCountP, "txCount", 1, "tx count")
		rootCmd.AddCommand(randSendCmd)
	}
}

func randSendTxs(fromAccountP string, toAccountsP []string, tpsP int, txCountP int) error {
	if fromAccountP == "" {
		return errors.New("invalid from account value")
	}
	if len(toAccountsP) <= 0 {
		return errors.New("invalid to account value")
	}
	if tpsP < 0 {
		return errors.New("invalid to tps value")
	}
	if txCountP < 0 {
		return errors.New("invalid to txCount value")
	}

	bytes, err := hex.DecodeString(fromAccountP)
	if err != nil {
		return err
	}
	fromAccount := types.NewAccount(bytes)

	var toAccounts []*types.Account
	for _, toAccountP := range toAccountsP {
		bytes, err := hex.DecodeString(toAccountP)
		if err != nil {
			return err
		}
		toAccounts = append(toAccounts, types.NewAccount(bytes))
	}

	err = transferBalanceToAccounts(fromAccount, toAccounts)
	if err != nil {
		return err
	}

	err = generateTxToAccounts(fromAccount, toAccounts, tpsP, txCountP)
	if err != nil {
		return err
	}
	return nil
}

func transferBalanceToAccounts(from *types.Account, toAccounts []*types.Account) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	var accountInfo *api.APIAccount
	err = client.Call(&accountInfo, "ledger_accountInfo", from.Address().String())
	if err != nil {
		return err
	}
	if accountInfo == nil {
		return nil
	}

	var totalBalance *types.Balance
	totalBalance = accountInfo.CoinBalance
	toCount := int64(len(toAccounts))
	amount, _ := totalBalance.Div(toCount)
	fmt.Printf("totalBalance is [%s], amount is [%s]\n", totalBalance, amount)
	if totalBalance.Compare(types.NewBalance(0)) != types.BalanceCompBigger {
		fmt.Printf("no need transfer balance to accounts\n")
		return nil
	}

	for _, toAcc := range toAccounts {
		fmt.Printf("transfer balance %s to %s\n", amount, toAcc.Address())
		para := api.APISendBlockPara{
			From:      from.Address(),
			TokenName: "QLC",
			To:        toAcc.Address(),
			Amount:    amount,
		}
		var sendBlock types.StateBlock
		err = client.Call(&sendBlock, "ledger_generateSendBlock", para, hex.EncodeToString(from.PrivateKey()))
		if err != nil {
			fmt.Println(err)
			continue
		}
		var sendHash types.Hash
		err = client.Call(&sendHash, "ledger_process", &sendBlock)
		if err != nil {
			fmt.Println(err)
			continue
		}

		var receiveBlock types.StateBlock
		err = client.Call(&receiveBlock, "ledger_generateReceiveBlock", &sendBlock, hex.EncodeToString(toAcc.PrivateKey()))
		if err != nil {
			fmt.Println(err)
			continue
		}
		var recvHash types.Hash
		err = client.Call(&recvHash, "ledger_process", &receiveBlock)
		if err != nil {
			fmt.Println(err)
			continue
		}

		var changeBlock types.StateBlock
		err = client.Call(&changeBlock, "ledger_generateChangeBlock", toAcc.Address(), toAcc.Address(), hex.EncodeToString(toAcc.PrivateKey()))
		if err != nil {
			fmt.Println(err)
			continue
		}
		var chgHash types.Hash
		err = client.Call(&chgHash, "ledger_process", &changeBlock)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	return nil
}

func generateTxToAccounts(from *types.Account, toAccounts []*types.Account, tpsP int, txCountP int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rand.Seed(time.Now().Unix())

	txCurNum := 0

	for txCurNum < txCountP {
		for _, fromAcc := range toAccounts {
			for _, toAcc := range toAccounts {
				if fromAcc == toAcc {
					continue
				}

				if err != nil {
					time.Sleep(time.Second)
					err = nil
				}

				amount := rand.Intn(1000) * 10e8
				if amount <= 0 {
					amount = 10e8
				}
				fmt.Printf("tx %d: fromAcc:%s, toAcc:%s, amount:%d\n", txCurNum, fromAcc.Address(), toAcc.Address(), amount)

				para := api.APISendBlockPara{
					From:      fromAcc.Address(),
					TokenName: "QLC",
					To:        toAcc.Address(),
					Amount:    types.NewBalance(int64(amount)),
				}

				var sendBlock types.StateBlock
				err = client.Call(&sendBlock, "ledger_generateSendBlock", para, hex.EncodeToString(fromAcc.PrivateKey()))
				if err != nil {
					fmt.Println(err)
					continue
				}
				var sendHash types.Hash
				err = client.Call(&sendHash, "ledger_process", &sendBlock)
				if err != nil {
					fmt.Println(err)
					continue
				}

				var receiveBlock types.StateBlock
				err = client.Call(&receiveBlock, "ledger_generateReceiveBlock", &sendBlock, hex.EncodeToString(toAcc.PrivateKey()))
				if err != nil {
					fmt.Println(err)
					continue
				}
				var recvHash types.Hash
				err = client.Call(&recvHash, "ledger_process", &receiveBlock)
				if err != nil {
					fmt.Println(err)
					continue
				}

				txCurNum++
				if txCurNum >= txCountP {
					return nil
				}

				if txCurNum%tpsP == 0 {
					time.Sleep(time.Second)
				}
			}
		}
	}

	return nil
}

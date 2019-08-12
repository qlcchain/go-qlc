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

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func generateTestPair() {
	var fromAccountP string
	var txAccountsP []string
	var onlyTxP bool
	var tpsP int
	var txCountP int
	var txTypeP string

	if interactive {
		from := util.Flag{
			Name:  "from",
			Must:  true,
			Usage: "send account private key",
			Value: "",
		}
		txAccounts := util.Flag{
			Name:  "txAccounts",
			Must:  true,
			Usage: "tx account private key",
			Value: "",
		}
		tps := util.Flag{
			Name:  "tps",
			Must:  true,
			Usage: "tx per sec",
			Value: 1,
		}
		onlyTx := util.Flag{
			Name:  "onlyTx",
			Must:  true,
			Usage: "only generate txs",
			Value: true,
		}
		txCount := util.Flag{
			Name:  "txCount",
			Must:  true,
			Usage: "tx count",
			Value: 1,
		}
		txType := util.Flag{
			Name:  "txType",
			Must:  true,
			Usage: "tx type, both/send",
			Value: "both",
		}
		c := &ishell.Cmd{
			Name: "generateTestPair",
			Help: "generate test pair send txs",
			Func: func(c *ishell.Context) {
				args := []util.Flag{from, txAccounts, onlyTx, tps, txCount, txType}
				if util.HelpText(c, args) {
					return
				}
				if err := util.CheckArgs(c, args); err != nil {
					util.Warn(err)
					return
				}
				fromAccountP = util.StringVar(c.Args, from)
				txAccountsP = util.StringSliceVar(c.Args, txAccounts)
				onlyTxP = util.BoolVar(c.Args, onlyTx)
				tpsP, _ = util.IntVar(c.Args, tps)
				txCountP, _ = util.IntVar(c.Args, txCount)
				txTypeP = util.StringVar(c.Args, txType)
				err := randSendTxs(fromAccountP, txAccountsP, onlyTxP, tpsP, txCountP, txTypeP)
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
				err := randSendTxs(fromAccountP, txAccountsP, onlyTxP, tpsP, txCountP, txTypeP)
				if err != nil {
					cmd.Println(err)
					return
				}
				cmd.Println("generate test pair send txs success")
			},
		}
		randSendCmd.Flags().StringVar(&fromAccountP, "from", "", "send account private key")
		randSendCmd.Flags().StringSliceVar(&txAccountsP, "txAccounts", nil, "tx account private key")
		randSendCmd.Flags().BoolVar(&onlyTxP, "onlyTx", true, "only generate txs")
		randSendCmd.Flags().IntVar(&tpsP, "tps", 1, "tx per sec")
		randSendCmd.Flags().IntVar(&txCountP, "txCount", 1, "tx count")
		randSendCmd.Flags().StringVar(&txTypeP, "txType", "both", "tx type, both/send")
		rootCmd.AddCommand(randSendCmd)
	}
}

func randSendTxs(fromAccountP string, txAccountsP []string, onlyTxP bool, tpsP int, txCountP int, txTypeP string) error {
	if !onlyTxP {
		if fromAccountP == "" {
			return errors.New("invalid from account value")
		}
	}

	if len(txAccountsP) <= 0 {
		return errors.New("invalid tx account value")
	}
	if tpsP < 0 {
		return errors.New("invalid tps value")
	}
	if txCountP < 0 {
		return errors.New("invalid txCount value")
	}

	var toAccounts []*types.Account
	for _, toAccountP := range txAccountsP {
		bytes, err := hex.DecodeString(toAccountP)
		if err != nil {
			return err
		}
		toAccounts = append(toAccounts, types.NewAccount(bytes))
	}

	if !onlyTxP {
		bytes, err := hex.DecodeString(fromAccountP)
		if err != nil {
			return err
		}
		fromAccount := types.NewAccount(bytes)

		err = transferBalanceToAccounts(fromAccount, toAccounts)
		if err != nil {
			return err
		}
	}

	err := generatePairTxs(toAccounts, tpsP, txCountP, txTypeP)
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

	var accountBalance *types.Balance
	accountBalance = accountInfo.CoinBalance
	fmt.Printf("accountBalance is [%s]\n", accountBalance)

	lockBalance := types.NewBalance(int64(50000000 * 1e8))
	if accountBalance.Compare(lockBalance) != types.BalanceCompBigger {
		fmt.Printf("no need transfer balance to accounts\n")
		return nil
	}
	totalBalance := *accountBalance
	freeBalance := totalBalance.Sub(lockBalance)

	toCount := int64(len(toAccounts))
	amount, _ := freeBalance.Div(toCount)
	fmt.Printf("freeBalance is [%s], amount is [%s]\n", freeBalance, amount)
	if amount.Compare(types.NewBalance(0)) != types.BalanceCompBigger {
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

		time.Sleep(time.Second)

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

		time.Sleep(time.Second)

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

		time.Sleep(time.Second)
	}

	return nil
}

func generatePairTxs(txAccountsP []*types.Account, tpsP int, txCountP int, txTypeP string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rand.Seed(time.Now().Unix())

	txCurNum := 0

	for txCurNum < txCountP {
		for _, fromAcc := range txAccountsP {
			for _, toAcc := range txAccountsP {
				if fromAcc == toAcc {
					continue
				}

				if err != nil {
					time.Sleep(time.Second)
					err = nil
				}

				amount := rand.Intn(1000) * 1e8
				if amount <= 0 {
					amount = 1e8
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

				if txTypeP == "both" {
					time.Sleep(time.Second)

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

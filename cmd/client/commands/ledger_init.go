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
	"math/big"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addLedgerInitLedgerByIshell(parentCmd *ishell.Cmd) {
	from := util.Flag{
		Name:  "from",
		Must:  true,
		Usage: "seed of genesis account",
		Value: "",
	}
	to := util.Flag{
		Name:  "to",
		Must:  true,
		Usage: "seed of receive accounts",
		Value: []string{},
	}
	args := []util.Flag{from, to}
	c := &ishell.Cmd{
		Name:                "init",
		Help:                "init ledger",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}
			fromSeedP := util.StringVar(c.Args, from)
			toSeedsP := util.StringSliceVar(c.Args, to)
			if len(toSeedsP) == 0 {
				util.Info("to account is empty, pls check")
				return
			}
			err := initLedger(fromSeedP, toSeedsP)
			if err != nil {
				util.Info(err)
				return
			}

			util.Info("init ledger success")
		},
	}
	parentCmd.AddCmd(c)
}

func addLedgerInityCobra(parentCmd *cobra.Command) {
	var fromAccountP string
	var toAccountsP []string
	var generateTestLedgerCmd = &cobra.Command{
		Use:   "init",
		Short: "init ledger",
		Run: func(cmd *cobra.Command, args []string) {
			err := initLedger(fromAccountP, toAccountsP)
			if err != nil {
				cmd.Println(err)
				return
			}
			cmd.Println("init ledger successful")
		},
	}
	generateTestLedgerCmd.Flags().StringVar(&fromAccountP, "from", "", "seed of genesis account")
	generateTestLedgerCmd.Flags().StringSliceVar(&toAccountsP, "to", []string{}, "seed of receive accounts")
	parentCmd.AddCommand(generateTestLedgerCmd)
}

func initLedger(fromSeed string, toSeeds []string) error {
	_, qlc, _ := types.KeypairFromSeed(fromSeed, 0)
	_, gqas, _ := types.KeypairFromSeed(fromSeed, 1)
	qlcAccount := types.NewAccount(qlc)
	qgasAccount := types.NewAccount(gqas)

	toAccounts := make([]*types.Account, len(toSeeds))
	for idx, seed := range toSeeds {
		_, prvKey, _ := types.KeypairFromSeed(seed, 0)
		to := types.NewAccount(prvKey)
		toAccounts[idx] = to
	}

	err := sendAndReceiveTokenAction(qlcAccount, toAccounts, "QLC", nil)
	if err != nil {
		return err
	}
	err = sendAndReceiveTokenAction(qgasAccount, toAccounts, "QGAS", big.NewInt(1e12))
	if err != nil {
		return err
	}
	return nil
}

func sendAndReceiveTokenAction(
	from *types.Account, toAccounts []*types.Account, token string, amount *big.Int,
) error {
	l := len(toAccounts)
	if l == 0 {
		return errors.New("can not find any to accounts")
	}
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
	var toAmount types.Balance
	if accountInfo != nil {
		var totalBalance *types.Balance
		for _, tm := range accountInfo.Tokens {
			if tm.TokenName == token {
				totalBalance = &tm.Balance
				break
			}
		}
		if amount != nil {
			exp := big.NewInt(0).Mul(amount, big.NewInt(int64(l)))
			if totalBalance == nil || totalBalance.Cmp(exp) < 0 {
				return fmt.Errorf("not enough balance, exp: %s, got: %s", exp, totalBalance)
			} else {
				toAmount = types.Balance{Int: amount}
			}
		} else {
			a, err := totalBalance.Div(int64(l))
			if err != nil {
				return err
			}
			toAmount = a
		}
		fmt.Printf("%s, totalBalance is [%s], amount is [%s]\n", token, totalBalance, toAmount)
	} else {
		return fmt.Errorf("can not find any account info of %s", from.Address().String())
	}

	for _, to := range toAccounts {
		var para api.APISendBlockPara
		para = api.APISendBlockPara{
			From:      from.Address(),
			TokenName: token,
			To:        to.Address(),
			Amount:    toAmount,
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
		confirmed := false
		sendHash := sendBlock.GetHash()

		for {
			if err := client.Call(&confirmed, "ledger_blockConfirmedStatus", sendHash); err != nil {
				return err
			}
			if confirmed {
				break
			}
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
		//var changeBlock types.StateBlock
		//err = client.Call(&changeBlock, "ledger_generateChangeBlock", to.Address(), to.Address(), hex.EncodeToString(to.PrivateKey()))
		//if err != nil {
		//	return err
		//}
		//var h2 types.Hash
		//err = client.Call(&h2, "ledger_process", &changeBlock)
		//if err != nil {
		//	return err
		//}
	}
	return nil
}

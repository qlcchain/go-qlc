/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/util"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/test/mock"

	"github.com/spf13/cobra"
	cmn "github.com/tendermint/tmlibs/common"
)

var (
	from   string
	to     string
	token  string
	amount string
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "send transaction",
	Run: func(cmd *cobra.Command, args []string) {
		err := sendAction()
		if err != nil {
			cmd.Println(err)
		}
	},
}

func init() {
	sendCmd.Flags().StringVarP(&from, "from", "f", "", "send account")
	sendCmd.Flags().StringVarP(&to, "to", "t", "", "receive account")
	sendCmd.Flags().StringVarP(&token, "token", "k", mock.GetChainTokenType().String(), "token hash for send action")
	sendCmd.Flags().StringVarP(&amount, "amount", "m", "", "send amount")
	rootCmd.AddCommand(sendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func sendAction() error {
	if from == "" || to == "" || amount == "" {
		fmt.Println("err transfer info")
		return errors.New("err transfer info")
	}
	source, err := types.HexToAddress(from)
	if err != nil {
		fmt.Println(err)
		return err
	}
	t, err := types.HexToAddress(to)
	if err != nil {
		fmt.Println(err)
		return err
	}
	tk, err := types.NewHash(token)
	if err != nil {
		fmt.Println(err)
		return err
	}

	am := types.StringToBalance(amount)
	if cfgPath == "" {
		cfgPath = config.DefaultDataDir()
	}
	cm := config.NewCfgManager(cfgPath)
	cfg, err := cm.Load()
	if err != nil {
		return err
	}
	err = initNode(source, pwd, cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	services, err = startNode()
	if err != nil {
		fmt.Println(err)
	}
	send(source, t, tk, am, pwd)
	cmn.TrapSignal(func() {
		stopNode(services)
	})
	return nil
}

func send(from, to types.Address, token types.Hash, amount types.Balance, password string) {
	w := ctx.Wallet.Wallet
	fmt.Println(from.String())
	session := w.NewSession(from)

	if b, err := session.VerifyPassword(password); b && err == nil {
		a, _ := mock.BalanceToRaw(amount, "QLC")
		sendBlock, err := session.GenerateSendBlock(from, token, to, a)
		if err != nil {
			fmt.Println(err)
		}

		client, err := ctx.RPC.RPC().Attach()
		defer client.Close()

		var h types.Hash
		err = client.Call(&h, "ledger_process", &sendBlock)
		if err != nil {
			fmt.Println(util.ToString(&sendBlock))
			fmt.Println("process block error", err)
		}
	} else {
		fmt.Println("invalid password ", err, " valid: ", b)
	}
}

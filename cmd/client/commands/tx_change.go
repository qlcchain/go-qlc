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

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
)

func addTxChangeCmdByShell(parentCmd *ishell.Cmd) {
	priKeyFlag := util.Flag{
		Name:  "priKey",
		Must:  true,
		Usage: "account private key",
		Value: "",
	}
	repAddrFlag := util.Flag{
		Name:  "repAddr",
		Must:  true,
		Usage: "representative address",
		Value: "",
	}
	args := []util.Flag{priKeyFlag, repAddrFlag}
	c := &ishell.Cmd{
		Name:                "change",
		Help:                "change representative",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}
			accountP := util.StringVar(c.Args, priKeyFlag)
			repP := util.StringVar(c.Args, repAddrFlag)
			err := changeAction(accountP, repP)
			if err != nil {
				util.Warn(err)
				return
			}
			util.Info("change representative success!")
		},
	}
	parentCmd.AddCmd(c)
}

func addTxChangeCmdByCobra(parentCmd *cobra.Command) {
	var accountP string
	var repP string

	var sendCmd = &cobra.Command{
		Use:   "change",
		Short: "change representative",
		Run: func(cmd *cobra.Command, args []string) {
			err := changeAction(accountP, repP)
			if err != nil {
				cmd.Println(err)
				return
			}
			fmt.Println("change representative success!")
		},
	}
	sendCmd.Flags().StringVarP(&accountP, "priKey", "k", "", "account private key")
	sendCmd.Flags().StringVarP(&repP, "repAddr", "r", "", "representative address")
	parentCmd.AddCommand(sendCmd)
}

func changeAction(accountP string, repP string) error {
	if accountP == "" || repP == "" {
		return errors.New("err change param values")
	}
	bytes, err := hex.DecodeString(accountP)
	if err != nil {
		return err
	}
	fromAccount := types.NewAccount(bytes)

	repAddr, err := types.HexToAddress(repP)
	if err != nil {
		return err
	}

	if err := sendChangeTx(fromAccount, repAddr); err != nil {
		return err
	}
	return nil
}

func sendChangeTx(account *types.Account, repAddr types.Address) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	var changeBlock types.StateBlock
	err = client.Call(&changeBlock, "ledger_generateChangeBlock", account.Address(), repAddr, hex.EncodeToString(account.PrivateKey()))
	if err != nil {
		fmt.Println(err)
		return err
	}

	var changeHash types.Hash
	err = client.Call(&changeHash, "ledger_process", &changeBlock)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("address", account.Address(), "changeHash", changeHash)

	return nil
}

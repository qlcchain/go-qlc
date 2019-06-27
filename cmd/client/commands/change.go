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

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/spf13/cobra"
)

func change() {
	var accountP string
	var repP string

	if interactive {
		account := util.Flag{
			Name:  "account",
			Must:  true,
			Usage: "account private key",
			Value: "",
		}
		rep := util.Flag{
			Name:  "representative",
			Must:  true,
			Usage: "representative address",
			Value: "",
		}
		c := &ishell.Cmd{
			Name: "change",
			Help: "change representative",
			Func: func(c *ishell.Context) {
				args := []util.Flag{account, rep}
				if util.HelpText(c, args) {
					return
				}
				if err := util.CheckArgs(c, args); err != nil {
					util.Warn(err)
					return
				}
				accountP := util.StringVar(c.Args, account)
				repP := util.StringVar(c.Args, rep)
				err := changeAction(accountP, repP)
				if err != nil {
					util.Warn(err)
					return
				}
				util.Info("change representative success!")
			},
		}
		shell.AddCmd(c)
	} else {
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
		sendCmd.Flags().StringVarP(&accountP, "account", "a", "", "account private key")
		sendCmd.Flags().StringVarP(&repP, "representative", "r", "", "representative address")
		rootCmd.AddCommand(sendCmd)
	}
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

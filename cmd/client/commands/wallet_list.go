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

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/common/types"
)

func addWalletListCmdByShell(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "list",
		Help: "return wallet list",
		Func: func(c *ishell.Context) {
			if util.HelpText(c, nil) {
				return
			}
			if err := util.CheckArgs(c, nil); err != nil {
				util.Warn(err)
				return
			}
			err := wallets()
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(c)
}

func addWalletListCmdByCobra(parentCmd *cobra.Command) {
	var wlCmd = &cobra.Command{
		Use:   "list",
		Short: "wallet address list",
		Run: func(cmd *cobra.Command, args []string) {
			err := wallets()
			if err != nil {
				cmd.Println(err)
				return
			}
		},
	}
	parentCmd.AddCommand(wlCmd)
}

func wallets() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()
	var addresses []types.Address
	err = client.Call(&addresses, "wallet_list")
	if err != nil {
		return err
	}

	if len(addresses) == 0 {
		return errors.New("no account ,you can try import one!")
	} else {
		for _, v := range addresses {
			if interactive {
				util.Info(v)
			} else {
				fmt.Println(v)
			}
		}
	}

	return nil
}

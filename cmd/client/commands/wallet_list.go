/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/spf13/cobra"
)

func walletList() {
	if interactive {
		c := &ishell.Cmd{
			Name: "walletlist",
			Help: "return wallet list",
			Func: func(c *ishell.Context) {
				if HelpText(c, nil) {
					return
				}
				addrs, err := wallets()
				if err != nil {
					Warn(err)
				} else {
					if len(addrs) == 0 {
						Warn("no account ,you can try import one!")
					} else {
						for _, v := range addrs {
							Info(v)
						}
					}

				}
			},
		}
		shell.AddCmd(c)
	} else {
		var wlCmd = &cobra.Command{
			Use:   "walletlist",
			Short: "wallet address list",
			Run: func(cmd *cobra.Command, args []string) {
				addrs, err := wallets()
				if err != nil {
					cmd.Println(err)
				} else {
					if len(addrs) == 0 {
						cmd.Println("no account ,you can try import one!")
					} else {
						for _, v := range addrs {
							cmd.Println(v)
						}
					}
				}
			},
		}
		rootCmd.AddCommand(wlCmd)
	}
}

func wallets() ([]types.Address, error) {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	var addresses []types.Address
	err = client.Call(&addresses, "wallet_list")
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

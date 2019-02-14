/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/spf13/cobra"
)

// wlCmd represents the wl command
var wlCmd = &cobra.Command{
	Use:   "walletlist",
	Short: "wallet address list",
	Run: func(cmd *cobra.Command, args []string) {
		addrs, err := walletList()
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

func init() {
	rootCmd.AddCommand(wlCmd)
}

func walletList() ([]types.Address, error) {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	var addresses []types.Address
	err = client.Call(&addresses, "wallet_list", pwd)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

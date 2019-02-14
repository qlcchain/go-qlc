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

// wcCmd represents the wc command
var wcCmd = &cobra.Command{
	Use:   "walletcreate",
	Short: "create a wallet for QLCChain node",
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := createWallet()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("create wallet: address=>%s, password=>%s success", addr.String(), pwd)
			cmd.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(wcCmd)
}

func createWallet() (types.Address, error) {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return types.ZeroAddress, err
	}
	defer client.Close()
	var addr types.Address
	err = client.Call(&addr, "wallet_newWallet", pwd)
	if err != nil {
		return types.ZeroAddress, err
	}
	return addr, nil
}

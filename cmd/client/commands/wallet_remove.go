/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/spf13/cobra"
)

// wrCmd represents the wr command
var wrCmd = &cobra.Command{
	Use:   "walletremove",
	Short: "remove wallet",
	Run: func(cmd *cobra.Command, args []string) {
		err := removeWallet()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("remove wallet: %s success", account)
			cmd.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(wrCmd)
}

func removeWallet() error {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return err
	}
	defer client.Close()
	err = client.Call(nil, "wallet_remove", account)
	if err != nil {
		return err
	}
	return nil
}

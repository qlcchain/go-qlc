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

var (
	addresses []string
)

// bcCmd represents the bc command
var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: " balance for accounts",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(endpoint)
		if err != nil {
			cmd.Println(err)
			return
		}
		defer client.Close()
		if len(addresses) < 1 {
			cmd.Println("err account")
			return
		}
		para := make([]string, 0)
		for _, a := range addresses {
			para = append(para, a)
		}
		var resp map[types.Address]map[string]map[string]types.Balance
		err = client.Call(&resp, "ledger_accountsBalances", para)
		if err != nil {
			cmd.Println(err)
			return
		}
		for key, value := range resp {
			cmd.Println(key)
			for k, v := range value {
				cmd.Printf("	%s, balance:%s, pending:%s", k, v["balance"], v["pending"])
				cmd.Println()
			}
		}
	},
}

func init() {
	balanceCmd.Flags().StringSliceVar(&addresses, "address", addresses, "address for accounts")
	rootCmd.AddCommand(balanceCmd)
}

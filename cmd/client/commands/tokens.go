/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/test/mock"
	"github.com/spf13/cobra"
)

// tlCmd represents the tl command
var tlCmd = &cobra.Command{
	Use:   "tokens",
	Short: "token list",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(endpoint)
		if err != nil {
			cmd.Println(err)
			return
		}
		defer client.Close()

		var tokeninfos []*mock.TokenInfo
		err = client.Call(&tokeninfos, "ledger_tokens")

		for _, v := range tokeninfos {
			cmd.Printf("TokenId:%s  TokenName:%s  TokenSymbol:%s  TotalSupply:%s  Decimals:%d  Owner:%s", v.TokenId, v.TokenName, v.TokenSymbol, v.TotalSupply, v.Decimals, v.Owner)
			cmd.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(tlCmd)
}

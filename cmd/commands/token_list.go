/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/qlcchain/go-qlc/test/mock"

	"github.com/spf13/cobra"
)

// tlCmd represents the tl command
var tlCmd = &cobra.Command{
	Use:   "tl",
	Short: "token list",
	Run: func(cmd *cobra.Command, args []string) {
		tokeninfos := getTokenList()
		for _, v := range tokeninfos {
			cmd.Printf("TokenId:%s  TokenName:%s  TokenSymbol:%s  TotalSupply:%s  Decimals:%d  Owner:%s", v.TokenId, v.TokenName, v.TokenSymbol, v.TotalSupply, v.Decimals, v.Owner)
			cmd.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(tlCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getTokenList() []mock.TokenInfo {
	scBlocks := mock.GetSmartContracts()
	var tokenInfos []mock.TokenInfo
	for _, v := range scBlocks {
		hash := v.GetHash()
		info, err := mock.GetTokenById(hash)
		if err != nil {
			continue
		}
		tokenInfos = append(tokenInfos, info)
	}
	return tokenInfos
}

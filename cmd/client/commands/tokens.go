/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func tokens() {
	if interactive {
		c := &ishell.Cmd{
			Name: "tokens",
			Help: "return token info list of chain",
			Func: func(c *ishell.Context) {
				if util.HelpText(c, nil) {
					return
				}
				if err := util.CheckArgs(c, nil); err != nil {
					util.Warn(err)
					return
				}
				err := tokensinfo()
				if err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(c)
	} else {
		var tlCmd = &cobra.Command{
			Use:   "tokens",
			Short: "return token info list of chain",
			Run: func(cmd *cobra.Command, args []string) {
				err := tokensinfo()
				if err != nil {
					cmd.Println(err)
					return
				}
			},
		}
		rootCmd.AddCommand(tlCmd)
	}
}

func tokensinfo() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	var tokeninfos []*types.TokenInfo
	err = client.Call(&tokeninfos, "ledger_tokens")
	if err != nil {
		return err
	}
	if interactive {
		util.Info(fmt.Sprintf("%d tokens found:", len(tokeninfos)))
	}
	for _, v := range tokeninfos {
		fmt.Printf("TokenId:%s  TokenName:%s  TokenSymbol:%s  TotalSupply:%s  Decimals:%d  Owner:%s", v.TokenId, v.TokenName, v.TokenSymbol, v.TotalSupply, v.Decimals, v.Owner)
		fmt.Println()
	}
	return nil
}

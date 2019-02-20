/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/test/mock"
)

func init() {
	c := &ishell.Cmd{
		Name: "tokens",
		Help: "return token info list of chain",
		Func: func(c *ishell.Context) {
			if HelpText(c, nil) {
				return
			}
			client, err := rpc.Dial(endpointP)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer client.Close()

			var tokeninfos []*mock.TokenInfo
			err = client.Call(&tokeninfos, "ledger_tokens")

			for _, v := range tokeninfos {
				fmt.Printf("TokenId:%s  TokenName:%s  TokenSymbol:%s  TotalSupply:%s  Decimals:%d  Owner:%s", v.TokenId, v.TokenName, v.TokenSymbol, v.TotalSupply, v.Decimals, v.Owner)
				fmt.Println()
			}
		},
	}
	shell.AddCmd(c)
}

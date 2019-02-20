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
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
)

func init() {
	address := Flag{
		Name:  "address",
		Must:  true,
		Usage: "address for account",
		Value: "",
	}
	s := &ishell.Cmd{
		Name: "balance",
		Help: "balance for accounts",
		Func: func(c *ishell.Context) {
			args := []Flag{address}
			if HelpText(c, args) {
				return
			}
			if err := CheckArgs(c, args); err != nil {
				Warn(err)
				return
			}
			addresses := StringSliceVar(c.Args, address)

			client, err := rpc.Dial(endpointP)
			if err != nil {
				Warn(err)
				return
			}
			defer client.Close()
			var resp map[types.Address]map[string]map[string]types.Balance
			err = client.Call(&resp, "ledger_accountsBalances", addresses)
			if err != nil {
				Warn(err)
				return
			}
			if len(resp) == 0 {
				Warn("can not find addresses in wallet")
			}
			for key, value := range resp {
				Info(key, ":")
				for k, v := range value {
					fmt.Printf("      %s: balance is %s, pending is %s", k, v["balance"], v["pending"])
					fmt.Println()
				}
			}
		},
	}
	shell.AddCmd(s)
}

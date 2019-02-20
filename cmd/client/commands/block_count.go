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
)

func init() {
	c := &ishell.Cmd{
		Name: "blockcount",
		Help: "return the total count of block in db",
		Func: func(c *ishell.Context) {
			if HelpText(c, nil) {
				return
			}
			client, err := rpc.Dial(endpointP)
			if err != nil {
				Warn(err)
				return
			}
			defer client.Close()

			var resp map[string]uint64
			err = client.Call(&resp, "ledger_transactionsCount")
			if err != nil {
				Warn(err)
				return
			}
			state := resp["count"]
			unchecked := resp["unchecked"]

			Info(fmt.Sprintf("total state block count is : %d, unchecked block count is: %d", state, unchecked))
		},
	}
	shell.AddCmd(c)
}

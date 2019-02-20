/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
)

func init() {
	c := &ishell.Cmd{
		Name: "walletlist",
		Help: "return wallet list",
		Func: func(c *ishell.Context) {
			if HelpText(c, nil) {
				return
			}
			addrs, err := walletList()
			if err != nil {
				Warn(err)
			} else {
				if len(addrs) == 0 {
					Warn("no account ,you can try import one!")
				} else {
					for _, v := range addrs {
						Info(v)
					}
				}

			}
		},
	}
	shell.AddCmd(c)
}

func walletList() ([]types.Address, error) {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	var addresses []types.Address
	err = client.Call(&addresses, "wallet_list")
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

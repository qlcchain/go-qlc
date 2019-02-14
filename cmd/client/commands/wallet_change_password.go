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

var newPwd string

// wcpCmd represents the wcp command
var wcpCmd = &cobra.Command{
	Use:   "changepassword",
	Short: "change wallet password",
	Run: func(cmd *cobra.Command, args []string) {
		err := changePassword()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("change password success for account: %s", account)
			cmd.Println()
		}
	},
}

func init() {
	wcpCmd.Flags().StringVarP(&newPwd, "newPassword", "n", "", "new password for wallet")
	rootCmd.AddCommand(wcpCmd)
}

func changePassword() error {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return err
	}
	defer client.Close()
	var addr types.Address
	err = client.Call(&addr, "wallet_changePassword", account, pwd, newPwd)
	if err != nil {
		return err
	}
	return nil
}

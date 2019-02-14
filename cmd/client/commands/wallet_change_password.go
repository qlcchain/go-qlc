// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
			cmd.Printf("change password success for account: %s\n", account)
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

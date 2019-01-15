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
	"github.com/qlcchain/go-qlc/config"
	"github.com/spf13/cobra"
)

var newPassword string

// wcpCmd represents the wcp command
var wcpCmd = &cobra.Command{
	Use:   "wcp",
	Short: "change wallet password",
	Run: func(cmd *cobra.Command, args []string) {
		err := changePassword()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("change password success for account: %s", account)
		}
	},
}

func init() {
	wcpCmd.Flags().StringVarP(&newPassword, "newPassword", "n", "", "new password for wallet")

	rootCmd.AddCommand(wcpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wcpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wcpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func changePassword() error {
	if cfgPath == "" {
		cfgPath = config.DefaultDataDir()
	}
	cm := config.NewCfgManager(cfgPath)
	cfg, err := cm.Load()
	if err != nil {
		return err
	}
	ac, err := types.HexToAddress(account)
	if err != nil {
		return err
	}
	err = initNode(ac, pwd, cfg)
	if err != nil {
		logger.Error(err)
		return err
	}
	w := ctx.Wallet.Wallet
	session := w.NewSession(ac)
	b, err := session.VerifyPassword(pwd)
	if b && err == nil {
		err = session.ChangePassword(newPassword)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

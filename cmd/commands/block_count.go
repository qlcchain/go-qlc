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

// bcCmd represents the bc command
var bcCmd = &cobra.Command{
	Use:   "bc",
	Short: "block count",
	Run: func(cmd *cobra.Command, args []string) {
		count, err := countBlock()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("total block count is : %d", count)
		}

	},
}

func init() {
	rootCmd.AddCommand(bcCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func countBlock() (uint64, error) {
	if cfgPath == "" {
		cfgPath = config.DefaultConfigFile()
	}
	cm := config.NewCfgManager(cfgPath)
	cfg, err := cm.Load()
	if err != nil {
		return 0, err
	}
	err = initNode(types.ZeroAddress, "", cfg)
	if err != nil {
		return 0, err
	}
	l := ctx.Ledger.Ledger

	count, err := l.CountBlocks()
	if err != nil {
		return 0, err
	}
	return count, nil
}

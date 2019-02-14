/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/spf13/cobra"
)

var (
	account  string
	pwd      string
	cfgPath  string
	endpoint string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "QLCC",
	Short: "CLI for QLCChain Client.",
	Long:  `QLC Chain is the next generation public blockchain designed for the NaaS.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := start()
		if err != nil {
			cmd.Println(err)
		}

	},
}

func start() error {
	return nil
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&account, "account", "a", "", "wallet address")
	rootCmd.PersistentFlags().StringVarP(&pwd, "password", "p", "", "password for wallet")
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "config file")
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "ws://0.0.0.0:9736", "endpoint for client")
}

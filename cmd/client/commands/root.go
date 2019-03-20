/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	"os"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	"github.com/qlcchain/go-qlc"
	"github.com/spf13/cobra"
)

var (
	shell       *ishell.Shell
	rootCmd     *cobra.Command
	interactive bool
)

type Flag struct {
	Name  string
	Usage string
	Must  bool
	Value interface{}
}

var (
	endpointP  string
	endpoint   Flag
	commonFlag []Flag
)

// set global variable
func init() {
	if goqlc.MAINNET {
		endpointP = "ws://0.0.0.0:9736"
	} else {
		endpointP = "ws://0.0.0.0:19736"
	}
	endpoint = Flag{
		Name:  "endpoint",
		Must:  false,
		Usage: "endpoint for client to connect to server",
		Value: endpointP,
	}
}

func Execute(osArgs []string) {
	interactive = IsInteractive(osArgs)
	if interactive {
		shell = ishell.NewWithConfig(&readline.Config{
			Prompt:      fmt.Sprintf("%c[1;0;32m%s%c[0m", 0x1B, ">> ", 0x1B),
			HistoryFile: "/tmp/readline.tmp",
			//AutoComplete:      completer,
			InterruptPrompt:   "^C",
			EOFPrompt:         "exit",
			HistorySearchFold: true,
			//FuncFilterInputRune: filterInput,
		})
		shell.Println("QLC Chain Client")
		//set common variable
		commonFlag = make([]Flag, 0)
		addcommands()
		// commonFlag = append(commonFlag, p)
		// run shell
		shell.Run()
	} else {
		rootCmd = &cobra.Command{
			Use:   "QLCC",
			Short: "CLI for QLCChain Client.",
			Long:  `QLC Chain is the next generation public blockchain designed for the NaaS.`,
			Run: func(cmd *cobra.Command, args []string) {
				//err := start()
				//if err != nil {
				//	cmd.Println(err)
				//}
			},
		}
		//rootCmd.PersistentFlags().StringVarP(&account, "account", "a", "", "wallet address")
		//rootCmd.PersistentFlags().StringVarP(&pwd, "password", "p", "", "password for wallet")
		//rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "config file")
		rootCmd.PersistentFlags().StringVarP(&endpointP, "endpoint", "e", endpointP, "endpoint for client")
		addcommands()
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func IsInteractive(osArgs []string) bool {
	if len(osArgs) == 2 && osArgs[1] == "-i" {
		return true
	}
	if len(osArgs) == 4 && osArgs[1] == "-i" && osArgs[2] == "--endpoint" {
		endpointP = osArgs[3]
		return true
	}
	return false
}

func addcommands() {
	account()
	balance()
	batchSend()
	blockCount()
	performance()
	send()
	tokens()
	version()
	changePassword()
	walletCreate()
	walletList()
	walletRemove()
	mintage()
}

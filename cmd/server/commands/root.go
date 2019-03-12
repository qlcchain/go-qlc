/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/cmd/client/commands"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/spf13/cobra"
)

var (
	shell       *ishell.Shell
	rootCmd     *cobra.Command
	interactive bool
)

var (
	//	accountP   string
	passwordP  string
	seedP      string
	cfgPathP   string
	isProfileP bool

	//account   commands.Flag
	password  commands.Flag
	seed      commands.Flag
	cfgPath   commands.Flag
	isProfile commands.Flag

	ctx      *chain.QlcContext
	services []common.Service
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(osArgs []string) {
	if len(osArgs) == 2 && osArgs[1] == "-i" {
		interactive = true
	}
	if interactive {
		shell = ishell.NewWithConfig(
			&readline.Config{
				Prompt:      fmt.Sprintf("%c[1;0;32m%s%c[0m", 0x1B, ">> ", 0x1B),
				HistoryFile: "/tmp/readline.tmp",
				//AutoComplete:      completer,
				InterruptPrompt:   "^C",
				EOFPrompt:         "exit",
				HistorySearchFold: true,
				//FuncFilterInputRune: filterInput,
			})
		shell.Println("QLC Chain Server")
		addcommand()
		shell.Run()
	} else {
		rootCmd = &cobra.Command{
			Use:   "QLCCChain",
			Short: "CLI for QLCChain Server",
			Long:  `QLC Chain is the next generation public blockchain designed for the NaaS.`,
			Run: func(cmd *cobra.Command, args []string) {
				err := start()
				if err != nil {
					cmd.Println(err)
				}

			},
		}
		rootCmd.PersistentFlags().StringVarP(&cfgPathP, "config", "c", "", "config file")
		//		rootCmd.PersistentFlags().StringVarP(&accountP, "account", "a", "", "wallet address,if is nil,just run a node")
		rootCmd.PersistentFlags().StringVarP(&passwordP, "password", "p", "", "password for wallet")
		rootCmd.PersistentFlags().StringVarP(&seedP, "seed", "s", "", "seed for accounts")
		rootCmd.PersistentFlags().BoolVar(&isProfileP, "profile", false, "enable profile")
		addcommand()
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

}

func addcommand() {
	if interactive {
		run()
	}
	walletimport()
}

func start() error {
	var seed types.Seed
	if cfgPathP == "" {
		cfgPathP = config.DefaultDataDir()
	}
	cm := config.NewCfgManager(cfgPathP)
	cfg, err := cm.Load()
	if err != nil {
		return err
	}
	if seedP == "" {
		seed = types.ZeroSeed
	} else {
		sByte, _ := hex.DecodeString(seedP)
		seedT, err := types.BytesToSeed(sByte)
		if err != nil {
			return err
		}
		seed = *seedT
	}

	if isProfileP {
		profDir := filepath.Join(cfg.DataDir, "pprof", time.Now().Format("2006-01-02T15-04"))
		_ = util.CreateDirIfNotExist(profDir)
		//CPU profile
		cpuProfile, err := os.Create(filepath.Join(profDir, "cpu.prof"))
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		} else {
			log.Println("create CPU profile: ", cpuProfile.Name())
		}

		runtime.SetCPUProfileRate(500)
		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		} else {
			log.Println("start CPU profile")
		}
		defer func() {
			_ = cpuProfile.Close()
			pprof.StopCPUProfile()
		}()

		//MEM profile
		memProfile, err := os.Create(filepath.Join(profDir, "mem.prof"))
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		} else {
			log.Println("create MEM profile: ", memProfile.Name())
		}
		runtime.GC()
		if err := pprof.WriteHeapProfile(memProfile); err != nil {
			log.Fatal("could not write memory profile: ", err)
		} else {
			log.Println("start MEM profile")
		}
		defer func() {
			_ = memProfile.Close()
		}()

		go func() {
			//view result in http://localhost:6060/debug/pprof/
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	err = runNode(seed, cfg)
	if err != nil {
		return err
	}
	return nil
}

func run() {
	//account = commands.Flag{
	//	Name:  "account",
	//	Must:  false,
	//	Usage: "wallet address,if is nil,just run a node",
	//	Value: "",
	//}
	password = commands.Flag{
		Name:  "password",
		Must:  false,
		Usage: "password for wallet",
		Value: "",
	}
	seed = commands.Flag{
		Name:  "seed",
		Must:  false,
		Usage: "seed for wallet,if is nil,just run a node",
		Value: "",
	}
	cfgPath = commands.Flag{
		Name:  "config",
		Must:  false,
		Usage: "config file path",
		Value: "",
	}

	isProfile = commands.Flag{
		Name:  "profile",
		Must:  false,
		Usage: "enable profile",
		Value: false,
	}

	s := &ishell.Cmd{
		Name: "run",
		Help: "start qlc server",
		Func: func(c *ishell.Context) {
			args := []commands.Flag{seed, cfgPath, isProfile}
			if commands.HelpText(c, args) {
				return
			}
			if err := commands.CheckArgs(c, args); err != nil {
				commands.Warn(err)
				return
			}
			//accountP = commands.StringVar(c.Args, account)
			passwordP = commands.StringVar(c.Args, password)
			seedP = commands.StringVar(c.Args, seed)
			cfgPathP = commands.StringVar(c.Args, cfgPath)
			isProfileP = commands.BoolVar(c.Args, isProfile)

			err := start()
			if err != nil {
				commands.Warn(err)
			}
		},
	}
	shell.AddCmd(s)
}

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
	ss "github.com/qlcchain/go-qlc/chain/services"
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
	seedP       string
	privateKeyP string
	accountP    string
	passwordP   string
	cfgPathP    string
	isProfileP  bool

	privateKey commands.Flag
	account    commands.Flag
	password   commands.Flag
	seed       commands.Flag
	cfgPath    commands.Flag
	isProfile  commands.Flag

	ctx            *chain.QlcContext
	services       []common.Service
	maxAccountSize = 100
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
		addCommand()
		shell.Run()
	} else {
		rootCmd = &cobra.Command{
			Use:   "gqlc",
			Short: "CLI for QLCChain Server",
			Long:  `QLC Chain is the next generation public block chain designed for the NaaS.`,
			Run: func(cmd *cobra.Command, args []string) {
				err := start()
				if err != nil {
					cmd.Println(err)
				}

			},
		}
		rootCmd.PersistentFlags().StringVar(&cfgPathP, "config", "", "config file")
		rootCmd.PersistentFlags().StringVar(&accountP, "account", "", "wallet address, if is nil,just run a node")
		rootCmd.PersistentFlags().StringVar(&passwordP, "password", "", "password for wallet")
		rootCmd.PersistentFlags().StringVar(&seedP, "seed", "", "seed for accounts")
		rootCmd.PersistentFlags().StringVar(&privateKeyP, "privateKey", "", "seed for accounts")
		rootCmd.PersistentFlags().BoolVar(&isProfileP, "profile", false, "enable profile")
		addCommand()
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func addCommand() {
	if interactive {
		run()
	}
	walletimport()
	version()
}

func start() error {
	var accounts []*types.Account
	if cfgPathP == "" {
		cfgPathP = config.DefaultDataDir()
	}
	cm := config.NewCfgManager(cfgPathP)
	cfg, err := cm.Load()
	if err != nil {
		return err
	}

	if len(seedP) > 0 {
		fmt.Println("run node SEED mode")
		sByte, _ := hex.DecodeString(seedP)
		tmp, err := seedToAccouts(sByte)
		if err != nil {
			return err
		}
		accounts = append(accounts, tmp...)
	} else if len(privateKeyP) > 0 {
		fmt.Println("run node PRIVATE KEY mode")
		bytes, err := hex.DecodeString(privateKeyP)
		if err != nil {
			return err
		}
		account := types.NewAccount(bytes)
		accounts = append(accounts, account)
	} else if len(accountP) > 0 {
		fmt.Println("run node WALLET mode")
		address, err := types.HexToAddress(accountP)
		if err != nil {
			return err
		}

		w := ss.NewWalletService(cfg).Wallet
		session := w.NewSession(address)
		defer func() {
			err := session.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()

		if b, err := session.VerifyPassword(passwordP); b && err == nil {
			bytes, err := session.GetSeed()
			tmp, err := seedToAccouts(bytes)
			if err != nil {
				return err
			}
			accounts = append(accounts, tmp...)
		} else {
			return fmt.Errorf("invalid wallet password of %s", accountP)
		}
	} else {
		fmt.Println("run node without account")
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

	err = runNode(accounts, cfg)
	if err != nil {
		return err
	}
	return nil
}

func seedToAccouts(data []byte) ([]*types.Account, error) {
	seed, err := types.BytesToSeed(data)
	if err != nil {
		return nil, err
	}
	var accounts []*types.Account
	for i := 0; i < maxAccountSize; i++ {
		account, _ := seed.Account(uint32(i))
		accounts = append(accounts, account)
	}

	return accounts, nil
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
			accountP = commands.StringVar(c.Args, account)
			passwordP = commands.StringVar(c.Args, password)
			privateKeyP = commands.StringVar(c.Args, privateKey)
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

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
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
)

var shell = ishell.NewWithConfig(
	&readline.Config{
		Prompt:      fmt.Sprintf("%c[1;0;32m%s%c[0m", 0x1B, ">> ", 0x1B),
		HistoryFile: "/tmp/readline.tmp",
		//AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		//FuncFilterInputRune: filterInput,
	})

var (
	//accountP  string
	account commands.Flag
	//passwordP string
	password commands.Flag
	//cfgPathP  string
	cfgPath commands.Flag

	ctx      *chain.QlcContext
	services []common.Service
)

func init() {
	account = commands.Flag{
		Name:  "account",
		Must:  false,
		Usage: "wallet address,if is nil,just run a node",
		Value: "",
	}
	password = commands.Flag{
		Name:  "password",
		Must:  false,
		Usage: "password for wallet",
		Value: "",
	}
	cfgPath = commands.Flag{
		Name:  "config",
		Must:  false,
		Usage: "config file path",
		Value: "",
	}

	isProfile := commands.Flag{
		Name:  "profile",
		Must:  false,
		Usage: "enable profile",
		Value: false,
	}

	s := &ishell.Cmd{
		Name: "run",
		Help: "start qlc server",
		Func: func(c *ishell.Context) {
			args := []commands.Flag{account, password, cfgPath, isProfile}
			if commands.HelpText(c, args) {
				return
			}
			if err := commands.CheckArgs(c, args); err != nil {
				commands.Warn(err)
				return
			}
			accountP := commands.StringVar(c.Args, account)
			passwordP := commands.StringVar(c.Args, password)
			cfgPathP := commands.StringVar(c.Args, cfgPath)
			isProfileP := commands.BoolVar(c.Args, isProfile)

			err := start(accountP, passwordP, cfgPathP, isProfileP)
			if err != nil {
				commands.Warn(err)
			}
		},
	}

	shell.AddCmd(s)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	shell.Println("QLC Chain Server")
	shell.Run()

}

func start(accountP, passwordP, cfgPathP string, isProfileP bool) error {
	var addr types.Address
	if cfgPathP == "" {
		cfgPathP = config.DefaultDataDir()
	}
	cm := config.NewCfgManager(cfgPathP)
	cfg, err := cm.Load()
	if err != nil {
		return err
	}
	if accountP == "" {
		addr = types.ZeroAddress
	} else {
		addr, err = types.HexToAddress(accountP)
		if err != nil {
			return err
		}
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

	err = runNode(addr, passwordP, cfg)
	if err != nil {
		return err
	}
	return nil
}

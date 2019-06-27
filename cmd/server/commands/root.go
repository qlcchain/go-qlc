/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	ss "github.com/qlcchain/go-qlc/chain/services"
	cmdutil "github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	qlclog "github.com/qlcchain/go-qlc/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	shell       *ishell.Shell
	rootCmd     *cobra.Command
	interactive bool
)

var (
	seedP         string
	privateKeyP   string
	accountP      string
	passwordP     string
	cfgPathP      string
	isProfileP    bool
	noBootstrapP  bool
	configParamsP string
	RunModeP      string

	privateKey   cmdutil.Flag
	account      cmdutil.Flag
	password     cmdutil.Flag
	seed         cmdutil.Flag
	cfgPath      cmdutil.Flag
	isProfile    cmdutil.Flag
	noBootstrap  cmdutil.Flag
	configParams cmdutil.Flag
	RunMode      cmdutil.Flag

	//ctx            *chain.QlcContext
	ledgerService    *ss.LedgerService
	walletService    *ss.WalletService
	netService       *ss.P2PService
	consensusService *ss.ConsensusService
	rPCService       *ss.RPCService
	sqliteService    *ss.SqliteService
	povService       *ss.PoVService
	minerService     *ss.MinerService
	services         []common.Service
	maxAccountSize   = 100
	logger           = qlclog.NewLogger("config_detail")
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
		rootCmd.PersistentFlags().BoolVar(&noBootstrapP, "nobootnode", false, "disable bootstrap node")
		rootCmd.PersistentFlags().StringVar(&configParamsP, "configParams", "", "parameter set that needs to be changed")
		rootCmd.PersistentFlags().StringVar(&RunModeP, "mode", common.RunModeNormalStr, "running mode")

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
	cfg, err := cmdutil.GetConfig(cfgPathP)
	if err != nil {
		return err
	}

	debug.SetGCPercent(10)

	if len(configParamsP) > 0 {
		fmt.Println("need set parameter")
		err = updateConfig(cfg, cfgPathP)
		if err != nil {
			return err
		}
	}
	if len(seedP) > 0 {
		fmt.Println("run node SEED mode")
		sByte, _ := hex.DecodeString(seedP)
		tmp, err := seedToAccounts(sByte)
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
		ledger.CloseLedger()
		session := w.NewSession(address)
		defer func() {
			err := w.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()

		if b, err := session.VerifyPassword(passwordP); b && err == nil {
			bytes, err := session.GetSeed()
			tmp, err := seedToAccounts(bytes)
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
			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}

	if noBootstrapP {
		//remove all p2p bootstrap node
		cfg.P2P.BootNodes = []string{}
	}

	configDetails := util.ToIndentString(cfg)
	log.Printf("%s\n", configDetails)

	if RunModeP != common.RunModeNormalStr && RunModeP != common.RunModeSimpleStr {
		log.Fatalf("invalid running node(%s), should be normal/simple", common.RunMode)
	} else {
		switch RunModeP {
		case common.RunModeNormalStr:
			common.RunMode = common.RunModeNormal
		case common.RunModeSimpleStr:
			common.RunMode = common.RunModeSimple
		}
		log.Printf("running node as %s mode", RunModeP)
	}

	if common.RunMode == common.RunModeSimple {
		go func() {
			cleanMem := time.NewTicker(30 * time.Second)
			for {
				select {
				case <-cleanMem.C:
					debug.FreeOSMemory()
				}
			}
		}()
	}

	err = runNode(accounts, cfg)
	if err != nil {
		return err
	}
	trapSignal()
	return nil
}

func trapSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	<-c

	sers := make([]common.Service, 0)
	for i := len(services) - 1; i >= 0; i-- {
		sers = append(sers, services[i])
	}
	stopNode(sers)
	fmt.Println("qlc node closed successfully")

	//bus := event.GetEventBus(cfg.LedgerDir())
	//if err := bus.Close(); err != nil {
	//	fmt.Println(err)
	//}
}

func seedToAccounts(data []byte) ([]*types.Account, error) {
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
	account = cmdutil.Flag{
		Name:  "account",
		Must:  false,
		Usage: "wallet address,if is nil,just run a node",
		Value: "",
	}
	password = cmdutil.Flag{
		Name:  "password",
		Must:  false,
		Usage: "password for wallet",
		Value: "",
	}
	seed = cmdutil.Flag{
		Name:  "seed",
		Must:  false,
		Usage: "seed for wallet,if is nil,just run a node",
		Value: "",
	}
	privateKey = cmdutil.Flag{
		Name:  "privateKey",
		Must:  false,
		Usage: "account private key",
		Value: "",
	}
	cfgPath = cmdutil.Flag{
		Name:  "config",
		Must:  false,
		Usage: "config file path",
		Value: "",
	}

	isProfile = cmdutil.Flag{
		Name:  "profile",
		Must:  false,
		Usage: "enable profile",
		Value: false,
	}

	noBootstrap = cmdutil.Flag{
		Name:  "nobootstrap",
		Must:  false,
		Usage: "disable p2p bootstrap node",
		Value: false,
	}

	configParams = cmdutil.Flag{
		Name:  "configParam",
		Must:  false,
		Usage: "parameter set that needs to be changed",
		Value: "",
	}

	RunMode = cmdutil.Flag{
		Name:  "mode",
		Must:  false,
		Usage: "running mode(normal/simple)",
		Value: common.RunModeNormal,
	}

	s := &ishell.Cmd{
		Name: "run",
		Help: "start qlc server",
		Func: func(c *ishell.Context) {
			args := []cmdutil.Flag{seed, cfgPath, isProfile, noBootstrap}
			if cmdutil.HelpText(c, args) {
				return
			}
			if err := cmdutil.CheckArgs(c, args); err != nil {
				cmdutil.Warn(err)
				return
			}
			accountP = cmdutil.StringVar(c.Args, account)
			passwordP = cmdutil.StringVar(c.Args, password)
			privateKeyP = cmdutil.StringVar(c.Args, privateKey)
			seedP = cmdutil.StringVar(c.Args, seed)
			cfgPathP = cmdutil.StringVar(c.Args, cfgPath)
			isProfileP = cmdutil.BoolVar(c.Args, isProfile)
			noBootstrapP = cmdutil.BoolVar(c.Args, noBootstrap)
			configParamsP = cmdutil.StringVar(c.Args, configParams)
			RunModeP = cmdutil.StringVar(c.Args, RunMode)

			err := start()
			if err != nil {
				cmdutil.Warn(err)
			}
		},
	}
	shell.AddCmd(s)
}

func updateConfig(cfg *config.Config, cfgPathP string) error {
	paramSlice := strings.Split(configParamsP, ":")
	var s []string
	if cfgPathP == "" {
		s = strings.Split(config.QlcConfigFile, ".")
	} else {
		s = strings.Split(filepath.Base(cfgPathP), ".")
	}
	if len(s) != 2 {
		return errors.New("split error")
	}
	viper.SetConfigName(s[0])
	viper.AddConfigPath(cfg.DataDir)
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	err = viper.ReadConfig(r)
	if err != nil {
		return err
	}

	for _, cp := range paramSlice {
		k := strings.Split(cp, "=")
		if len(k) != 2 || len(k[0]) == 0 || len(k[1]) == 0 {
			continue
		}
		if oldValue := viper.Get(k[0]); oldValue != nil {
			viper.Set(k[0], k[1])
		}
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return err
	}
	return nil
}

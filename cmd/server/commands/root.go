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
	"math/big"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/qlcchain/go-qlc/common"

	"github.com/qlcchain/go-qlc/config"

	"github.com/qlcchain/go-qlc/log"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/chain/context"
	cmdutil "github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/wallet"
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
	configParamsP string
	genesisSeedP  string

	privateKey   cmdutil.Flag
	account      cmdutil.Flag
	password     cmdutil.Flag
	seed         cmdutil.Flag
	cfgPath      cmdutil.Flag
	isProfile    cmdutil.Flag
	configParams cmdutil.Flag
	genesisSeed  cmdutil.Flag
	//chainContext   *context.ChainContext
	maxAccountSize = 100
	//logger         = qlclog.NewLogger("config_detail")
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(osArgs []string) {
	if len(osArgs) == 2 && osArgs[1] == "-s" {
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
		rootCmd.PersistentFlags().StringVar(&configParamsP, "configParams", "", "parameter set that needs to be changed")
		rootCmd.PersistentFlags().StringVar(&genesisSeedP, "genesisSeed", "", "genesis seed")
		addCommand()
		if err := rootCmd.Execute(); err != nil {
			log.Root.Info(err)
			os.Exit(1)
		}
	}
}

func addCommand() {
	if interactive {
		run()
	}
	addWalletCmd()
	chainVersion()
	removeDB()
	purgePov()
}

func start() error {
	var accounts []*types.Account
	chainContext := context.NewChainContext(cfgPathP)
	cm, err := chainContext.ConfigManager(func(cm *config.CfgManager) error {
		cfg, _ := cm.Config()
		if genesisSeedP != "" {
			cfg.Genesis.GenesisBlocks = []*config.GenesisInfo{}
			if err := generateChainTokenGenesisBlock(genesisSeedP, cfg); err != nil {
				return err
			}
			if err := generateGasTokenGenesisBlock(genesisSeedP, cfg); err != nil {
				return err
			}
			_ = cm.Save()
		}
		loadGenesisAccount(cfg)
		if len(configParamsP) > 0 {
			params := strings.Split(configParamsP, ";")

			if len(params) > 0 {
				_, err := cm.PatchParams(params, cfg)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	cfg, err := cm.Config()
	if err != nil {
		return err
	}

	log.Root.Info("Run node id: ", chainContext.Id())

	if len(seedP) > 0 {
		log.Root.Info("run node SEED mode")
		sByte, _ := hex.DecodeString(seedP)
		tmp, err := seedToAccounts(sByte)
		if err != nil {
			return err
		}
		accounts = append(accounts, tmp...)
	} else if len(privateKeyP) > 0 {
		log.Root.Info("run node PRIVATE KEY mode")
		bytes, err := hex.DecodeString(privateKeyP)
		if err != nil {
			return err
		}
		account := types.NewAccount(bytes)
		accounts = append(accounts, account)
	} else if len(accountP) > 0 {
		log.Root.Info("run node WALLET mode")
		address, err := types.HexToAddress(accountP)
		if err != nil {
			return err
		}

		w := wallet.NewWalletStore(cm.ConfigFile)
		defer func() {
			if w != nil {
				_ = w.Close()
			}
			ledger.CloseLedger()
		}()

		session := w.NewSession(address)
		defer func() {
			err := w.Close()
			if err != nil {
				log.Root.Info(err)
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
		log.Root.Info("run node without account")
	}

	if isProfileP {
		profDir := filepath.Join(cfg.DataDir, "pprof", time.Now().Format("2006-01-02T15-04"))
		_ = util.CreateDirIfNotExist(profDir)
		//CPU profile
		cpuProfile, err := os.Create(filepath.Join(profDir, "cpu.prof"))
		if err != nil {
			log.Root.Error("could not create CPU profile: ", err)
		} else {
			log.Root.Info("create CPU profile: ", cpuProfile.Name())
		}

		runtime.SetCPUProfileRate(500)
		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Root.Error("could not start CPU profile: ", err)
		} else {
			log.Root.Info("start CPU profile")
		}
		defer func() {
			_ = cpuProfile.Close()
			pprof.StopCPUProfile()
		}()

		//MEM profile
		memProfile, err := os.Create(filepath.Join(profDir, "mem.prof"))
		if err != nil {
			log.Root.Error("could not create memory profile: ", err)
		} else {
			log.Root.Info("create MEM profile: ", memProfile.Name())
		}
		runtime.GC()
		if err := pprof.WriteHeapProfile(memProfile); err != nil {
			log.Root.Error("could not write memory profile: ", err)
		} else {
			log.Root.Info("start MEM profile")
		}
		defer func() {
			_ = memProfile.Close()
		}()

		go func() {
			//view result in http://localhost:6060/debug/pprof/
			log.Root.Info(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// save accounts to context
	chainContext.SetAccounts(accounts)
	// start all services by chain context
	err = chainContext.Init(func() error {
		return chain.RegisterServices(chainContext)
	})
	if err != nil {
		log.Root.Error(err)
		return err
	}
	err = chainContext.Start()

	if err != nil {
		return err
	}
	trapSignal()
	return nil
}

func trapSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	chainContext := context.NewChainContext(cfgPathP)
	err := chainContext.Stop()
	if err != nil {
		log.Root.Info(err)
	}

	log.Root.Info("qlc node closed successfully")
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

	configParams = cmdutil.Flag{
		Name:  "configParam",
		Must:  false,
		Usage: "parameter set that needs to be changed",
		Value: "",
	}

	genesisSeed = cmdutil.Flag{
		Name:  "genesisSeed",
		Must:  false,
		Usage: "genesis seed",
		Value: "",
	}

	s := &ishell.Cmd{
		Name: "run",
		Help: "start qlc server",
		Func: func(c *ishell.Context) {
			args := []cmdutil.Flag{seed, cfgPath, isProfile}
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
			configParamsP = cmdutil.StringVar(c.Args, configParams)
			genesisSeedP = cmdutil.StringVar(c.Args, genesisSeed)

			err := start()
			if err != nil {
				cmdutil.Warn(err)
			}
		},
	}
	shell.AddCmd(s)
}

func generateChainTokenGenesisBlock(seedString string, cfg *config.Config) error {
	chainTokenInfos := &config.GenesisInfo{
		ChainToken: true,
		GasToken:   false,
	}
	genesisTime := time.Unix(1573208071, 0)
	bytes, _ := hex.DecodeString(seedString)
	seed, _ := types.BytesToSeed(bytes)
	account, _ := seed.Account(0)
	tokenName := "QLC"
	tokenSymbol := "QLC"
	decimals := uint8(8)
	address := account.Address()
	tokenHash := cabi.NewTokenHash(address, types.ZeroHash, tokenName)
	var totalSupply = big.NewInt(6e16)
	mintageData, err := cabi.MintageABI.PackMethod(cabi.MethodNameMintage, tokenHash, tokenName, tokenSymbol, totalSupply, decimals, address, "")
	if err != nil {
		return err
	}
	send := types.StateBlock{
		Type:           types.ContractSend,
		Address:        types.MintageAddress,
		Link:           address.ToHash(),
		Balance:        types.ZeroBalance,
		Vote:           types.ZeroBalance,
		Network:        types.ZeroBalance,
		Storage:        types.ZeroBalance,
		Oracle:         types.ZeroBalance,
		Token:          tokenHash,
		Data:           mintageData,
		Representative: address,
		Timestamp:      genesisTime.Add(time.Second * 1).Unix(),
	}
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()
	h1 := send.GetHash()
	send.Signature = account.Sign(h1)
	chainTokenInfos.Mintage = send

	genesisData, err := cabi.MintageABI.PackVariable(cabi.VariableNameToken, tokenHash, tokenName, tokenSymbol, totalSupply,
		decimals, address, big.NewInt(0), int64(0), address, "")

	if err != nil {
		return err
	}

	receive := types.StateBlock{
		Type:           types.ContractReward,
		Previous:       types.ZeroHash,
		Address:        address,
		Link:           h1,
		Balance:        types.Balance{Int: totalSupply},
		Vote:           types.ZeroBalance,
		Network:        types.ZeroBalance,
		Storage:        types.ZeroBalance,
		Oracle:         types.ZeroBalance,
		Token:          tokenHash,
		Data:           genesisData,
		Representative: address,
		Timestamp:      genesisTime.Add(time.Second * 10).Unix(),
	}
	//var w types.Work
	worker2, _ := types.NewWorker(w, receive.Root())
	receive.Work = worker2.NewWork()
	h2 := receive.GetHash()
	receive.Signature = account.Sign(h2)
	chainTokenInfos.Genesis = receive
	cfg.Genesis.GenesisBlocks = append(cfg.Genesis.GenesisBlocks, chainTokenInfos)

	return nil
}

func generateGasTokenGenesisBlock(seedString string, cfg *config.Config) error {
	genesisGasTokenInfos := &config.GenesisInfo{
		ChainToken: false,
		GasToken:   true,
	}
	genesisTime := time.Unix(1573208071, 0)
	bytes, _ := hex.DecodeString(seedString)
	seed, _ := types.BytesToSeed(bytes)
	repAccount, _ := seed.Account(0)
	account, _ := seed.Account(1)
	tokenName := "QGAS"
	tokenSymbol := "QGAS"
	decimals := uint8(8)
	repAddress := repAccount.Address()
	address := account.Address()
	tokenHash := cabi.NewTokenHash(address, types.ZeroHash, tokenName)
	var totalSupply = big.NewInt(1e16)
	mintageData, err := cabi.MintageABI.PackMethod(cabi.MethodNameMintage, tokenHash, tokenName, tokenSymbol, totalSupply, decimals, address, "")
	if err != nil {
		return err
	}
	send := types.StateBlock{
		Type:           types.ContractSend,
		Address:        types.MintageAddress,
		Link:           address.ToHash(),
		Balance:        types.ZeroBalance,
		Vote:           types.ZeroBalance,
		Network:        types.ZeroBalance,
		Storage:        types.ZeroBalance,
		Oracle:         types.ZeroBalance,
		Token:          tokenHash,
		Data:           mintageData,
		Representative: repAddress,
		Timestamp:      genesisTime.Add(time.Second * 1).Unix(),
	}
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()
	h1 := send.GetHash()
	send.Signature = account.Sign(h1)
	genesisGasTokenInfos.Mintage = send

	genesisData, err := cabi.MintageABI.PackVariable(cabi.VariableNameToken, tokenHash, tokenName, tokenSymbol, totalSupply,
		decimals, address, big.NewInt(0), int64(0), address, "")

	if err != nil {
		return err
	}

	receive := types.StateBlock{
		Type:           types.ContractReward,
		Previous:       types.ZeroHash,
		Address:        address,
		Link:           h1,
		Balance:        types.Balance{Int: totalSupply},
		Vote:           types.ZeroBalance,
		Network:        types.ZeroBalance,
		Storage:        types.ZeroBalance,
		Oracle:         types.ZeroBalance,
		Token:          tokenHash,
		Data:           genesisData,
		Representative: repAddress,
		Timestamp:      genesisTime.Add(time.Second * 10).Unix(),
	}
	//var w types.Work
	worker2, _ := types.NewWorker(w, receive.Root())
	receive.Work = worker2.NewWork()
	h2 := receive.GetHash()
	receive.Signature = account.Sign(h2)
	genesisGasTokenInfos.Genesis = receive
	cfg.Genesis.GenesisBlocks = append(cfg.Genesis.GenesisBlocks, genesisGasTokenInfos)

	return nil
}

func loadGenesisAccount(cfg *config.Config) {
	if len(common.GenesisInfos) > 0 {
		common.GenesisInfos = common.GenesisInfos[:0]
	}

	for _, v := range cfg.Genesis.GenesisBlocks {
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: v.Mintage,
			GenesisBlock:        v.Genesis,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}
}

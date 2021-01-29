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
	"strings"
	"syscall"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/chain/context"
	cmdutil "github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/common/vmcontract/mintage"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
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
	profPortP     int16
	isSingleP     bool
	configParamsP string
	genesisSeedP  string

	privateKey   cmdutil.Flag
	account      cmdutil.Flag
	password     cmdutil.Flag
	seed         cmdutil.Flag
	cfgPath      cmdutil.Flag
	isProfile    cmdutil.Flag
	profPort     cmdutil.Flag
	isSingle     cmdutil.Flag
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
		rootCmd.PersistentFlags().Int16Var(&profPortP, "profPort", 6060, "profile port")
		rootCmd.PersistentFlags().StringVar(&configParamsP, "configParams", "", "parameter set that needs to be changed")
		rootCmd.PersistentFlags().StringVar(&genesisSeedP, "genesisSeed", "", "genesis seed")
		rootCmd.PersistentFlags().BoolVar(&isSingleP, "single", false, "single node, no PoV, no bootnodes")
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
		if len(configParamsP) > 0 {
			params := strings.Split(configParamsP, ";")

			if len(params) > 0 {
				_, err := cm.PatchParams(params, cfg)
				if err != nil {
					return err
				}
			}
		}

		if isSingleP {
			// clear boot nodes
			cfg.P2P.BootNodes = cfg.P2P.BootNodes[:0]
			log.Root.Debug("clear all boot nodes...")
			//enable RPC
			cfg.RPC.Enable = true
			// run genesisSeed account
			if seedP == "" && genesisSeedP != "" {
				seedP = genesisSeedP
			}
		}

		return nil
	})
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
		go func() {
			//view result in http://localhost:6060/debug/pprof/
			addr := fmt.Sprintf("localhost:%d", profPortP)
			log.Root.Info(http.ListenAndServe(addr, nil))
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

	if isSingleP {
		cfg, _ := cm.Config()
		chainContext.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
		chainContext.EventBus().Publish(topic.EventAddP2PStream, &topic.EventAddP2PStreamMsg{PeerID: cfg.P2P.ID.PeerID})
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
	profPort = cmdutil.Flag{
		Name:  "profPort",
		Must:  false,
		Usage: "prof port ",
		Value: "",
	}
	args := []cmdutil.Flag{account, seed, cfgPath, isProfile, profPort, password, privateKey,
		configParams, genesisSeed}
	s := &ishell.Cmd{
		Name:                "run",
		Help:                "start qlc server",
		CompleterWithPrefix: cmdutil.OptsCompleter(args),
		Func: func(c *ishell.Context) {
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
			isSingleP = cmdutil.BoolVar(c.Args, isSingle)
			profPortTmp, _ := cmdutil.IntVar(c.Args, profPort)
			if profPortTmp > 0 {
				profPortP = int16(profPortTmp)
			} else {
				profPortTmp = 6060
			}
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
	tokenHash := mintage.NewTokenHash(address, types.ZeroHash, tokenName)
	var totalSupply = big.NewInt(6e16)
	mintageData, err := mintage.MintageABI.PackMethod(mintage.MethodNameMintage, tokenHash, tokenName, tokenSymbol, totalSupply, decimals, address, "")
	if err != nil {
		return err
	}
	send := types.StateBlock{
		Type:    types.ContractSend,
		Address: contractaddress.MintageAddress,
		Link:    address.ToHash(),
		Balance: types.ZeroBalance,
		//Vote:           types.ZeroBalance,
		//Network:        types.ZeroBalance,
		//Storage:        types.ZeroBalance,
		//Oracle:         types.ZeroBalance,
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

	genesisData, err := mintage.MintageABI.PackVariable(mintage.VariableNameToken, tokenHash, tokenName, tokenSymbol, totalSupply,
		decimals, address, big.NewInt(0), int64(0), address, "")

	if err != nil {
		return err
	}

	receive := types.StateBlock{
		Type:     types.ContractReward,
		Previous: types.ZeroHash,
		Address:  address,
		Link:     h1,
		Balance:  types.Balance{Int: totalSupply},
		//Vote:           types.ZeroBalance,
		//Network:        types.ZeroBalance,
		//Storage:        types.ZeroBalance,
		//Oracle:         types.ZeroBalance,
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
	tokenHash := mintage.NewTokenHash(address, types.ZeroHash, tokenName)
	var totalSupply = big.NewInt(1e16)
	mintageData, err := mintage.MintageABI.PackMethod(mintage.MethodNameMintage, tokenHash, tokenName, tokenSymbol, totalSupply, decimals, address, "")
	if err != nil {
		return err
	}
	send := types.StateBlock{
		Type:    types.ContractSend,
		Address: contractaddress.MintageAddress,
		Link:    address.ToHash(),
		Balance: types.ZeroBalance,
		//Vote:           types.ZeroBalance,
		//Network:        types.ZeroBalance,
		//Storage:        types.ZeroBalance,
		//Oracle:         types.ZeroBalance,
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

	genesisData, err := mintage.MintageABI.PackVariable(mintage.VariableNameToken, tokenHash, tokenName, tokenSymbol, totalSupply,
		decimals, address, big.NewInt(0), int64(0), address, "")

	if err != nil {
		return err
	}

	receive := types.StateBlock{
		Type:     types.ContractReward,
		Previous: types.ZeroHash,
		Address:  address,
		Link:     h1,
		Balance:  types.Balance{Int: totalSupply},
		//Vote:           types.ZeroBalance,
		//Network:        types.ZeroBalance,
		//Storage:        types.ZeroBalance,
		//Oracle:         types.ZeroBalance,
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

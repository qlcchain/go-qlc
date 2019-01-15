package commands

import (
	"fmt"
	"os"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/log"

	"github.com/spf13/cobra"
)

var (
	account  string
	pwd      string
	cfgPath  string
	ctx      *chain.QlcContext
	services []common.Service
	logger   = log.NewLogger("cli")
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
	Use:   "QLCChain",
	Short: "CLI for QLCChain.",
	Long:  `QLC Chain is the next generation public blockchain designed for the NaaS.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := start()
		if err != nil {
			cmd.Println(err)
		}

	},
}

func start() error {
	var addr types.Address
	if cfgPath == "" {
		cfgPath = config.DefaultDataDir()
	}
	cm := config.NewCfgManager(cfgPath)
	cfg, err := cm.Load()
	if err != nil {
		return err
	}
	if account == "" {
		addr = types.ZeroAddress
	} else {
		addr, err = types.HexToAddress(account)
		if err != nil {
			logger.Fatal(err)
		}
	}
	err = runNode(addr, pwd, cfg)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "config file")
	rootCmd.PersistentFlags().StringVarP(&account, "account", "a", "", "wallet address,if is nil,just run a node")
	rootCmd.PersistentFlags().StringVarP(&pwd, "password", "p", "", "password for wallet")
}

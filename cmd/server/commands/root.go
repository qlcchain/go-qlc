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

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/spf13/cobra"
)

var (
	account   string
	pwd       string
	isProfile bool
	cfgPath   string
	ctx       *chain.QlcContext
	services  []common.Service
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
			fmt.Println(err)
		}
	}

	if isProfile {
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
	rootCmd.PersistentFlags().BoolVar(&isProfile, "profile", false, "enable profile")
}

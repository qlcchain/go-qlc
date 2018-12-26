/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"flag"
	"fmt"
	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/wallet"
	"os"
	"os/signal"
	"reflect"
	"strconv"
)

var (
	version   string
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
	ctx       *chain.QlcContext
	services  []common.Service
	logger    = log.NewLogger("main")
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	fmt.Printf("qo-qlc %s-%s.%s", version, sha1ver, buildTime)

	//	var h bool
	//	var file string
	//	var password string
	//	var accString string
	//	var account types.Address
	//	flag.BoolVar(&h, "h", false, "print help message")
	//	flag.StringVar(&file, "config", "", "config file")
	//	flag.StringVar(&accString, "account", "", "wallet account")
	//	flag.StringVar(&password, "pwd", "", "wallet account password")
	//	flag.Parse()
	//
	//	if len(file) == 0 {
	//		file = config.DefaultConfigFile()
	//	}
	//
	//	if len(accString) > 0 {
	//		var err error
	//		account, err = types.HexToAddress(accString)
	//		if err != nil {
	//			logger.Error(err)
	//		}
	//	}
	//
	//	err := initNode(file, account, password)
	//	if err != nil {
	//		logger.Error(err)
	//	}
	//
	//	flag.Usage = func() {
	//		_, _ = fmt.Fprintf(os.Stdout, `gqlc version: %s-%s.%s
	//Usage: gqlc [-config filename] [-h]
	//
	//Options:
	//`, version, sha1ver, buildTime)
	//		flag.PrintDefaults()
	//	}
	//
	//	if h {
	//		flag.Usage()
	//	}

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	pwd := createCmd.String("pwd", "", "password")

	importCmd := flag.NewFlagSet("import", flag.ExitOnError)
	seed := importCmd.String("seed", "", "seed")
	importPwd := importCmd.String("pwd", "", "password")

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCmd.String("from", "", "transfer from")
	sendTo := sendCmd.String("to", "", "transfer to")
	sendToken := sendCmd.String("token", "", "transfer token")
	sendAmount := sendCmd.String("amount", "", "transfer amount")
	sendPwd := sendCmd.String("pwd", "", "password")

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	addr := runCmd.String("addr", "", "account string")
	runPwd := runCmd.String("pwd", "", "")

	var services []common.Service
	//if len(os.Args) < 2 {
	//	logger.Error("invalid args.")
	//	return
	//}
	switch os.Args[1] {
	case "create":
		err := createCmd.Parse(os.Args[2:])
		if err != nil {
			logger.Fatal(err)
		}
	case "import":
		err := importCmd.Parse(os.Args[2:])
		if err != nil {
			logger.Fatal(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			logger.Fatal(err)
		}
	case "run":
		err := runCmd.Parse(os.Args[2:])
		if err != nil {
			logger.Fatal(err)
		}
	default:
		//logger.Errorf("invalid args %s", os.Args[1])
		return
	}

	if createCmd.Parsed() {
		err := initNode(types.ZeroAddress, "")
		if err != nil {
			logger.Error(err)
		}
		err = Create(*pwd)
		if err != nil {
			logger.Error(err)
		}
		return
	}

	if importCmd.Parsed() {
		if len(*seed) == 0 {
			logger.Fatal("invalid seed")
		}
		err := initNode(types.ZeroAddress, "")
		if err != nil {
			logger.Error(err)
		}
		err = Import(*seed, *importPwd)
		if err != nil {
			logger.Error(err)
		}
		return
	}

	if runCmd.Parsed() {
		if len(*addr) == 0 {
			logger.Fatal("invalid account")
		}

		account, err := types.HexToAddress(*addr)
		if err != nil {
			logger.Fatal(err)
		}

		err = initNode(account, *runPwd)
		if err != nil {
			logger.Fatal(err)
		}

		services, err := startNode()
		// Block until a signal is received.
		s := <-c
		fmt.Println("Got signal:", s)
		stopNode(services)
		return
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == "" || *sendToken == "" {
			logger.Fatal("err transfer info")
		}
		source, err := types.HexToAddress(*sendFrom)
		if err != nil {
			logger.Error(err)
		}
		to, err := types.HexToAddress(*sendTo)
		if err != nil {
			logger.Error(err)
		}
		token, err := types.NewHash(*sendToken)
		if err != nil {
			logger.Error(err)
		}

		u, err := strconv.ParseUint(*sendAmount, 10, 64)
		if err != nil {
			logger.Error(err)
		}

		amount := types.ParseBalanceInts(uint64(0), u)
		err = initNode(source, *sendPwd)
		if err != nil {
			logger.Error(err)
		}
		services, err = startNode()
		if err != nil {
			logger.Error(err)
		}

		send(source, to, token, amount, *sendPwd)
		// Block until a signal is received.
		s := <-c
		fmt.Println("Got signal:", s)
		stopNode(services)
	}
}

func initNode(account types.Address, password string) error {
	cm := config.NewCfgManager(config.DefaultConfigFile())
	cfg, err := cm.Load()

	ctx, err = chain.New(cfg)
	if err != nil {
		logger.Fatal()
	}
	ctx.Ledger = ledger.NewLedgerService(cfg)
	ctx.Wallet = wallet.NewWalletService(cfg)
	ctx.NetService, err = p2p.NewQlcService(cfg)
	logService := log.NewLogService(cfg)
	if err != nil {
		return err
	}
	ctx.DPosService, err = consensus.NewDposService(cfg, ctx.NetService, account, password)
	if err != nil {
		return err
	}

	if !account.IsZero() {
		_ = ctx.NetService.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventPublish, func(v interface{}) {

			if b, ok := v.(*types.StateBlock); ok {
				if b.Address.ToHash() != b.Link {
					s := ctx.Wallet.Wallet.NewSession(account)
					if isValid, err := s.VerifyPassword(password); isValid && err == nil {
						accounts, err := s.GetAccounts()
						if err != nil {
							logger.Error(err)
						}

						for _, a := range accounts {
							if a.ToHash() == b.Link {
								logger.Debugf("receive block from [%s] to[%s] amount[%d]", b.Address.String(), a.String(), b.Balance.BigInt())
								break
							}
						}
					}
				}
			}
		})
	}

	//ctx.NetService.MessageEvent().GetEvent("consensus").UnSubscribe(p2p.EventPublish, sub)

	services = []common.Service{logService, ctx.NetService, ctx.DPosService, ctx.Ledger, ctx.Wallet}

	return nil
}

func startNode() ([]common.Service, error) {
	for _, service := range services {
		err := service.Init()
		if err != nil {
			return nil, err
		}
		err = service.Start()
		if err != nil {
			return nil, err
		}
		logger.Debugf("%s start successful.", reflect.TypeOf(service))
	}

	return services, nil
}

func stopNode(services []common.Service) {
	for _, service := range services {
		err := service.Stop()
		if err != nil {
			logger.Error(err)
		}
	}
}

func Create(password string) error {
	w := ctx.Wallet.Wallet
	address, err := w.NewWallet()
	if err != nil {
		return err
	}

	if len(password) > 0 {
		if err := w.NewSession(address).ChangePassword(password); err != nil {
			return err
		}
	}
	logger.Info("create wallet: address=>%s, password=>%s", address.String(), password)
	return nil
}

func Import(seed string, password string) error {
	w := ctx.Wallet.Wallet

	if addr, err := w.NewWalletBySeed(seed, password); err != nil {
		return err
	} else {
		logger.Infof("import seed[%s] password[%s] => %s", seed, password, addr.String())
	}
	return nil
}

func send(from, to types.Address, token types.Hash, amount types.Balance, password string) {
	w := ctx.Wallet.Wallet
	l := ctx.Ledger.Ledger
	logger.Debug(from.String())
	session := w.NewSession(from)

	if b, err := session.VerifyPassword(password); b && err == nil {
		sendblock, err := session.GenerateSendBlock(from, token, to, amount)
		if err != nil {
			logger.Fatal(err)
		}
		if l.Process(sendblock) == ledger.Other {
			logger.Fatal()
		}
		logger.Info("send block, ", sendblock.GetHash())

		net, err := p2p.NewQlcService(ctx.Config)
		blockBytes, err := sendblock.MarshalMsg(nil)
		net.Broadcast(p2p.PublishReq, blockBytes)
	} else {
		logger.Error("invalid password ", err, " valid: ", b)
	}
}

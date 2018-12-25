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
	"os"
	"time"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/wallet"
)

var logger = log.NewLogger("main")

var ctx *chain.QlcContext

var (
	version   string
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

func main() {
	var h bool
	var file string
	flag.BoolVar(&h, "h", false, "print help message")
	flag.StringVar(&file, "config", "", "config file")
	//flag.StringVar(&file, "c", "", "config file(shorthand)")
	flag.Parse()

	if len(file) == 0 {
		file = config.DefaultConfigFile()
	}
	fmt.Println("cfg file: ", file)

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stdout, `gqlc version: %s-%s.%s
Usage: gqlc [-config filename] [-h]

Options:
`, version, sha1ver, buildTime)
		flag.PrintDefaults()
	}

	if h {
		flag.Usage()
	}

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	seed := createCmd.String("seed", "", "seed")
	pwd := createCmd.String("pwd", "", "password")

	importCmd := flag.NewFlagSet("import", flag.ExitOnError)
	password := importCmd.String("password", "", "password")

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCmd.String("from", "", "transfer from")
	sendTo := sendCmd.String("to", "", "transfer to")
	sendToken := sendCmd.String("token", "", "transfer token")
	sendAmount := sendCmd.String("amount", "", "transfer amount")

	switch os.Args[1] {
	case "create":
		err := createCmd.Parse(os.Args[2:])
		if err != nil {
			logger.Fatal()
		}
	case "import":
		err := importCmd.Parse(os.Args[2:])
		if err != nil {
			logger.Fatal()
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			logger.Fatal()
		}
	default:

	}

	setQlcContext()

	if createCmd.Parsed() {
		Create(*seed, *pwd)
	}

	if importCmd.Parsed() {
		if *password == "" {
			logger.Fatal("password is nil")
		}
		Import(*password)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == "" || *sendToken == "" {
			logger.Fatal("err transfer info")
		}
		go send(*sendFrom, *sendTo, *sendToken, *sendAmount)
	}

	go Receive()

	select {}
}

func setQlcContext() {
	dir := "inter-test"
	logger.Info(dir)
	cfg, err := config.DefaultConfig()
	if err != nil {
		logger.Fatal()
	}
	cfg.DataDir = dir
	ctx, err = chain.New(cfg)
	if err != nil {
		logger.Fatal()
	}
	ctx.Ledger = ledger.NewLedgerService(cfg)
	ctx.Wallet = wallet.NewWalletService(cfg)

}

func getService() (*wallet.Session, *ledger.Ledger) {
	wallet := ctx.Wallet.Wallet
	id, _ := wallet.NewWallet()
	session := wallet.NewSession(id)
	ledger := ctx.Ledger.Ledger
	return session, ledger

}

func Create(seed, password string) error {
	logger.Info(seed, password)
	wallet, _ := getService()

	if seed != "" {
		if err := wallet.SetSeed([]byte(seed)); err != nil {
			return err
		}
	}
	if password != "" {
		if err := wallet.ChangePassword(password); err != nil {
			return err
		}
	}
	return nil
}

func Import(password string) error {
	logger.Info(password)

	wallet, _ := getService()

	if err := wallet.EnterPassword(password); err != nil {
		return err
	}
	return nil
}

func send(fromStr, toStr, tokenStr, amountStr string) {
	logger.Info(fromStr, toStr, tokenStr, amountStr)
	wallet, l := getService()

	source, err := types.HexToAddress(fromStr)
	if err != nil {
		logger.Fatal()
	}
	to, err := types.HexToAddress(toStr)
	if err != nil {
		logger.Fatal()
	}
	token, err := types.NewHash(tokenStr)
	if err != nil {
		logger.Fatal()
	}
	amount, err := types.ParseBalanceString(amountStr)
	if err != nil {
		logger.Fatal()
	}
	sendblock, err := wallet.GenerateSendBlock(source, token, to, amount)
	if err != nil {
		logger.Fatal()
	}
	if l.Process(sendblock) == ledger.Other {
		logger.Fatal()
	}
	logger.Info("receive block, ", sendblock.GetHash())

	net, err := p2p.NewQlcService(ctx.Config, ctx.Ledger.Ledger)
	blockBytes, err := sendblock.MarshalMsg(nil)
	net.Broadcast(p2p.PublishReq, blockBytes)
}

func Receive() {
	wallet, l := getService()

	addrs, err := wallet.GetAccounts()
	if err != nil {
		logger.Fatal()
	}

	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		for _, addr := range addrs {
			pendings, err := l.Pending(addr)
			if err != nil {
				logger.Fatal()
			}
			for _, pending := range pendings {
				sendhash := pending.Hash
				sendblock, err := l.GetBlock(sendhash)
				if err != nil {
					logger.Fatal()
				}
				receiveblock, err := wallet.GenerateReceiveBlock(sendblock)
				if err != nil {
					logger.Fatal()
				}

				if l.Process(receiveblock) == ledger.Other {
					logger.Fatal()
				}
				logger.Info("receive block, ", receiveblock.GetHash())

				net, err := p2p.NewQlcService(ctx.Config, ctx.Ledger.Ledger)
				blockBytes, err := receiveblock.MarshalMsg(nil)
				net.Broadcast(p2p.PublishReq, blockBytes)
			}
		}
	}
}

// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"errors"
	"fmt"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/test/mock"

	"github.com/spf13/cobra"
	cmn "github.com/tendermint/tmlibs/common"
)

var (
	from   string
	to     string
	token  string
	amount string
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "send transaction",
	Run: func(cmd *cobra.Command, args []string) {
		err := sendAction()
		if err != nil {
			cmd.Println(err)
		}
	},
}

func init() {
	sendCmd.Flags().StringVarP(&from, "from", "f", "", "send account")
	sendCmd.Flags().StringVarP(&to, "to", "t", "", "receive account")
	sendCmd.Flags().StringVarP(&token, "token", "k", mock.GetChainTokenType().String(), "token hash for send action")
	sendCmd.Flags().StringVarP(&amount, "amount", "m", "", "send amount")
	rootCmd.AddCommand(sendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func sendAction() error {
	if from == "" || to == "" || amount == "" {
		fmt.Println("err transfer info")
		return errors.New("err transfer info")
	}
	source, err := types.HexToAddress(from)
	if err != nil {
		fmt.Println(err)
		return err
	}
	t, err := types.HexToAddress(to)
	if err != nil {
		fmt.Println(err)
		return err
	}
	tk, err := types.NewHash(token)
	if err != nil {
		fmt.Println(err)
		return err
	}

	am := types.StringToBalance(amount)
	if cfgPath == "" {
		cfgPath = config.DefaultDataDir()
	}
	cm := config.NewCfgManager(cfgPath)
	cfg, err := cm.Load()
	if err != nil {
		return err
	}
	err = initNode(source, pwd, cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	services, err = startNode()
	if err != nil {
		fmt.Println(err)
	}
	send(source, t, tk, am, pwd)
	cmn.TrapSignal(func() {
		stopNode(services)
	})
	return nil
}

func send(from, to types.Address, token types.Hash, amount types.Balance, password string) {
	w := ctx.Wallet.Wallet
	l := ctx.Ledger.Ledger
	n := ctx.NetService
	fmt.Println(from.String())
	session := w.NewSession(from)

	if b, err := session.VerifyPassword(password); b && err == nil {
		a, _ := mock.BalanceToRaw(amount, "QLC")
		sendBlock, err := session.GenerateSendBlock(from, token, to, a)
		if err != nil {
			fmt.Println(err)
		}

		if r, err := l.Process(sendBlock); err != nil || r == ledger.Other {
			fmt.Println(jsoniter.MarshalToString(&sendBlock))
			fmt.Println("process block error", err)
		} else {
			fmt.Println("send block, ", sendBlock.GetHash())

			meta, err := l.GetAccountMeta(from)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(jsoniter.MarshalToString(&meta))
			pushBlock := protos.PublishBlock{
				Blk: sendBlock,
			}
			bytes, err := protos.PublishBlockToProto(&pushBlock)
			if err != nil {
				fmt.Println(err)
			} else {
				n.Broadcast(p2p.PublishReq, bytes)
			}
		}
	} else {
		fmt.Println("invalid password ", err, " valid: ", b)
	}
}

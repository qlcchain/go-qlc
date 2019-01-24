/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"strings"

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
	fromAccount string
	toAccounts  []string
	sendToken   string
	sendAmount  string
)

// sendCmd represents the send command
var batchSendCmd = &cobra.Command{
	Use:   "batchsend",
	Short: "batch send transaction",
	Run: func(cmd *cobra.Command, args []string) {
		err := batchSendAction()
		if err != nil {
			cmd.Println(err)
		}
	},
}

func init() {
	batchSendCmd.Flags().StringVar(&fromAccount, "from", "", "send account private key")
	batchSendCmd.Flags().StringSliceVar(&toAccounts, "to", toAccounts, "receive accounts")
	batchSendCmd.Flags().StringVar(&sendToken, "token", mock.GetChainTokenType().String(), "token hash for send action")
	batchSendCmd.Flags().StringVar(&sendAmount, "amount", "", "send amount")
	rootCmd.AddCommand(batchSendCmd)
}

func batchSendAction() error {
	if fromAccount == "" || len(toAccounts) == 0 || len(sendAmount) == 0 {
		fmt.Println("err transfer info")
		return errors.New("err transfer info")
	}
	bytes, err := hex.DecodeString(fromAccount)
	if err != nil {
		fmt.Println(err)
		return err
	}
	account := types.NewAccount(bytes)

	tk, err := types.NewHash(sendToken)
	if err != nil {
		fmt.Println(err)
		return err
	}

	am := types.StringToBalance(sendAmount)
	if cfgPath == "" {
		cfgPath = config.DefaultDataDir()
	}

	cm := config.NewCfgManager(cfgPath)
	cfg, err := cm.Load()
	if err != nil {
		return err
	}
	err = initNode(types.Address{}, "", cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	services, err = startNode()
	if err != nil {
		fmt.Println(err)
	}

	for _, toAccount := range toAccounts {
		t, err := types.HexToAddress(strings.TrimSpace(toAccount))
		if err != nil {
			fmt.Println(err)
			continue
		}

		sendTx(t, tk, am, account)
	}

	cmn.TrapSignal(func() {
		stopNode(services)
	})
	return nil
}

func sendTx(to types.Address, token types.Hash, amount types.Balance, account *types.Account) {
	l := ctx.Ledger.Ledger
	n := ctx.NetService

	a, _ := mock.BalanceToRaw(amount, "QLC")
	sendBlock, err := generateSendBlock(account, token, to, a)
	from := account.Address()
	fmt.Printf("send from %s token[%s] to %s\n", from.String(), token.String(), to.String())

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(jsoniter.MarshalToString(&sendBlock))

	if r, err := l.Process(sendBlock); err != nil || r != ledger.Progress {
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
}

// TODO: remove
func generateSendBlock(account *types.Account, token types.Hash, to types.Address, amount types.Balance) (types.Block, error) {
	l := ctx.Ledger.Ledger
	source := account.Address()
	tm, err := l.GetTokenMeta(source, token)
	if err != nil {
		return nil, err
	}
	balance, err := l.TokenBalance(source, token)
	if err != nil {
		return nil, err
	}

	if balance.Compare(amount) == types.BalanceCompBigger {
		newBalance := balance.Sub(amount)
		sendBlock, _ := types.NewBlock(types.State)

		if sb, ok := sendBlock.(*types.StateBlock); ok {
			sb.Address = source
			sb.Token = token
			sb.Link = to.ToHash()
			sb.Balance = newBalance
			sb.Previous = tm.Header
			sb.Representative = tm.Representative
			sb.Work = generateWork(sb.Root())
			sb.Signature = account.Sign(sb.GetHash())
		}
		return sendBlock, nil
	} else {
		return nil, fmt.Errorf("%s not enought balance(%s) of %s", source.String(), balance, amount)
	}
}

// TODO: remove
func generateWork(hash types.Hash) types.Work {
	var work types.Work
	worker, _ := types.NewWorker(work, hash)
	return worker.NewWork()
}

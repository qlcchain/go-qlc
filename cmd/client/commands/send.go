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

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/rpc/api"
	"github.com/spf13/cobra"
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
	sendCmd.Flags().StringVarP(&from, "from", "f", "", "send account private key")
	sendCmd.Flags().StringVarP(&to, "to", "t", "", "receive account")
	sendCmd.Flags().StringVarP(&token, "token", "k", "QLC", "token for send action")
	sendCmd.Flags().StringVarP(&amount, "amount", "m", "", "send amount")
	rootCmd.AddCommand(sendCmd)
}

func sendAction() error {
	if from == "" || to == "" || amount == "" {
		fmt.Println("err transfer info")
		return errors.New("err transfer info")
	}
	bytes, err := hex.DecodeString(from)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fromAccount := types.NewAccount(bytes)

	t, err := types.HexToAddress(to)
	if err != nil {
		fmt.Println(err)
		return err
	}

	am := types.StringToBalance(amount)
	sendTx(fromAccount, t, token, am)
	return nil
}

func sendTx(account *types.Account, to types.Address, token string, amount types.Balance) error {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return err
	}
	defer client.Close()

	//a, _ := mock.BalanceToRaw(amount, "QLC")
	//fmt.Println(a)
	para := api.APISendBlockPara{
		Send:      account.Address(),
		TokenName: token,
		To:        to,
		Amount:    amount,
	}
	var sendBlock types.StateBlock
	err = client.Call(&sendBlock, "ledger_generateSendBlock", para, hex.EncodeToString(account.PrivateKey()))
	if err != nil {
		fmt.Println(err)
		return err
	}
	from := account.Address()
	fmt.Printf("send from %s token[%s] to %s\n", from.String(), token, to.String())

	fmt.Println(sendBlock.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &sendBlock)
	if err != nil {
		fmt.Println("process block error, ", err)
		return err
	}
	return nil
}

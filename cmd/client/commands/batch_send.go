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
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/spf13/cobra"
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
		cmd.Println("batch send done")
	},
}

func init() {
	batchSendCmd.Flags().StringVar(&fromAccount, "from", "", "send account private key")
	batchSendCmd.Flags().StringSliceVar(&toAccounts, "to", toAccounts, "receive accounts")
	batchSendCmd.Flags().StringVar(&sendToken, "token", "QLC", "token name for send action")
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
		return err
	}

	account := types.NewAccount(bytes)
	am := types.StringToBalance(sendAmount)

	for _, toAccount := range toAccounts {
		t, err := types.HexToAddress(strings.TrimSpace(toAccount))
		if err != nil {
			fmt.Println(err)
			continue
		}
		sendTx(account, t, sendToken, am)
	}
	return nil
}

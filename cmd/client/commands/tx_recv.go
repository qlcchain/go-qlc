package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/cmd/util"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/spf13/cobra"
)

func addTxRecvCmdByShell(parentCmd *ishell.Cmd) {
	priKey := util.Flag{
		Name:  "priKey",
		Must:  true,
		Usage: "account private hex string",
	}
	sendHash := util.Flag{
		Name:  "hash",
		Must:  true,
		Usage: "send block hash string",
	}

	cmd := &ishell.Cmd{
		Name: "recv",
		Help: "recv pending balance",
		Func: func(c *ishell.Context) {
			args := []util.Flag{priKey, sendHash}
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			priKeyP := util.StringVar(c.Args, priKey)
			sendHashP := util.StringVar(c.Args, sendHash)

			if err := recvAction(priKeyP, sendHashP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func addTxRecvCmdByCobra(parentCmd *cobra.Command) {
	var priKeyP string
	var sendHashP string

	var cmd = &cobra.Command{
		Use:   "recv",
		Short: "recv pending balance",
		Run: func(cmd *cobra.Command, args []string) {
			err := recvAction(priKeyP, sendHashP)
			if err != nil {
				cmd.Println(err)
			}
		},
	}
	cmd.Flags().StringVar(&priKeyP, "priKey", "", "account private hex string")
	cmd.Flags().StringVar(&sendHashP, "hash", "", "send block hash string")
	parentCmd.AddCommand(cmd)
}

func recvAction(priKeyP, sendHashP string) error {
	if priKeyP == "" {
		return errors.New("invalid priKey value")
	}

	if sendHashP == "" {
		return errors.New("invalid hash value")
	}

	accBytes, err := hex.DecodeString(priKeyP)
	if err != nil {
		return err
	}
	account := types.NewAccount(accBytes)
	if account == nil {
		return errors.New("can not new account")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	recv := types.StateBlock{}
	err = client.Call(&recv, "ledger_generateReceiveBlockByHash", sendHashP)
	if err != nil {
		return err
	}

	var w2 types.Work
	worker2, _ := types.NewWorker(w2, recv.Root())
	recv.Work = worker2.NewWork()

	recvHash := recv.GetHash()
	recv.Signature = account.Sign(recvHash)

	fmt.Printf("ReceiveBlock:\n%s\n", cutil.ToIndentString(recv))
	fmt.Println("address", recv.Address, "recvHash", recvHash)

	err = client.Call(nil, "ledger_process", &recv)
	if err != nil {
		return err
	}

	fmt.Printf("success to recv balance, account balance %s\n", recv.Balance)

	return nil
}

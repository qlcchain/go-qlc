package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/spf13/cobra"
)

func recv() {
	var accountP string
	var sendHashP string

	if interactive {
		account := util.Flag{
			Name:  "account",
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
				args := []util.Flag{account, sendHash}
				if util.HelpText(c, args) {
					return
				}
				err := util.CheckArgs(c, args)
				if err != nil {
					util.Warn(err)
					return
				}

				accountP = util.StringVar(c.Args, account)
				sendHashP = util.StringVar(c.Args, sendHash)

				if err := recvAction(accountP, sendHashP); err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "recv",
			Short: "recv pending balance",
			Run: func(cmd *cobra.Command, args []string) {
				err := recvAction(accountP, sendHashP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		cmd.Flags().StringVar(&accountP, "account", "", "account private hex string")
		cmd.Flags().StringVar(&sendHashP, "hash", "", "send block hash string")
		rootCmd.AddCommand(cmd)
	}
}

func recvAction(accountP, sendHashP string) error {
	if accountP == "" {
		return errors.New("invalid account value")
	}

	if sendHashP == "" {
		return errors.New("invalid hash value")
	}

	accBytes, err := hex.DecodeString(accountP)
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

package commands

import (
	"encoding/hex"
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
)

func addKYCAdminUpdateCmdByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "account to register (private key in hex string)",
		Value: "",
	}
	successor := util.Flag{
		Name:  "successor",
		Must:  true,
		Usage: "admin hand over to",
		Value: "",
	}
	comment := util.Flag{
		Name:  "comment",
		Must:  false,
		Usage: "admin comment",
		Value: "",
	}
	args := []util.Flag{account, successor, comment}
	c := &ishell.Cmd{
		Name:                "adminUpdate",
		Help:                "update admin comment or hand over admin",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}

			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			accountP := util.StringVar(c.Args, account)
			successorP := util.StringVar(c.Args, successor)
			commentP := util.StringVar(c.Args, comment)

			err := adminUpdate(accountP, successorP, commentP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func adminUpdate(account, successorP, comment string) error {
	if account == "" {
		return fmt.Errorf("account can not be null")
	}

	if successorP == "" {
		return fmt.Errorf("successor can not be null")
	}

	accBytes, err := hex.DecodeString(account)
	if err != nil {
		return err
	}

	acc := types.NewAccount(accBytes)
	if acc == nil {
		return fmt.Errorf("account format err")
	}

	successor, err := types.HexToAddress(successorP)
	if err != nil {
		return err
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	param := &api.KYCAdminUpdateParam{
		Admin:     acc.Address(),
		Successor: successor,
		Comment:   comment,
	}

	var block types.StateBlock
	err = client.Call(&block, "KYC_getAdminHandoverBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("send block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}

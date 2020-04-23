package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPtmKeyDeleteByAccountByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "account of ptmkey user (private key in hex string)",
		Value: "",
	}
	btype := util.Flag{
		Name:  "btype",
		Must:  false,
		Usage: "ptmkey bussiness type(default/invalid)",
		Value: "",
	}
	args := []util.Flag{account, btype}
	c := &ishell.Cmd{
		Name:                "delete",
		Help:                "delete ptm key",
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
			btypeP := util.StringVar(c.Args, btype)

			err := ptmKeyDelete(accountP, btypeP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func ptmKeyDelete(accountP, vBtypeP string) error {
	if accountP == "" {
		return fmt.Errorf("account can not be null")
	}

	accBytes, err := hex.DecodeString(accountP)
	if err != nil {
		return err
	}

	acc := types.NewAccount(accBytes)
	if acc == nil {
		return fmt.Errorf("account format err")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	param := &api.PtmKeyDeleteParam{
		Account: acc.Address(),
		Btype:   vBtypeP,
	}

	var block types.StateBlock
	err = client.Call(&block, "ptmkey_getPtmKeyDeleteBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("ptmkey update block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}

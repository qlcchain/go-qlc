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

func addPtmKeyUpdateCmdByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "account of ptmkey user(private key in hex string)",
		Value: "",
	}
	btype := util.Flag{
		Name:  "btype",
		Must:  true,
		Usage: "ptmkey bussiness type(default/invalid)",
		Value: "",
	}
	key := util.Flag{
		Name:  "key",
		Must:  true,
		Usage: "ptm's public key",
		Value: "",
	}
	args := []util.Flag{account, btype, key}
	c := &ishell.Cmd{
		Name:                "update",
		Help:                "update ptm key",
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
			keyP := util.StringVar(c.Args, key)

			err := ptmKeyUpdate(accountP, btypeP, keyP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func ptmKeyUpdate(accountP, vBtypeP, vKeyP string) error {
	if accountP == "" {
		return fmt.Errorf("account can not be null")
	}

	if vBtypeP == "" {
		return fmt.Errorf("ptmkey btype can not be null")
	}

	if vKeyP == "" {
		return fmt.Errorf("ptmkey key can not be null")
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
	//fmt.Printf("ptmkey update input:\naccount(%s)\nvBtypeP(%s)\nvKeyP(%s)\n",accountP,vBtypeP,vKeyP)
	param := &api.PtmKeyUpdateParam{
		Account: acc.Address(),
		Btype:   vBtypeP,
		Pubkey:  vKeyP,
	}

	var block types.StateBlock
	err = client.Call(&block, "ptmkey_getPtmKeyUpdateBlock", param)
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

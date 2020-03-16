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

func addUnPublishCmdByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "account to publish (private key in hex string)",
		Value: "",
	}
	typ := util.Flag{
		Name:  "type",
		Must:  true,
		Usage: "unPublish id type (email/weChat)",
		Value: "",
	}
	id := util.Flag{
		Name:  "id",
		Must:  true,
		Usage: "unPublish id (email address/weChat id)",
		Value: "",
	}
	kt := util.Flag{
		Name:  "kt",
		Must:  true,
		Usage: "publish public key type(ed25519/rsa4096)",
		Value: "",
	}
	pk := util.Flag{
		Name:  "pk",
		Must:  true,
		Usage: "unPublish public key",
		Value: "",
	}
	hash := util.Flag{
		Name:  "hash",
		Must:  true,
		Usage: "unPublish hash",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "unPublish",
		Help: "unPublish id and key",
		Func: func(c *ishell.Context) {
			args := []util.Flag{account, typ, id, kt, pk, hash}
			if util.HelpText(c, args) {
				return
			}

			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			accountP := util.StringVar(c.Args, account)
			typeP := util.StringVar(c.Args, typ)
			idP := util.StringVar(c.Args, id)
			ktP := util.StringVar(c.Args, kt)
			pkP := util.StringVar(c.Args, pk)
			hashP := util.StringVar(c.Args, hash)

			err := unPublish(accountP, typeP, idP, ktP, pkP, hashP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func unPublish(accountP, typeP, idP, ktP, pkP, hashP string) error {
	if accountP == "" {
		return fmt.Errorf("account can not be null")
	}

	if typeP == "" {
		return fmt.Errorf("unpublish type can not be null")
	}

	if idP == "" {
		return fmt.Errorf("unpublish id can not be null")
	}

	if ktP == "" {
		return fmt.Errorf("unpublish public key type can not be null")
	}

	if pkP == "" {
		return fmt.Errorf("unpublish public key can not be null")
	}

	if hashP == "" {
		return fmt.Errorf("unpublish hash can not be null")
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

	param := &api.UnPublishParam{
		Account: acc.Address(),
		PType:   typeP,
		PID:     idP,
		KeyType: ktP,
		PubKey:  pkP,
		Hash:    hashP,
	}

	var block types.StateBlock
	err = client.Call(&block, "dpki_getUnPublishBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("unPublish block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}

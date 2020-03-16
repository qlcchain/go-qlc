package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPublishCmdByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "account to publish (private key in hex string)",
		Value: "",
	}
	typ := util.Flag{
		Name:  "type",
		Must:  true,
		Usage: "publish id type (email/weChat)",
		Value: "",
	}
	id := util.Flag{
		Name:  "id",
		Must:  true,
		Usage: "publish id (email address/weChat id)",
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
		Usage: "publish public key",
		Value: "",
	}
	fee := util.Flag{
		Name:  "fee",
		Must:  true,
		Usage: "publish fee (at least 5 QGAS)",
		Value: "",
	}
	verifiers := util.Flag{
		Name:  "verifiers",
		Must:  true,
		Usage: "verifiers",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "publish",
		Help: "publish id and key",
		Func: func(c *ishell.Context) {
			args := []util.Flag{account, typ, id, kt, pk, fee, verifiers}
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
			feeP := util.StringVar(c.Args, fee)
			verifiersP := util.StringVar(c.Args, verifiers)

			err := publish(accountP, typeP, idP, ktP, pkP, feeP, verifiersP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func publish(accountP, typeP, idP, ktP, pkP, feeP, verifiersP string) error {
	if accountP == "" {
		return fmt.Errorf("account can not be null")
	}

	if typeP == "" {
		return fmt.Errorf("publish type can not be null")
	}

	if idP == "" {
		return fmt.Errorf("publish id can not be null")
	}

	if ktP == "" {
		return fmt.Errorf("publish public key type can not be null")
	}

	if pkP == "" {
		return fmt.Errorf("publish public key can not be null")
	}

	if feeP == "" {
		return fmt.Errorf("publish fee can not be null")
	}

	if verifiersP == "" {
		return fmt.Errorf("verifiers can not be null")
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

	verifiers := make([]types.Address, 0)
	err = json.Unmarshal([]byte(verifiersP), &verifiers)
	if err != nil {
		return err
	}

	param := &api.PublishParam{
		Account:   acc.Address(),
		PType:     typeP,
		PID:       idP,
		KeyType:   ktP,
		PubKey:    pkP,
		Fee:       types.StringToBalance(feeP),
		Verifiers: verifiers,
	}

	var blockInfo api.PublishRet
	err = client.Call(&blockInfo, "dpki_getPublishBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	block := blockInfo.Block
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("publish block:\n%s\nhash[%s]\nverifiers:\n%s\n", cutil.ToIndentString(block), block.GetHash(),
		cutil.ToIndentString(blockInfo.Verifiers))

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}

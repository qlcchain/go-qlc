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

func addVerifierRegisterCmdByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "account to register (private key in hex string)",
		Value: "",
	}
	vType := util.Flag{
		Name:  "type",
		Must:  true,
		Usage: "verifier type(email/weChat)",
		Value: "",
	}
	vInfo := util.Flag{
		Name:  "info",
		Must:  true,
		Usage: "verifiers address(email address/weChat ID)",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "register",
		Help: "register verifier",
		Func: func(c *ishell.Context) {
			args := []util.Flag{account, vType, vInfo}
			if util.HelpText(c, args) {
				return
			}

			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			accountP := util.StringVar(c.Args, account)
			vTypeP := util.StringVar(c.Args, vType)
			vInfoP := util.StringVar(c.Args, vInfo)

			err := verifierRegister(accountP, vTypeP, vInfoP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func verifierRegister(accountP, vTypeP, vInfoP string) error {
	if accountP == "" {
		return fmt.Errorf("account can not be null")
	}

	if vTypeP == "" {
		return fmt.Errorf("verifier type can not be null")
	}

	if vInfoP == "" {
		return fmt.Errorf("verifier info can not be null")
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

	param := &api.VerifierRegParam{
		Account: acc.Address(),
		VType:   vTypeP,
		VInfo:   vInfoP,
	}

	var block types.StateBlock
	err = client.Call(&block, "pkd_getVerifierRegisterBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("register block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}

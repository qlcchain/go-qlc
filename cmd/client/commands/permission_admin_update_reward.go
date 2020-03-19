package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
)

func addPermissionAdminUpdateRewardCmdByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "reward account (private key in hex string)",
		Value: "",
	}
	hash := util.Flag{
		Name:  "hash",
		Must:  true,
		Usage: "send block hash",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "adminUpdateReward",
		Help: "update admin comment or hand over admin",
		Func: func(c *ishell.Context) {
			args := []util.Flag{account, hash}
			if util.HelpText(c, args) {
				return
			}

			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			accountP := util.StringVar(c.Args, account)
			hashP := util.StringVar(c.Args, hash)

			err := adminUpdateReward(accountP, hashP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func adminUpdateReward(accountP, sendHashP string) error {
	if accountP == "" {
		return fmt.Errorf("account can not be null")
	}

	if sendHashP == "" {
		return fmt.Errorf("hash can not be null")
	}

	accBytes, err := hex.DecodeString(accountP)
	if err != nil {
		return err
	}

	acc := types.NewAccount(accBytes)
	if acc == nil {
		return fmt.Errorf("account format err")
	}

	sendHash, err := types.NewHash(sendHashP)
	if err != nil {
		return err
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	var block types.StateBlock
	err = client.Call(&block, "permission_getAdminUpdateRewardBlockBySendHash", sendHash)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("reward block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}

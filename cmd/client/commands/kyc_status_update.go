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

func addKYCStatusUpdateCmdByShell(parentCmd *ishell.Cmd) {
	operator := util.Flag{
		Name:  "operator",
		Must:  true,
		Usage: "operator user (private key in hex string)",
		Value: "",
	}
	address := util.Flag{
		Name:  "address",
		Must:  true,
		Usage: "user's chain address",
		Value: "",
	}
	status := util.Flag{
		Name:  "status",
		Must:  true,
		Usage: "kyc status",
		Value: "",
	}
	args := []util.Flag{operator, address, status}
	c := &ishell.Cmd{
		Name:                "statusUpdate",
		Help:                "update kyc status",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}

			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			operatorP := util.StringVar(c.Args, operator)
			addressP := util.StringVar(c.Args, address)
			statusP := util.StringVar(c.Args, status)

			err := statusUpdate(operatorP, addressP, statusP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func statusUpdate(operator, addressP, status string) error {
	if operator == "" {
		return fmt.Errorf("operator can not be null")
	}

	if addressP == "" {
		return fmt.Errorf("address can not be null")
	}

	accBytes, err := hex.DecodeString(operator)
	if err != nil {
		return err
	}

	acc := types.NewAccount(accBytes)
	if acc == nil {
		return fmt.Errorf("account format err")
	}

	address, err := types.HexToAddress(addressP)
	if err != nil {
		return fmt.Errorf("address format err")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	param := &api.KYCUpdateStatusParam{
		Operator:     acc.Address(),
		ChainAddress: address,
		Status:       status,
	}

	var block types.StateBlock
	err = client.Call(&block, "KYC_getUpdateStatusBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("node update block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}

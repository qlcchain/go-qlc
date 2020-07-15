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

func addKYCStatusUpdateCmdByShell(parentCmd *ishell.Cmd) {
	admin := util.Flag{
		Name:  "admin",
		Must:  true,
		Usage: "admin user (private key in hex string)",
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
	args := []util.Flag{admin, address, status}
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

			adminP := util.StringVar(c.Args, admin)
			addressP := util.StringVar(c.Args, address)
			statusP := util.StringVar(c.Args, status)

			err := statusUpdate(adminP, addressP, statusP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func statusUpdate(admin, addressP, status string) error {
	if admin == "" {
		return fmt.Errorf("admin can not be null")
	}

	if addressP == "" {
		return fmt.Errorf("address can not be null")
	}

	accBytes, err := hex.DecodeString(admin)
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
		Admin:        acc.Address(),
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

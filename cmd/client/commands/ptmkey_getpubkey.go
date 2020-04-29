package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	_ "github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
	_ "github.com/qlcchain/go-qlc/vm/contract/abi"
)

func addPtmKeyGetPubkeyCmdByShell(parentCmd *ishell.Cmd) {
	address := util.Flag{
		Name:  "address",
		Must:  true,
		Usage: "account address hex string",
		Value: "",
	}
	btype := util.Flag{
		Name:  "btype",
		Must:  false,
		Usage: "business type",
		Value: "default",
	}
	args := []util.Flag{address, btype}
	cmd := &ishell.Cmd{
		Name:                "getPtmkey",
		Help:                "get Ptm pubkey",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			addressP := util.StringVar(c.Args, address)
			btypeP := util.StringVar(c.Args, btype)

			if err := ptmKeyGetPubkeyAction(addressP, btypeP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func ptmKeyGetPubkeyAction(addressP, btypeP string) error {
	fmt.Println(addressP, btypeP)

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rsppks := make([]*api.PtmKeyUpdateParam, 0)
	pubAddr, err := types.HexToAddress(addressP)
	if err != nil {
		return err
	}
	if btypeP != "" {
		err = client.Call(&rsppks, "ptmkey_getPtmKeyByAccountAndBtype", pubAddr, btypeP)
	} else {
		err = client.Call(&rsppks, "ptmkey_getPtmKeyByAccount", pubAddr)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Total: 	%d\n", len(rsppks))
	for idx, pi := range rsppks {
		fmt.Printf("Index:	%d\n", idx)
		fmt.Printf("Account:%s\n", pi.Account)
		fmt.Printf("Btype: 	%s\n", pi.Btype)
		fmt.Printf("Pubkey:	%s\n", pi.Pubkey)
		fmt.Printf("\n")
	}

	return nil
}

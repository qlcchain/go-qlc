package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addDpkiGetPublishInfoCmdByShell(parentCmd *ishell.Cmd) {
	address := util.Flag{
		Name:  "address",
		Must:  false,
		Usage: "account address hex string",
		Value: "",
	}
	stype := util.Flag{
		Name:  "stype",
		Must:  false,
		Usage: "service type",
		Value: "email",
	}
	sid := util.Flag{
		Name:  "sid",
		Must:  false,
		Usage: "service id",
		Value: "",
	}
	args := []util.Flag{address, stype, sid}
	cmd := &ishell.Cmd{
		Name:                "getPublishInfo",
		Help:                "get publish info",
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
			stypeP := util.StringVar(c.Args, stype)
			sidP := util.StringVar(c.Args, sid)

			if err := dpkiGetPublishInfoAction(addressP, stypeP, sidP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func dpkiGetPublishInfoAction(addressP, stypeP, sidP string) error {
	fmt.Println(addressP, stypeP, sidP)

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspPubInfos := make([]*api.PublishInfoState, 0)

	if sidP != "" {
		err = client.Call(&rspPubInfos, "dpki_getPubKeyByTypeAndID", stypeP, sidP)
	} else if addressP != "" {
		pubAddr, err := types.HexToAddress(addressP)
		if err != nil {
			return err
		}
		err = client.Call(&rspPubInfos, "dpki_getPublishInfosByAccountAndType", pubAddr, stypeP)
	} else {
		err = client.Call(&rspPubInfos, "dpki_getPublishInfosByType", stypeP)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Total: %d\n", len(rspPubInfos))
	for idx, pi := range rspPubInfos {
		fmt.Printf("Index:       %d\n", idx)
		fmt.Printf("Account:     %s\n", pi.Account)
		fmt.Printf("ServiceType: %s\n", pi.PType)
		fmt.Printf("ServiceID:   %s\n", pi.PID)
		fmt.Printf("PubKey:      %s\n", pi.PubKey)
		fmt.Printf("Hash:        %s\n", pi.Hash)
		if pi.State != nil {
			fmt.Printf("VerifyStatus:\n")
			fmt.Printf(" VerifiedHeight: %d\n", pi.State.VerifiedHeight)
			fmt.Printf(" VerifiedStatus: %d\n", pi.State.VerifiedStatus)
			fmt.Printf(" BonusFee:       %d\n", pi.State.BonusFee)
			fmt.Printf(" OracleAccounts: %d", len(pi.State.OracleAccounts))
			for _, oa := range pi.State.OracleAccounts {
				fmt.Printf(",%s", oa)
			}
			fmt.Printf("\n")
		}
		fmt.Printf("\n")
	}

	return nil
}

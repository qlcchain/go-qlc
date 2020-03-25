package commands

import (
	"errors"
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPledgeGetInfoCmdByShell(parentCmd *ishell.Cmd) {
	pldAddrFlag := util.Flag{
		Name:  "pledgeAddress",
		Must:  false,
		Usage: "pledge account address hex string",
		Value: "",
	}
	bnfAddrFlag := util.Flag{
		Name:  "beneficialAddress",
		Must:  false,
		Usage: "beneficial account address hex string",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "getInfo",
		Help: "get pledge info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{pldAddrFlag, bnfAddrFlag}
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			pldAddr := util.StringVar(c.Args, pldAddrFlag)
			bnfAddr := util.StringVar(c.Args, bnfAddrFlag)

			if err := runPledgeGetInfoCmd(pldAddr, bnfAddr); err != nil {
				util.Warn(err)
				return
			}
		},
	}

	parentCmd.AddCmd(cmd)
}

func runPledgeGetInfoCmd(pldAddr, bnfAddr string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PledgeInfos)

	if pldAddr != "" {
		err = client.Call(&rspInfo, "pledge_getPledgeInfosByPledgeAddress", pldAddr)
	} else if bnfAddr != "" {
		err = client.Call(&rspInfo, "pledge_getBeneficialPledgeInfosByAddress", bnfAddr)
	} else {
		return errors.New("invalid param value")
	}
	if err != nil {
		return err
	}

	fmt.Printf("TotalInfos: %d\n", len(rspInfo.PledgeInfo))
	fmt.Printf("TotalAmounts: %d\n", rspInfo.TotalAmounts)

	for _, pi := range rspInfo.PledgeInfo {
		fmt.Printf("\n")
		fmt.Printf("PledgeAddress: %s\n", pi.PledgeAddress)
		fmt.Printf("Beneficial:    %s\n", pi.Beneficial)
		fmt.Printf("PType:         %s\n", pi.PType)
		fmt.Printf("Amount:        %s\n", pi.Amount)
		fmt.Printf("NEP5TxId:      %s\n", pi.NEP5TxId)
		fmt.Printf("WithdrawTime:  %s\n", pi.WithdrawTime)
	}

	return nil
}

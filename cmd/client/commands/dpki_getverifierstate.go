package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addDpkiGetVerifierStateListCmdByShell(parentCmd *ishell.Cmd) {
	hashFlag := util.Flag{
		Name:  "hash",
		Must:  false,
		Usage: "hash of pov block",
		Value: "",
	}
	heightFlag := util.Flag{
		Name:  "height",
		Must:  false,
		Usage: "height of pov block",
		Value: -1,
	}
	args := []util.Flag{hashFlag, heightFlag}

	cmd := &ishell.Cmd{
		Name:                "getVerifierStateList",
		Help:                "get verifier state list",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			hash := util.StringVar(c.Args, hashFlag)
			height, _ := util.IntVar(c.Args, heightFlag)

			err := runDpkiGetVerifierStateListCmd(hash, height)
			if err != nil {
				c.Println(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runDpkiGetVerifierStateListCmd(hash string, height int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspData := new(api.PKDVerifierStateList)
	if len(hash) > 0 {
		err = client.Call(rspData, "dpki_getAllVerifierStatesByBlockHash", hash)
		if err != nil {
			return err
		}
	} else if height >= 0 {
		err = client.Call(rspData, "dpki_getAllVerifierStatesByBlockHeight", height)
		if err != nil {
			return err
		}
	} else {
		hdrInfo := new(api.PovApiHeader)
		err = client.Call(hdrInfo, "pov_getLatestHeader")
		if err != nil {
			return err
		}
		height = int(hdrInfo.GetHeight())

		err = client.Call(rspData, "dpki_getAllVerifierStatesByBlockHeight", height)
		if err != nil {
			return err
		}
	}

	fmt.Printf("TotalNum: %d\n", rspData.VerifierNum)
	fmt.Printf("%-64s %-12s %-12s\n", "Account", "TotalVerify", "TotalReward")
	for addr, vs := range rspData.AllVerifiers {
		fmt.Printf("%-64s %-12d %-12s\n", addr, vs.TotalVerify, vs.TotalReward)
	}

	return nil
}

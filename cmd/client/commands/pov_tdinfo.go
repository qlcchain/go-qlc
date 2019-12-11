package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPovTdInfoCmdByShell(parentCmd *ishell.Cmd) {
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

	cmd := &ishell.Cmd{
		Name: "getTdInfo",
		Help: "get total difficulty (work sum) info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{hashFlag, heightFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			hashStr := util.StringVar(c.Args, hashFlag)
			height, _ := util.IntVar(c.Args, heightFlag)

			err := runPovTdInfoCmd(hashStr, height)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovTdInfoCmd(hashStr string, height int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovApiTD)
	if hashStr != "" {
		err = client.Call(rspInfo, "pov_getBlockTDByHash", hashStr)
	} else if height > 0 {
		err = client.Call(rspInfo, "pov_getBlockTDByHeight", height)
	} else {
		hdrInfo := new(api.PovApiHeader)
		err = client.Call(hdrInfo, "pov_getLatestHeader")
		if err != nil {
			return err
		}

		err = client.Call(rspInfo, "pov_getBlockTDByHeight", hdrInfo.GetHeight())
	}

	if err != nil {
		return err
	}

	fmt.Printf("Chain:   %s(0x%s)\n", rspInfo.TD.Chain.String(), rspInfo.TD.Chain.Text(16))
	fmt.Printf("Sha256d: %s(0x%s)\n", rspInfo.TD.Sha256d.String(), rspInfo.TD.Sha256d.Text(16))
	fmt.Printf("Scrypt:  %s(0x%s)\n", rspInfo.TD.Scrypt.String(), rspInfo.TD.Scrypt.Text(16))
	fmt.Printf("X11:     %s(0x%s)\n", rspInfo.TD.X11.String(), rspInfo.TD.X11.Text(16))

	return nil
}

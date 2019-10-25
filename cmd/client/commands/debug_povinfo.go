package commands

import (
	"encoding/json"
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
)

func addDebugPovInfoCmdByShell(parentCmd *ishell.Cmd) {
	cmd := &ishell.Cmd{
		Name: "getPovInfo",
		Help: "get pov info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{}
			if util.HelpText(c, args) {
				return
			}

			err := runDebugPovInfoCmd()
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runDebugPovInfoCmd() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(map[string]interface{})
	err = client.Call(rspInfo, "debug_getPovInfo")
	if err != nil {
		return err
	}

	infoBytes, err := json.MarshalIndent(rspInfo, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(infoBytes))

	return nil
}

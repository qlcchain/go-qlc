package commands

import (
	"encoding/json"
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
)

func addDebugPrivacyInfoCmdByShell(parentCmd *ishell.Cmd) {
	cmd := &ishell.Cmd{
		Name: "getPrivacyInfo",
		Help: "get privacy info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{}
			if util.HelpText(c, args) {
				return
			}

			err := runDebugPrivacyInfoCmd()
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runDebugPrivacyInfoCmd() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(map[string]interface{})
	err = client.Call(rspInfo, "debug_getPrivacyInfo")
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

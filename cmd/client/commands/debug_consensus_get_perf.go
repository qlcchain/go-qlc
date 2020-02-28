package commands

import (
	"encoding/json"
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/cmd/util"
	rpc "github.com/qlcchain/jsonrpc2"
)

func addDebugConsensusGetPerfCmdByShell(parentCmd *ishell.Cmd) {
	cmd := &ishell.Cmd{
		Name: "getConsPerf",
		Help: "get consensus perf",
		Func: func(c *ishell.Context) {
			err := runDebugConsensusGetPerfCmd()
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runDebugConsensusGetPerfCmd() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(map[string]interface{})
	err = client.Call(rspInfo, "debug_getConsPerf")
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

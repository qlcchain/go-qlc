package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/cmd/util"
	rpc "github.com/qlcchain/jsonrpc2"
)

func addDebugGCCmdByShell(parentCmd *ishell.Cmd) {
	cmd := &ishell.Cmd{
		Name: "gc",
		Help: "memory gc",
		Func: func(c *ishell.Context) {
			err := runDebugGCCmd()
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runDebugGCCmd() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(map[string]interface{})
	err = client.Call(rspInfo, "debug_gc")
	if err != nil {
		return err
	}

	return nil
}

package commands

import (
	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
)

func addDebugFeedConsensusCmdByShell(parentCmd *ishell.Cmd) {
	cmd := &ishell.Cmd{
		Name: "feedConsensus",
		Help: "feed blocks to consensus",
		Func: func(c *ishell.Context) {
			args := []util.Flag{}
			if util.HelpText(c, args) {
				return
			}

			err := runDebugFeedConsensusCmd()
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runDebugFeedConsensusCmd() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(map[string]interface{})
	err = client.Call(rspInfo, "debug_feedConsensus")
	if err != nil {
		return err
	}

	return nil
}

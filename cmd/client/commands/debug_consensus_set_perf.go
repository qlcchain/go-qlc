package commands

import (
	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
)

func addDebugConsensusSetPerfCmdByShell(parentCmd *ishell.Cmd) {
	op := util.Flag{
		Name:  "op",
		Must:  true,
		Usage: "set perf operation(0:close 1:bl 2:bp 3:all 4:export)",
		Value: "0",
	}
	args := []util.Flag{op}
	cmd := &ishell.Cmd{
		Name:                "setConsPerf",
		Help:                "set consensus perf",
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

			opP, err := util.IntVar(c.Args, op)
			if err != nil {
				util.Warn(err)
			}

			err = runDebugConsensusSetPerfCmd(opP)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runDebugConsensusSetPerfCmd(op int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(map[string]interface{})
	err = client.Call(rspInfo, "debug_setConsPerf", op)
	if err != nil {
		return err
	}

	return nil
}

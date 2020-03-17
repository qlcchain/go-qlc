package commands

import (
	"bytes"
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addDebugUncheckAnalysisCmdByShell(parentCmd *ishell.Cmd) {
	hash := util.Flag{
		Name:  "hash",
		Must:  false,
		Usage: "block hash",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "uncheckAnalysis",
		Help: "analyse unchecked blocks",
		Func: func(c *ishell.Context) {
			args := []util.Flag{hash}
			if util.HelpText(c, args) {
				return
			}

			hashP := util.StringVar(c.Args, hash)

			err := runDebugUncheckAnalysisCmd(hashP)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runDebugUncheckAnalysisCmd(hashStr string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new([]*api.UncheckInfo)

	if len(hashStr) > 0 {
		hash, err := types.NewHash(hashStr)
		if err != nil {
			return err
		}

		err = client.Call(rspInfo, "debug_uncheckBlock", hash)
		if err != nil {
			return err
		}
	} else {
		err = client.Call(rspInfo, "debug_uncheckAnalysis")
		if err != nil {
			return err
		}
	}

	fmt.Printf("%-64s\t%-16s\t\t%-64s\t%s\n", "block", "gapType", "gapHash", "gapHeight")
	fmt.Printf("%s\n", bytes.Repeat([]byte("-"), 200))
	for _, r := range *rspInfo {
		fmt.Printf("%-64s\t%-16s\t\t%-64s\t%d\n", r.Hash, r.GapType, r.GapHash, r.GapHeight)
	}
	fmt.Printf("%s\n", bytes.Repeat([]byte("-"), 200))

	return nil
}

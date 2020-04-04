package commands

import (
	"fmt"
	"sort"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPovRepStateListCmdByShell(parentCmd *ishell.Cmd) {
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
	sortFlag := util.Flag{
		Name:  "sort",
		Must:  false,
		Usage: "sort type, total/vote/height",
		Value: "balance",
	}
	args := []util.Flag{hashFlag, heightFlag, sortFlag}
	cmd := &ishell.Cmd{
		Name:                "getAllRepStates",
		Help:                "get all rep state list",
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
			sortType := util.StringVar(c.Args, sortFlag)

			err := runPovRepStateListCmd(hash, height, sortType)
			if err != nil {
				c.Println(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovRepStateListCmd(hash string, height int, sortType string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	repStateList := new(api.PovApiRepState)
	if len(hash) > 0 {
		err = client.Call(repStateList, "pov_getAllRepStatesByBlockHash", hash)
		if err != nil {
			return err
		}
	} else if height >= 0 {
		err = client.Call(repStateList, "pov_getAllRepStatesByBlockHeight", height)
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

		err = client.Call(repStateList, "pov_getAllRepStatesByBlockHeight", height)
		if err != nil {
			return err
		}
	}

	var sortReps []*types.PovRepState
	for _, rep := range repStateList.Reps {
		sortReps = append(sortReps, rep)
	}
	sort.Slice(sortReps, func(i, j int) bool {
		if sortType == "vote" {
			if sortReps[i].Vote.Compare(sortReps[j].Vote) == types.BalanceCompBigger {
				return true
			}
		} else if sortType == "height" {
			if sortReps[i].Height > sortReps[j].Height {
				return true
			}
		} else {
			if sortReps[i].Total.Compare(sortReps[j].Total) == types.BalanceCompBigger {
				return true
			}
		}
		return false
	})

	fmt.Printf("TotalNum: %d\n", len(repStateList.Reps))
	fmt.Printf("%-64s %-12s %-12s %-6s %-10s\n", "Account", "Total", "Vote", "Status", "Height")
	for _, rep := range sortReps {
		fmt.Printf("%-64s %-12s %-12s %-6d %-10d\n",
			rep.Account,
			formatPovReward(rep.Total),
			formatPovReward(rep.Vote),
			rep.Status, rep.Height)
	}

	return nil
}

package commands

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
	"strconv"
	"time"
)

func addPovHeaderInfoCmdByShell(parentCmd *ishell.Cmd) {
	idFlag := util.Flag{
		Name:  "id",
		Must:  true,
		Usage: "height or hash of pov block",
		Value: "-1",
	}

	cmd := &ishell.Cmd{
		Name: "getHeaderInfo",
		Help: "get pov header info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{idFlag}
			if util.HelpText(c, args) {
				return
			}
			idStr := util.StringVar(c.Args, idFlag)

			err := runPovHeaderInfoCmd(idStr)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovHeaderInfoCmd(idStr string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovApiHeader)

	if len(idStr) == 64 {
		err = client.Call(rspInfo, "pov_getHeaderByHash", idStr)
		if err != nil {
			return err
		}
	} else {
		var height int64
		if idStr == "latest" {
			height = int64(-1)
		} else {
			height, err = strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return err
			}
		}

		if height < 0 {
			err = client.Call(rspInfo, "pov_getLatestHeader")
		} else {
			err = client.Call(rspInfo, "pov_getHeaderByHeight", height)
		}
		if err != nil {
			return err
		}
	}

	fmt.Printf("Hash:       %s\n", rspInfo.BasHdr.Hash)
	fmt.Printf("Height:     %d\n", rspInfo.BasHdr.Height)
	fmt.Printf("Version:    %d(0x%x)\n", rspInfo.BasHdr.Version, rspInfo.BasHdr.Version)
	fmt.Printf("Previous:   %s\n", rspInfo.BasHdr.Previous)
	fmt.Printf("MerkleRoot: %s\n", rspInfo.BasHdr.MerkleRoot)
	fmt.Printf("Time:       %d(%s)\n", rspInfo.BasHdr.Timestamp, time.Unix(int64(rspInfo.BasHdr.Timestamp), 0))
	fmt.Printf("Bits:       %d(0x%x)\n", rspInfo.BasHdr.Bits, rspInfo.BasHdr.Bits)
	fmt.Printf("Nonce:      %d(0x%x)\n", rspInfo.BasHdr.Nonce, rspInfo.BasHdr.Nonce)
	fmt.Printf("TxNum:      %d\n", rspInfo.CbTx.TxNum)
	fmt.Printf("StateHash:  %s\n", rspInfo.CbTx.StateHash)

	fmt.Printf("\n")
	fmt.Printf("AlgoName:       %s\n", rspInfo.AlgoName)
	fmt.Printf("AlgoEfficiency: %d\n", rspInfo.AlgoEfficiency)
	fmt.Printf("AlgoDifficulty: %s\n", formatPovDifficulty(rspInfo.AlgoDifficulty))
	fmt.Printf("NormDifficulty: %s\n", formatPovDifficulty(rspInfo.NormDifficulty))

	fmt.Printf("\n")
	if rspInfo.AuxHdr == nil {
		fmt.Printf("AuxPoW: false\n")
	} else {
		fmt.Printf("AuxPoW:      true\n")
		fmt.Printf("ParentHash:  %s\n", rspInfo.AuxHdr.ParentHash)
		parHeader := rspInfo.AuxHdr.ParBlockHeader
		fmt.Printf("ParentTime:  %d(%s)\n", parHeader.Timestamp, time.Unix(int64(parHeader.Timestamp), 0))
		fmt.Printf("ParentBits:  %d(0x%x)\n", parHeader.Bits, parHeader.Bits)
		fmt.Printf("ParentNonce: %d(0x%x)\n", parHeader.Nonce, parHeader.Nonce)
	}

	fmt.Printf("\n")
	if rspInfo.CbTx != nil && len(rspInfo.CbTx.TxIns) > 0 {
		fmt.Printf("CoinbaseHash:  %s\n", rspInfo.CbTx.Hash)
		fmt.Printf("CoinbaseExtra: %s\n", rspInfo.CbTx.TxIns[0].Extra)
		fmt.Printf("Miner:         %s\n", rspInfo.GetMinerAddr())
		fmt.Printf("Reward(M):     %s QGAS\n", formatPovReward(rspInfo.GetMinerReward()))
		fmt.Printf("Reward(R):     %s QGAS\n", formatPovReward(rspInfo.GetRepReward()))
	}

	return nil
}

func addPovHeaderListCmdByShell(parentCmd *ishell.Cmd) {
	heightFlag := util.Flag{
		Name:  "height",
		Must:  false,
		Usage: "height of pov blocks",
		Value: -1,
	}
	countFlag := util.Flag{
		Name:  "count",
		Must:  false,
		Usage: "count of pov blocks",
		Value: 10,
	}
	ascFlag := util.Flag{
		Name:  "ascend",
		Must:  false,
		Usage: "order is ascend or descend",
		Value: false,
	}

	cmd := &ishell.Cmd{
		Name: "getHeaderList",
		Help: "get pov header list",
		Func: func(c *ishell.Context) {
			args := []util.Flag{heightFlag, countFlag, ascFlag}
			if util.HelpText(c, args) {
				return
			}
			height, _ := util.IntVar(c.Args, heightFlag)
			count, _ := util.IntVar(c.Args, countFlag)
			asc := util.BoolVar(c.Args, ascFlag)

			err := runPovHeaderListCmd(height, count, asc)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovHeaderListCmd(height int, count int, asc bool) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	if height < 0 {
		rspLatestInfo := new(api.PovApiHeader)
		err = client.Call(rspLatestInfo, "pov_getLatestHeader")
		if err != nil {
			return err
		}

		height = int(rspLatestInfo.GetHeight())
	}

	rspInfo := new(api.PovApiBatchHeader)

	err = client.Call(rspInfo, "pov_batchGetHeadersByHeight", height, count, asc)
	if err != nil {
		return err
	}

	fmt.Printf("%-10s %-64s %-6s %-7s %-3s %-7s %-19s\n", "Height", "Hash", "TxNum", "Algo", "Aux", "Diff", "Time")
	for _, apiHeader := range rspInfo.Headers {
		isAux := 0
		if apiHeader.AuxHdr != nil {
			isAux = 1
		}

		blkTm := time.Unix(int64(apiHeader.GetTimestamp()), 0)

		fmt.Printf("%-10d %-64s %-6d %-7s %-3d %-7s %-19s\n",
			apiHeader.GetHeight(), apiHeader.GetHash(), apiHeader.GetTxNum(),
			apiHeader.AlgoName, isAux,
			formatPovDifficulty(apiHeader.NormDifficulty),
			blkTm.Format("2006-01-02 15:04:05"))
	}

	return nil
}

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

func addPovBlockInfoCmdByShell(parentCmd *ishell.Cmd) {
	idFlag := util.Flag{
		Name:  "id",
		Must:  true,
		Usage: "height or hash of pov block",
		Value: "-1",
	}
	txOffsetFlag := util.Flag{
		Name:  "txOffset",
		Must:  false,
		Usage: "tx offset of pov block",
		Value: 0,
	}
	txCountFlag := util.Flag{
		Name:  "txCount",
		Must:  false,
		Usage: "tx count of pov block",
		Value: 100,
	}

	cmd := &ishell.Cmd{
		Name: "getBlockInfo",
		Help: "get pov block info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{idFlag, txOffsetFlag, txCountFlag}
			if util.HelpText(c, args) {
				return
			}
			idStr := util.StringVar(c.Args, idFlag)
			txOffset, _ := util.IntVar(c.Args, txOffsetFlag)
			txCount, _ := util.IntVar(c.Args, txCountFlag)

			err := runPovBlockInfoCmd(idStr, txOffset, txCount)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovBlockInfoCmd(idStr string, txOffset int, txCount int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovApiBlock)

	if len(idStr) == 64 {
		err = client.Call(rspInfo, "pov_getBlockByHash", idStr, txOffset, txCount)
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
			err = client.Call(rspInfo, "pov_getLatestBlock", txOffset, txCount)
		} else {
			err = client.Call(rspInfo, "pov_getBlockByHeight", height, txOffset, txCount)
		}
		if err != nil {
			return err
		}
	}

	header := &(rspInfo.Header)

	fmt.Printf("Hash:       %s\n", header.BasHdr.Hash)
	fmt.Printf("Height:     %d\n", header.BasHdr.Height)
	fmt.Printf("Version:    %d(0x%x)\n", header.BasHdr.Version, header.BasHdr.Version)
	fmt.Printf("Previous:   %s\n", header.BasHdr.Previous)
	fmt.Printf("MerkleRoot: %s\n", header.BasHdr.MerkleRoot)
	fmt.Printf("Time:       %d(%s)\n", header.BasHdr.Timestamp, time.Unix(int64(header.BasHdr.Timestamp), 0))
	fmt.Printf("Bits:       %d(0x%x)\n", header.BasHdr.Bits, header.BasHdr.Bits)
	fmt.Printf("Nonce:      %d(0x%x)\n", header.BasHdr.Nonce, header.BasHdr.Nonce)
	fmt.Printf("TxNum:      %d\n", header.CbTx.TxNum)
	fmt.Printf("StateHash:  %s\n", header.CbTx.StateHash)

	fmt.Printf("\n")
	fmt.Printf("AlgoName:       %s\n", rspInfo.AlgoName)
	fmt.Printf("AlgoEfficiency: %d\n", rspInfo.AlgoEfficiency)
	fmt.Printf("AlgoDifficulty: %s\n", formatPovDifficulty(rspInfo.AlgoDifficulty))
	fmt.Printf("NormDifficulty: %s\n", formatPovDifficulty(rspInfo.NormDifficulty))

	fmt.Printf("\n")
	if header.AuxHdr == nil {
		fmt.Printf("AuxPoW: false\n")
	} else {
		fmt.Printf("AuxPoW:      true\n")
		fmt.Printf("ParentHash:  %s\n", header.AuxHdr.ParentHash)
		parHeader := header.AuxHdr.ParBlockHeader
		fmt.Printf("ParentTime:  %d(%s)\n", parHeader.Timestamp, time.Unix(int64(parHeader.Timestamp), 0))
		fmt.Printf("ParentBits:  %d(0x%x)\n", parHeader.Bits, parHeader.Bits)
		fmt.Printf("ParentNonce: %d(0x%x)\n", parHeader.Nonce, parHeader.Nonce)
	}

	fmt.Printf("\n")
	if header.CbTx != nil && len(header.CbTx.TxIns) > 0 {
		fmt.Printf("CoinbaseHash:  %s\n", header.CbTx.Hash)
		fmt.Printf("CoinbaseExtra: %s\n", header.CbTx.TxIns[0].Extra)
		fmt.Printf("Miner:         %s\n", header.GetMinerAddr())
		fmt.Printf("Reward(M):     %s QGAS\n", formatPovReward(header.GetMinerReward()))
		fmt.Printf("Reward(R):     %s QGAS\n", formatPovReward(header.GetRepReward()))
	}

	body := &(rspInfo.Body)
	if body != nil && len(body.Txs) > 0 {
		fmt.Printf("\n")
		fmt.Printf("%-7s %-64s\n", "TxIndex", "TxHash")
		for txIdx, txPov := range body.Txs {
			fmt.Printf("%-7d %-64s\n", txOffset+txIdx, txPov.GetHash())
		}
	}

	return nil
}

func addPovBlockListCmdByShell(parentCmd *ishell.Cmd) {
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
		Name: "getBlockList",
		Help: "get pov block list",
		Func: func(c *ishell.Context) {
			args := []util.Flag{heightFlag, countFlag, ascFlag}
			if util.HelpText(c, args) {
				return
			}
			height, _ := util.IntVar(c.Args, heightFlag)
			count, _ := util.IntVar(c.Args, countFlag)
			asc := util.BoolVar(c.Args, ascFlag)

			err := runPovBlockListCmd(height, count, asc)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovBlockListCmd(height int, count int, asc bool) error {
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

package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/qlcchain/go-qlc/cmd/util"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addMinerRewardCmdByShell(parentCmd *ishell.Cmd) {
	cbPriKey := util.Flag{
		Name:  "cbPriKey",
		Must:  true,
		Usage: "coinbase account private hex string",
	}
	bnfPriKey := util.Flag{
		Name:  "bnfPriKey",
		Must:  false,
		Usage: "beneficial account private hex string",
		Value: "",
	}
	bnfAddr := util.Flag{
		Name:  "bnfAddr",
		Must:  false,
		Usage: "beneficial account address hex string",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "reward",
		Help: "miner reward (gas token)",
		Func: func(c *ishell.Context) {
			args := []util.Flag{cbPriKey, bnfPriKey, bnfAddr}
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			cbPriKeyP := util.StringVar(c.Args, cbPriKey)
			bnfPriKeyP := util.StringVar(c.Args, bnfPriKey)
			bnfAddrP := util.StringVar(c.Args, bnfAddr)

			if err := minerRewardAction(cbPriKeyP, bnfPriKeyP, bnfAddrP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func addMinerRewardCmdByCobra(parentCmd *cobra.Command) {
	var cbPriKeyP string
	var bnfPriKeyP string
	var bnfAddrP string

	var cmd = &cobra.Command{
		Use:   "reward",
		Short: "miner reward (gas token)",
		Run: func(cmd *cobra.Command, args []string) {
			err := minerRewardAction(cbPriKeyP, bnfPriKeyP, bnfAddrP)
			if err != nil {
				cmd.Println(err)
			}
		},
	}
	cmd.Flags().StringVar(&cbPriKeyP, "cbPriKey", "", "coinbase account private hex string")
	cmd.Flags().StringVar(&bnfPriKeyP, "bnfPriKey", "", "beneficial account private hex string")
	cmd.Flags().StringVar(&bnfAddrP, "bnfAddr", "", "beneficial account address hex string")
	parentCmd.AddCommand(cmd)
}

func minerRewardAction(cbPriKeyP string, bnfPriKeyP string, bnfAddrHexP string) error {
	if cbPriKeyP == "" {
		return errors.New("invalid coinbase value")
	}
	if bnfPriKeyP == "" && bnfAddrHexP == "" {
		return errors.New("invalid beneficial value, private key or address must have one")
	}

	cbBytes, err := hex.DecodeString(cbPriKeyP)
	if err != nil {
		return err
	}
	cbAcc := types.NewAccount(cbBytes)
	if cbAcc == nil {
		return errors.New("can not new coinbase account")
	}

	var bnfAcc *types.Account
	var bnfAddr types.Address
	if bnfPriKeyP != "" {
		bnfBytes, err := hex.DecodeString(bnfPriKeyP)
		if err != nil {
			return err
		}
		bnfAcc = types.NewAccount(bnfBytes)
		if bnfAcc == nil {
			return errors.New("can not new beneficial account")
		}
		bnfAddr = bnfAcc.Address()
	} else {
		bnfAddr, err = types.HexToAddress(bnfAddrHexP)
		if err != nil {
			return err
		}
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	// calc avail reward info
	rspRewardInfo := new(api.MinerAvailRewardInfo)
	err = client.Call(rspRewardInfo, "miner_getAvailRewardInfo", cbAcc.Address())
	if err != nil {
		return err
	}
	fmt.Printf("AvailRewardInfo:\n%s\n", cutil.ToIndentString(rspRewardInfo))

	if !rspRewardInfo.NeedCallReward {
		return errors.New("can not call reward contract because no available reward height")
	}

	rewardParam := api.RewardParam{
		Coinbase:     cbAcc.Address(),
		Beneficial:   bnfAddr,
		StartHeight:  rspRewardInfo.AvailStartHeight,
		EndHeight:    rspRewardInfo.AvailEndHeight,
		RewardBlocks: rspRewardInfo.AvailRewardBlocks,
		RewardAmount: rspRewardInfo.AvailRewardAmount.Int,
	}

	// generate contract send block
	sendBlock := types.StateBlock{}
	err = client.Call(&sendBlock, "miner_getRewardSendBlock", &rewardParam)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, sendBlock.Root())
	sendBlock.Work = worker.NewWork()

	sendHash := sendBlock.GetHash()
	sendBlock.Signature = cbAcc.Sign(sendHash)

	fmt.Printf("SendBlock:\n%s\n", cutil.ToIndentString(sendBlock))
	fmt.Println("address", sendBlock.Address, "sendHash", sendHash)

	err = client.Call(nil, "ledger_process", &sendBlock)
	if err != nil {
		return err
	}
	fmt.Printf("success to send reward, delta balance %s blocks %d\n", rewardParam.RewardAmount.String(), rewardParam.RewardBlocks)

	// generate contract recv block if we have beneficial prikey
	rewardBlock := types.StateBlock{}
	if bnfAcc != nil {
		time.Sleep(3 * time.Second)

		err = client.Call(&rewardBlock, "miner_getRewardRecvBlockBySendHash", sendHash)
		if err != nil {
			return err
		}

		var w2 types.Work
		worker2, _ := types.NewWorker(w2, rewardBlock.Root())
		rewardBlock.Work = worker2.NewWork()

		rewardHash := rewardBlock.GetHash()
		rewardBlock.Signature = bnfAcc.Sign(rewardHash)

		fmt.Printf("RewardBlock:\n%s\n", cutil.ToIndentString(rewardBlock))
		fmt.Println("address", rewardBlock.Address, "rewardHash", rewardHash)

		err = client.Call(nil, "ledger_process", &rewardBlock)
		if err != nil {
			return err
		}
		fmt.Printf("success to recv reward, account balance %s\n", rewardBlock.Balance)
	}

	return nil
}

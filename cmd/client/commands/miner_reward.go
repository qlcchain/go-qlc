package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	"github.com/spf13/cobra"
)

func minerReward() {
	var cbPriKeyP string
	var bnfPriKeyP string
	var bnfAddrP string

	if interactive {
		cbPriKey := util.Flag{
			Name:  "cbprikey",
			Must:  true,
			Usage: "coinbase account private hex string",
		}
		bnfPriKey := util.Flag{
			Name:  "bnfprikey",
			Must:  false,
			Usage: "beneficial account private hex string",
		}
		bnfAddr := util.Flag{
			Name:  "bnfaddr",
			Must:  false,
			Usage: "beneficial account address hex string",
		}

		cmd := &ishell.Cmd{
			Name: "minerreward",
			Help: "miner get reward (gas token)",
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

				cbPriKeyP = util.StringVar(c.Args, cbPriKey)
				bnfPriKeyP = util.StringVar(c.Args, bnfPriKey)
				bnfAddrP = util.StringVar(c.Args, bnfAddr)

				if err := minerRewardAction(cbPriKeyP, bnfPriKeyP, bnfAddrP); err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "minerreward",
			Short: "miner get reward (gas token)",
			Run: func(cmd *cobra.Command, args []string) {
				err := minerRewardAction(cbPriKeyP, bnfPriKeyP, bnfAddrP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		cmd.Flags().StringVar(&cbPriKeyP, "cbprikey", "", "coinbase account private hex string")
		cmd.Flags().StringVar(&bnfPriKeyP, "bnfprikey", "", "beneficial account private hex string")
		cmd.Flags().StringVar(&bnfAddrP, "bnfaddr", "", "beneficial account address hex string")
		rootCmd.AddCommand(cmd)
	}
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
	}

	rewardAmount := common.PovMinerRewardPerBlockBalance.Mul(int64(rewardParam.RewardBlocks))

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

	// generate contract recv block if we have beneficial prikey
	rewardBlock := types.StateBlock{}
	if bnfAcc != nil {
		err = client.Call(&rewardBlock, "miner_getRewardRecvBlock", &sendBlock)
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
	}

	// publish all blocks to node
	err = client.Call(nil, "ledger_process", &sendBlock)
	if err != nil {
		return err
	}
	fmt.Printf("success to send reward, balance %s blocks %d\n", rewardAmount, rewardParam.RewardBlocks)

	if bnfAcc != nil {
		err = client.Call(nil, "ledger_process", &rewardBlock)
		if err != nil {
			return err
		}
		fmt.Printf("success to recv reward\n")
	} else {
		fmt.Printf("please to recv reward\n")
	}

	return nil
}

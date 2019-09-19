package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/spf13/cobra"
)

func repReward() {
	var repPriKeyP string
	var bnfPriKeyP string
	var bnfAddrP string

	if interactive {
		repPriKey := util.Flag{
			Name:  "repPriKey",
			Must:  true,
			Usage: "representative account private hex string",
		}
		bnfPriKey := util.Flag{
			Name:  "bnfPriKey",
			Must:  false,
			Usage: "beneficial account private hex string",
		}
		bnfAddr := util.Flag{
			Name:  "bnfAddr",
			Must:  false,
			Usage: "beneficial account address hex string",
		}

		cmd := &ishell.Cmd{
			Name: "repReward",
			Help: "representative get reward (gas token)",
			Func: func(c *ishell.Context) {
				args := []util.Flag{repPriKey, bnfPriKey, bnfAddr}
				if util.HelpText(c, args) {
					return
				}

				err := util.CheckArgs(c, args)
				if err != nil {
					util.Warn(err)
					return
				}

				repPriKeyP = util.StringVar(c.Args, repPriKey)
				bnfPriKeyP = util.StringVar(c.Args, bnfPriKey)
				bnfAddrP = util.StringVar(c.Args, bnfAddr)

				if err := repRewardAction(repPriKeyP, bnfPriKeyP, bnfAddrP); err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "repReward",
			Short: "representative get reward (gas token)",
			Run: func(cmd *cobra.Command, args []string) {
				err := repRewardAction(repPriKeyP, bnfPriKeyP, bnfAddrP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		cmd.Flags().StringVar(&repPriKeyP, "repPriKey", "", "representative account private hex string")
		cmd.Flags().StringVar(&bnfPriKeyP, "bnfPriKey", "", "beneficial account private hex string")
		cmd.Flags().StringVar(&bnfAddrP, "bnfAddr", "", "beneficial account address hex string")
		rootCmd.AddCommand(cmd)
	}
}

func repRewardAction(repPriKeyP string, bnfPriKeyP string, bnfAddrHexP string) error {
	if repPriKeyP == "" {
		return errors.New("invalid representative private key")
	}

	if bnfPriKeyP == "" && bnfAddrHexP == "" {
		return errors.New("invalid beneficial value, private key or address must have one")
	}

	repPriKeyBytes, err := hex.DecodeString(repPriKeyP)
	if err != nil {
		return err
	}
	repAcc := types.NewAccount(repPriKeyBytes)
	if repAcc == nil {
		return errors.New("get representative account err")
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
	rspRewardInfo := new(api.RepAvailRewardInfo)
	err = client.Call(rspRewardInfo, "rep_getAvailRewardInfo", repAcc.Address())
	if err != nil {
		return err
	}
	fmt.Println("rep address:", repAcc.Address())
	fmt.Printf("AvailRewardInfo:\n%s\n", cutil.ToIndentString(rspRewardInfo))

	if !rspRewardInfo.NeedCallReward {
		return errors.New("can not call reward contract because no available reward height")
	}

	rewardParam := api.RepRewardParam{
		Account:      repAcc.Address(),
		Beneficial:   bnfAddr,
		StartHeight:  rspRewardInfo.AvailStartHeight,
		EndHeight:    rspRewardInfo.AvailEndHeight,
		RewardBlocks: rspRewardInfo.AvailRewardBlocks,
		RewardAmount: rspRewardInfo.AvailRewardAmount,
	}

	// generate contract send block
	sendBlock := types.StateBlock{}
	err = client.Call(&sendBlock, "rep_getRewardSendBlock", &rewardParam)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, sendBlock.Root())
	sendBlock.Work = worker.NewWork()

	sendHash := sendBlock.GetHash()
	sendBlock.Signature = repAcc.Sign(sendHash)

	fmt.Printf("SendBlock:\n%s\n", cutil.ToIndentString(sendBlock))
	fmt.Println("address", sendBlock.Address, "sendHash", sendHash)

	// generate contract recv block if we have beneficial prikey
	rewardBlock := types.StateBlock{}
	if bnfAcc != nil {
		err = client.Call(&rewardBlock, "rep_getRewardRecvBlock", &sendBlock)
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
	fmt.Printf("success to send reward, delta balance %s blocks %d\n", rewardParam.RewardAmount.String(), rewardParam.RewardBlocks)

	if bnfAcc != nil {
		err = client.Call(nil, "ledger_process", &rewardBlock)
		if err != nil {
			return err
		}
		fmt.Printf("success to recv reward, account balance %s\n", rewardBlock.Balance)
	} else {
		fmt.Printf("please to recv reward, send hash %s\n", sendHash)
	}

	return nil
}

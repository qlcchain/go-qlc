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
	var coinbaseP string
	var beneficialP string

	if interactive {
		coinbase := util.Flag{
			Name:  "coinbase",
			Must:  true,
			Usage: "coinbase account private hex string",
		}
		beneficial := util.Flag{
			Name:  "beneficial",
			Must:  false,
			Usage: "beneficial account private hex string",
		}

		cmd := &ishell.Cmd{
			Name: "minerreward",
			Help: "miner get reward (gas token)",
			Func: func(c *ishell.Context) {
				args := []util.Flag{coinbase, beneficial}
				if util.HelpText(c, args) {
					return
				}
				err := util.CheckArgs(c, args)
				if err != nil {
					util.Warn(err)
					return
				}

				coinbaseP = util.StringVar(c.Args, coinbase)
				beneficialP = util.StringVar(c.Args, beneficial)

				if err := minerRewardAction(coinbaseP, beneficialP); err != nil {
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
				err := minerRewardAction(coinbaseP, beneficialP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		cmd.Flags().StringVar(&coinbaseP, "coinbase", "", "coinbase account private hex string")
		cmd.Flags().StringVar(&beneficialP, "beneficial", "", "beneficial account private hex string")
		rootCmd.AddCommand(cmd)
	}
}

func minerRewardAction(coinbaseP, beneficialP string) error {
	if coinbaseP == "" {
		return errors.New("invalid coinbase value")
	}

	if beneficialP == "" {
		return errors.New("invalid beneficial value")
	}

	cbBytes, err := hex.DecodeString(coinbaseP)
	if err != nil {
		return err
	}
	coinbaseAcc := types.NewAccount(cbBytes)
	if coinbaseAcc == nil {
		return errors.New("can not new coinbase account")
	}

	benBytes, err := hex.DecodeString(beneficialP)
	if err != nil {
		return err
	}
	beneficialAcc := types.NewAccount(benBytes)
	if beneficialAcc == nil {
		return errors.New("can not new beneficial account")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspRewardInfo := new(api.MinerAvailRewardInfo)
	err = client.Call(rspRewardInfo, "miner_getAvailRewardInfo", coinbaseAcc.Address())
	if err != nil {
		return err
	}
	fmt.Printf("AvailRewardInfo:\n%s\n", cutil.ToIndentString(rspRewardInfo))

	if !rspRewardInfo.NeedCallReward {
		return errors.New("can not call reward contract because no available reward height")
	}

	rewardParam := api.RewardParam{
		Coinbase:     coinbaseAcc.Address(),
		Beneficial:   beneficialAcc.Address(),
		StartHeight:  rspRewardInfo.AvailStartHeight,
		EndHeight:    rspRewardInfo.AvailEndHeight,
		RewardBlocks: rspRewardInfo.AvailRewardBlocks,
	}

	rewardAmount := common.PovMinerRewardPerBlockBalance.Mul(int64(rewardParam.RewardBlocks))

	send := types.StateBlock{}
	err = client.Call(&send, "miner_getRewardSendBlock", &rewardParam)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	sendHash := send.GetHash()
	send.Signature = coinbaseAcc.Sign(sendHash)

	fmt.Printf("SendBlock:\n%s\n", cutil.ToIndentString(send))
	fmt.Println("address", send.Address, "sendHash", sendHash)

	reward := types.StateBlock{}
	err = client.Call(&reward, "miner_getRewardRecvBlock", &send)
	if err != nil {
		return err
	}

	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	rewardHash := reward.GetHash()
	reward.Signature = beneficialAcc.Sign(rewardHash)

	fmt.Printf("RewardBlock:\n%s\n", cutil.ToIndentString(reward))
	fmt.Println("address", reward.Address, "rewardHash", rewardHash)

	err = client.Call(nil, "ledger_process", &send)
	if err != nil {
		return err
	}

	err = client.Call(nil, "ledger_process", &reward)
	if err != nil {
		return err
	}

	fmt.Printf("success to reward balance %s with blocks %d\n", rewardAmount, rewardParam.RewardBlocks)

	return nil
}

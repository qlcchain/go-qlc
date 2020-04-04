package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addDpkiRewardCmdByShell(parentCmd *ishell.Cmd) {
	rwdPriKey := util.Flag{
		Name:  "rwdPriKey",
		Must:  true,
		Usage: "reward account private hex string",
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
		Help: "withdraw verifier reward (gas token)",
		Func: func(c *ishell.Context) {
			args := []util.Flag{rwdPriKey, bnfPriKey, bnfAddr}
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			rwdPriKeyP := util.StringVar(c.Args, rwdPriKey)
			bnfPriKeyP := util.StringVar(c.Args, bnfPriKey)
			bnfAddrP := util.StringVar(c.Args, bnfAddr)

			if err := dpkiRewardAction(rwdPriKeyP, bnfPriKeyP, bnfAddrP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func dpkiRewardAction(rwdPriKeyP string, bnfPriKeyP string, bnfAddrHexP string) error {
	if rwdPriKeyP == "" {
		return errors.New("invalid reward account value")
	}
	if bnfPriKeyP == "" && bnfAddrHexP == "" {
		return errors.New("invalid beneficial account value, private key or address must have one")
	}

	rwdBytes, err := hex.DecodeString(rwdPriKeyP)
	if err != nil {
		return err
	}
	rwdAcc := types.NewAccount(rwdBytes)
	if rwdAcc == nil {
		return errors.New("can not new reward account")
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
	rspRewardInfo := new(api.PKDAvailRewardInfo)
	err = client.Call(rspRewardInfo, "dpki_getAvailRewardInfo", rwdAcc.Address())
	if err != nil {
		return err
	}
	fmt.Printf("AvailRewardInfo:\n%s\n", cutil.ToIndentString(rspRewardInfo))

	if !rspRewardInfo.NeedCallReward {
		return errors.New("can not call reward contract because no available reward height")
	}

	rewardParam := api.PKDRewardParam{
		Account:      rwdAcc.Address(),
		Beneficial:   bnfAddr,
		EndHeight:    rspRewardInfo.AvailEndHeight,
		RewardAmount: rspRewardInfo.AvailRewardAmount.Int,
	}

	// generate contract send block
	sendBlock := types.StateBlock{}
	err = client.Call(&sendBlock, "dpki_getRewardSendBlock", &rewardParam)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, sendBlock.Root())
	sendBlock.Work = worker.NewWork()

	sendHash := sendBlock.GetHash()
	sendBlock.Signature = rwdAcc.Sign(sendHash)

	fmt.Printf("SendBlock:\n%s\n", cutil.ToIndentString(sendBlock))
	fmt.Println("address", sendBlock.Address, "sendHash", sendHash)

	err = client.Call(nil, "ledger_process", &sendBlock)
	if err != nil {
		return err
	}
	fmt.Printf("success to send reward, delta balance %s\n", rewardParam.RewardAmount.String())

	// generate contract recv block if we have beneficial prikey
	recvBlock := types.StateBlock{}
	if bnfAcc != nil {
		time.Sleep(3 * time.Second)

		err = client.Call(&recvBlock, "dpki_getRewardRecvBlockBySendHash", sendHash)
		if err != nil {
			return err
		}

		var w2 types.Work
		worker2, _ := types.NewWorker(w2, recvBlock.Root())
		recvBlock.Work = worker2.NewWork()

		rewardHash := recvBlock.GetHash()
		recvBlock.Signature = bnfAcc.Sign(rewardHash)

		fmt.Printf("ReceiveBlock:\n%s\n", cutil.ToIndentString(recvBlock))
		fmt.Println("address", recvBlock.Address, "rewardHash", rewardHash)

		err = client.Call(nil, "ledger_process", &recvBlock)
		if err != nil {
			return err
		}
		fmt.Printf("success to recv reward, account balance %s\n", recvBlock.Balance)
	}

	return nil
}

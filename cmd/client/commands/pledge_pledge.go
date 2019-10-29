package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPledgePledgeCmdByShell(parentCmd *ishell.Cmd) {
	beneficialAccount := util.Flag{
		Name:  "beneficialAccount",
		Must:  false,
		Usage: "beneficial account private hex string",
		Value: "",
	}
	beneficialAddress := util.Flag{
		Name:  "beneficialAddress",
		Must:  false,
		Usage: "beneficial account address hex string",
		Value: "",
	}
	pledgeAccount := util.Flag{
		Name:  "pledgeAccount",
		Must:  true,
		Usage: "pledge account private hex string",
	}
	amount := util.Flag{
		Name:  "amount",
		Must:  true,
		Usage: "pledge amount",
	}
	pType := util.Flag{
		Name:  "pType",
		Must:  true,
		Usage: "pledge type",
	}

	cmd := &ishell.Cmd{
		Name: "pledge",
		Help: "pledge token",
		Func: func(c *ishell.Context) {
			args := []util.Flag{beneficialAccount, beneficialAddress, pledgeAccount, amount, pType}
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			beneficialAccountP := util.StringVar(c.Args, beneficialAccount)
			beneficialAddressP := util.StringVar(c.Args, beneficialAddress)
			pledgeAccountP := util.StringVar(c.Args, pledgeAccount)
			amountP := util.StringVar(c.Args, amount)
			pTypeP := util.StringVar(c.Args, pType)

			fmt.Println(beneficialAccountP, pledgeAccountP, amountP, pTypeP)
			if err := pledgeAction(beneficialAccountP, beneficialAddressP, pledgeAccountP, amountP, pTypeP); err != nil {
				util.Warn(err)
				return
			}
		},
	}

	parentCmd.AddCmd(cmd)
}

func addPledgePledgeCmdByCobra(parentCmd *cobra.Command) {
	var beneficialAccountP string
	var beneficialAddressP string
	var pledgeAccountP string
	var amountP string
	var pTypeP string

	var cmd = &cobra.Command{
		Use:   "pledge",
		Short: "pledge token",
		Run: func(cmd *cobra.Command, args []string) {
			err := pledgeAction(beneficialAccountP, beneficialAddressP, pledgeAccountP, amountP, pTypeP)
			if err != nil {
				cmd.Println(err)
			}
		},
	}
	cmd.Flags().StringVar(&pledgeAccountP, "pAccount", "", "pledge account private hex string")
	cmd.Flags().StringVar(&beneficialAccountP, "bAccount", "", "beneficial account private hex string")
	cmd.Flags().StringVar(&beneficialAddressP, "bAddress", "", "beneficial account address hex string")
	cmd.Flags().StringVar(&amountP, "amount", "", "pledge amount")
	cmd.Flags().StringVar(&pTypeP, "pledgeType", "", "pledge type")
	parentCmd.AddCommand(cmd)
}

func pledgeAction(beneficialAccount, beneficialAddressP, pledgeAccount, amount, pType string) error {
	pBytes, err := hex.DecodeString(pledgeAccount)
	if err != nil {
		return err
	}
	p := types.NewAccount(pBytes)

	var bnfAccount *types.Account
	var bnfAddr types.Address
	if beneficialAccount != "" {
		bBytes, err := hex.DecodeString(beneficialAccount)
		if err != nil {
			return err
		}
		bnfAccount = types.NewAccount(bBytes)
		bnfAddr = bnfAccount.Address()
	} else if beneficialAddressP != "" {
		bnfAddr, err = types.HexToAddress(beneficialAddressP)
		if err != nil {
			return err
		}
	} else {
		return errors.New("beneficial account or address is empty")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	am := types.StringToBalance(amount)
	NEP5tTxId := random.RandomHexString(32)
	pledgeParam := api.PledgeParam{
		Beneficial: bnfAddr, PledgeAddress: p.Address(), Amount: am,
		PType: pType, NEP5TxId: NEP5tTxId,
	}
	fmt.Printf("\nPledgeAddress:%s, Beneficial:%s, NEP5tTxId:%s\n", pledgeParam.PledgeAddress, pledgeParam.Beneficial, NEP5tTxId)

	// send block
	send := types.StateBlock{}
	err = client.Call(&send, "pledge_getPledgeBlock", &pledgeParam)
	if err != nil {
		return err
	}
	sendHash := send.GetHash()
	send.Signature = p.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	// reward block
	var rewardPtr *types.StateBlock
	if bnfAccount != nil {
		reward := types.StateBlock{}
		rewardPtr = &reward
		err = client.Call(&reward, "pledge_getPledgeRewardBlock", &send)
		if err != nil {
			return err
		}
		rewardHash := reward.GetHash()
		reward.Signature = bnfAccount.Sign(rewardHash)
		var w2 types.Work
		worker2, _ := types.NewWorker(w2, reward.Root())
		reward.Work = worker2.NewWork()
	}

	//TODO: batch process send/reward
	fmt.Printf("sendHash:%s\n", sendHash)
	for try := 0; try < 3; try++ {
		err = client.Call(nil, "ledger_process", &send)
		if err != nil {
			fmt.Printf("send block, try %d err %s\n", try, err)
			time.Sleep(1 * time.Second)
			continue
		}

		break
	}

	if rewardPtr != nil {
		rewardHash := rewardPtr.GetHash()
		fmt.Printf("rewardHash:%s\n", rewardHash)
		for try := 0; try < 3; try++ {
			err = client.Call(nil, "ledger_process", rewardPtr)
			if err != nil {
				fmt.Printf("reward block, try %d err %s\n", try, err)
				time.Sleep(1 * time.Second)
				continue
			}

			break
		}
	}

	return nil
}

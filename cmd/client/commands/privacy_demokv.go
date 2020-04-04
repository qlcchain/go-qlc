package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPrivacyDemoKVSetCmdByShell(parentCmd *ishell.Cmd) {
	priKeyFlag := util.Flag{
		Name:  "priKey",
		Must:  true,
		Usage: "account private key",
		Value: "",
	}

	keyFlag := util.Flag{
		Name:  "key",
		Must:  true,
		Usage: "key of KV, default format is string, 0x will be HEX",
		Value: "",
	}
	valFlag := util.Flag{
		Name:  "value",
		Must:  true,
		Usage: "value of KV, default format is string, 0x will be HEX",
		Value: "",
	}

	priFromFlag := util.Flag{
		Name:  "privateFrom",
		Must:  false,
		Usage: "private from pubkey",
		Value: "",
	}
	priForFlag := util.Flag{
		Name:  "privateFor",
		Must:  false,
		Usage: "private for (destination) pubkey list",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "setDemoKV",
		Help: "set demo kv info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{priKeyFlag, priFromFlag, priForFlag, keyFlag, valFlag}
			if util.HelpText(c, args) {
				return
			}

			priKeyStr := util.StringVar(c.Args, priKeyFlag)
			keyStr := util.StringVar(c.Args, keyFlag)
			valStr := util.StringVar(c.Args, valFlag)

			priFromStr := util.StringVar(c.Args, priFromFlag)
			priForStr := util.StringVar(c.Args, priForFlag)

			err := runPrivacySetDemoKVCmd(priKeyStr, keyStr, valStr, priFromStr, priForStr)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPrivacySetDemoKVCmd(priKeyStr, keyStr, valStr, priFromStr, priForStr string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	if priKeyStr == "" {
		return errors.New("private key is nil")
	}
	pkBytes, err := hex.DecodeString(priKeyStr)
	if err != nil {
		return err
	}
	acc := types.NewAccount(pkBytes)

	if len(keyStr) == 0 {
		return errors.New("key is nil")
	}
	if len(valStr) == 0 {
		return errors.New("value is nil")
	}

	var keyBytes []byte
	if strings.HasPrefix(keyStr, "0x") {
		decRet, err := hex.DecodeString(keyStr[2:])
		if err != nil {
			return err
		}
		keyBytes = decRet
	} else {
		keyBytes = []byte(keyStr)
	}

	var valBytes []byte
	if strings.HasPrefix(valStr, "0x") {
		decRet, err := hex.DecodeString(valStr[2:])
		if err != nil {
			return err
		}
		valBytes = decRet
	} else {
		valBytes = []byte(valStr)
	}

	fmt.Printf("Key:\n")
	fmt.Printf("  HEX:    %s\n", hex.EncodeToString(keyBytes))
	fmt.Printf("  String: %s\n", string(keyBytes))
	fmt.Printf("  []byte: %v\n", keyBytes)
	fmt.Printf("Value:\n")
	fmt.Printf("  HEX:    %s\n", hex.EncodeToString(valBytes))
	fmt.Printf("  String: %s\n", string(valBytes))
	fmt.Printf("  []byte: %v\n", valBytes)

	paraStrList := []string{
		hex.EncodeToString(keyBytes),
		hex.EncodeToString(valBytes),
	}
	fmt.Printf("paraStrList: %s\n", paraStrList)

	var contractRspData []byte
	err = client.Call(&contractRspData, "contract_packChainContractData", contractaddress.PrivacyDemoKVAddress, "PrivacyDemoKVSet", paraStrList)
	if err != nil {
		return err
	}
	fmt.Printf("PackChainContractData:\n%s\n", hex.EncodeToString(contractRspData))

	contractSendPara := api.ContractSendBlockPara{
		Address:   acc.Address(),
		To:        contractaddress.PrivacyDemoKVAddress,
		TokenName: "QLC",
		Amount:    types.NewBalance(0),
		Data:      contractRspData,

		PrivateFrom: priFromStr,
		PrivateFor:  strings.Split(priForStr, ","),
	}
	var contractRspBlk types.StateBlock
	err = client.Call(&contractRspBlk, "contract_generateSendBlock", &contractSendPara)
	if err != nil {
		return err
	}
	blkHash := contractRspBlk.GetHash()
	contractRspBlk.Signature = acc.Sign(blkHash)
	fmt.Printf("GenerateSendBlock:\n%s, %s\n", blkHash, cutil.ToIndentString(contractRspBlk))

	var rspHash types.Hash
	err = client.Call(&rspHash, "ledger_process", contractRspBlk)
	if err != nil {
		return err
	}
	fmt.Printf("Process:\n%s\n", rspHash)

	return nil
}

func addPrivacyDemoKVGetCmdByShell(parentCmd *ishell.Cmd) {
	keyFlag := util.Flag{
		Name:  "key",
		Must:  true,
		Usage: "key of KV, default format is string, 0x will be HEX",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "getDemoKV",
		Help: "get demo kv info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{keyFlag}
			if util.HelpText(c, args) {
				return
			}

			keyStr := util.StringVar(c.Args, keyFlag)

			err := runPrivacyGetDemoKVCmd(keyStr)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPrivacyGetDemoKVCmd(keyStr string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	if len(keyStr) == 0 {
		return errors.New("key is nil")
	}

	var keyBytes []byte
	if strings.HasPrefix(keyStr, "0x") {
		decKey, err := hex.DecodeString(keyStr[2:])
		if err != nil {
			return err
		}
		keyBytes = decKey
	} else {
		keyBytes = []byte(keyStr)
	}
	fmt.Printf("Key:\n")
	fmt.Printf("  Bytes:  %v\n", keyBytes)
	fmt.Printf("  HEX:    %s\n", hex.EncodeToString(keyBytes))
	fmt.Printf("  String: %s\n", string(keyBytes))

	var valBytes []byte
	err = client.Call(&valBytes, "privacy_getDemoKV", keyBytes)
	if err != nil {
		return err
	}

	fmt.Printf("Value:\n")
	fmt.Printf("  Bytes:  %v\n", valBytes)
	fmt.Printf("  HEX:    %s\n", hex.EncodeToString(valBytes))
	fmt.Printf("  String: %s\n", string(valBytes))

	return nil
}

package commands

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPrivacyDistributePayloadCmdByShell(parentCmd *ishell.Cmd) {
	dataFlag := util.Flag{
		Name:  "data",
		Must:  true,
		Usage: "raw data of private transaction",
		Value: "",
	}
	priFromFlag := util.Flag{
		Name:  "privateFrom",
		Must:  true,
		Usage: "private from pubkey",
		Value: "",
	}
	priForFlag := util.Flag{
		Name:  "privateFor",
		Must:  true,
		Usage: "private for (destination) pubkey list",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "distributePayload",
		Help: "set demo kv info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{dataFlag, priFromFlag, priForFlag}
			if util.HelpText(c, args) {
				return
			}

			dataStr := util.StringVar(c.Args, dataFlag)
			priFromStr := util.StringVar(c.Args, priFromFlag)
			priForStr := util.StringVar(c.Args, priForFlag)

			err := runPrivacyDistributePayloadCmd(dataStr, priFromStr, priForStr)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPrivacyDistributePayloadCmd(dataStr, priFromStr, priForStr string) error {
	if len(dataStr) == 0 {
		return errors.New("data is nil")
	}
	if len(priFromStr) == 0 {
		return errors.New("privateFrom is nil")
	}
	if len(priForStr) == 0 {
		return errors.New("privateFor is nil")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	var dataBytes []byte
	if strings.HasPrefix(dataStr, "0x") {
		decKey, err := hex.DecodeString(dataStr[2:])
		if err != nil {
			return err
		}
		dataBytes = decKey
	} else {
		dataBytes, err = base64.StdEncoding.DecodeString(dataStr)
		if err != nil {
			dataBytes = []byte(dataStr)
		}
	}

	fmt.Printf("Data:\n")
	fmt.Printf("  HEX:    %s\n", hex.EncodeToString(dataBytes))
	fmt.Printf("  Base64: %s\n", base64.StdEncoding.EncodeToString(dataBytes))

	para := &api.PrivacyDistributeParam{
		RawPayload:  dataBytes,
		PrivateFrom: priFromStr,
		PrivateFor:  strings.Split(priForStr, ","),
	}

	var enclaveKeyBytes []byte
	err = client.Call(&enclaveKeyBytes, "privacy_distributeRawPayload", para)
	if err != nil {
		return err
	}

	fmt.Printf("EnclaveKey:\n")
	fmt.Printf("  HEX:    %s\n", hex.EncodeToString(enclaveKeyBytes))
	fmt.Printf("  Base64: %s\n", base64.StdEncoding.EncodeToString(enclaveKeyBytes))

	return nil
}

func addPrivacyGetPayloadCmdByShell(parentCmd *ishell.Cmd) {
	keyFlag := util.Flag{
		Name:  "enclaveKey",
		Must:  false,
		Usage: "enclave key of private transaction",
		Value: "",
	}
	hashFlag := util.Flag{
		Name:  "hash",
		Must:  false,
		Usage: "hash of private transaction",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "getRawPayload",
		Help: "get raw payload of private transaction",
		Func: func(c *ishell.Context) {
			args := []util.Flag{keyFlag, hashFlag}
			if util.HelpText(c, args) {
				return
			}

			keyStr := util.StringVar(c.Args, keyFlag)
			hashStr := util.StringVar(c.Args, hashFlag)

			var err error
			if hashStr != "" {
				err = runPrivacyGetPayloadByHashCmd(hashStr)
			} else {
				err = runPrivacyGetPayloadCmd(keyStr)
			}
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPrivacyGetPayloadCmd(keyStr string) error {
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
		keyBytes, err = base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			keyBytes = []byte(keyStr)
		}
	}
	fmt.Printf("EnclaveKey:\n")
	fmt.Printf("  HEX:    %s\n", hex.EncodeToString(keyBytes))
	fmt.Printf("  Base64: %s\n", base64.StdEncoding.EncodeToString(keyBytes))

	var valBytes []byte
	err = client.Call(&valBytes, "privacy_getRawPayload", keyBytes)
	if err != nil {
		return err
	}

	fmt.Printf("Payload:\n")
	fmt.Printf("  HEX:    %s\n", hex.EncodeToString(valBytes))
	fmt.Printf("  Base64: %s\n", base64.StdEncoding.EncodeToString(valBytes))

	return nil
}

func runPrivacyGetPayloadByHashCmd(hashStr string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	if len(hashStr) == 0 {
		return errors.New("hash is nil")
	}

	var valBytes []byte
	err = client.Call(&valBytes, "privacy_getBlockPrivatePayload", hashStr)
	if err != nil {
		return err
	}

	fmt.Printf("Payload:\n")
	fmt.Printf("  HEX:    %s\n", hex.EncodeToString(valBytes))
	fmt.Printf("  Base64: %s\n", base64.StdEncoding.EncodeToString(valBytes))

	return nil
}

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type ContractApi struct {
	l      ledger.Store
	eb     event.EventBus
	cc     *chainctx.ChainContext
	logger *zap.SugaredLogger
}

type ContractPrivacyParam struct {
	PrivateFrom    string   `json:"privateFrom"`
	PrivateFor     []string `json:"privateFor"`
	PrivateGroupID string   `json:"privateGroupID"`
}

func NewContractApi(cc *chainctx.ChainContext, l ledger.Store) *ContractApi {
	return &ContractApi{cc: cc, eb: cc.EventBus(), l: l, logger: log.NewLogger("api_contract")}
}

func (c *ContractApi) GetAbiByContractAddress(address types.Address) (string, error) {
	return contract.GetAbiByContractAddress(address)
}

func (c *ContractApi) PackContractData(abiStr string, methodName string, params []string) ([]byte, error) {
	abiContract, err := abi.JSONToABIContract(strings.NewReader(abiStr))
	if err != nil {
		return nil, err
	}
	method, ok := abiContract.Methods[methodName]
	if !ok {
		return nil, errors.New("method name not found")
	}
	arguments, err := convert(params, method.Inputs)
	if err != nil {
		return nil, err
	}
	return abiContract.PackMethod(methodName, arguments...)
}

func (c *ContractApi) PackChainContractData(contractAddress types.Address, methodName string, params []string) ([]byte, error) {
	abiStr, err := c.GetAbiByContractAddress(contractAddress)
	if err != nil {
		return nil, err
	}
	return c.PackContractData(abiStr, methodName, params)
}

func (c *ContractApi) ContractAddressList() []types.Address {
	return contractaddress.ChainContractAddressList
}

type ContractSendBlockPara struct {
	Address   types.Address `json:"address"`
	TokenName string        `json:"tokenName"`
	To        types.Address `json:"to"`
	Amount    types.Balance `json:"amount"`
	Data      []byte        `json:"data"`

	PrivateFrom    string   `json:"privateFrom,omitempty"`
	PrivateFor     []string `json:"privateFor,omitempty"`
	PrivateGroupID string   `json:"privateGroupID,omitempty"`
	EnclaveKey     []byte   `json:"enclaveKey,omitempty"`
}

func (c *ContractApi) GenerateSendBlock(para *ContractSendBlockPara) (*types.StateBlock, error) {
	if !c.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	// check parameters
	if para == nil {
		return nil, ErrParameterNil
	}
	if para.Amount.Int == nil || para.Address.IsZero() || para.To.IsZero() || para.TokenName == "" || len(para.Data) == 0 {
		return nil, errors.New("invalid transaction parameter")
	}

	if len(para.Data) > 0 && len(para.EnclaveKey) > 0 {
		return nil, errors.New("invalid Data and EnclaveKey parameter")
	}

	// check private parameters
	if len(para.PrivateFrom) > 0 {
		if len(para.PrivateFor) == 0 && para.PrivateGroupID == "" {
			return nil, errors.New("invalid PrivateFor and PrivateGroupID parameter")
		}
	}

	rawPayload := para.Data
	blkFillData := para.Data

	if len(para.EnclaveKey) > 0 {
		if len(para.PrivateFrom) == 0 {
			return nil, errors.New("invalid PrivateFrom parameter when EnclaveKey not nil")
		}

		// fetch raw payload by enclave key for private txs
		msgReq := &topic.EventPrivacyRecvReqMsg{
			EnclaveKey: para.EnclaveKey,

			RspChan: make(chan *topic.EventPrivacyRecvRspMsg, 1),
		}

		retPayload, err := privacyGetRawPayload(c.cc, msgReq)
		if err != nil {
			return nil, err
		}
		rawPayload = retPayload

		blkFillData = para.EnclaveKey
	} else if len(para.PrivateFrom) > 0 {
		// convert raw payload to enclave key for private txs
		msgReq := &topic.EventPrivacySendReqMsg{
			RawPayload:     para.Data,
			PrivateFrom:    para.PrivateFrom,
			PrivateFor:     para.PrivateFor,
			PrivateGroupID: para.PrivateGroupID,

			RspChan: make(chan *topic.EventPrivacySendRspMsg, 1),
		}

		enclaveKey, err := privacyDistributeRawPayload(c.cc, msgReq)
		if err != nil {
			return nil, err
		}

		blkFillData = enclaveKey
	}

	// check contract method whether exist or not
	cm, ok, err := contract.GetChainContract(para.To, rawPayload)
	if err != nil {
		return nil, err
	}
	if !ok || cm == nil {
		return nil, errors.New("chain contract method not exist")
	}

	// check account metas
	info, err := c.l.GetTokenByName(para.TokenName)
	if err != nil {
		return nil, err
	}
	tm, err := c.l.GetTokenMeta(para.Address, info.TokenId)
	if err != nil {
		return nil, errors.New("token not found")
	}
	prev, err := c.l.GetStateBlock(tm.Header)
	if err != nil {
		return nil, err
	}

	povHeader, err := c.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	// fill fields by api parameters
	sendBlk := &types.StateBlock{
		Address: para.Address,
		Token:   info.TokenId,
		Link:    para.To.ToHash(),
		Data:    blkFillData,

		PrivateFrom:    para.PrivateFrom,
		PrivateFor:     para.PrivateFor,
		PrivateGroupID: para.PrivateGroupID,
	}
	if len(para.PrivateFrom) > 0 {
		sendBlk.SetPrivatePayload(rawPayload)
	}

	// fill fields by ledger account metas
	sendBlk.Type = types.ContractSend
	sendBlk.Balance = tm.Balance.Sub(para.Amount)
	sendBlk.Previous = tm.Header
	sendBlk.Representative = tm.Representative
	sendBlk.Vote = prev.Vote
	sendBlk.Network = prev.Network
	sendBlk.Oracle = prev.Oracle
	sendBlk.Storage = prev.Storage

	sendBlk.Timestamp = common.TimeNow().Unix()
	sendBlk.PoVHeight = povHeader.GetHeight()

	// pre-running contract method send action
	vmCtx := vmstore.NewVMContext(c.l, &para.To)
	cd := cm.GetDescribe()
	if cd.GetVersion() == contract.SpecVer1 {
		if err := cm.DoSend(vmCtx, sendBlk); err != nil {
			return nil, err
		}

		h := vmstore.TrieHash(vmCtx)
		if h != nil {
			sendBlk.Extra = h
		}
	} else if cd.GetVersion() == contract.SpecVer2 {
		_, _, err := cm.ProcessSend(vmCtx, sendBlk)
		if err != nil {
			return nil, err
		}

		h := vmstore.TrieHash(vmCtx)
		if h != nil {
			sendBlk.Extra = h
		}
	} else {
		return nil, errors.New("invalid contract version")
	}

	return sendBlk, nil
}

type ContractRewardBlockPara struct {
	SendHash types.Hash `json:"sendHash"`
	Data     []byte     `json:"data"`

	PrivateFrom    string   `json:"privateFrom,omitempty"`
	PrivateFor     []string `json:"privateFor,omitempty"`
	PrivateGroupID string   `json:"privateGroupID,omitempty"`
}

func (c *ContractApi) GenerateRewardBlock(para *ContractRewardBlockPara) (*types.StateBlock, error) {
	if !c.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	// check parameters
	if para == nil {
		return nil, ErrParameterNil
	}

	if para.SendHash.IsZero() {
		return nil, errors.New("invalid transaction parameter")
	}

	if len(para.Data) == 0 && len(para.PrivateFrom) > 0 {
		para.PrivateFrom = ""
	}

	sendBlk, err := c.l.GetStateBlockConfirmed(para.SendHash)
	if err != nil {
		return nil, err
	}

	// make sure private raw data exist
	if len(sendBlk.PrivateFrom) > 0 && len(sendBlk.GetPayload()) == 0 {
		msgReq := &topic.EventPrivacyRecvReqMsg{
			EnclaveKey: sendBlk.GetData(),

			RspChan: make(chan *topic.EventPrivacyRecvRspMsg, 1),
		}

		rawData, err := privacyGetRawPayload(c.cc, msgReq)
		if err != nil {
			return nil, err
		}
		if len(rawData) == 0 {
			return nil, errors.New("send is private but this node is not recipient")
		}

		sendBlk.SetPrivatePayload(rawData)
	}

	// check contract method whether exist or not
	ca, err := types.BytesToAddress(sendBlk.Link.Bytes())
	if err != nil {
		return nil, err
	}
	cm, ok, err := contract.GetChainContract(ca, sendBlk.GetPayload())
	if err != nil {
		return nil, err
	}
	if !ok || cm == nil {
		return nil, errors.New("chain contract method not exist")
	}

	povHeader, err := c.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	rawPayload := para.Data
	blkFillData := para.Data

	if len(para.Data) > 0 && len(para.PrivateFrom) > 0 {
		// convert raw payload to enclave key for private txs
		msgReq := &topic.EventPrivacySendReqMsg{
			RawPayload:     para.Data,
			PrivateFrom:    para.PrivateFrom,
			PrivateFor:     para.PrivateFor,
			PrivateGroupID: para.PrivateGroupID,

			RspChan: make(chan *topic.EventPrivacySendRspMsg, 1),
		}

		enclaveKey, err := privacyDistributeRawPayload(c.cc, msgReq)
		if err != nil {
			return nil, err
		}

		blkFillData = enclaveKey
	}

	// fill default fields in reward block, but all fields can be set by DoReceive action
	recvBlk := &types.StateBlock{
		Type:           types.ContractReward,
		Link:           para.SendHash,
		Data:           blkFillData,
		PrivateFrom:    para.PrivateFrom,
		PrivateFor:     para.PrivateFor,
		PrivateGroupID: para.PrivateGroupID,
		Timestamp:      common.TimeNow().Unix(),
		PoVHeight:      povHeader.GetHeight(),
	}

	if len(para.Data) > 0 && len(para.PrivateFrom) > 0 {
		recvBlk.SetPrivatePayload(rawPayload)
	}

	// pre-running contract method receive action
	vmCtx := vmstore.NewVMContextWithBlock(c.l, sendBlk)
	if vmCtx == nil {
		return nil, errors.New("generate reward block: can not get vm context")
	}
	g, err := cm.DoReceive(vmCtx, recvBlk, sendBlk)
	if err != nil {
		return nil, err
	}
	if len(g) == 0 {
		return nil, errors.New("run DoReceive got empty block")
	}

	h := vmstore.TrieHash(g[0].VMContext)
	if h != nil {
		recvBlk.Extra = h
	}

	return recvBlk, nil
}

func convert(params []string, arguments abi.Arguments) ([]interface{}, error) {
	if len(params) != len(arguments) {
		return nil, errors.New("argument size not match")
	}
	resultList := make([]interface{}, len(params))
	for i, argument := range arguments {
		result, err := convertOne(params[i], argument.Type)
		if err != nil {
			return nil, err
		}
		resultList[i] = result
	}
	return resultList, nil
}

func convertOne(param string, t abi.Type) (interface{}, error) {
	typeString := t.String()
	if strings.Contains(typeString, "[") {
		return convertToArray(param, t)
	} else if typeString == "bool" {
		return convertToBool(param)
	} else if strings.HasPrefix(typeString, "int") {
		return convertToInt(param, t.Size)
	} else if strings.HasPrefix(typeString, "uint") {
		return convertToUint(param, t.Size)
	} else if typeString == "address" {
		return types.HexToAddress(param)
	} else if typeString == "tokenId" || typeString == "hash" {
		return types.NewHash(param)
	} else if typeString == "signature" {
		return types.NewSignature(param)
	} else if typeString == "string" {
		return param, nil
	} else if typeString == "bytes" {
		return convertToDynamicBytes(param)
	} else if strings.HasPrefix(typeString, "bytes") {
		return convertToFixedBytes(param, t.Size)
	}
	return nil, errors.New("unknown type " + typeString)
}

func convertToArray(param string, t abi.Type) (interface{}, error) {
	if t.Elem.Elem != nil {
		return nil, errors.New(t.String() + " type not supported")
	}
	typeString := t.Elem.String()
	if typeString == "bool" {
		return convertToBoolArray(param)
	} else if strings.HasPrefix(typeString, "int") {
		return convertToIntArray(param, *t.Elem)
	} else if strings.HasPrefix(typeString, "uint") {
		return convertToUintArray(param, *t.Elem)
	} else if typeString == "address" {
		return convertToAddressArray(param)
	} else if typeString == "tokenId" || typeString == "hash" {
		return convertToTokenIdArray(param)
	} else if typeString == "string" {
		return convertToStringArray(param)
	} else if typeString == "signature" {
		return convertToSignatureArray(param)
	}
	return nil, errors.New(typeString + " array type not supported")
}

func convertToBoolArray(param string) (interface{}, error) {
	resultList := make([]bool, 0)
	if err := json.Unmarshal([]byte(param), &resultList); err != nil {
		return nil, err
	}
	return resultList, nil
}

func convertToIntArray(param string, t abi.Type) (interface{}, error) {
	size := t.Size
	if size == 8 {
		resultList := make([]int8, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 16 {
		resultList := make([]int16, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 32 {
		resultList := make([]int32, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 64 {
		resultList := make([]int64, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else {
		resultList := make([]*big.Int, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	}
}
func convertToUintArray(param string, t abi.Type) (interface{}, error) {
	size := t.Size
	if size == 8 {
		resultList := make([]uint8, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 16 {
		resultList := make([]uint16, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 32 {
		resultList := make([]uint32, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else if size == 64 {
		resultList := make([]uint64, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	} else {
		resultList := make([]*big.Int, 0)
		if err := json.Unmarshal([]byte(param), &resultList); err != nil {
			return nil, err
		}
		return resultList, nil
	}
}
func convertToAddressArray(param string) (interface{}, error) {
	resultList := make([]types.Address, 0)
	if err := json.Unmarshal([]byte(param), &resultList); err != nil {
		return nil, err
	}
	return resultList, nil
}

func convertToTokenIdArray(param string) (interface{}, error) {
	resultList := make([]types.Hash, 0)
	if err := json.Unmarshal([]byte(param), &resultList); err != nil {
		return nil, err
	}
	return resultList, nil
}

func convertToSignatureArray(param string) (interface{}, error) {
	resultList := make([]types.Signature, 0)
	if err := json.Unmarshal([]byte(param), &resultList); err != nil {
		return nil, err
	}
	return resultList, nil
}

func convertToStringArray(param string) (interface{}, error) {
	resultList := make([]string, 0)
	if err := json.Unmarshal([]byte(param), &resultList); err != nil {
		return nil, err
	}
	return resultList, nil
}

func convertToBool(param string) (interface{}, error) {
	if param == "true" {
		return true, nil
	} else {
		return false, nil
	}
}

func convertToInt(param string, size int) (interface{}, error) {
	bigInt, ok := new(big.Int).SetString(param, 0)
	if !ok || bigInt.BitLen() > size-1 {
		return nil, errors.New(param + " convert to int failed")
	}
	if size == 8 {
		return int8(bigInt.Int64()), nil
	} else if size == 16 {
		return int16(bigInt.Int64()), nil
	} else if size == 32 {
		return int32(bigInt.Int64()), nil
	} else if size == 64 {
		return int64(bigInt.Int64()), nil
	} else {
		return bigInt, nil
	}
}

func convertToUint(param string, size int) (interface{}, error) {
	bigInt, ok := new(big.Int).SetString(param, 0)
	if !ok || bigInt.BitLen() > size {
		return nil, errors.New(param + " convert to uint failed")
	}
	if size == 8 {
		return uint8(bigInt.Uint64()), nil
	} else if size == 16 {
		return uint16(bigInt.Uint64()), nil
	} else if size == 32 {
		return uint32(bigInt.Uint64()), nil
	} else if size == 64 {
		return uint64(bigInt.Uint64()), nil
	} else {
		return bigInt, nil
	}
}

func convertToFixedBytes(param string, size int) (interface{}, error) {
	if len(param) != size*2 {
		return nil, errors.New(param + " is not valid bytes")
	}
	return hex.DecodeString(param)
}
func convertToDynamicBytes(param string) (interface{}, error) {
	return hex.DecodeString(param)
}

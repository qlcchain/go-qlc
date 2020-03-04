/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/validator.v2"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonSettlement = `[
  {
    "type": "function",
    "name": "CreateContract",
    "inputs": [
        { "name": "partyA", "type": "address" },
        { "name": "partyAName", "type": "string" },
        { "name": "partyB", "type": "address" },
        { "name": "partyBName", "type": "string" },
        { "name": "previous", "type": "hash" },
        { "name": "serviceId", "type": "string" },
        { "name": "mcc", "type": "uint64" },
        { "name": "mnc", "type": "uint64" },
        { "name": "totalAmount", "type": "uint64" },
        { "name": "unitPrice", "type": "uint64" },
        { "name": "currency", "type": "string" },
        { "name": "signDate", "type": "int64" },
        { "name": "signatureA", "type": "signature" }
    ]
  }, {
    "type": "function",
    "name": "SignContract",
    "inputs": [
        { "name": "contractAddress", "type": "address" },
        { "name": "confirmDate", "type": "int64" },
        { "name": "signatureB", "type": "signature" }
    ]
  }, {
    "type": "function",
    "name": "ProcessCDR",
    "inputs": [
        { "name": "index", "type": "uint64" },
        { "name": "smsDt", "type": "int64" },
        { "name": "sender", "type": "string" },
        { "name": "destination", "type": "string" },
        { "name": "dstCountry", "type": "string" },
        { "name": "dstOperator", "type": "string" },
        { "name": "dstMcc", "type": "string" },
        { "name": "dstMnc", "type": "string" },
        { "name": "sellPrice", "type": "string" },
        { "name": "sellCurrency", "type": "string" },
        { "name": "customerName", "type": "string" },
        { "name": "customerID", "type": "string" },
        { "name": "sendingStatus", "type": "string" },
        { "name": "dlrStatus", "type": "string" }
    ]
  }, {
    "type": "function",
    "name": "AddPreStop",
    "inputs": [
        { "name": "stop", "type": "string" }
    ]
  }, {
    "type": "function",
    "name": "RemovePreStop",
    "inputs": [
        { "name": "stop", "type": "string" }
    ]
  },{
    "type": "function",
    "name": "UpdatePreStop",
    "inputs": [
        { "name": "old", "type": "string" },
        { "name": "new", "type": "string" }
    ]
  }, {
    "type": "function",
    "name": "AddNextStop",
    "inputs": [
        { "name": "stop", "type": "string" }
    ]
  }, {
    "type": "function",
    "name": "RemoveNextStop",
    "inputs": [
        { "name": "stop", "type": "string" }
    ]
  },{
    "type": "function",
    "name": "UpdateNextStop",
    "inputs": [
        { "name": "old", "type": "string" },
        { "name": "new", "type": "string" }
    ]
  },{
    "type": "function",
    "name": "TerminateContract",
    "inputs": [
        { "name": "contractAddress", "type": "address" }
    ]
  }
]
`
	MethodNameCreateContract    = "CreateContract"
	MethodNameSignContract      = "SignContract"
	MethodNameTerminateContract = "TerminateContract"
	MethodNameProcessCDR        = "ProcessCDR"
	MethodNameAddPreStop        = "AddPreStop"
	MethodNameRemovePreStop     = "RemovePreStop"
	MethodNameUpdatePreStop     = "UpdatePreStop"
	MethodNameAddNextStop       = "AddNextStop"
	MethodNameRemoveNextStop    = "RemoveNextStop"
	MethodNameUpdateNextStop    = "UpdateNextStop"
)

var (
	SettlementABI, _ = abi.JSONToABIContract(strings.NewReader(JsonSettlement))
	keySize          = types.AddressSize*2 + 1
	mappingPrefix    = []byte{33}
)

type ABIer interface {
	ToABI() ([]byte, error)
	FromABI(data []byte) error
}

type Verifier interface {
	Verify() error
}

//go:generate msgp
type SignContractParam struct {
	ContractAddress types.Address `msg:"a,extension" json:"contractAddress"`
	ConfirmDate     int64         `msg:"cd" json:"confirmDate"`
}

func (z *SignContractParam) Verify() (bool, error) {
	if z.ContractAddress.IsZero() {
		return false, fmt.Errorf("invalid contract address %s", z.ContractAddress.String())
	}
	if z.ConfirmDate == 0 {
		return false, errors.New("invalid contract confirm date")
	}
	return true, nil
	//if h, err := types.HashBytes(z.ContractAddress[:], util.BE_Int2Bytes(z.ConfirmDate)); err != nil {
	//	return false, err
	//} else {
	//	return addr.Verify(h[:], z.SignatureB[:]), nil
	//}
}

func (z *SignContractParam) ToABI() ([]byte, error) {
	//return SettlementABI.PackMethod(MethodNameSignContract, z.ContractAddress, z.ConfirmDate, z.SignatureB)
	id := SettlementABI.Methods[MethodNameSignContract].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *SignContractParam) FromABI(data []byte) error {
	//return SettlementABI.UnpackMethod(z, MethodNameSignContract, data)
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

//go:generate msgp
type TerminateParam struct {
	ContractAddress types.Address `msg:"a,extension" json:"contractAddress"`
}

func (z *TerminateParam) ToABI() ([]byte, error) {
	id := SettlementABI.Methods[MethodNameTerminateContract].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *TerminateParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *TerminateParam) Verify() error {
	if z.ContractAddress.IsZero() {
		return errors.New("invalid contract address")
	}
	return nil
}

func (z *TerminateParam) String() string {
	return util.ToIndentString(z)
}

//go:generate msgp
type Contractor struct {
	Address types.Address `msg:"a,extension" json:"address"`
	Name    string        `msg:"n" json:"name"`
}

func (z *Contractor) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *Contractor) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

//go:generate msgp
type ContractService struct {
	ServiceId   string  `msg:"id" json:"serviceId" validate:"nonzero"`
	Mcc         uint64  `msg:"mcc" json:"mcc"`
	Mnc         uint64  `msg:"mnc" json:"mnc"`
	TotalAmount uint64  `msg:"t" json:"totalAmount" validate:"min=1"`
	UnitPrice   float64 `msg:"u" json:"unitPrice" validate:"nonzero"`
	Currency    string  `msg:"c" json:"currency" validate:"nonzero"`
}

func (z *ContractService) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *ContractService) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *ContractService) Balance() (types.Balance, error) {
	f := z.UnitPrice * 1e8
	if mul, b := util.SafeMul(z.TotalAmount, uint64(f)); b {
		return types.ZeroBalance, fmt.Errorf("overflow when mul %d and %f", z.TotalAmount, z.UnitPrice)
	} else {
		return types.Balance{Int: new(big.Int).SetUint64(mul)}, nil
	}
}

//go:generate msgp
type CreateContractParam struct {
	PartyA    Contractor        `msg:"pa" json:"partyA"`
	PartyB    Contractor        `msg:"pb" json:"partyB"`
	Previous  types.Hash        `msg:"pre,extension" json:"previous"`
	Services  []ContractService `msg:"s" json:"services"`
	SignDate  int64             `msg:"t1" json:"signDate"`
	StartDate int64             `msg:"t3" json:"startDate"`
	EndDate   int64             `msg:"t4" json:"endDate"`
	//SignatureA *types.Signature  `msg:"sa,extension" json:"signatureA"`
}

func (z *CreateContractParam) Verify() (bool, error) {
	if z.PartyA.Address.IsZero() || len(z.PartyA.Name) == 0 {
		return false, fmt.Errorf("invalid partyA params")
	}

	if z.PartyB.Address.IsZero() || len(z.PartyB.Name) == 0 {
		return false, fmt.Errorf("invalid partyB params")
	}

	if z.Previous.IsZero() {
		return false, errors.New("invalid previous hash")
	}

	if len(z.Services) == 0 {
		return false, errors.New("empty contract services")
	}

	for _, s := range z.Services {
		if err := validator.Validate(s); err != nil {
			return false, err
		}
	}

	if z.SignDate <= 0 {
		return false, fmt.Errorf("invalid sign date %d", z.SignDate)
	}

	if z.StartDate <= 0 {
		return false, fmt.Errorf("invalid start date %d", z.StartDate)
	}

	if z.EndDate <= 0 {
		return false, fmt.Errorf("invalid end date %d", z.EndDate)
	}

	return true, nil
}

func (z *CreateContractParam) ToContractParam() *ContractParam {
	return &ContractParam{
		CreateContractParam: *z,
		ConfirmDate:         0,
		Status:              ContractStatusActiveStage1,
	}
}

func (z *CreateContractParam) ToABI() ([]byte, error) {
	//return SettlementABI.PackMethod(MethodNameCreateContract, z.PartyA, z.PartyAName, z.PartyB, z.PartyBName, z.Previous,
	//	z.ServiceId, z.Mcc, z.Mnc, z.TotalAmount, z.UnitPrice, z.Currency, z.SignDate, z.SignatureA)
	id := SettlementABI.Methods[MethodNameCreateContract].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *CreateContractParam) FromABI(data []byte) error {
	//return SettlementABI.UnpackMethod(z, MethodNameCreateContract, data)
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *CreateContractParam) String() string {
	return util.ToIndentString(z)
}

func ParseContractParam(v []byte) (*ContractParam, error) {
	cp := &ContractParam{}
	if err := cp.FromABI(v); err != nil {
		return nil, err
	} else {
		return cp, nil
	}
}

func (z *CreateContractParam) Address() (types.Address, error) {
	if _, err := z.Verify(); err != nil {
		return types.ZeroAddress, err
	}

	var result []byte
	if data, err := z.PartyA.ToABI(); err != nil {
		return types.ZeroAddress, err
	} else {
		result = append(result, data...)
	}
	if data, err := z.PartyB.ToABI(); err != nil {
		return types.ZeroAddress, err
	} else {
		result = append(result, data...)
	}

	result = append(result, z.Previous[:]...)
	for _, s := range z.Services {
		if data, err := s.ToABI(); err != nil {
			return types.ZeroAddress, err
		} else {
			result = append(result, data...)
		}
	}
	result = append(result, util.BE_Int2Bytes(z.SignDate)...)

	hash := types.HashData(result)

	return types.BytesToAddress(hash[:])
}

func (z *CreateContractParam) Balance() (types.Balance, error) {
	total := types.ZeroBalance
	for _, service := range z.Services {
		if b, err := service.Balance(); err != nil {
			return types.ZeroBalance, err
		} else {
			total = total.Add(b)
		}
	}
	return total, nil
}

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
ActiveStage1
Activated
DestroyStage1
Destroyed
Rejected
)
*/
type ContractStatus int

//go:generate msgp
type ContractParam struct {
	CreateContractParam
	PreStops    []string       `msg:"pre" json:"preStops,omitempty"`
	NextStops   []string       `msg:"nex" json:"nextStops,omitempty"`
	ConfirmDate int64          `msg:"t2" json:"confirmDate"`
	Status      ContractStatus `msg:"s" json:"status"`
	Terminator  *types.Address `msg:"t,extension" json:"terminator,omitempty"`
}

func (z *ContractParam) IsPreStop(n string) bool {
	if len(z.PreStops) == 0 {
		return false
	}
	for _, stop := range z.PreStops {
		if stop == n {
			return true
		}
	}
	return false
}

func (z *ContractParam) IsNextStop(n string) bool {
	if len(z.NextStops) == 0 {
		return false
	}
	for _, stop := range z.NextStops {
		if stop == n {
			return true
		}
	}
	return false
}

func (z *ContractParam) IsContractor(addr types.Address) bool {
	return z.PartyA.Address == addr || z.PartyB.Address == addr
}

func (z *ContractParam) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *ContractParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *ContractParam) Equal(cp *CreateContractParam) (bool, error) {
	if cp == nil {
		return false, errors.New("invalid input value")
	}

	a1, err := z.Address()
	if err != nil {
		return false, err
	}

	a2, err := cp.Address()
	if err != nil {
		return false, err
	}

	if a1 == a2 {
		return true, nil
	} else {
		return false, fmt.Errorf("invalid address, exp: %s,act: %s", a1.String(), a2.String())
	}
}

func (z *ContractParam) DoActive(operator types.Address) error {
	if z.PartyB.Address != operator {
		return fmt.Errorf("invalid partyB, exp: %s, act: %s", z.PartyB.Address.String(), operator.String())
	}

	if z.Status == ContractStatusActiveStage1 {
		z.Status = ContractStatusActivated
		return nil
	} else if z.Status == ContractStatusDestroyed {
		return errors.New("contract has been destroyed")
	} else {
		return fmt.Errorf("invalid contract status, %s", z.Status.String())
	}
}

func (z *ContractParam) DoTerminate(operator types.Address) error {
	if b := z.IsContractor(operator); !b {
		return fmt.Errorf("permission denied, only contractor can terminate it, exp: %s or %s, act: %s",
			z.PartyA.Address.String(), z.PartyB.Address.String(), operator.String())
	}

	if z.Terminator != nil {
		if reflect.DeepEqual(z.Terminator, &operator) {
			return fmt.Errorf("%s already terminated contract", operator.String())
		}
	} else {
		z.Terminator = &operator
	}

	if z.Status == ContractStatusActiveStage1 {
		if z.PartyA.Address == operator {
			z.Status = ContractStatusDestroyed
		} else {
			z.Status = ContractStatusRejected
		}
	} else if z.Status == ContractStatusActivated {
		z.Status = ContractStatusDestroyStage1
	} else if z.Status == ContractStatusDestroyStage1 {
		z.Status = ContractStatusDestroyed
	} else {
		return fmt.Errorf("invalid contract status, %s", z.Status.String())
	}

	return nil
}

func (z *ContractParam) String() string {
	return util.ToIndentString(z)
}

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
Sent
Error
Empty
)
*/
type SendingStatus int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
Delivered
Rejected
Unknown
Undelivered
Empty
)
*/
type DLRStatus int

// TODO:
// we should make sure that can use CDR data to match to a specific settlement contract
//go:generate msgp
type CDRParam struct {
	ContractAddress types.Address `msg:"a" json:"contractAddress"`
	Index           uint64        `msg:"i" json:"index" validate:"min=1"`
	SmsDt           int64         `msg:"dt" json:"smsDt" validate:"min=1"`
	Sender          string        `msg:"tx" json:"sender" validate:"nonzero"`
	Destination     string        `msg:"d" json:"destination" validate:"nonzero"`
	//DstCountry    string        `msg:"dc" json:"dstCountry" validate:"nonzero"`
	//DstOperator   string        `msg:"do" json:"dstOperator" validate:"nonzero"`
	//DstMcc        uint64        `msg:"mcc" json:"dstMcc"`
	//DstMnc        uint64        `msg:"mnc" json:"dstMnc"`
	//SellPrice     float64       `msg:"p" json:"sellPrice" validate:"nonzero"`
	//SellCurrency  string        `msg:"c" json:"sellCurrency" validate:"nonzero"`
	//CustomerName  string        `msg:"cn" json:"customerName" validate:"nonzero"`
	//CustomerID    string        `msg:"cid" json:"customerID" validate:"nonzero"`
	SendingStatus SendingStatus `msg:"s" json:"sendingStatus"`
	DlrStatus     DLRStatus     `msg:"ds" json:"dlrStatus"`
	PreStop       string        `msg:"ps" json:"preStop" `
	NextStop      string        `msg:"ns" json:"nextStop" `
	//MessageID    string  `msg:"id" json:"messageID"`
	//connection         string
	//clientIp           string
	//failureCode        string
	//dlrDt              string
	//buyPrice           float64
	//buyCurrency        string
	//mnpPrice           string
	//mnpCurrency        string
	//supplierGateName   string
	//supplierGateID     string
	//supplierCustomerID string
	//briefMessage       string
	//foreignMessageID   string
}

func (z *CDRParam) ToABI() ([]byte, error) {
	id := SettlementABI.Methods[MethodNameProcessCDR].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *CDRParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *CDRParam) String() string {
	return util.ToIndentString(z)
}

func (z *CDRParam) Status() bool {
	switch z.DlrStatus {
	case DLRStatusDelivered:
		return true
	case DLRStatusUndelivered:
		return false
	case DLRStatusUnknown:
		fallthrough
	case DLRStatusEmpty:
		switch z.SendingStatus {
		case SendingStatusSent:
			return true
		default:
			return false
		}
	}
	return false
}

func (z *CDRParam) Verify() error {
	if errs := validator.Validate(z); errs != nil {
		return errs
	}
	return nil
}

func (z *CDRParam) ToHash() (types.Hash, error) {
	return types.HashBytes(util.BE_Uint64ToBytes(z.Index), []byte(z.Sender), []byte(z.Destination))
}

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
unknown
stage1
success
failure
missing
duplicate
)
*/
type SettlementStatus int

type SettlementCDR struct {
	CDRParam
	From types.Address `msg:"f,extension" json:"from"`
}

//go:generate msgp
type CDRStatus struct {
	Params map[string][]CDRParam `msg:"p" json:"params"`
	Status SettlementStatus      `msg:"s" json:"status"`
}

func (z *CDRStatus) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *CDRStatus) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *CDRStatus) ToHash() (types.Hash, error) {
	if len(z.Params) > 0 {
		for _, v := range z.Params {
			if len(v) > 0 {
				return v[0].ToHash()
			}
		}
		//keys := reflect.ValueOf(z.Params).MapKeys()
		//return z.Params[keys[0].Interface().(string)][0].ToHash()
	}
	return types.ZeroHash, errors.New("no cdr record")
}

func (z *CDRStatus) State(addr *types.Address) (string, bool, error) {
	if params, ok := z.Params[addr.String()]; ok {
		switch size := len(params); {
		case size == 1:
			return params[0].Sender, params[0].Status(), nil
		case size > 1: //upload multi-times or normalize time error
			return params[0].Sender, false, nil
		default:
			return "", false, nil
		}
	} else {
		return "", false, fmt.Errorf("can not find data of %s", addr.String())
	}
}

func (z *CDRStatus) IsInCycle(start, end int64) bool {
	if len(z.Params) == 0 {
		return false
	}

	if start != 0 && end != 0 {
		i := 0
		for _, params := range z.Params {
			if len(params) > 0 {
				param := params[0]
				if param.SmsDt >= start && param.SmsDt <= end {
					return true
				}
			}
			i++
		}
		if i == len(z.Params) {
			return false
		}
	}
	return true
}

func (z *CDRStatus) ExtractID() (dt int64, sender, destination string, err error) {
	if len(z.Params) > 0 {
		for _, v := range z.Params {
			if len(v) >= 1 {
				return v[0].SmsDt, v[0].Sender, v[0].Destination, nil
			}
		}
	}

	return 0, "", "", errors.New("can not find any CDR param")
}

func (z *CDRStatus) String() string {
	return util.ToIndentString(z)
}

// DoSettlement process settlement
// @param cdr  cdr data
func (z *CDRStatus) DoSettlement(cdr SettlementCDR) (err error) {
	if z.Params == nil {
		z.Params = make(map[string][]CDRParam, 0)
	}

	from := cdr.From.String()
	if params, ok := z.Params[from]; ok {
		params = append(params, cdr.CDRParam)
		z.Params[from] = params
	} else {
		z.Params[from] = []CDRParam{cdr.CDRParam}
	}

	switch size := len(z.Params); {
	//case size == 0:
	//	z.Status = SettlementStatusUnknown
	//	break
	case size == 1:
		z.Status = SettlementStatusStage1
		break
	case size == 2:
		z.Status = SettlementStatusSuccess
		b := true
		// combine all status
		for _, params := range z.Params {
			//for _, param := range params {
			//	b = b && param.Status()
			//}
			switch l := len(params); {
			case l > 1:
				z.Status = SettlementStatusDuplicate
				return
			case l == 1:
				b = b && params[0].Status()
				break
			}
		}
		if !b {
			z.Status = SettlementStatusFailure
		}
	case size > 2:
		err = fmt.Errorf("invalid params size %d", size)
	}
	return err
}

func ParseCDRStatus(v []byte) (*CDRStatus, error) {
	state := &CDRStatus{}
	if err := state.FromABI(v); err != nil {
		return nil, err
	} else {
		return state, nil
	}
}

//go:generate msgp
type StopParam struct {
	ContractAddress types.Address `msg:"ca" json:"contractAddress"`
	StopName        string        `msg:"n" json:"stopName" validate:"nonzero"`
}

func (z *StopParam) ToABI(methodName string) ([]byte, error) {
	id := SettlementABI.Methods[methodName].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *StopParam) FromABI(methodName string, data []byte) error {
	if method, err := SettlementABI.MethodById(data[:4]); err == nil && method.Name == methodName {
		_, err := z.UnmarshalMsg(data[4:])
		return err
	} else {
		return fmt.Errorf("could not locate named method: %s", methodName)
	}
}

func (z *StopParam) Verify() error {
	return validator.Validate(z)
}

//go:generate msgp
type UpdateStopParam struct {
	ContractAddress types.Address `msg:"ca" json:"contractAddress"`
	StopName        string        `msg:"n" json:"stopName" validate:"nonzero"`
	New             string        `msg:"n2" json:"newName" validate:"nonzero"`
}

func (z *UpdateStopParam) ToABI(methodName string) ([]byte, error) {
	id := SettlementABI.Methods[methodName].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *UpdateStopParam) FromABI(methodName string, data []byte) error {
	if method, err := SettlementABI.MethodById(data[:4]); err == nil && method.Name == methodName {
		_, err := z.UnmarshalMsg(data[4:])
		return err
	} else {
		return fmt.Errorf("could not locate named method: %s", methodName)
	}
}

func (z *UpdateStopParam) Verify() error {
	return validator.Validate(z)
}

// IsContractAvailable check contract status by contract ID
func IsContractAvailable(ctx *vmstore.VMContext, addr *types.Address) bool {
	if value, err := ctx.GetStorage(types.SettlementAddress[:], addr[:]); err == nil {
		if param, err := ParseContractParam(value); err == nil {
			unix := common.TimeNow().Unix()
			return param.Status == ContractStatusActivated && unix >= param.StartDate && unix <= param.EndDate
		}
	}
	return false
}

// GetCDRStatus
// @param addr settlement contract address
// @param CDR data hash
func GetCDRStatus(ctx *vmstore.VMContext, addr *types.Address, hash types.Hash) (*CDRStatus, error) {
	logger := log.NewLogger("GetCDRStatus")
	defer func() {
		_ = logger.Sync()
	}()

	if storage, err := ctx.GetStorage(addr[:], hash[:]); err != nil {
		return nil, err
	} else {
		status := &CDRStatus{}
		if err := status.FromABI(storage); err != nil {
			return nil, err
		} else {
			return status, nil
		}
	}
}

//go:generate msgp
type ContractAddressList struct {
	AddressList []*types.Address `msg:"a" json:"addressList"`
}

func newContractAddressList(address *types.Address) *ContractAddressList {
	return &ContractAddressList{AddressList: []*types.Address{address}}
}

func (z *ContractAddressList) Append(address *types.Address) bool {
	ret := false
	for _, a := range z.AddressList {
		if a == address {
			ret = true
			break
		}
	}

	if !ret {
		z.AddressList = append(z.AddressList, address)
		return true
	}
	return false
}

func (z *ContractAddressList) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *ContractAddressList) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *ContractAddressList) String() string {
	return util.ToIndentString(z.AddressList)
}

func SaveCDRStatus(ctx *vmstore.VMContext, addr *types.Address, hash *types.Hash, state *CDRStatus) error {
	if data, err := state.ToABI(); err != nil {
		return err
	} else {
		if err := ctx.SetStorage(addr[:], hash[:], data); err != nil {
			return err
		} else {
			// save keymap
			if err := saveCDRMapping(ctx, addr, hash); err != nil {
				return err
			}
		}
	}

	return nil
}

func saveCDRMapping(ctx *vmstore.VMContext, addr *types.Address, hash *types.Hash) error {
	if storage, err := ctx.GetStorage(mappingPrefix, hash[:]); err != nil {
		if err == vmstore.ErrStorageNotFound {
			cl := newContractAddressList(addr)
			if data, err := cl.ToABI(); err != nil {
				return err
			} else {
				if err := ctx.SetStorage(mappingPrefix, hash[:], data); err != nil {
					return err
				}
			}
		} else {
			return err
		}
	} else {
		cl := &ContractAddressList{}
		if err := cl.FromABI(storage); err != nil {
			return err
		} else {
			if b := cl.Append(addr); b {
				if data, err := cl.ToABI(); err != nil {
					return err
				} else {
					if err := ctx.SetStorage(mappingPrefix, hash[:], data); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func GetCDRMapping(ctx *vmstore.VMContext, hash *types.Hash) ([]*types.Address, error) {
	if storage, err := ctx.GetStorage(mappingPrefix, hash[:]); err != nil {
		return nil, err
	} else {
		cl := &ContractAddressList{}
		if err := cl.FromABI(storage); err != nil {
			return nil, err
		} else {
			return cl.AddressList, nil
		}
	}
}

// GetAllCDRStatus get all CDR records of the specific settlement contract
// @param addr settlement smart contract
func GetAllCDRStatus(ctx *vmstore.VMContext, addr *types.Address) ([]*CDRStatus, error) {
	return GetCDRStatusByDate(ctx, addr, 0, 0)
}

//FindSettlementContract query settlement contract by user address and CDR data
func FindSettlementContract(ctx *vmstore.VMContext, addr *types.Address, param *CDRParam) (*ContractParam, error) {
	if contracts, err := queryContractParamByAddress(ctx, "FindSettlementContract", func(cp *ContractParam) bool {
		if cp.PartyA.Address == *addr {
			return len(param.NextStop) > 0 && cp.IsNextStop(param.NextStop)
		} else if cp.PartyB.Address == *addr {
			return len(param.PreStop) > 0 && cp.IsPreStop(param.PreStop)
		}
		return false
	}); err != nil {
		return nil, err
	} else {
		switch size := len(contracts); {
		case size == 0:
			return nil, fmt.Errorf("can not find settlement contract related with %s", addr.String())
		case size > 1:
			return nil, fmt.Errorf("find mutilple(%d) settlement contract", len(contracts))
		default:
			return contracts[0], nil
		}
	}
}

func GetPreStopNames(ctx *vmstore.VMContext, addr *types.Address) ([]string, error) {
	var result []string
	if contracts, err := queryContractParamByAddress(ctx, "GetPreStopNames", func(cp *ContractParam) bool {
		return cp.PartyB.Address == *addr && (cp.Status == ContractStatusActivated || cp.Status == ContractStatusActiveStage1)
	}); err != nil {
		return nil, err
	} else {
		for _, c := range contracts {
			if len(c.PreStops) > 0 {
				result = append(result, c.PreStops...)
			}
		}
	}

	return result, nil
}

func GetNextStopNames(ctx *vmstore.VMContext, addr *types.Address) ([]string, error) {
	var result []string
	if contracts, err := queryContractParamByAddress(ctx, "GetPreStopNames", func(cp *ContractParam) bool {
		return cp.PartyA.Address == *addr && (cp.Status == ContractStatusActivated || cp.Status == ContractStatusActiveStage1)
	}); err != nil {
		return nil, err
	} else {
		for _, c := range contracts {
			if len(c.NextStops) > 0 {
				result = append(result, c.NextStops...)
			}
		}
	}

	return result, nil
}

// GetSettlementContract
// @param addr smart contract address
func GetSettlementContract(ctx *vmstore.VMContext, addr *types.Address) (*ContractParam, error) {
	logger := log.NewLogger("GetSettlementContract")
	defer func() {
		_ = logger.Sync()
	}()

	if storage, err := ctx.GetStorage(types.SettlementAddress[:], addr[:]); err != nil {
		return nil, err
	} else {
		cp := &ContractParam{}
		if err := cp.FromABI(storage); err != nil {
			return nil, err
		} else {
			return cp, nil
		}
	}
}

func GetAllSettlementContract(ctx *vmstore.VMContext) ([]*ContractParam, error) {
	return queryContractParamByAddress(ctx, "GetContractsByAddress", func(cp *ContractParam) bool {
		return true
	})
}

// GetContractsByAddress get all contract data by address both Party A and Party B
func GetContractsByAddress(ctx *vmstore.VMContext, addr *types.Address) ([]*ContractParam, error) {
	return queryContractParamByAddress(ctx, "GetContractsByAddress", func(cp *ContractParam) bool {
		return cp.PartyA.Address == *addr || cp.PartyB.Address == *addr
	})
}

// GetContractsIDByAddressAsPartyA get all contracts ID as Party A
func GetContractsIDByAddressAsPartyA(ctx *vmstore.VMContext, addr *types.Address) ([]*ContractParam, error) {
	return queryContractParamByAddress(ctx, "GetContractsIDByAddressAsPartyA", func(cp *ContractParam) bool {
		return cp.PartyA.Address == *addr
	})
}

// GetContractsIDByAddressAsPartyB get all contracts ID as Party B
func GetContractsIDByAddressAsPartyB(ctx *vmstore.VMContext, addr *types.Address) ([]*ContractParam, error) {
	return queryContractParamByAddress(ctx, "GetContractsIDByAddressAsPartyB", func(cp *ContractParam) bool {
		return cp.PartyB.Address == *addr
	})
}

type SummaryRecord struct {
	Total   uint64  `json:"total"`
	Success uint64  `json:"success"`
	Fail    uint64  `json:"fail"`
	Result  float64 `json:"result"`
}

func (z *SummaryRecord) DoCalculate() *SummaryRecord {
	z.Total = z.Fail + z.Success

	if z.Total > 0 {
		z.Result = float64(z.Success) / float64(z.Total)
	}

	return z
}

func (z *SummaryRecord) String() string {
	return util.ToIndentString(z)
}

type MatchingRecord struct {
	Orphan   SummaryRecord `json:"orphan"`
	Matching SummaryRecord `json:"matching"`
}

type CompareRecord struct {
	PartyA MatchingRecord `json:"partyA"`
	PartyB MatchingRecord `json:"partyB"`
}

func (z *CompareRecord) UpdateCounter(isPartyA, isMatching, state bool) {
	if isPartyA {
		if isMatching {
			if state {
				z.PartyA.Matching.Success++
			} else {
				z.PartyA.Matching.Fail++
			}
		} else {
			if state {
				z.PartyA.Orphan.Success++
			} else {
				z.PartyA.Orphan.Fail++
			}
		}
	} else {
		if isMatching {
			if state {
				z.PartyB.Matching.Success++
			} else {
				z.PartyB.Matching.Fail++
			}
		} else {
			if state {
				z.PartyB.Orphan.Success++
			} else {
				z.PartyB.Orphan.Fail++
			}
		}
	}
}

func (z *CompareRecord) DoCalculate() {
	z.PartyA.Matching.DoCalculate()
	z.PartyA.Orphan.DoCalculate()
	z.PartyB.Matching.DoCalculate()
	z.PartyB.Orphan.DoCalculate()
}

type SummaryResult struct {
	Contract *ContractParam            `json:"contract"`
	Records  map[string]*CompareRecord `json:"records"`
	Total    CompareRecord             `json:"total"`
}

func NewSummaryResult() *SummaryResult {
	return &SummaryResult{
		Records: make(map[string]*CompareRecord),
	}
}

func (z *SummaryResult) UpdateState(name string, isPartyA, isMatching, state bool) {
	if name != "" {
		if _, ok := z.Records[name]; !ok {
			z.Records[name] = &CompareRecord{}
		}
		z.Records[name].UpdateCounter(isPartyA, isMatching, state)
	}
	z.Total.UpdateCounter(isPartyA, isMatching, state)
}

func (z *SummaryResult) DoCalculate() {
	for k := range z.Records {
		z.Records[k].DoCalculate()
	}
	z.Total.DoCalculate()
}

func (z *SummaryResult) String() string {
	return util.ToIndentString(z)
}

// GetSummaryReport
// addr settlement contract address
func GetSummaryReport(ctx *vmstore.VMContext, addr *types.Address, start, end int64) (*SummaryResult, error) {
	records, err := GetCDRStatusByDate(ctx, addr, start, end)
	if err != nil {
		return nil, err
	}

	c, err := GetSettlementContract(ctx, addr)
	if err != nil {
		return nil, err
	}
	result := NewSummaryResult()
	result.Contract = c

	partyA := c.PartyA.Address
	partyB := c.PartyB.Address

	for _, status := range records {
		//party A
		if s1, b1, err := status.State(&partyA); err == nil {
			result.UpdateState(s1, true, status.Status == SettlementStatusSuccess, b1)
		}

		//party B
		if s2, b2, err := status.State(&partyB); err == nil {
			result.UpdateState(s2, false, status.Status == SettlementStatusSuccess, b2)
		}
	}

	result.DoCalculate()

	return result, nil
}

// GetCDRStatusByDate
func GetCDRStatusByDate(ctx *vmstore.VMContext, addr *types.Address, start, end int64) ([]*CDRStatus, error) {
	logger := log.NewLogger("GetCDRStatusByDate")
	defer func() {
		_ = logger.Sync()
	}()

	var result []*CDRStatus

	if err := ctx.Iterator(addr[:], func(key []byte, value []byte) error {
		if len(key) == keySize && len(value) > 0 {
			status := &CDRStatus{}
			if err := status.FromABI(value); err != nil {
				logger.Error(err)
			} else {
				if status.IsInCycle(start, end) {
					result = append(result, status)
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

type InvoiceRecord struct {
	Address                  types.Address `json:"contractAddress"`
	StartDate                int64         `json:"startDate"`
	EndDate                  int64         `json:"endDate"`
	Customer                 string        `json:"customer"`
	CustomerSr               string        `json:"customerSr"`
	Country                  string        `json:"country"`
	Operator                 string        `json:"operator"`
	ServiceId                string        `json:"serviceId"`
	MCC                      uint64        `json:"mcc"`
	MNC                      uint64        `json:"mnc"`
	Currency                 string        `json:"currency"`
	UnitPrice                float64       `json:"unitPrice"`
	SumOfBillableSMSCustomer uint64        `json:"sumOfBillableSMSCustomer"`
	SumOfTOTPrice            float64       `json:"sumOfTOTPrice"`
}

// GenerateInvoicesByContract
// addr settlement contract address
func GenerateInvoicesByContract(ctx *vmstore.VMContext, addr *types.Address, start, end int64) ([]*InvoiceRecord, error) {
	logger := log.NewLogger("GenerateInvoicesByContract")
	defer func() {
		_ = logger.Sync()
	}()

	c, err := GetSettlementContract(ctx, addr)
	if err != nil {
		return nil, err
	}
	var result []*InvoiceRecord

	contractAddr, err := c.Address()
	if err != nil {
		logger.Error(err)
	}

	cache := make(map[string]int)
	if cdrs, err := GetCDRStatusByDate(ctx, &contractAddr, start, end); err == nil {
		for _, cdr := range cdrs {
			if cdr.Status == SettlementStatusSuccess {
				if _, sender, _, err := cdr.ExtractID(); err == nil {
					if _, ok := cache[sender]; ok {
						cache[sender]++
					} else {
						cache[sender] = 1
					}
				}
			}
		}
	} else {
		logger.Error(err)
	}

	// TODO: how to match service???
	service := c.Services[0]
	for k, v := range cache {
		if v > 0 {
			invoice := &InvoiceRecord{
				Address:                  contractAddr,
				StartDate:                c.StartDate,
				EndDate:                  c.EndDate,
				Customer:                 k,
				CustomerSr:               "",
				Country:                  "",
				Operator:                 c.PartyB.Name,
				ServiceId:                service.ServiceId,
				MCC:                      service.Mcc,
				MNC:                      service.Mnc,
				Currency:                 service.Currency,
				UnitPrice:                service.UnitPrice,
				SumOfBillableSMSCustomer: uint64(v),
				SumOfTOTPrice:            service.UnitPrice * float64(v),
			}
			result = append(result, invoice)
		}
	}

	if len(result) > 0 {
		sort.Slice(result, func(i, j int) bool {
			r1 := result[i]
			r2 := result[j]
			if r1.StartDate < r2.EndDate {
				return true
			}
			if r1.StartDate > r2.EndDate {
				return false
			}
			return r1.EndDate < r2.EndDate
		})
	}

	return result, nil
}

// GenerateInvoices
// @param addr user qlcchain address
func GenerateInvoices(ctx *vmstore.VMContext, addr *types.Address, start, end int64) ([]*InvoiceRecord, error) {
	logger := log.NewLogger("GenerateInvoices")
	defer func() {
		_ = logger.Sync()
	}()

	contracts, err := GetContractsIDByAddressAsPartyA(ctx, addr)
	if err != nil {
		return nil, err
	}
	var result []*InvoiceRecord
	for _, c := range contracts {
		contractAddr, err := c.Address()
		if err != nil {
			logger.Error(err)
			continue
		}

		cache := make(map[string]int)
		if cdrs, err := GetCDRStatusByDate(ctx, &contractAddr, start, end); err == nil {
			if len(cdrs) == 0 {
				continue
			}

			for _, cdr := range cdrs {
				if cdr.Status == SettlementStatusSuccess {
					if _, sender, _, err := cdr.ExtractID(); err == nil {
						if _, ok := cache[sender]; ok {
							cache[sender]++
						} else {
							cache[sender] = 1
						}
					}
				}
			}
		} else {
			logger.Error(err)
		}

		// TODO: how to match service???
		service := c.Services[0]
		for k, v := range cache {
			if v > 0 {
				invoice := &InvoiceRecord{
					Address:                  contractAddr,
					StartDate:                c.StartDate,
					EndDate:                  c.EndDate,
					Customer:                 k,
					CustomerSr:               "",
					Country:                  "",
					Operator:                 c.PartyB.Name,
					ServiceId:                service.ServiceId,
					MCC:                      service.Mcc,
					MNC:                      service.Mnc,
					Currency:                 service.Currency,
					UnitPrice:                service.UnitPrice,
					SumOfBillableSMSCustomer: uint64(v),
					SumOfTOTPrice:            service.UnitPrice * float64(v),
				}
				result = append(result, invoice)
			}
		}
	}

	if len(result) > 0 {
		sort.Slice(result, func(i, j int) bool {
			r1 := result[i]
			r2 := result[j]
			if r1.StartDate < r2.EndDate {
				return true
			}
			if r1.StartDate > r2.EndDate {
				return false
			}
			return r1.EndDate < r2.EndDate
		})
	}

	return result, nil
}

// GenerateMultiPartyInvoice
// addr settlement contract address
func GenerateMultiPartyInvoice(ctx *vmstore.VMContext, addr *types.Address, start, end int64) ([]*InvoiceRecord, error) {
	logger := log.NewLogger("GenerateMultiPartyInvoice")
	defer func() {
		_ = logger.Sync()
	}()

	c, err := GetSettlementContract(ctx, addr)
	if err != nil {
		return nil, err
	}
	var result []*InvoiceRecord

	//if !IsContractAvailable(ctx, addr) {
	//	return nil, fmt.Errorf("contract %s is invalid", addr.String())
	//}

	cache := make(map[string]int)
	if cdrs, err := GetCDRStatusByDate(ctx, addr, start, end); err == nil {
		for _, cdr := range cdrs {
			if cdr.Status == SettlementStatusSuccess {
				// make sure that all contract status
				if hash, err := cdr.ToHash(); err != nil {
					logger.Error(err)
				} else {
					if addressList, err := GetCDRMapping(ctx, &hash); err != nil {
						logger.Error(err)
					} else {
						counter := 0
						size := len(addressList)
						for _, a := range addressList {
							logger.Infof("%s==>%s, %t", addr.String(), a.String())
							if status, err := GetCDRStatus(ctx, a, hash); err != nil {
								logger.Error(err)
							} else {
								if status.Status == SettlementStatusSuccess {
									counter++
								}
							}
						}
						if size > 1 && counter == size {
							if _, sender, _, err := cdr.ExtractID(); err == nil {
								if _, ok := cache[sender]; ok {
									cache[sender]++
								} else {
									cache[sender] = 1
								}
							}
						} else {
							logger.Infof("size: %d, counter: %d, flag: %t, %t", size, counter, size > 1, counter+1 == size)
						}
					}
				}
			}
		}
	} else {
		logger.Error(err)
	}

	// TODO: how to match service???
	service := c.Services[0]
	for k, v := range cache {
		if v > 0 {
			invoice := &InvoiceRecord{
				Address:                  *addr,
				StartDate:                c.StartDate,
				EndDate:                  c.EndDate,
				Customer:                 k,
				CustomerSr:               "",
				Country:                  "",
				Operator:                 c.PartyB.Name,
				ServiceId:                service.ServiceId,
				MCC:                      service.Mcc,
				MNC:                      service.Mnc,
				Currency:                 service.Currency,
				UnitPrice:                service.UnitPrice,
				SumOfBillableSMSCustomer: uint64(v),
				SumOfTOTPrice:            service.UnitPrice * float64(v),
			}
			result = append(result, invoice)
		}
	}

	if len(result) > 0 {
		sort.Slice(result, func(i, j int) bool {
			r1 := result[i]
			r2 := result[j]
			if r1.StartDate < r2.EndDate {
				return true
			}
			if r1.StartDate > r2.EndDate {
				return false
			}
			return r1.EndDate < r2.EndDate
		})
	}

	return result, nil
}

func queryContractParamByAddress(ctx *vmstore.VMContext, name string, fn func(cp *ContractParam) bool) ([]*ContractParam, error) {
	logger := log.NewLogger(name)
	defer func() {
		_ = logger.Sync()
	}()

	var result []*ContractParam

	if err := ctx.Iterator(types.SettlementAddress[:], func(key []byte, value []byte) error {
		if len(key) == keySize && len(value) > 0 {
			cp := &ContractParam{}
			if err := cp.FromABI(value); err != nil {
				logger.Error(err)
			} else {
				if fn(cp) {
					result = append(result, cp)
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(result) > 0 {
		sort.Slice(result, func(i, j int) bool {
			r1 := result[i]
			r2 := result[j]
			if r1.StartDate < r2.StartDate {
				return true
			}
			if r1.StartDate > r2.StartDate {
				return false
			}

			return r1.EndDate < r2.EndDate
		})
	}

	return result, nil
}

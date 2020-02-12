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
	"strings"

	"github.com/qlcchain/go-qlc/common"

	"gopkg.in/validator.v2"

	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/common/util"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
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
  }
]
`
	MethodNameCreateContract = "CreateContract"
	MethodNameSignContract   = "SignContract"
	MethodNameProcessCDR     = "ProcessCDR"
	MethodNameAddPreStop     = "AddPreStop"
	MethodNameRemovePreStop  = "RemovePreStop"
	MethodNameUpdatePreStop  = "UpdatePreStop"
	MethodNameAddNextStop    = "AddNextStop"
	MethodNameRemoveNextStop = "RemoveNextStop"
	MethodNameUpdateNextStop = "UpdateNextStop"
)

var (
	SettlementABI, _ = abi.JSONToABIContract(strings.NewReader(JsonSettlement))
	keySize          = types.AddressSize*2 + 1
)

type ABIer interface {
	ToABI() ([]byte, error)
	FromABI(data []byte) error
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

//func (z *SignContractParam) Sign(account *types.Account) error {
//	if z.ConfirmDate <= 0 {
//		return fmt.Errorf("invalid confirm date[%d]", z.ConfirmDate)
//	}
//	h, err := types.HashBytes(z.ContractAddress[:], util.BE_Int2Bytes(z.ConfirmDate))
//	if err != nil {
//		return err
//	}
//	z.SignatureB = account.Sign(h)
//	return nil
//}

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
	EndData   int64             `msg:"t4" json:"endData"`
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

	if z.EndData <= 0 {
		return false, fmt.Errorf("invalid end date %d", z.EndData)
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
Actived
DestroyStage1
Destroyed
)
*/
type ContractStatus int

//go:generate msgp
type ContractParam struct {
	CreateContractParam
	PreStops    []string       `msg:"pre" json:"preStops"`
	NextStops   []string       `msg:"nex" json:"nextStops"`
	ConfirmDate int64          `msg:"t2" json:"confirmDate"`
	Status      ContractStatus `msg:"s" json:"status"`
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

func (z *ContractParam) String() string {
	return util.ToIndentString(z)
}

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
Send
Error
Empty
)
*/
type SendingStatus int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
Delivered
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
		case SendingStatusSend:
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
	Params []SettlementCDR  `msg:"p" json:"params"`
	Status SettlementStatus `msg:"s" json:"status"`
}

func (z *CDRStatus) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *CDRStatus) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *CDRStatus) String() string {
	return util.ToIndentString(z)
}

// TODO:
// DoSettlement process settlement
// @param cdr  cdr data
func (z *CDRStatus) DoSettlement(cdr SettlementCDR) (err error) {
	z.Params = append(z.Params, cdr)

	switch size := len(z.Params); {
	//case size == 0:
	//	z.Status = SettlementStatusUnknown
	//	break
	case size == 1:
		z.Status = SettlementStatusStage1
		break
	case size >= 2:
		z.Status = SettlementStatusSuccess
		b := true
		// combine all status
		for _, param := range z.Params {
			b = b && param.Status()
		}
		if !b {
			z.Status = SettlementStatusFailure
		}
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
			return param.Status == ContractStatusActived && unix >= param.StartDate && unix <= param.EndData
		}
	}
	return false
}

// GetCDRStatus
// @param addr settlement contract address
// @param CDR data hash
func GetCDRStatus(ctx *vmstore.VMContext, addr *types.Address, hash types.Hash) (*CDRStatus, error) {
	logger := log.NewLogger("GetContracts")
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

// GetAllCDRStatus get all CDR records of the specific settlement contract
// @param addr settlement smart contract
func GetAllCDRStatus(ctx *vmstore.VMContext, addr *types.Address) ([]*CDRStatus, error) {
	logger := log.NewLogger("GetAllCDRStatus")
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
				result = append(result, status)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

//GetSettlementContract query settlement contract by user address and CDR data
func GetSettlementContract(ctx *vmstore.VMContext, addr *types.Address, param *CDRParam) (*ContractParam, error) {
	if contracts, err := queryContractParamByAddress(ctx, "GetSettlementContract", func(cp *ContractParam) bool {
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

// GetContracts
// @param addr smart contract address
func GetContracts(ctx *vmstore.VMContext, addr *types.Address) (*ContractParam, error) {
	logger := log.NewLogger("GetContracts")
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

	return result, nil
}

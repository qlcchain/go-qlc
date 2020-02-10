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
      { "name": "smsDt", "type": "int64" },
      { "name": "messageID", "type": "string" },
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
  }
]
`
	MethodNameCreateContract = "CreateContract"
	MethodNameSignContract   = "SignContract"
	MethodNameProcessCDR     = "ProcessCDR"
)

var (
	SettlementABI, _ = abi.JSONToABIContract(strings.NewReader(JsonSettlement))
	keySize          = types.AddressSize*2 + 1
)

type SignContractParam struct {
	ContractAddress types.Address   `msg:"a,extension" json:"contractAddress"`
	ConfirmDate     int64           `msg:"cd" json:"confirmDate"`
	SignatureB      types.Signature `msg:"sb,extension" json:"signatureB"`
}

func (z *SignContractParam) Verify(addr types.Address) (bool, error) {
	if z.ContractAddress.IsZero() {
		return false, fmt.Errorf("invalid contract address %s", z.ContractAddress.String())
	}
	if z.ConfirmDate == 0 {
		return false, errors.New("invalid contract confirm date")
	}
	if h, err := types.HashBytes(z.ContractAddress[:], util.BE_Int2Bytes(z.ConfirmDate)); err != nil {
		return false, err
	} else {
		return addr.Verify(h[:], z.SignatureB[:]), nil
	}
}

func (z *SignContractParam) ToABI() ([]byte, error) {
	//return SettlementABI.PackMethod(MethodNameSignContract, z.ContractAddress, z.ConfirmDate, z.SignatureB)
	return z.MarshalMsg(nil)
}

func (z *SignContractParam) FromABI(data []byte) error {
	//return SettlementABI.UnpackMethod(z, MethodNameSignContract, data)
	_, err := z.UnmarshalMsg(data)
	return err
}

//go:generate msgp
type CreateContractParam struct {
	PartyA      types.Address   `msg:"pa,extension" json:"partyA"`
	PartyAName  string          `msg:"an" json:"partyAName"`
	PartyB      types.Address   `msg:"pb,extension" json:"partyB"`
	PartyBName  string          `msg:"bn" json:"partyBName"`
	Previous    types.Hash      `msg:"pre,extension" json:"previous"`
	ServiceId   string          `msg:"id" json:"serviceId"`
	Mcc         uint64          `msg:"mcc" json:"mcc"`
	Mnc         uint64          `msg:"mnc" json:"mnc"`
	TotalAmount uint64          `msg:"t" json:"totalAmount"`
	UnitPrice   uint64          `msg:"u" json:"unitPrice"`
	Currency    string          `msg:"c" json:"currency"`
	SignDate    int64           `msg:"t1" json:"signDate"`
	SignatureA  types.Signature `msg:"sa,extension" json:"signatureA"`
}

func (z *CreateContractParam) Verify() (bool, error) {
	if _, err := z.verifyParam(); err != nil {
		return false, err
	}

	if !z.SignatureA.IsZero() {
		a, _ := z.Address()
		if verify := z.PartyA.Verify(a[:], z.SignatureA[:]); !verify {
			return false, fmt.Errorf("invalid signature %s of %s", z.SignatureA, z.PartyA.String())
		}
	} else {
		return false, fmt.Errorf("invalid signature")
	}

	return true, nil
}

func (z *CreateContractParam) verifyParam() (bool, error) {
	if z.PartyA.IsZero() || len(z.PartyAName) == 0 {
		return false, fmt.Errorf("invalid partyA params")
	}

	if z.PartyB.IsZero() || len(z.PartyBName) == 0 {
		return false, fmt.Errorf("invalid partyB params")
	}

	if z.Previous.IsZero() {
		return false, errors.New("invalid preivous hash")
	}

	if len(z.ServiceId) == 0 {
		return false, errors.New("invalid service ID")
	}

	if z.TotalAmount == 0 {
		return false, errors.New("invalid total amount")
	}

	if z.UnitPrice == 0 {
		return false, errors.New("invalid unit price")
	}

	if len(z.Currency) == 0 {
		return false, errors.New("invalid currency")
	}

	if z.SignDate <= 0 {
		return false, fmt.Errorf("invalid sign date %d", z.SignDate)
	}

	return true, nil
}

func (z *CreateContractParam) Balance() (types.Balance, error) {
	if mul, b := util.SafeMul(z.TotalAmount, z.UnitPrice); b {
		return types.ZeroBalance, fmt.Errorf("overflow when mul %d and %d", z.TotalAmount, z.UnitPrice)
	} else {
		return types.Balance{Int: new(big.Int).SetUint64(mul)}, nil
	}
}

func (z *CreateContractParam) ToContractParam() *ContractParam {
	return &ContractParam{
		CreateContractParam: *z,
		ConfirmDate:         0,
		SignatureB:          nil,
	}
}

func (z *CreateContractParam) ToABI() ([]byte, error) {
	//return SettlementABI.PackMethod(MethodNameCreateContract, z.PartyA, z.PartyAName, z.PartyB, z.PartyBName, z.Previous,
	//	z.ServiceId, z.Mcc, z.Mnc, z.TotalAmount, z.UnitPrice, z.Currency, z.SignDate, z.SignatureA)
	return z.MarshalMsg(nil)
}

func (z *CreateContractParam) FromABI(data []byte) error {
	//return SettlementABI.UnpackMethod(z, MethodNameCreateContract, data)
	_, err := z.UnmarshalMsg(data)
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
	if _, err := z.verifyParam(); err != nil {
		return types.ZeroAddress, err
	}

	hash, err := types.HashBytes(z.PartyA[:], []byte(z.PartyAName), z.PartyB[:], []byte(z.PartyBName),
		z.Previous[:], []byte(z.ServiceId), util.BE_Uint64ToBytes(z.Mcc),
		util.BE_Uint64ToBytes(z.Mnc), util.BE_Uint64ToBytes(z.TotalAmount), util.BE_Uint64ToBytes(z.UnitPrice),
		[]byte(z.Currency), util.BE_Int2Bytes(z.SignDate))
	if err != nil {
		return types.ZeroAddress, err
	}
	return types.BytesToAddress(hash[:])
}

func (z *CreateContractParam) Sign(account *types.Account) error {
	address, err := z.Address()
	if err != nil {
		return err
	}
	h, err := types.BytesToHash(address[:])
	if err != nil {
		return err
	}
	z.SignatureA = account.Sign(h)
	return nil
}

//go:generate msgp
type ContractParam struct {
	CreateContractParam
	ConfirmDate int64            `msg:"t2" json:"confirmDate"`
	SignatureB  *types.Signature `msg:"sb,extension" json:"signatureB,omitempty"`
}

func (z *ContractParam) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *ContractParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *ContractParam) Sign(account *types.Account) error {
	if z.ConfirmDate <= 0 {
		return fmt.Errorf("invalid confirm date[%d]", z.ConfirmDate)
	}
	address, err := z.Address()
	if err != nil {
		return err
	}
	h, err := types.HashBytes(address[:], util.BE_Int2Bytes(z.ConfirmDate))
	if err != nil {
		return err
	}
	s := account.Sign(h)
	z.SignatureB = &s
	return nil
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

	if z.SignatureA.IsZero() || cp.SignatureA.IsZero() {
		return false, errors.New("empty signature")
	}

	if a1 == a2 {
		if b := z.PartyA.Verify(a1[:], z.SignatureA[:]); b {
			return true, nil
		} else {
			return false, fmt.Errorf("invalid signature: %s", z.SignatureA)
		}
	} else {
		return false, fmt.Errorf("invalid address, exp: %s,act: %s", a1.String(), a2.String())
	}
}

func (z *ContractParam) String() string {
	return util.ToIndentString(z)
}

// TODO:
// we should make sure that can use CDR data to match to a specific settlement contract
type CDRParam struct {
	Index        uint64
	SmsDt        string
	MessageID    string
	Sender       string
	Destination  string
	DstCountry   string
	DstOperator  string
	DstMcc       string
	DstMnc       string
	SellPrice    float64
	SellCurrency string
	//connection         string
	CustomerName  string
	CustomerID    string
	SendingStatus string
	DlrStatus     string
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

// IsContractAvailable check contract status by contract ID
func IsContractAvailable(ctx *vmstore.VMContext, addr *types.Address) bool {
	if value, err := ctx.GetStorage(types.SettlementAddress[:], addr[:]); err == nil {
		if param, err := ParseContractParam(value); err == nil {
			// TODO: verify end date??
			if param.SignatureB != nil && !param.SignatureB.IsZero() && !param.SignatureA.IsZero() {
				return true
			}
		}
	}
	return false
}

// GetContractsByAddress get all contract data by address both Party A and Party B
func GetContractsByAddress(ctx *vmstore.VMContext, addr *types.Address) ([]*ContractParam, error) {
	return queryContractParamByAddress(ctx, "GetContractsByAddress", func(cp *ContractParam) bool {
		return cp.PartyA == *addr || cp.PartyB == *addr
	})
}

// GetContractsIDByAddressAsPartyA get all contracts ID as Party A
func GetContractsIDByAddressAsPartyA(ctx *vmstore.VMContext, addr *types.Address) ([]*ContractParam, error) {
	return queryContractParamByAddress(ctx, "GetContractsIDByAddressAsPartyA", func(cp *ContractParam) bool {
		return cp.PartyA == *addr
	})
}

// GetContractsIDByAddressAsPartyB get all contracts ID as Party B
func GetContractsIDByAddressAsPartyB(ctx *vmstore.VMContext, addr *types.Address) ([]*ContractParam, error) {
	return queryContractParamByAddress(ctx, "GetContractsIDByAddressAsPartyB", func(cp *ContractParam) bool {
		return cp.PartyB == *addr
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

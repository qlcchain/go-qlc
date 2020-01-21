/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"strings"

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
      { "name": "mcc", "type": "int64" },
      { "name": "mnc", "type": "int64" },
      { "name": "amount", "type": "uint256" },
      { "name": "unit"", "type": "uint256" },
      { "name": "currency", "type": "string" },
      { "name": "signDate", "type": "int64" },
      { "name": "signA", "type": "signature" }
    ]
  },
  {
    "type": "function",
    "name": "SignContract",
    "inputs": [
      { "name": "contractID", "type": "hash" },
      { "name": "confirmDate", "type": "int64" },
      { "name": "signB", "type": "signature" }
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
)

//go:generate msgp
type Contractor struct {
	Address types.Address `msg:"a,extension" json:"address"`
	Name    string        `msg:"n" json:"name"`
}

//go:generate msgp
type CreateContractParam struct {
	PartyA      Contractor       `msg:"pa,extension" json:"partyA"`
	PartyB      Contractor       `msg:"pb,extension" json:"partyB"`
	PreHash     types.Hash       `msg:"pre,extension" json:"previous"`
	ServiceId   string           `msg:"s" json:"serviceId"`
	MCC         int              `msg:"mcc,extension" json:"mcc"`
	MNC         int              `msg:"mnc,extension" json:"mnc"`
	TotalAmount uint64           `msg:"t" json:"totalAmount"`
	UnitPrice   uint64           `msg:"u" json:"unitPrice"`
	Currency    string           `msg:"c" json:"currency"`
	SignDate    int64            `msg:"t1" json:"signDate"`
	SignatureA  *types.Signature `msg:"sa,extension" json:"signatureA"`
}

//go:generate msgp
type ContractParam struct {
	CreateContractParam
	ConfirmDate int64            `msg:"t2" json:"confirmDate"`
	SignatureB  *types.Signature `msg:"sb,extension" json:"signatureB,omitempty"`
}

// TODO:
// we should make sure that can use CDR data to match to a specific settlement contract
type CDRParam struct {
	smsDt        string
	messageID    string
	sender       string
	destination  string
	dstCountry   string
	dstOperator  string
	dstMcc       string
	dstMnc       string
	sellPrice    float64
	sellCurrency string
	//connection         string
	customerName  string
	customerID    string
	sendingStatus string
	dlrStatus     string
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
func IsContractAvailable(hash *types.Hash) bool {
	return false
}

// GetContractsByAddress get all contract data by address both Party A and Party B
func GetContractsByAddress(addr *types.Address) []*ContractParam {
	return nil
}

// GetContractsIDByAddressAsPartyA get all contracts ID as Party A
func GetContractsIDByAddressAsPartyA(addr *types.Address) []*types.Hash {
	return nil
}

// GetContractsIDByAddressAsPartyB get all contracts ID as Party B
func GetContractsIDByAddressAsPartyB(addr *types.Address) []*types.Hash {
	return nil
}

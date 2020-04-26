/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"
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
  },{
    "type": "function",
    "name": "RegisterAsset",
    "inputs": [
        { "name": "asset", "type": "string" }
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
	MethodNameRegisterAsset     = "RegisterAsset"
)

const (
	ContractFlag byte = iota
	AssetFlag
	AddressMappingFlag
)

var (
	SettlementABI, _ = abi.JSONToABIContract(strings.NewReader(JsonSettlement))
	keySize          = types.AddressSize * 2
	assetKeySize     = keySize + 1
	contractKeySize  = keySize + 1
	contractPrefix   = append(contractaddress.SettlementAddress[:], ContractFlag)
	assetKeyPrefix   = append(contractaddress.SettlementAddress[:], AssetFlag)
	mappingPrefix    = append(contractaddress.SettlementAddress[:], AddressMappingFlag)
)

func SaveContractParam(ctx *vmstore.VMContext, addr *types.Address, bts []byte) error {
	if err := ctx.SetStorage(contractPrefix, addr[:], bts); err != nil {
		return err
	}
	return nil
}

func GetContractParam(ctx *vmstore.VMContext, addr *types.Address) ([]byte, error) {
	if storage, err := ctx.GetStorage(contractPrefix, addr[:]); err != nil {
		return nil, err
	} else {
		return storage, nil
	}
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
func GetAllCDRStatus(store ledger.Store, addr *types.Address) ([]*CDRStatus, error) {
	return GetCDRStatusByDate(store, addr, 0, 0)
}

// GetMultiPartyCDRStatus get all CDR records belong to firstAddr and secondAddr
func GetMultiPartyCDRStatus(store ledger.Store, firstAddr, secondAddr *types.Address) (map[types.Address][]*CDRStatus, error) {
	if data1, err := GetCDRStatusByDate(store, firstAddr, 0, 0); err != nil {
		return nil, err
	} else {
		if data2, err := GetCDRStatusByDate(store, secondAddr, 0, 0); err != nil {
			return nil, err
		} else {
			return map[types.Address][]*CDRStatus{*firstAddr: data1, *secondAddr: data2}, nil
		}
	}
}

//FindSettlementContract query settlement contract by user address and CDR data
func FindSettlementContract(store ledger.Store, addr *types.Address, param *CDRParam) (*ContractParam, error) {
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
	if contracts, err := queryContractParam(ctx, "FindSettlementContract", func(cp *ContractParam) bool {
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
			return nil, fmt.Errorf("can not find settlement contract related with %s by %s", addr.String(), param.String())
		case size > 1:
			return nil, fmt.Errorf("find mutilple(%d) settlement contract", len(contracts))
		default:
			return contracts[0], nil
		}
	}
}

func GetPreStopNames(ctx *vmstore.VMContext, addr *types.Address) ([]string, error) {
	var result []string
	if contracts, err := queryContractParam(ctx, "GetPreStopNames", func(cp *ContractParam) bool {
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
	if contracts, err := queryContractParam(ctx, "GetPreStopNames", func(cp *ContractParam) bool {
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

	if storage, err := ctx.GetStorage(contractPrefix, addr[:]); err != nil {
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

func GetAllSettlementContract(store ledger.Store) ([]*ContractParam, error) {
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
	return queryContractParam(ctx, "GetContractsByAddress", func(cp *ContractParam) bool {
		return true
	})
}

// GetContractsByAddress get all contract data by address both Party A and Party B
func GetContractsByAddress(store ledger.Store, addr *types.Address) ([]*ContractParam, error) {
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
	return queryContractParam(ctx, "GetContractsByAddress", func(cp *ContractParam) bool {
		return cp.IsContractor(*addr)
	})
}

// GetContractsByStatus get all contract data by address both Party A and Party B
func GetContractsByStatus(store ledger.Store, addr *types.Address, status ContractStatus) ([]*ContractParam, error) {
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
	return queryContractParam(ctx, "GetContractsByAddress", func(cp *ContractParam) bool {
		return cp.IsContractor(*addr) && status == cp.Status
	})
}

// GetContractsByStatus get all expired contract data by address both Party A and Party B
func GetExpiredContracts(store ledger.Store, addr *types.Address) ([]*ContractParam, error) {
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
	return queryContractParam(ctx, "GetContractsByAddress", func(cp *ContractParam) bool {
		return cp.IsContractor(*addr) && cp.IsExpired()
	})
}

// GetContractsIDByAddressAsPartyA get all contracts ID as Party A
func GetContractsIDByAddressAsPartyA(store ledger.Store, addr *types.Address) ([]*ContractParam, error) {
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
	return queryContractParam(ctx, "GetContractsIDByAddressAsPartyA", func(cp *ContractParam) bool {
		return cp.PartyA.Address == *addr
	})
}

// GetContractsIDByAddressAsPartyB get all contracts ID as Party B
func GetContractsIDByAddressAsPartyB(store ledger.Store, addr *types.Address) ([]*ContractParam, error) {
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
	return queryContractParam(ctx, "GetContractsIDByAddressAsPartyB", func(cp *ContractParam) bool {
		return cp.PartyB.Address == *addr
	})
}

func GetContractsAddressByPartyANextStop(store ledger.Store, addr *types.Address, nextStop string) (*types.Address, error) {
	return queryContractByAddressAndStopName(store, "GetContractsAddressByPartyANextStop", func(cp *ContractParam) bool {
		var b bool
		if cp.PartyA.Address == *addr {
			for _, n := range cp.NextStops {
				if n == nextStop {
					b = true
					break
				}
			}
		}
		return b
	})
}

func GetContractsAddressByPartyBPreStop(store ledger.Store, addr *types.Address, preStop string) (*types.Address, error) {
	return queryContractByAddressAndStopName(store, "GetContractsAddressByPartyBPreStop", func(cp *ContractParam) bool {
		var b bool
		if cp.PartyB.Address == *addr {
			for _, n := range cp.PreStops {
				if n == preStop {
					b = true
					break
				}
			}
		}
		return b
	})
}

// GetSummaryReport
// addr settlement contract address
func GetSummaryReport(store ledger.Store, addr *types.Address, start, end int64) (*SummaryResult, error) {
	return generateSummaryReport(store, addr, customerFn, start, end)
}

func GetSummaryReportByAccount(store ledger.Store, addr *types.Address, account string, start, end int64) (*SummaryResult, error) {
	if account == "" {
		return nil, errors.New("empty account")
	}
	return generateSummaryReport(store, addr, func(status *CDRStatus) (s string, err error) {
		_, a, _, err := status.ExtractAccount()
		if err != nil {
			return "", err
		}
		if a == account {
			return a, nil
		}

		return "", fmt.Errorf("invalid account, exp:%s, got: %s", account, a)
	}, start, end)
}

func GetSummaryReportByCustomer(store ledger.Store, addr *types.Address, customer string, start, end int64) (*SummaryResult, error) {
	if customer == "" {
		return nil, errors.New("empty customer")
	}
	return generateSummaryReport(store, addr, func(status *CDRStatus) (s string, err error) {
		_, c, _, err := status.ExtractCustomer()
		if err != nil {
			return "", err
		}
		if c == customer {
			return c, nil
		}

		return "", fmt.Errorf("invalid customer, exp:%s, got: %s", customer, c)
	}, start, end)
}

func generateSummaryReport(store ledger.Store, addr *types.Address, fn func(status *CDRStatus) (string, error), start, end int64) (*SummaryResult, error) {
	records, err := GetCDRStatusByDate(store, addr, start, end)
	if err != nil {
		return nil, err
	}
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
	c, err := GetSettlementContract(ctx, addr)
	if err != nil {
		return nil, err
	}
	result := newSummaryResult()
	result.Contract = c

	partyA := c.PartyA.Address
	partyB := c.PartyB.Address

	addrs := []*types.Address{&partyA, &partyB}
	for _, status := range records {
		isMatching := status.IsMatching(addrs)
		//party A
		if s1, b1, err := status.State(&partyA, fn); err == nil {
			result.UpdateState(s1, "partyA", isMatching, b1)
		}

		//party B
		if s2, b2, err := status.State(&partyB, fn); err == nil {
			result.UpdateState(s2, "partyB", isMatching, b2)
		}
	}

	result.DoCalculate()

	return result, nil
}

// GetCDRStatusByDate
func GetCDRStatusByDate(store ledger.Store, addr *types.Address, start, end int64) ([]*CDRStatus, error) {
	logger := log.NewLogger("GetCDRStatusByDate")
	defer func() {
		_ = logger.Sync()
	}()

	var result []*CDRStatus
	iterator := store.NewVMIterator(&contractaddress.SettlementAddress)
	if err := iterator.Next(addr[:], func(key []byte, value []byte) error {
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

func GenerateInvoicesByAccount(store ledger.Store, addr *types.Address, account string, start, end int64) ([]*InvoiceRecord, error) {
	if account == "" {
		return nil, errors.New("empty account")
	}
	return generateInvoices(store, addr, func(status *CDRStatus) (s string, err error) {
		if _, a, _, err := status.ExtractAccount(); err == nil {
			if account == a {
				return a, nil
			} else {
				return "", fmt.Errorf("invalid account,exp: %s, got: %s", account, a)
			}
		} else {
			return "", err
		}
	}, start, end)
}

func GenerateInvoicesByCustomer(store ledger.Store, addr *types.Address, customer string, start, end int64) ([]*InvoiceRecord, error) {
	if customer == "" {
		return nil, errors.New("empty customer")
	}
	return generateInvoices(store, addr, func(status *CDRStatus) (s string, err error) {
		if _, c, _, err := status.ExtractCustomer(); err == nil {
			if customer == c {
				return c, nil
			} else {
				return "", fmt.Errorf("invalid customer,exp: %s, got: %s", customer, c)
			}
		} else {
			return "", err
		}
	}, start, end)
}

func generateInvoices(store ledger.Store, addr *types.Address, fn func(*CDRStatus) (string, error), start, end int64) ([]*InvoiceRecord, error) {
	logger := log.NewLogger("generateInvoices")
	defer func() {
		_ = logger.Sync()
	}()
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
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
	if cdrs, err := GetCDRStatusByDate(store, &contractAddr, start, end); err == nil {
		for _, cdr := range cdrs {
			if cdr.Status == SettlementStatusSuccess {
				if sender, err := fn(cdr); err == nil {
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
			return sortInvoiceFun(result[i], result[j])
		})
	}

	return result, nil
}

// GenerateInvoicesByContract
// addr settlement contract address
func GenerateInvoicesByContract(store ledger.Store, addr *types.Address, start, end int64) ([]*InvoiceRecord, error) {
	return generateInvoices(store, addr, func(status *CDRStatus) (s string, err error) {
		if _, sender, _, err := status.ExtractID(); err == nil {
			return sender, nil
		}
		return "", err
	}, start, end)
}

// GenerateInvoices
// @param addr user qlcchain address
func GenerateInvoices(store ledger.Store, addr *types.Address, start, end int64) ([]*InvoiceRecord, error) {
	logger := log.NewLogger("GenerateInvoices")
	defer func() {
		_ = logger.Sync()
	}()

	contracts, err := GetContractsIDByAddressAsPartyA(store, addr)
	if err != nil {
		return nil, err
	}
	var result []*InvoiceRecord
	for _, c := range contracts {
		if contractAddr, err := c.Address(); err == nil {
			if records, err := GenerateInvoicesByContract(store, &contractAddr, start, end); err == nil {
				for _, record := range records {
					result = append(result, record)
				}
			} else {
				logger.Error(err)
			}
		}
	}

	if len(result) > 1 {
		sort.Slice(result, func(i, j int) bool {
			return sortInvoiceFun(result[i], result[j])
		})
	}

	return result, nil
}

// GetMultiPartySummaryReport
func GetMultiPartySummaryReport(store ledger.Store, firstAddr, secondAddr *types.Address, start, end int64) (*MultiPartySummaryResult, error) {
	logger := log.NewLogger("GetMultiPartySummaryReport")
	defer func() {
		_ = logger.Sync()
	}()

	records, err := GetCDRStatusByDate(store, firstAddr, start, end)
	if err != nil {
		return nil, err
	}

	c1, c2, _, err := verifyMultiPartyAddress(store, firstAddr, secondAddr)
	if err != nil {
		return nil, err
	}

	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)

	result := newMultiPartySummaryResult()
	result.Contracts = []*ContractParam{c1, c2}

	partyA := c1.PartyA.Address
	partyB := c1.PartyB.Address
	partyC := c2.PartyB.Address

	addrs := []*types.Address{&partyA, &partyB, &partyC}
	for _, record := range records {
		hash, err := record.ToHash()
		if err != nil {
			logger.Error(err)
			continue
		}

		if stat2, err := GetCDRStatus(ctx, secondAddr, hash); err == nil {
			record.Merge(stat2)
		} else {
			logger.Errorf("%s[%s], err %s", secondAddr.String(), hash.String(), err)
		}

		isMatching := record.IsMatching(addrs)
		// party A
		if s1, b1, err := record.State(&partyA, customerFn); err == nil {
			result.UpdateState(s1, "partyA", isMatching, b1)
		}

		// party B
		if s2, b2, err := record.State(&partyB, customerFn); err == nil {
			result.UpdateState(s2, "partyB", isMatching, b2)
		}

		// party C
		if s3, b3, err := record.State(&partyC, customerFn); err == nil {
			result.UpdateState(s3, "partyC", isMatching, b3)
		}
	}

	result.DoCalculate()

	return result, nil
}

// GenerateMultiPartyInvoice
// addr settlement contract address
func GenerateMultiPartyInvoice(store ledger.Store, firstAddr, secondAddr *types.Address, start, end int64) ([]*InvoiceRecord, error) {
	if _, _, _, err := verifyMultiPartyAddress(store, firstAddr, secondAddr); err != nil {
		return nil, err
	}

	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)

	return generateInvoices(store, firstAddr, func(status *CDRStatus) (s string, err error) {
		hash, err := status.ToHash()
		if err != nil {
			return "", fmt.Errorf("tohash: %s, %s", status.String(), err)
		}

		if stat2, err := GetCDRStatus(ctx, secondAddr, hash); err == nil {
			if stat2.Status == SettlementStatusSuccess {
				_, sender, _, err := stat2.ExtractID()
				if err != nil {
					return "", fmt.Errorf("extractid: %s, %s", stat2.String(), err)
				}
				return sender, nil
			} else {
				return "", fmt.Errorf("invalid status, (%s)%s:%s", secondAddr.String(), hash.String(), status.Status.String())
			}
		} else {
			return "", err
		}
	}, start, end)
}

func GetAssetParam(ctx *vmstore.VMContext, h types.Address) (*AssetParam, error) {
	if storage, err := ctx.GetStorage(assetKeyPrefix, h[:]); err != nil {
		return nil, err
	} else {
		if param, err := ParseAssertParam(storage); err != nil {
			return nil, err
		} else {
			return param, nil
		}
	}
}

func SaveAssetParam(ctx *vmstore.VMContext, bts []byte) error {
	param, err := ParseAssertParam(bts)
	if err != nil {
		return err
	}

	if err := param.Verify(); err != nil {
		return err
	}

	h, err := param.ToAddress()
	if err != nil {
		return err
	}

	if storage, err := ctx.GetStorage(assetKeyPrefix, h[:]); err != nil {
		if err != vmstore.ErrStorageNotFound {
			return err
		} else {
			if err = ctx.SetStorage(assetKeyPrefix, h[:], bts); err != nil {
				return err
			}
		}
	} else if len(storage) > 0 {
		if !bytes.Equal(bts, storage) {
			return fmt.Errorf("invalid saved data of %s", h.String())
		}
	}

	return nil
}

func GetAllAsserts(store ledger.Store) ([]*AssetParam, error) {
	return queryAsserts(store, "GetAllAssert", func(param *AssetParam) bool {
		return true
	})
}

func GetAssertsByAddress(store ledger.Store, owner *types.Address) ([]*AssetParam, error) {
	return queryAsserts(store, "GetAssertsByAddress", func(param *AssetParam) bool {
		return param.Owner.Address == *owner
	})
}

func verifyMultiPartyAddress(store ledger.Store, firstAddr, secondAddr *types.Address) (*ContractParam, *ContractParam, bool, error) {
	ctx := vmstore.NewVMContext(store, &contractaddress.SettlementAddress)
	c1, err := GetSettlementContract(ctx, firstAddr)
	if err != nil {
		return nil, nil, false, err
	}

	c2, err := GetSettlementContract(ctx, secondAddr)
	if err != nil {
		return nil, nil, false, err
	}

	if c1.PartyB.Address != c2.PartyA.Address {
		return c1, c2, false, fmt.Errorf("%s's PartyB should be %s's PartyA,exp: %s, but go %s", firstAddr.String(),
			secondAddr.String(), c1.PartyB.String(), c2.PartyA.String())
	}
	return c1, c2, true, nil
}

func sortContractParam(r1, r2 *ContractParam) bool {
	if r1.StartDate < r2.StartDate {
		return true
	}
	if r1.StartDate > r2.StartDate {
		return false
	}

	return r1.EndDate < r2.EndDate
}

func queryContractParam(ctx *vmstore.VMContext, name string, fn func(cp *ContractParam) bool) ([]*ContractParam, error) {
	logger := log.NewLogger(name)
	defer func() {
		_ = logger.Sync()
	}()

	var result []*ContractParam
	if err := ctx.Iterator(contractPrefix, func(key []byte, value []byte) error {
		if len(key) == contractKeySize && len(value) > 0 {
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

	if len(result) > 1 {
		sort.Slice(result, func(i, j int) bool {
			return sortContractParam(result[i], result[j])
		})
	}

	return result, nil
}

func queryContractByAddressAndStopName(store ledger.Store, name string, fn func(cp *ContractParam) bool) (*types.Address, error) {
	logger := log.NewLogger(name)
	defer func() {
		_ = logger.Sync()
	}()

	var result *ContractParam
	iterator := store.NewVMIterator(&contractaddress.SettlementAddress)
	if err := iterator.Next(contractPrefix, func(key []byte, value []byte) error {
		if len(key) == contractKeySize && len(value) > 0 {
			cp := &ContractParam{}
			if err := cp.FromABI(value); err != nil {
				logger.Error(err)
			} else {
				if fn(cp) {
					result = cp
					return nil
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if result != nil {
		addr, err := result.Address()
		if err != nil {
			return nil, err
		}
		return &addr, nil
	}
	return nil, fmt.Errorf("can not find any contract")
}

func queryAsserts(store ledger.Store, name string, fn func(param *AssetParam) bool) ([]*AssetParam, error) {
	logger := log.NewLogger(name)
	defer func() {
		_ = logger.Sync()
	}()

	var result []*AssetParam
	iterator := store.NewVMIterator(&contractaddress.SettlementAddress)
	if err := iterator.Next(assetKeyPrefix, func(key []byte, value []byte) error {
		if len(key) == assetKeySize && len(value) > 0 {
			if param, err := ParseAssertParam(value); err != nil {
				logger.Error(err)
			} else if fn(param) {
				result = append(result, param)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(result) > 1 {
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

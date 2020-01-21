/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

type SettlementAPI struct {
}

func NewSettlement() *SettlementAPI {
	return &SettlementAPI{}
}

// GetCreateContractBlock
// generate ContractSend block to call smart contract for generating settlement contract as PartyA
// @param addr common settlement smart contract address
// @param param smart contract params
// @return state block to be processed
func (s *SettlementAPI) GetCreateContractBlock(addr *types.Address, param *abi.CreateContractParam) (*types.StateBlock, error) {
	return nil, nil
}

// GetSignContractBlock
// generate ContractSend block to call smart contract for signing settlement contract as PartyB
// smart contract address = hash(*abi.CreateContractParam)
// param.SignatureB = sign(hash(*abi.CreateContractParam) + param.ConfirmDate)
// @param param sign settlement contract address created by PartyA
// @return state block(without signature) to be processed
func (s *SettlementAPI) GetSignContractBlock(param *abi.ContractParam) (*types.StateBlock, error) {
	return nil, nil
}

// SettlementContract settlement contract for RPC
type SettlementContract struct {
	abi.ContractParam               //contract params
	Address           types.Address // settlement smart contract address
}

// GetContractsByAddress query all related settlement contracts info by address
// @param addr user qlcchain address
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
// @return all settlement contract
func (s *SettlementAPI) GetContractsByAddress(addr *types.Address, count int, offset *int) ([]*SettlementContract, error) {
	return nil, nil
}

// GetContractsAsPartyA query all settlement contracts as Party A info by address
// @param addr user qlcchain address
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
// @return all settlement contract as PartyA
func (s *SettlementAPI) GetContractsAsPartyA(addr *types.Address, count int, offset *int) ([]*SettlementContract, error) {
	return nil, nil
}

// GetContractsAsPartyA query all settlement contracts as Party B info by address
// @param addr user qlcchain address
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
// @return all settlement contract as PartyB
func (s *SettlementAPI) GetContractsAsPartyB(addr *types.Address, count int, offset *int) ([]*SettlementContract, error) {
	return nil, nil
}

// TODO: to be confirmed
type CDRParam struct {
	abi.CDRParam
}

// GetProcessCDRBlock save CDR data for the settlement
// @param addr settlement smart contract address
// @param params CDR params to be processed
// @return contract send block to be processed
func (s *SettlementAPI) GetProcessCDRBlock(addr *types.Address, params []*CDRParam) ([]*types.StateBlock, error) {
	return nil, nil
}

// TODO: report data te be confirmed
type SettlementRecord struct {
	Address                  types.Address // smart contract address
	Customer                 string
	CustomerSr               string
	Country                  string
	Operator                 string
	ServiceId                string
	MCC                      int
	MNC                      int
	Currency                 string
	UnitPrice                float64
	SumOfBillableSMSCustomer int64
	SumOfTOTPrice            float64
}

// TODO: query condition to be confirmed, maybe more interface?
// GenerateReport Generate reports for specified contracts based on start and end date
// @param addr user qlcchain address
// @param start report start date
// @param end report end data
// @return settlement report
func (s *SettlementAPI) GenerateReport(addr *types.Address, start, end time.Time) (*[]SettlementRecord, error) {
	return nil, nil
}

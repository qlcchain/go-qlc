/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import "github.com/qlcchain/go-qlc/common/types"

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

func sortInvoiceFun(r1, r2 *InvoiceRecord) bool {
	if r1.StartDate < r2.EndDate {
		return true
	}
	if r1.StartDate > r2.EndDate {
		return false
	}
	return r1.EndDate < r2.EndDate
}

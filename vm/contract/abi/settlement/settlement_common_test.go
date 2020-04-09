/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"time"

	"github.com/qlcchain/go-qlc/mock"
)

var (
	createContractParam = CreateContractParam{
		PartyA: Contractor{
			Address: mock.Address(),
			Name:    "PCCWG",
		},
		PartyB: Contractor{
			Address: mock.Address(),
			Name:    "HKTCSL",
		},
		Previous: mock.Hash(),
		Services: []ContractService{
			{
				ServiceId:   mock.Hash().String(),
				Mcc:         1,
				Mnc:         2,
				TotalAmount: 100,
				UnitPrice:   2,
				Currency:    "USD",
			}, {
				ServiceId:   mock.Hash().String(),
				Mcc:         22,
				Mnc:         1,
				TotalAmount: 300,
				UnitPrice:   4,
				Currency:    "USD",
			},
		},
		SignDate:  time.Now().AddDate(0, 0, -5).Unix(),
		StartDate: time.Now().AddDate(0, 0, -2).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 2).Unix(),
	}

	cdrParam = CDRParam{
		Index:         1,
		SmsDt:         time.Now().Unix(),
		Sender:        "PCCWG",
		Destination:   "85257***343",
		SendingStatus: SendingStatusSent,
		DlrStatus:     DLRStatusDelivered,
	}

	assetParam = AssetParam{
		Owner: Contractor{
			Address: mock.Address(),
			Name:    "HKT-CSL",
		},
		Previous: mock.Hash(),
		Assets: []*Asset{
			{
				Mcc:         42,
				Mnc:         5,
				TotalAmount: 1000,
				SLAs: []*SLA{
					NewLatency(60*time.Second, []*Compensation{
						{
							Low:  50,
							High: 60,
							Rate: 10,
						},
						{
							Low:  60,
							High: 80,
							Rate: 20.5,
						},
					}),
					NewDeliveredRate(float32(0.95), []*Compensation{
						{
							Low:  0.8,
							High: 0.9,
							Rate: 5,
						},
						{
							Low:  0.7,
							High: 0.8,
							Rate: 5.5,
						},
					}),
				},
			},
		},
		SignDate:  time.Now().Unix(),
		StartDate: time.Now().AddDate(0, 0, 1).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 1).Unix(),
		Status:    AssetStatusActivated,
	}
)

func buildContractParam() (param *ContractParam) {
	cp := createContractParam
	cp.PartyA.Address = mock.Address()
	cp.PartyB.Address = mock.Address()

	cd := time.Now().Unix()
	param = &ContractParam{
		CreateContractParam: cp,
		PreStops:            []string{"PCCWG", "test1"},
		NextStops:           []string{"HKTCSL", "test2"},
		ConfirmDate:         cd,
		Status:              ContractStatusActivated,
	}
	return
}

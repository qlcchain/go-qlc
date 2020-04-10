/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/util"
)

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
	records map[string]*MatchingRecord
}

func newCompareRecord() *CompareRecord {
	return &CompareRecord{records: make(map[string]*MatchingRecord)}
}

func (z *CompareRecord) MarshalJSON() ([]byte, error) {
	return json.Marshal(z.records)
}

func (z *CompareRecord) UpdateCounter(party string, isMatching, state bool) {
	if _, ok := z.records[party]; !ok {
		z.records[party] = &MatchingRecord{}
	}

	v := z.records[party]
	if isMatching {
		if state {
			v.Matching.Success++
		} else {
			v.Matching.Fail++
		}
	} else {
		if state {
			v.Orphan.Success++
		} else {
			v.Orphan.Fail++
		}
	}
}

func (z *CompareRecord) DoCalculate() {
	for _, v := range z.records {
		v.Matching.DoCalculate()
		v.Orphan.DoCalculate()
	}
}

type SummaryResult struct {
	Contract *ContractParam            `json:"contract"`
	Records  map[string]*CompareRecord `json:"records"`
	Total    *CompareRecord            `json:"total"`
}

func newSummaryResult() *SummaryResult {
	return &SummaryResult{
		Records: make(map[string]*CompareRecord),
		Total:   newCompareRecord(),
	}
}

func (z *SummaryResult) UpdateState(sender, party string, isMatching, state bool) {
	if sender != "" {
		if _, ok := z.Records[sender]; !ok {
			z.Records[sender] = newCompareRecord()
		}
		z.Records[sender].UpdateCounter(party, isMatching, state)
	}
	z.Total.UpdateCounter(party, isMatching, state)
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

type MultiPartySummaryResult struct {
	Contracts []*ContractParam          `json:"contracts"`
	Records   map[string]*CompareRecord `json:"records"`
	Total     *CompareRecord            `json:"total"`
}

func newMultiPartySummaryResult() *MultiPartySummaryResult {
	return &MultiPartySummaryResult{
		Records: make(map[string]*CompareRecord),
		Total:   newCompareRecord(),
	}
}

func (z *MultiPartySummaryResult) DoCalculate() {
	for k := range z.Records {
		z.Records[k].DoCalculate()
	}
	z.Total.DoCalculate()
}

func (z *MultiPartySummaryResult) UpdateState(sender, party string, isMatching, state bool) {
	if sender != "" {
		if _, ok := z.Records[sender]; !ok {
			z.Records[sender] = newCompareRecord()
		}
		z.Records[sender].UpdateCounter(party, isMatching, state)
	}
	z.Total.UpdateCounter(party, isMatching, state)
}

func (z *MultiPartySummaryResult) String() string {
	return util.ToIndentString(z)
}

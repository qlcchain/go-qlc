/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

//go:generate msgp
type VmLog struct {
	Topics []Hash `msg:"topics" json:"topics"`
	Data   []byte `msg:"data" json:"data"`
}

//go:generate msgp
type VmLogs struct {
	Logs []*VmLog `msg:"logs" json:"logs"`
}

func (vl *VmLogs) Hash() *Hash {
	if len(vl.Logs) == 0 {
		return nil
	}
	var source []byte

	for _, vmLog := range vl.Logs {
		for _, topic := range vmLog.Topics {
			source = append(source, topic[:]...)
		}
		source = append(source, vmLog.Data...)
	}

	hash := HashData(source)
	return &hash
}

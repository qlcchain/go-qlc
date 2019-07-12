package types

import "encoding/json"

//go:generate msgp
type Benefit struct {
	Balance Balance `msg:"balance,extension" json:"balance"`
	Vote    Balance `msg:"vote,extension" json:"vote"`
	Network Balance `msg:"network,extension" json:"network"`
	Storage Balance `msg:"storage,extension" json:"storage"`
	Oracle  Balance `msg:"oracle,extension" json:"oracle"`
	Total   Balance `msg:"total,extension" json:"total"`
	Hash    Balance `msg:"hash,extension" json:"hash"`
}

func (b *Benefit) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}

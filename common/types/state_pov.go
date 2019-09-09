package types

import (
	"fmt"
	"strings"
)

const (
	PovStatePrefixAcc = byte(1)
	PovStatePrefixRep = byte(2)
)

func PovCreateStateKey(prefix byte, rawKey []byte) []byte {
	key := make([]byte, 1+len(rawKey))
	key[0] = prefix
	copy(key[1:], rawKey)
	return key
}

func PovCreateAccountStateKey(address Address) []byte {
	return PovCreateStateKey(PovStatePrefixAcc, address.Bytes())
}

func PovCreateRepStateKey(address Address) []byte {
	return PovCreateStateKey(PovStatePrefixRep, address.Bytes())
}

//go:generate msgp

type PovAccountState struct {
	Balance     Balance          `msg:"b,extension" json:"balance"`
	Vote        Balance          `msg:"v,extension" json:"vote"`
	Network     Balance          `msg:"n,extension" json:"network"`
	Storage     Balance          `msg:"s,extension" json:"storage"`
	Oracle      Balance          `msg:"o,extension" json:"oracle"`
	TokenStates []*PovTokenState `msg:"ts" json:"tokenStates"`
}

func NewPovAccountState() *PovAccountState {
	return &PovAccountState{
		Balance: NewBalance(0),
		Vote:    NewBalance(0),
		Network: NewBalance(0),
		Storage: NewBalance(0),
		Oracle:  NewBalance(0),
	}
}

func (as *PovAccountState) GetTokenState(tokenType Hash) *PovTokenState {
	for _, token := range as.TokenStates {
		if token.Type == tokenType {
			return token
		}
	}
	return nil
}

func (as *PovAccountState) TotalBalance() Balance {
	return as.Balance.Add(as.Vote).Add(as.Storage).Add(as.Network).Add(as.Oracle)
}

func (as *PovAccountState) Serialize() ([]byte, error) {
	return as.MarshalMsg(nil)
}

func (as *PovAccountState) Deserialize(text []byte) error {
	_, err := as.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (as *PovAccountState) Clone() *PovAccountState {
	newAs := *as

	newAs.TokenStates = nil
	for _, ts := range as.TokenStates {
		newTs := ts.Clone()
		newAs.TokenStates = append(newAs.TokenStates, newTs)
	}

	return &newAs
}

func (as *PovAccountState) String() string {
	sb := strings.Builder{}
	sb.WriteString("{")

	asInfo := fmt.Sprintf("Balance:%s, Vote:%s, Network:%s, Storage:%s, Oracle:%s",
		as.Balance, as.Vote, as.Network, as.Storage, as.Oracle)
	sb.WriteString(asInfo)

	sb.WriteString(", TokenState:[")
	for _, ts := range as.TokenStates {
		sb.WriteString(ts.String())
	}
	sb.WriteString("]")

	sb.WriteString("}")
	return sb.String()
}

type PovTokenState struct {
	Type           Hash    `msg:"t,extension" json:"type"`
	Hash           Hash    `msg:"h,extension" json:"hash"`
	Representative Address `msg:"r,extension" json:"representative"`
	Balance        Balance `msg:"b,extension" json:"balance"`
}

func NewPovTokenState() *PovTokenState {
	return &PovTokenState{
		Balance: NewBalance(0),
	}
}

func (ts *PovTokenState) Serialize() ([]byte, error) {
	return ts.MarshalMsg(nil)
}

func (ts *PovTokenState) Deserialize(text []byte) error {
	_, err := ts.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (ts *PovTokenState) Clone() *PovTokenState {
	newTs := *ts
	return &newTs
}

func (ts *PovTokenState) String() string {
	return fmt.Sprintf("{Type:%s, Hash:%s, Rep:%s, Balance:%s}",
		ts.Type, ts.Hash, ts.Representative, ts.Balance)
}

type PovRepState struct {
	Balance Balance `msg:"b,extension" json:"balance"`
	Vote    Balance `msg:"v,extension" json:"vote"`
	Network Balance `msg:"n,extension" json:"network"`
	Storage Balance `msg:"s,extension" json:"storage"`
	Oracle  Balance `msg:"o,extension" json:"oracle"`
	Total   Balance `msg:"t,extension" json:"total"`

	Status    uint32 `msg:"st" json:"status"`
	PovHeight uint64 `msg:"ph" json:"povHeight"`
}

func NewPovRepState() *PovRepState {
	return &PovRepState{
		Balance: NewBalance(0),
		Vote:    NewBalance(0),
		Network: NewBalance(0),
		Storage: NewBalance(0),
		Oracle:  NewBalance(0),
		Total:   NewBalance(0),
	}
}

func (rs *PovRepState) Serialize() ([]byte, error) {
	return rs.MarshalMsg(nil)
}

func (rs *PovRepState) Deserialize(text []byte) error {
	_, err := rs.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (rs *PovRepState) Clone() *PovRepState {
	newRs := *rs
	return &newRs
}

func (rs *PovRepState) CalcTotal() Balance {
	return rs.Balance.Add(rs.Vote).Add(rs.Network).Add(rs.Oracle).Add(rs.Storage)
}

func (rs *PovRepState) String() string {
	return fmt.Sprintf("{Balance:%s, Vote:%s, Network:%s, Storage:%s, Oracle:%s, Total:%s}",
		rs.Balance, rs.Vote, rs.Network, rs.Storage, rs.Oracle, rs.Total)
}

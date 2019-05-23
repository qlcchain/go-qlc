package types

import (
	"fmt"
	"strings"
)

//go:generate msgp

type PovAccountState struct {
	Hash        Hash             `msg:"hash,extension" json:"hash"`
	Balance     Balance          `msg:"balance,extension" json:"balance"`
	Vote        Balance          `msg:"vote,extension" json:"vote"`
	Network     Balance          `msg:"network,extension" json:"network"`
	Storage     Balance          `msg:"storage,extension" json:"storage"`
	Oracle      Balance          `msg:"oracle,extension" json:"oracle"`
	TokenStates []*PovTokenState `msg:"tokenStates" json:"tokenStates"`
	RepState    *PovRepState     `msg:"repState" json:"repState"`
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

	if as.RepState != nil {
		newAs.RepState = as.RepState.Clone()
	}

	return &newAs
}

func (as *PovAccountState) String() string {
	sb := strings.Builder{}
	sb.WriteString("{")

	asInfo := fmt.Sprintf("Hash:%s, Balance:%s, Vote:%s, Network:%s, Storage:%s, Oracle:%s",
		as.Hash, as.Balance, as.Vote, as.Network, as.Storage, as.Oracle)
	sb.WriteString(asInfo)

	sb.WriteString(", TokenState:[")
	for _, ts := range as.TokenStates {
		sb.WriteString(ts.String())
	}
	sb.WriteString("]")

	if as.RepState != nil {
		sb.WriteString(", RepState:{")
		sb.WriteString(as.RepState.String())
		sb.WriteString("}")
	}

	sb.WriteString("}")
	return sb.String()
}

type PovTokenState struct {
	Type           Hash    `msg:"type,extension" json:"type"`
	Representative Address `msg:"rep,extension" json:"representative"`
	Balance        Balance `msg:"balance,extension" json:"balance"`
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
	return fmt.Sprintf("{Type:%s, Rep:%s, Balance:%s}",
		ts.Type, ts.Representative, ts.Balance)
}

type PovRepState struct {
	Balance Balance `msg:"balance,extension" json:"balance"`
	Vote    Balance `msg:"vote,extension" json:"vote"`
	Network Balance `msg:"network,extension" json:"network"`
	Storage Balance `msg:"storage,extension" json:"storage"`
	Oracle  Balance `msg:"oracle,extension" json:"oracle"`
	Total   Balance `msg:"total,extension" json:"total"`
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

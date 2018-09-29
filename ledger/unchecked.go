package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
)

//go:generate msgp
type PendingKey struct {
	Address types.Address `msg:"account" json:"account"`
	Hash    types.Hash    `msg:"hash" json:"hash"`
}

//go:generate msgp
type PendingInfo struct {
	Source types.Address `msg:"source" json:"source"`
	Amount types.Balance `msg:"amount" json:"amount"`
	Type   types.Hash    `msg:"type" json:"type"`
}

type UncheckedKind byte

const (
	UncheckedKindPrevious UncheckedKind = iota
	UncheckedKindSource
)

type UncheckedBlockWalkFunc func(block Block, kind UncheckedKind) error

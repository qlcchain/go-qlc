package abi

import (
	"github.com/qlcchain/go-qlc/common/types"
)

const (
	PermissionDataAdmin uint8 = iota
	PermissionDataNode
	PermissionDataNodeIndex
)

const (
	PermissionAdminStatusActive uint8 = iota
	PermissionAdminStatusHandOver
)

const (
	PermissionNodeKindIPPort uint8 = iota
	PermissionNodeKindPeerID
	PermissionNodeKindInvalid
)

const (
	PermissionCommentMaxLen = 128
)

func PermissionAdminStatusString(status uint8) string {
	switch status {
	case PermissionAdminStatusActive:
		return "active"
	case PermissionAdminStatusHandOver:
		return "in hand over"
	default:
		return ""
	}
}

//go:generate msgp
type AdminAccount struct {
	Addr    types.Address `msg:"a,extension" json:"account"`
	Comment string        `msg:"c" json:"comment"`
	Status  uint8         `msg:"s" json:"status"`
}

//go:generate msgp
type PermNode struct {
	Index   uint32 `msg:"-" json:"index"`
	Kind    uint8  `msg:"k" json:"kind"`
	Node    string `msg:"n" json:"node"`
	Comment string `msg:"c" json:"comment"`
	Valid   bool   `msg:"s" json:"valid"`
}

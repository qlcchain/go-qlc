package abi

import (
	"github.com/qlcchain/go-qlc/common/types"
)

const (
	PermissionDataAdmin uint8 = iota
	PermissionDataNode
)

const (
	PermissionCommentMaxLen = 128
)

//go:generate msgp
type AdminAccount struct {
	Account types.Address `msg:"-" json:"account"`
	Comment string        `msg:"c" json:"comment"`
	Valid   bool          `msg:"v" json:"valid"`
}

//go:generate msgp
type PermNode struct {
	NodeId  string `msg:"-" json:"nodeId"`
	NodeUrl string `msg:"nu" json:"nodeUrl"`
	Comment string `msg:"c" json:"comment"`
	Valid   bool   `msg:"s" json:"valid"`
}

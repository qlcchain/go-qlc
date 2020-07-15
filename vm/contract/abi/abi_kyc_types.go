package abi

import (
	"errors"
	"github.com/qlcchain/go-qlc/common/types"
)

const (
	KYCDataAdmin uint8 = iota
	KYCDataStatus
	KYCDataAddress
	KYCDataTradeAddress
)

const (
	KYCTradeAddressAdd uint8 = iota
	KYCTradeAddressRemove
	KYCTradeAddressInvalid
)

const (
	KYCCommentMaxLen = 128
)

//go:generate msgp
type KYCAdminAccount struct {
	Account types.Address `msg:"-" json:"account"`
	Comment string        `msg:"c" json:"comment"`
	Valid   bool          `msg:"v" json:"valid"`
}

//go:generate msgp
type KYCStatus struct {
	ChainAddress types.Address `msg:"-" json:"chainAddress"`
	Status       string        `msg:"s" json:"status"`
	Valid        bool          `msg:"v" json:"valid"`
}

//go:generate msgp
type KYCAddress struct {
	ChainAddress types.Address `msg:"-" json:"chainAddress"`
	Action       uint8         `msg:"-" json:"action"`
	TradeAddress string        `msg:"t" json:"tradeAddress"`
	Comment      string        `msg:"c" json:"comment"`
	Valid        bool          `msg:"v" json:"valid"`
}

func (z *KYCAddress) GetMixKey() []byte {
	key := make([]byte, 0)
	key = append(key, z.ChainAddress.Bytes()...)
	taKey, _ := types.Sha256HashData([]byte(z.TradeAddress))
	key = append(key, taKey.Bytes()...)
	return key
}

func (z *KYCAddress) GetKey() []byte {
	var taKey types.Hash
	taKey, _ = types.Sha256HashData([]byte(z.TradeAddress))
	return taKey[:]
}

func KYCTradeAddressActionFromString(action string) (uint8, error) {
	switch action {
	case "add":
		return KYCTradeAddressAdd, nil
	case "remove":
		return KYCTradeAddressRemove, nil
	default:
		return KYCTradeAddressInvalid, errors.New("wrong action")
	}
}

func KYCTradeAddressActionToString(action uint8) string {
	switch action {
	case KYCTradeAddressAdd:
		return "add"
	case KYCTradeAddressRemove:
		return "remove"
	default:
		return "wrong action"
	}
}

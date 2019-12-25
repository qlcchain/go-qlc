package abi

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"math/big"
	"strings"
)

const (
	JsonPublisher = `
	[
		{"type":"function","name":"Publish","inputs":[
			{"name":"account","type":"address"},
			{"name":"pType","type":"uint32"},
			{"name":"pID","type":"hash"},
			{"name":"pubKey","type":"uint8[]"},
			{"name":"verifiers","type":"address[]"},
			{"name":"codes","type":"hash[]"},
			{"name":"fee","type":"uint256"}
		]},
		{"type":"function","name":"UnPublish","inputs":[
			{"name":"account","type":"address"},
			{"name":"Type","type":"uint32"},
			{"name":"ID","type":"hash"}
		]},
		{"type":"variable","name":"PublishInfo","inputs":[
			{"name":"verifiers","type":"address[]"},
			{"name":"codes","type":"hash[]"},
			{"name":"fee","type":"uint256"}
		]}
	]`

	MethodNamePublish       = "Publish"
	MethodNameUnPublish     = "UnPublish"
	VariableNamePublishInfo = "PublishInfo"
)

var (
	PublisherABI, _ = abi.JSONToABIContract(strings.NewReader(JsonPublisher))
)

const (
	// prefix + contractAddr + type + id + pk + account => key + verifiers + codes + fee
	PublishTypeIndexS = 1 + types.AddressSize
	PublishTypeIndexE = PublishTypeIndexS + 4
	PublishIdIndexS   = PublishTypeIndexE
	PublishIdIndexE   = PublishIdIndexS + sha256.Size
	PublishPkIndexS   = PublishIdIndexE
	PublishPkIndexE   = PublishPkIndexS + ed25519.PublicKeySize
	PublishAccIndexS  = PublishPkIndexE
	PublishAccIndexE  = PublishAccIndexS + 32
)

type PublishInfo struct {
	Account   types.Address
	PType     uint32
	PID       types.Hash
	PubKey    []byte
	Verifiers []types.Address
	Codes     []types.Hash
	Fee       *big.Int
}

type UnPublishInfo struct {
	Account types.Address
	PType   uint32
	PID     types.Hash
}

type PubKeyInfo struct {
	Verifiers []types.Address
	Codes     []types.Hash
	Fee       *big.Int
}

func PublishInfoCheck(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash, pk []byte, fee types.Balance) error {
	switch pt {
	case types.OracleTypeEmail, types.OracleTypeWeChat:
		publishCost := types.Balance{Int: types.PublishCost}
		if fee.Compare(publishCost) == types.BalanceCompSmaller {
			return fmt.Errorf("fee is not enough")
		}

		if len(pk) != ed25519.PublicKeySize {
			return fmt.Errorf("pk len err")
		}

		err := CheckPublishKeyRegistered(ctx, account, pt, id, pk)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("puslish type(%s) err", types.OracleTypeToString(pt))
	}

	return nil
}

func UnPublishInfoCheck(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash) error {
	switch pt {
	case types.OracleTypeEmail, types.OracleTypeWeChat:
		if !CheckPublishInfoExist(ctx, account, pt, id) {
			return fmt.Errorf("there is no valid kind(%s) of id(%s) for(%s)", types.OracleTypeToString(pt), id, account)
		}
	default:
		return fmt.Errorf("puslish type(%s) err", types.OracleTypeToString(pt))
	}

	return nil
}

func CheckPublishKeyRegistered(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash, pk []byte) error {
	typeStr := types.OracleTypeToString(pt)

	pis := GetPublishInfoByTypeAndId(ctx, pt, id)
	if pis != nil {
		piNum := len(pis)
		if piNum == 1 {
			if bytes.Equal(account[:], pis[0].Account[:]) && bytes.Equal(pk, pis[0].PubKey) {
				return fmt.Errorf("you have regisered the pubkey(%s) for id(%s) of kind(%s)",
					types.NewHexBytesFromData(pk), id, typeStr)
			}
		} else if piNum > 1 {
			return fmt.Errorf("id(%s) of kind(%s) has been registered by more than one user", id, typeStr)
		}
	}

	return nil
}

func CheckPublishInfoExist(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash) bool {
	pk := GetPublishKeyByAccountAndTypeAndID(ctx, account, pt, id)
	if pk == nil {
		return false
	}

	key := append(util.BE_Uint32ToBytes(pt), id[:]...)
	key = append(key, pk...)
	key = append(key, account[:]...)
	_, err := ctx.GetStorage(types.PublisherAddress[:], key)
	if err != nil {
		return false
	}
	return true
}

func GetPublishInfoByTypeAndId(ctx *vmstore.VMContext, pt uint32, id types.Hash) []*PublishInfo {
	pis := make([]*PublishInfo, 0)

	itKey := append(types.PublisherAddress[:], util.BE_Uint32ToBytes(pt)...)
	itKey = append(itKey, id[:]...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
		if err != nil {
			return err
		}

		var info PubKeyInfo
		err = PublisherABI.UnpackVariable(&info, VariableNamePublishInfo, value)
		if err != nil {
			return err
		}

		pi := &PublishInfo{
			Account:   addr,
			PType:     pt,
			PID:       id,
			PubKey:    key[PublishPkIndexS:PublishPkIndexE],
			Verifiers: info.Verifiers,
			Codes:     info.Codes,
			Fee:       info.Fee,
		}

		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		return nil
	}

	return pis
}

func GetAllPublishInfo(ctx *vmstore.VMContext) []*PublishInfo {
	pis := make([]*PublishInfo, 0)

	err := ctx.Iterator(types.PublisherAddress[:], func(key []byte, value []byte) error {
		pt := util.BE_BytesToUint32(key[PublishTypeIndexS:PublishTypeIndexE])

		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
		if err != nil {
			return err
		}

		var info PubKeyInfo
		err = PublisherABI.UnpackVariable(&info, VariableNamePublishInfo, value)
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[PublishIdIndexS:PublishIdIndexE])
		if err != nil {
			return err
		}

		pi := &PublishInfo{
			Account:   addr,
			PType:     pt,
			PID:       id,
			PubKey:    key[PublishPkIndexS:PublishPkIndexE],
			Verifiers: info.Verifiers,
			Codes:     info.Codes,
			Fee:       info.Fee,
		}

		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		return nil
	}

	return pis
}

func GetPublishInfoByType(ctx *vmstore.VMContext, pt uint32) []*PublishInfo {
	pis := make([]*PublishInfo, 0)

	itKey := append(types.PublisherAddress[:], util.BE_Uint32ToBytes(pt)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
		if err != nil {
			return err
		}

		var info PubKeyInfo
		err = PublisherABI.UnpackVariable(&info, VariableNamePublishInfo, value)
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[PublishIdIndexS:PublishIdIndexE])
		if err != nil {
			return err
		}

		pi := &PublishInfo{
			Account:   addr,
			PType:     pt,
			PID:       id,
			PubKey:    key[PublishPkIndexS:PublishPkIndexE],
			Verifiers: info.Verifiers,
			Codes:     info.Codes,
			Fee:       info.Fee,
		}

		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		return nil
	}

	return pis
}

func GetPublishInfoByAccount(ctx *vmstore.VMContext, account types.Address) []*PublishInfo {
	pis := make([]*PublishInfo, 0)

	err := ctx.Iterator(types.PublisherAddress[:], func(key []byte, value []byte) error {
		if !bytes.Equal(key[PublishAccIndexS:PublishAccIndexE], account[:]) {
			return nil
		}

		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
		if err != nil {
			return err
		}

		var info PubKeyInfo
		err = PublisherABI.UnpackVariable(&info, VariableNamePublishInfo, value)
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[PublishIdIndexS:PublishIdIndexE])
		if err != nil {
			return err
		}

		pi := &PublishInfo{
			Account:   addr,
			PType:     util.BE_BytesToUint32(key[PublishTypeIndexS:PublishTypeIndexE]),
			PID:       id,
			PubKey:    key[PublishPkIndexS:PublishPkIndexE],
			Verifiers: info.Verifiers,
			Codes:     info.Codes,
			Fee:       info.Fee,
		}

		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		return nil
	}

	return pis
}

func GetPublishInfoByAccountAndType(ctx *vmstore.VMContext, account types.Address, pt uint32) []*PublishInfo {
	pis := make([]*PublishInfo, 0)

	itKey := append(types.PublisherAddress[:], util.BE_Uint32ToBytes(pt)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		if !bytes.Equal(key[PublishAccIndexS:PublishAccIndexE], account[:]) {
			return nil
		}

		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
		if err != nil {
			return err
		}

		var info PubKeyInfo
		err = PublisherABI.UnpackVariable(&info, VariableNamePublishInfo, value)
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[PublishIdIndexS:PublishIdIndexE])
		if err != nil {
			return err
		}

		pi := &PublishInfo{
			Account:   addr,
			PType:     util.BE_BytesToUint32(key[PublishTypeIndexS:PublishTypeIndexE]),
			PID:       id,
			PubKey:    key[PublishPkIndexS:PublishPkIndexE],
			Verifiers: info.Verifiers,
			Codes:     info.Codes,
			Fee:       info.Fee,
		}

		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		return nil
	}

	return pis
}

func GetPublishKeyByAccountAndTypeAndID(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash) []byte {
	var ret []byte

	itKey := append(types.PublisherAddress[:], util.BE_Uint32ToBytes(pt)...)
	itKey = append(itKey, id[:]...)
	_ = ctx.Iterator(itKey, func(key []byte, value []byte) error {
		if len(ret) > 0 {
			return nil
		}

		if !bytes.Equal(key[PublishAccIndexS:PublishAccIndexE], account[:]) {
			return nil
		}

		ret = key[PublishPkIndexS:PublishPkIndexE]
		return nil
	})

	if len(ret) > 0 {
		return ret
	}

	return nil
}

func GetPublishInfo(ctx *vmstore.VMContext, pt uint32, id types.Hash, pk []byte) []*PublishInfo {
	pis := make([]*PublishInfo, 0)

	itKey := append(types.PublisherAddress[:], util.BE_Uint32ToBytes(pt)...)
	itKey = append(itKey, id[:]...)
	itKey = append(itKey, pk...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
		if err != nil {
			return err
		}

		var info PubKeyInfo
		err = PublisherABI.UnpackVariable(&info, VariableNamePublishInfo, value)
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[PublishIdIndexS:PublishIdIndexE])
		if err != nil {
			return err
		}

		pi := &PublishInfo{
			Account:   addr,
			PType:     util.BE_BytesToUint32(key[PublishTypeIndexS:PublishTypeIndexE]),
			PID:       id,
			PubKey:    key[PublishPkIndexS:PublishPkIndexE],
			Verifiers: info.Verifiers,
			Codes:     info.Codes,
			Fee:       info.Fee,
		}

		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		return nil
	}

	return pis
}

package abi

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"strings"
)

const (
	JsonOracle = `
	[
		{"type":"function","name":"Oracle","inputs":[
			{"name":"account","type":"address"},
			{"name":"oType","type":"uint32"},
			{"name":"oID","type":"hash"},
			{"name":"pubKey","type":"uint8[]"},
			{"name":"code","type":"string"},
			{"name":"hash","type":"hash"}
		]},
		{"type":"variable","name":"OracleInfo","inputs":[
			{"name":"code","type":"string"},
			{"name":"hash","type":"hash"}
		]}
	]`

	MethodNameOracle       = "Oracle"
	VariableNameOracleInfo = "OracleInfo"
)

var (
	OracleABI, _ = abi.JSONToABIContract(strings.NewReader(JsonOracle))
)

const (
	// prefix + contractAddr + type + id + pk + account => code
	OracleTypeIndexS = 1 + types.AddressSize
	OracleTypeIndexE = OracleTypeIndexS + 4
	OracleIdIndexS   = OracleTypeIndexE
	OracleIdIndexE   = OracleIdIndexS + sha256.Size
	OraclePkIndexS   = OracleIdIndexE
	OraclePkIndexE   = OraclePkIndexS + ed25519.PublicKeySize
	OracleAccIndexS  = OraclePkIndexE
	OracleAccIndexE  = OracleAccIndexS + 32
)

type OracleInfo struct {
	Account types.Address
	OType   uint32
	OID     types.Hash
	PubKey  []byte
	Code    string
	Hash    types.Hash
}

type CodeInfo struct {
	Code string
	Hash types.Hash
}

func OracleInfoCheck(ctx *vmstore.VMContext, account types.Address, ot uint32, id types.Hash, pk []byte, code string, hash types.Hash) error {
	switch ot {
	case types.OracleTypeEmail, types.OracleTypeWeChat:
		if len(pk) != ed25519.PublicKeySize {
			return fmt.Errorf("pk len err")
		}

		err := VerifierPledgeCheck(ctx, account)
		if err != nil {
			return err
		}

		_, err = GetVerifierInfoByAccountAndType(ctx, account, ot)
		if err != nil {
			return err
		}

		blk, err := ctx.Ledger.GetStateBlockConfirmed(hash)
		if err != nil {
			return err
		}

		var info PublishInfo
		err = PublisherABI.UnpackMethod(&info, MethodNamePublish, blk.GetData())
		if err != nil {
			return err
		}

		if id != info.PID || ot != info.PType || !bytes.Equal(pk, info.PubKey) {
			return fmt.Errorf("wrong oracle info with hash(%s)", hash)
		}

		index := -1
		for i, v := range info.Verifiers {
			if v == account {
				index = i
				break
			}
		}

		if index == -1 {
			return fmt.Errorf("get verifier err")
		}

		codeComb := append([]byte(types.NewHexBytesFromData(pk).String()), []byte(code)...)
		codeHash, err := types.Sha256HashData(codeComb)
		if err != nil {
			return err
		}

		if codeHash != info.Codes[index] {
			return fmt.Errorf("verification code err")
		}
	default:
		return fmt.Errorf("puslish type(%s) err", types.OracleTypeToString(ot))
	}

	return nil
}

func GetAllOracleInfo(ctx *vmstore.VMContext) []*OracleInfo {
	ois := make([]*OracleInfo, 0)

	err := ctx.Iterator(types.OracleAddress[:], func(key []byte, value []byte) error {
		ot := util.BE_BytesToUint32(key[OracleTypeIndexS:OracleTypeIndexE])

		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = OracleABI.UnpackVariable(&info, VariableNameOracleInfo, value)
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[OracleIdIndexS:OracleIdIndexE])
		if err != nil {
			return err
		}

		oi := &OracleInfo{
			Account: addr,
			OType:   ot,
			OID:     id,
			PubKey:  key[OraclePkIndexS:OraclePkIndexE],
			Code:    info.Code,
			Hash:    info.Hash,
		}

		ois = append(ois, oi)
		return nil
	})
	if err != nil {
		return nil
	}

	return ois
}

func GetOracleInfoByType(ctx *vmstore.VMContext, ot uint32) []*OracleInfo {
	ois := make([]*OracleInfo, 0)

	itKey := append(types.OracleAddress[:], util.BE_Uint32ToBytes(ot)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = OracleABI.UnpackVariable(&info, VariableNameOracleInfo, value)
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[OracleIdIndexS:OracleIdIndexE])
		if err != nil {
			return err
		}

		oi := &OracleInfo{
			Account: addr,
			OType:   ot,
			OID:     id,
			PubKey:  key[OraclePkIndexS:OraclePkIndexE],
			Code:    info.Code,
			Hash:    info.Hash,
		}

		ois = append(ois, oi)
		return nil
	})
	if err != nil {
		return nil
	}

	return ois
}

func GetOracleInfoByTypeAndID(ctx *vmstore.VMContext, ot uint32, id types.Hash) []*OracleInfo {
	ois := make([]*OracleInfo, 0)

	itKey := append(types.OracleAddress[:], util.BE_Uint32ToBytes(ot)...)
	itKey = append(itKey, id[:]...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = OracleABI.UnpackVariable(&info, VariableNameOracleInfo, value)
		if err != nil {
			return err
		}

		oi := &OracleInfo{
			Account: addr,
			OType:   ot,
			OID:     id,
			PubKey:  key[OraclePkIndexS:OraclePkIndexE],
			Code:    info.Code,
			Hash:    info.Hash,
		}

		ois = append(ois, oi)
		return nil
	})
	if err != nil {
		return nil
	}

	return ois
}

func GetOracleInfoByTypeAndIDAndPk(ctx *vmstore.VMContext, ot uint32, id types.Hash, pk []byte) []*OracleInfo {
	ois := make([]*OracleInfo, 0)

	itKey := append(types.OracleAddress[:], util.BE_Uint32ToBytes(ot)...)
	itKey = append(itKey, id[:]...)
	itKey = append(itKey, pk...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = OracleABI.UnpackVariable(&info, VariableNameOracleInfo, value)
		if err != nil {
			return err
		}

		oi := &OracleInfo{
			Account: addr,
			OType:   ot,
			OID:     id,
			PubKey:  pk,
			Code:    info.Code,
			Hash:    info.Hash,
		}

		ois = append(ois, oi)
		return nil
	})
	if err != nil {
		return nil
	}

	return ois
}

func GetOracleInfoByAccount(ctx *vmstore.VMContext, account types.Address) []*OracleInfo {
	ois := make([]*OracleInfo, 0)

	err := ctx.Iterator(types.OracleAddress[:], func(key []byte, value []byte) error {
		if !bytes.Equal(key[OracleAccIndexS:OracleAccIndexE], account[:]) {
			return nil
		}

		ot := util.BE_BytesToUint32(key[OracleTypeIndexS:OracleTypeIndexE])

		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = OracleABI.UnpackVariable(&info, VariableNameOracleInfo, value)
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[OracleIdIndexS:OracleIdIndexE])
		if err != nil {
			return err
		}

		oi := &OracleInfo{
			Account: addr,
			OType:   ot,
			OID:     id,
			PubKey:  key[OraclePkIndexS:OraclePkIndexE],
			Code:    info.Code,
			Hash:    info.Hash,
		}

		ois = append(ois, oi)
		return nil
	})
	if err != nil {
		return nil
	}

	return ois
}

func GetOracleInfoByAccountAndType(ctx *vmstore.VMContext, account types.Address, ot uint32) []*OracleInfo {
	ois := make([]*OracleInfo, 0)

	itKey := append(types.OracleAddress[:], util.BE_Uint32ToBytes(ot)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		if !bytes.Equal(key[OracleAccIndexS:OracleAccIndexE], account[:]) {
			return nil
		}

		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = OracleABI.UnpackVariable(&info, VariableNameOracleInfo, value)
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[OracleIdIndexS:OracleIdIndexE])
		if err != nil {
			return err
		}

		oi := &OracleInfo{
			Account: addr,
			OType:   ot,
			OID:     id,
			PubKey:  key[OraclePkIndexS:OraclePkIndexE],
			Code:    info.Code,
			Hash:    info.Hash,
		}

		ois = append(ois, oi)
		return nil
	})
	if err != nil {
		return nil
	}

	return ois
}

func GetOracleInfoByHash(ctx *vmstore.VMContext, hash types.Hash) []*OracleInfo {
	ois := make([]*OracleInfo, 0)
	block, err := ctx.Ledger.GetStateBlockConfirmed(hash)
	if err != nil {
		return nil
	}

	var pi PublishInfo
	err = PublisherABI.UnpackMethod(&pi, MethodNamePublish, block.GetData())
	if err != nil {
		return nil
	}

	itKey := append(types.OracleAddress[:], util.BE_Uint32ToBytes(pi.PType)...)
	itKey = append(itKey, pi.PID[:]...)
	itKey = append(itKey, pi.PubKey...)
	err = ctx.Iterator(itKey, func(key []byte, value []byte) error {
		var info CodeInfo
		err = OracleABI.UnpackVariable(&info, VariableNameOracleInfo, value)
		if err != nil {
			return err
		}

		if info.Hash != hash {
			return nil
		}

		ot := util.BE_BytesToUint32(key[OracleTypeIndexS:OracleTypeIndexE])

		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		id, err := types.BytesToHash(key[OracleIdIndexS:OracleIdIndexE])
		if err != nil {
			return err
		}

		oi := &OracleInfo{
			Account: addr,
			OType:   ot,
			OID:     id,
			PubKey:  key[OraclePkIndexS:OraclePkIndexE],
			Code:    info.Code,
			Hash:    info.Hash,
		}

		ois = append(ois, oi)
		return nil
	})
	if err != nil {
		return nil
	}

	return ois
}

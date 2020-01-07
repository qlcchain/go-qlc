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
	"math/big"
	"strings"
)

const (
	JsonPublicKeyDistribution = `
	[
		{"type":"function","name":"PKDVerifierRegister","inputs":[
			{"name":"vType","type":"uint32"},
			{"name":"vInfo","type":"string"}
		]},
		{"type":"function","name":"PKDVerifierUnregister","inputs":[
			{"name":"vType","type":"uint32"}
		]},
		{"type":"function","name":"PKDOracle","inputs":[
			{"name":"oType","type":"uint32"},
			{"name":"oID","type":"hash"},
			{"name":"pubKey","type":"uint8[]"},
			{"name":"code","type":"string"},
			{"name":"hash","type":"hash"}
		]},
		{"type":"function","name":"PKDPublish","inputs":[
			{"name":"pType","type":"uint32"},
			{"name":"pID","type":"hash"},
			{"name":"pubKey","type":"uint8[]"},
			{"name":"verifiers","type":"address[]"},
			{"name":"codes","type":"hash[]"},
			{"name":"fee","type":"uint256"}
		]},
		{"type":"function","name":"PKDUnPublish","inputs":[
			{"name":"pType","type":"uint32"},
			{"name":"pID","type":"hash"}
		]},
		{"type":"variable","name":"PKDVerifierInfo","inputs":[
			{"name":"vInfo","type":"string"},
			{"name":"valid","type":"bool"}
		]},
		{"type":"variable","name":"PKDOracleInfo","inputs":[
			{"name":"code","type":"string"},
			{"name":"hash","type":"hash"}
		]},
		{"type":"variable","name":"PKDPublishInfo","inputs":[
			{"name":"verifiers","type":"address[]"},
			{"name":"codes","type":"hash[]"},
			{"name":"fee","type":"uint256"},
			{"name":"valid","type":"bool"}
		]}
	]`

	MethodNamePKDVerifierRegister   = "PKDVerifierRegister"
	MethodNamePKDVerifierUnregister = "PKDVerifierUnregister"
	MethodNamePKDOracle             = "PKDOracle"
	MethodNamePKDPublish            = "PKDPublish"
	MethodNamePKDUnPublish          = "PKDUnPublish"
	VariableNamePKDPublishInfo      = "PKDPublishInfo"
	VariableNamePKDOracleInfo       = "PKDOracleInfo"
	VariableNamePKDVerifierInfo     = "PKDVerifierInfo"
)

var (
	PublicKeyDistributionABI, _ = abi.JSONToABIContract(strings.NewReader(JsonPublicKeyDistribution))
)

const (
	PKDStorageTypeVerifier byte = iota
	PKDStorageTypePublisher
	PKDStorageTypeOracle
)

const (
	// prefix(1) + contractAddr(32) + pkdType(1) + type(4) + account(32) => info + valid
	VerifierTypeIndexS = 1 + types.AddressSize + 1
	VerifierTypeIndexE = VerifierTypeIndexS + 4
	VerifierAccIndexS  = VerifierTypeIndexE
	VerifierAccIndexE  = VerifierAccIndexS + 32
)

type VerifierRegInfo struct {
	Account types.Address
	VType   uint32
	VInfo   string
}

type VerifierUnRegInfo struct {
	Account types.Address
	VType   uint32
}

type VerifierStorage struct {
	VInfo string
	Valid bool
}

func VerifierRegInfoCheck(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) error {
	switch vType {
	case types.OracleTypeEmail:
		if !util.VerifyEmailFormat(vInfo) {
			return fmt.Errorf("invalid email format (%s)", vInfo)
		}

		if CheckVerifierInfoExist(ctx, account, vType, vInfo) {
			return fmt.Errorf("email has been registered (%s)", vInfo)
		}
	case types.OracleTypeWeChat:
		if CheckVerifierInfoExist(ctx, account, vType, vInfo) {
			return fmt.Errorf("weChat id has been registered (%s)", vInfo)
		}
	default:
		return fmt.Errorf("invalid verifier type(%s)", types.OracleTypeToString(vType))
	}

	return nil
}

func VerifierUnRegInfoCheck(ctx *vmstore.VMContext, account types.Address, vType uint32) error {
	switch vType {
	case types.OracleTypeEmail, types.OracleTypeWeChat:
		if !CheckVerifierExist(ctx, account, vType) {
			return fmt.Errorf("there is no valid registered info")
		}
	default:
		return fmt.Errorf("invalid verifier type(%s)", types.OracleTypeToString(vType))
	}
	return nil
}

func VerifierPledgeCheck(ctx *vmstore.VMContext, account types.Address) error {
	am, err := ctx.Ledger.GetAccountMeta(account)
	if err != nil {
		return err
	}

	minPledgeAmount := types.Balance{Int: types.MinVerifierPledgeAmount}
	if am.CoinOracle.Compare(minPledgeAmount) == types.BalanceCompSmaller {
		return fmt.Errorf("%s have not enough oracle pledge %s", account, am.CoinOracle)
	}

	return nil
}

func CheckVerifierInfoExist(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) bool {
	var key []byte
	key = append(key, PKDStorageTypeVerifier)
	key = append(key, util.BE_Uint32ToBytes(vType)...)
	key = append(key, account[:]...)
	val, err := ctx.GetStorage(types.PubKeyDistributionAddress[:], key)
	if err != nil {
		return false
	}

	var info VerifierStorage
	err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDVerifierInfo, val)
	if err != nil {
		return false
	}

	if info.Valid && info.VInfo == vInfo {
		return true
	}

	return false
}

func CheckVerifierExist(ctx *vmstore.VMContext, account types.Address, vType uint32) bool {
	var key []byte
	key = append(key, PKDStorageTypeVerifier)
	key = append(key, util.BE_Uint32ToBytes(vType)...)
	key = append(key, account[:]...)
	val, err := ctx.GetStorage(types.PubKeyDistributionAddress[:], key)
	if err != nil {
		return false
	}

	var info VerifierStorage
	err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDVerifierInfo, val)
	if err != nil {
		return false
	}

	return info.Valid
}

func GetVerifierInfoByAccountAndType(ctx *vmstore.VMContext, account types.Address, vType uint32) (*VerifierStorage, error) {
	var key []byte
	key = append(key, PKDStorageTypeVerifier)
	key = append(key, util.BE_Uint32ToBytes(vType)...)
	key = append(key, account[:]...)
	val, err := ctx.GetStorage(types.PubKeyDistributionAddress[:], key)
	if err != nil {
		return nil, err
	}

	vs := new(VerifierStorage)
	err = PublicKeyDistributionABI.UnpackVariable(vs, VariableNamePKDVerifierInfo, val)
	if err != nil || !vs.Valid {
		return nil, err
	}

	return vs, nil
}

func GetAllVerifiers(ctx *vmstore.VMContext) ([]*VerifierRegInfo, error) {
	vrs := make([]*VerifierRegInfo, 0)

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeVerifier)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		var vs VerifierStorage
		err := PublicKeyDistributionABI.UnpackVariable(&vs, VariableNamePKDVerifierInfo, value)
		if err != nil || !vs.Valid {
			return err
		}

		addr, err := types.BytesToAddress(key[VerifierAccIndexS:VerifierAccIndexE])
		if err != nil {
			return err
		}

		vr := &VerifierRegInfo{
			Account: addr,
			VType:   util.BE_BytesToUint32(key[VerifierTypeIndexS:VerifierTypeIndexE]),
			VInfo:   vs.VInfo,
		}

		vrs = append(vrs, vr)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return vrs, nil
}

func GetVerifiersByType(ctx *vmstore.VMContext, vType uint32) ([]*VerifierRegInfo, error) {
	vrs := make([]*VerifierRegInfo, 0)

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeVerifier)
	itKey = append(itKey, util.BE_Uint32ToBytes(vType)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		var vs VerifierStorage
		err := PublicKeyDistributionABI.UnpackVariable(&vs, VariableNamePKDVerifierInfo, value)
		if err != nil || !vs.Valid {
			return err
		}

		addr, err := types.BytesToAddress(key[VerifierAccIndexS:VerifierAccIndexE])
		if err != nil {
			return err
		}

		vr := &VerifierRegInfo{
			Account: addr,
			VType:   util.BE_BytesToUint32(key[VerifierTypeIndexS:VerifierTypeIndexE]),
			VInfo:   vs.VInfo,
		}

		vrs = append(vrs, vr)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return vrs, nil
}

func GetVerifiersByAccount(ctx *vmstore.VMContext, account types.Address) ([]*VerifierRegInfo, error) {
	vrs := make([]*VerifierRegInfo, 0)

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeVerifier)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		if !bytes.Equal(key[VerifierAccIndexS:VerifierAccIndexE], account[:]) {
			return nil
		}

		var vs VerifierStorage
		err := PublicKeyDistributionABI.UnpackVariable(&vs, VariableNamePKDVerifierInfo, value)
		if err != nil || !vs.Valid {
			return err
		}

		vr := &VerifierRegInfo{
			Account: account,
			VType:   util.BE_BytesToUint32(key[VerifierTypeIndexS:VerifierTypeIndexE]),
			VInfo:   vs.VInfo,
		}

		vrs = append(vrs, vr)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return vrs, nil
}

const (
	// prefix(1) + contractAddr(32) + pkdType(1) + type(4) + id(32) + pk(32) + account(32) => verifiers + codes + fee
	OracleTypeIndexS = 1 + types.AddressSize + 1
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

		blk, err := ctx.Ledger.GetStateBlockConfirmed(hash)
		if err != nil {
			return err
		}

		var info PublishInfo
		err = PublicKeyDistributionABI.UnpackMethod(&info, MethodNamePKDPublish, blk.GetData())
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeOracle)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		ot := util.BE_BytesToUint32(key[OracleTypeIndexS:OracleTypeIndexE])

		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDOracleInfo, value)
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeOracle)
	itKey = append(itKey, util.BE_Uint32ToBytes(ot)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDOracleInfo, value)
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeOracle)
	itKey = append(itKey, util.BE_Uint32ToBytes(ot)...)
	itKey = append(itKey, id[:]...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDOracleInfo, value)
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeOracle)
	itKey = append(itKey, util.BE_Uint32ToBytes(ot)...)
	itKey = append(itKey, id[:]...)
	itKey = append(itKey, pk...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDOracleInfo, value)
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeOracle)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		if !bytes.Equal(key[OracleAccIndexS:OracleAccIndexE], account[:]) {
			return nil
		}

		ot := util.BE_BytesToUint32(key[OracleTypeIndexS:OracleTypeIndexE])

		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDOracleInfo, value)
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeOracle)
	itKey = append(itKey, util.BE_Uint32ToBytes(ot)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		if !bytes.Equal(key[OracleAccIndexS:OracleAccIndexE], account[:]) {
			return nil
		}

		addr, err := types.BytesToAddress(key[OracleAccIndexS:OracleAccIndexE])
		if err != nil {
			return err
		}

		var info CodeInfo
		err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDOracleInfo, value)
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
	err = PublicKeyDistributionABI.UnpackMethod(&pi, MethodNamePKDPublish, block.GetData())
	if err != nil {
		return nil
	}

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypeOracle)
	itKey = append(itKey, util.BE_Uint32ToBytes(pi.PType)...)
	itKey = append(itKey, pi.PID[:]...)
	itKey = append(itKey, pi.PubKey...)
	err = ctx.Iterator(itKey, func(key []byte, value []byte) error {
		var info CodeInfo
		err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDOracleInfo, value)
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

const (
	// prefix(1) + contractAddr(32) + pkdType(1) + type(4) + id(32) + pk(32) + account(32) => verifiers + codes + fee
	PublishTypeIndexS = 1 + types.AddressSize + 1
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
	Valid     bool
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

	var key []byte
	key = append(key, PKDStorageTypePublisher)
	key = append(key, util.BE_Uint32ToBytes(pt)...)
	key = append(key, id[:]...)
	key = append(key, pk...)
	key = append(key, account[:]...)
	data, err := ctx.GetStorage(types.PubKeyDistributionAddress[:], key)
	if err != nil {
		return false
	}

	var info PubKeyInfo
	err = PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDPublishInfo, data)
	if err != nil || !info.Valid {
		return false
	}

	return true
}

func GetPublishInfoByTypeAndId(ctx *vmstore.VMContext, pt uint32, id types.Hash) []*PublishInfo {
	pis := make([]*PublishInfo, 0)

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypePublisher)
	itKey = append(itKey, util.BE_Uint32ToBytes(pt)...)
	itKey = append(itKey, id[:]...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		var info PubKeyInfo
		err := PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDPublishInfo, value)
		if err != nil || !info.Valid {
			return err
		}

		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypePublisher)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		var info PubKeyInfo
		err := PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDPublishInfo, value)
		if err != nil || !info.Valid {
			return err
		}

		pt := util.BE_BytesToUint32(key[PublishTypeIndexS:PublishTypeIndexE])

		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypePublisher)
	itKey = append(itKey, util.BE_Uint32ToBytes(pt)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		var info PubKeyInfo
		err := PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDPublishInfo, value)
		if err != nil || !info.Valid {
			return err
		}

		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypePublisher)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		if !bytes.Equal(key[PublishAccIndexS:PublishAccIndexE], account[:]) {
			return nil
		}

		var info PubKeyInfo
		err := PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDPublishInfo, value)
		if err != nil || !info.Valid {
			return err
		}

		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypePublisher)
	itKey = append(itKey, util.BE_Uint32ToBytes(pt)...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		if !bytes.Equal(key[PublishAccIndexS:PublishAccIndexE], account[:]) {
			return nil
		}

		var info PubKeyInfo
		err := PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDPublishInfo, value)
		if err != nil || !info.Valid {
			return err
		}

		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
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

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypePublisher)
	itKey = append(itKey, util.BE_Uint32ToBytes(pt)...)
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

func GetPublishInfo(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash) []*PublishInfo {
	pis := make([]*PublishInfo, 0)

	itKey := append(types.PubKeyDistributionAddress[:], PKDStorageTypePublisher)
	itKey = append(itKey, util.BE_Uint32ToBytes(pt)...)
	itKey = append(itKey, id[:]...)
	err := ctx.Iterator(itKey, func(key []byte, value []byte) error {
		if !bytes.Equal(key[PublishAccIndexS:PublishAccIndexE], account[:]) {
			return nil
		}

		var info PubKeyInfo
		err := PublicKeyDistributionABI.UnpackVariable(&info, VariableNamePKDPublishInfo, value)
		if err != nil || !info.Valid {
			return err
		}

		addr, err := types.BytesToAddress(key[PublishAccIndexS:PublishAccIndexE])
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

package abi

import (
	"errors"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonDoDBilling = `[
		{"type":"function","name":"DoDCreateAccount","inputs":[
			{"name":"accountName","type":"string"},
			{"name":"accountInfo","type":"string"},
			{"name":"accountType","type":"string"}
		]},
		{"type":"function","name":"DoDCoupleAccount","inputs":[
			{"name":"uuid","type":"string"},
			{"name":"accountName","type":"string"}
		]},
		{"type":"function","name":"DoDSetService","inputs":[
			{"name":"connectionID","type":"string"},
			{"name":"serviceType","type":"string"},
			{"name":"chargeType","type":"uint8"},
			{"name":"paidRule","type":"uint8"},
			{"name":"buyMode","type":"uint8"},
			{"name":"location","type":"string"},
			{"name":"price","type":"float64"},
			{"name":"unit","type":"string"},
			{"name":"currency","type":"string"},
			{"name":"balance","type":"float64"},
			{"name":"bandwidth","type":"string"},
			{"name":"startTime","type":"string"},
			{"name":"endTime","type":"string"},
			{"name":"tempStartTime","type":"uint64"},
			{"name":"tempEndTime","type":"uint64"},
			{"name":"tempPrice","type":"float64"},
			{"name":"tempBandwidth","type":"string"},
			{"name":"quota","type":"float64"},
			{"name":"usageLimitation","type":"float64"},
			{"name":"minBandwidth","type":"string"},
			{"name":"expireTime","type":"uint64"}
		]},
		{"type":"function","name":"DoDUpdateUsage","inputs":[
			{"name":"connectionID","type":"string"},
			{"name":"chargeType","type":"uint8"},
			{"name":"sourceType","type":"string"},
			{"name":"sourceDevice","type":"string"},
			{"name":"ip","type":"string"},
			{"name":"usage","type":"uint64"},
			{"name":"timestamp","type":"uint64"}
		]}
	]`

	MethodNameDoDCreateAccount = "DoDCreateAccount"
	MethodNameDoDCoupleAccount = "DoDCoupleAccount"
	MethodNameDoDSetService    = "DoDSetService"
	MethodNameDoDUpdateUsage   = "DoDUpdateUsage"
)

var (
	DoDBillingABI, _ = abi.JSONToABIContract(strings.NewReader(JsonDoDBilling))
)

func GetAccount(ctx *vmstore.VMContext, accountName string) *DoDAccount {
	var key []byte
	ah := types.Sha256DHashData([]byte(accountName))
	key = append(key, DoDDataTypeAccount)
	key = append(key, ah.Bytes()...)

	val, err := ctx.GetStorage(contractaddress.DoDBillingAddress.Bytes(), key)
	if err != nil {
		return nil
	}

	account := new(DoDAccount)
	_, err = account.UnmarshalMsg(val)
	if err != nil {
		return nil
	}

	return account
}

func GetConnection(ctx *vmstore.VMContext, connectionID string) *DoDConnection {
	var key []byte
	key = append(key, DoDDataTypeConnection)
	key = append(key, connectionID...)

	val, err := ctx.GetStorage(contractaddress.DoDBillingAddress.Bytes(), key)
	if err != nil {
		return nil
	}

	conn := new(DoDConnection)
	_, err = conn.UnmarshalMsg(val)
	if err != nil {
		return nil
	}

	return conn
}

func GetConnectionUsage(ctx *vmstore.VMContext, connectionID string, startTime, endTime uint64) ([]*DoDUsage, error) {
	usage := make([]*DoDUsage, 0)

	pre := append(contractaddress.DoDBillingAddress.Bytes(), DoDDataTypeUsage)
	pre = append(pre, connectionID...)

	if err := ctx.IteratorAll(pre, func(key []byte, value []byte) error {
		timestamp := util.BE_BytesToUint64(key[1+32+len(connectionID):])
		if timestamp >= startTime && timestamp <= endTime {
			u := new(DoDUsage)
			_, err := u.UnmarshalMsg(value)
			if err != nil {
				return err
			}

			usage = append(usage, u)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return usage, nil
}

func GenerateInvoice(ctx *vmstore.VMContext, connectionID string, startTime, endTime uint64) (*DoDInvoice, error) {
	invoice := new(DoDInvoice)
	invoice.ConnectionID = connectionID
	invoice.StartTime = startTime
	invoice.EndTime = endTime

	conn := GetConnection(ctx, connectionID)
	if conn == nil {
		return nil, errors.New("connection not exist")
	}

	pre := append(contractaddress.DoDBillingAddress.Bytes(), DoDDataTypeUsage)
	pre = append(pre, connectionID...)

	usage := float64(0)

	if err := ctx.IteratorAll(pre, func(key []byte, value []byte) error {
		timestamp := util.BE_BytesToUint64(key[1+32+len(connectionID):])
		if timestamp >= startTime && timestamp <= endTime {
			u := new(DoDUsage)
			_, err := u.UnmarshalMsg(value)
			if err != nil {
				return err
			}

			usage += u.Usage
		}

		return nil
	}); err != nil {
		return nil, err
	}

	invoice.Usage = usage
	invoice.Price = conn.Price
	invoice.Currency = conn.Currency
	invoice.TotalPrice = usage * conn.Price

	return invoice, nil
}

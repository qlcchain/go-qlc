/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"fmt"
	"math/big"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	units = map[string]*big.Int{
		"qlc":  big.NewInt(1),
		"Kqlc": big.NewInt(1e5),
		"QLC":  big.NewInt(1e8),
		"MQLC": big.NewInt(1e11),
	}
)

func BalanceToRaw(b types.Balance, unit string) (types.Balance, error) {
	if v, ok := units[unit]; ok {
		//v = v.Div(v, units["raw"])
		return types.Balance{Int: new(big.Int).Mul(b.Int, v)}, nil
	}
	return b, fmt.Errorf("invalid unit %s", unit)
}

func RawToBalance(b types.Balance, unit string) (types.Balance, error) {
	if v, ok := units[unit]; ok {
		//v = v.Div(v, units["raw"])
		return types.Balance{Int: new(big.Int).Div(b.Int, v)}, nil
	}
	return b, fmt.Errorf("invalid unit %s", unit)
}

func RawToBalanceFloat(b types.Balance, unit string) (*big.Float, error) {
	if v, ok := units[unit]; ok {
		b := new(big.Float).SetInt(big.NewInt(b.Int64()))
		d := new(big.Float).SetInt(v)
		r := new(big.Float).Quo(b, d)
		return r, nil
	}
	return nil, fmt.Errorf("invalid unit %s", unit)
}

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"fmt"

	"github.com/qlcchain/go-qlc/vm/compiler"
)

//TODO: implement
func ModuleToABIContract(code []byte) (ABIContract, error) {
	m, err := compiler.LoadModule(code)
	if err != nil {
		return ABIContract{}, err
	}

	fmt.Println(m)
	return ABIContract{}, nil
}

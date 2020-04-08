/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"reflect"
	"testing"
)

func TestContractor_FromABI(t *testing.T) {
	c := createContractParam.PartyA
	abi, err := c.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	c2 := &Contractor{}
	if err = c2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&c, c2) {
			t.Fatalf("invalid contractor, %v, %v", &c, c2)
		} else {
			t.Log(c.String())
		}
	}
}

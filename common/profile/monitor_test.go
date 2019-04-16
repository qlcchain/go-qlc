/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package profile

import (
	"math/big"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	defer Duration(time.Now(), "TestDuration")

	x := big.NewInt(2)
	y := big.NewInt(1)
	for one := big.NewInt(1); x.Sign() > 0; x.Sub(x, one) {
		y.Mul(y, x)
	}
}

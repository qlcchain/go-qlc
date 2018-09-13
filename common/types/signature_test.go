/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"strings"
	"testing"
)

func TestSignature(t *testing.T) {
	const s = "148AA79F002D747E4E262B0CC2F7B5FAB121C9362C8DB5906DC40B91147A57DAA827DF4321D0D8DED972C2469C72B4191E3AF9A69A67FC893462DCE19E9E7005"
	var sign Signature
	err := sign.Of(s)
	upper := strings.ToUpper(sign.String())
	if err != nil || upper != s {
		t.Errorf("sign missmatch. expect: %s but: %s", s, upper)
	}
}

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import "testing"

func TestAddressToken_String(t *testing.T) {
	at := &AddressToken{}
	t.Log(at.String())
}

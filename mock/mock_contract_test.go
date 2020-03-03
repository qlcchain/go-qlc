/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package mock

import (
	"testing"
)

func TestContractBlocks(t *testing.T) {
	if blocks := ContractBlocks(); len(blocks) == 0 {
		t.Fatal()
	}
}

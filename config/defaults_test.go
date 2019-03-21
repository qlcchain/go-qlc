/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"testing"
)

func Test_identityConfig(t *testing.T) {
	pk, id, err := identityConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(pk) == 0 {
		t.Fatal("invalid pk")
	}

	if len(id) == 0 {
		t.Fatal("invalid pk ")
	}
}

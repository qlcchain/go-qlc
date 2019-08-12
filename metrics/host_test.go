/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package metrics

import (
	"github.com/qlcchain/go-qlc/common/util"
	"testing"
)

func TestHost(t *testing.T) {
	h, err := Host()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(h))
}

func TestUser(t *testing.T) {
	users, err := Users()
	if err != nil {
		t.Fatal(err)
	}

	for idx, user := range users {
		t.Log(idx, util.ToIndentString(user))
	}
}

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

func TestHash(t *testing.T) {
	var hash Hash
	s := "2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"
	err := hash.Of(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	upper := strings.ToUpper(hash.String())
	if upper != s {
		t.Errorf("expect:%s but %s", s, upper)
	}
}

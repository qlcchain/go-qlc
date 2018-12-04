/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import "testing"

func TestQlcDir(t *testing.T) {
	dir := QlcDir("test", "badger")
	t.Logf("%s", dir)
}

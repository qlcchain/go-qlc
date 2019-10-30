/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package version

import (
	"testing"
	"time"
)

func TestVersionString(t *testing.T) {
	t.Log(VersionString())
}

func TestShortVersion1(t *testing.T) {
	Mode = "TestNet"
	BuildTime = time.Now().Format("2006-01-02 15:04:05-0700")
	GitRev = "c095c91ad2df"
	Version = "999"
	t.Log(ShortVersion())
}

func TestShortVersion2(t *testing.T) {
	t.Log(ShortVersion())
}

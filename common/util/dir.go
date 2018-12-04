/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import (
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

const homeDir = "~/.qlcchain/"

//QlcDir calculate qlcchain folder
func QlcDir(path ...string) string {
	h, _ := homedir.Expand(homeDir)
	t := []string{h}
	t = append(t, path...)

	return filepath.Join(t...)
}

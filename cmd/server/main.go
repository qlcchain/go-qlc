/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"os"

	cmd "github.com/qlcchain/go-qlc/cmd/server/commands"
)

func main() {
	cmd.Execute(os.Args)
}

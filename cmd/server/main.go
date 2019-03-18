/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"fmt"
	"github.com/qlcchain/go-qlc"
	"os"
	"strconv"

	cmd "github.com/qlcchain/go-qlc/cmd/server/commands"
)

func main() {
	fmt.Println("main net: ", strconv.FormatBool(goqlc.MAINNET))
	cmd.Execute(os.Args)
}

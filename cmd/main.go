/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"os"

	client "github.com/qlcchain/go-qlc/cmd/client/commands"
	server "github.com/qlcchain/go-qlc/cmd/server/commands"
)

func main() {
	args := os.Args
	if len(args) > 1 && (args[1] == "-i" || args[1] == "--endpoint") {
		client.Execute(os.Args)
	} else {
		server.Execute(os.Args)
	}
}

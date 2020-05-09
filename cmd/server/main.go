/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import "github.com/qlcchain/go-qlc/rpc/grpc/server"

func main() {
	server.Listen()
	server.Listen2()
	select {}
	//cmd.Execute(os.Args)
}

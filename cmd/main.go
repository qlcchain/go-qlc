/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"fmt"
	"github.com/qlcchain/go-qlc/chain/version"
	"os"

	client "github.com/qlcchain/go-qlc/cmd/client/commands"
	server "github.com/qlcchain/go-qlc/cmd/server/commands"
)

func main() {
	fmt.Println(`          $$\                     $$\                 $$\           
          $$ |                    $$ |                \__|          
 $$$$$$\  $$ | $$$$$$$\  $$$$$$$\ $$$$$$$\   $$$$$$\  $$\ $$$$$$$\  
$$  __$$\ $$ |$$  _____|$$  _____|$$  __$$\  \____$$\ $$ |$$  __$$\ 
$$ /  $$ |$$ |$$ /      $$ /      $$ |  $$ | $$$$$$$ |$$ |$$ |  $$ |
$$ |  $$ |$$ |$$ |      $$ |      $$ |  $$ |$$  __$$ |$$ |$$ |  $$ |
\$$$$$$$ |$$ |\$$$$$$$\ \$$$$$$$\ $$ |  $$ |\$$$$$$$ |$$ |$$ |  $$ |
 \____$$ |\__| \_______| \_______|\__|  \__| \_______|\__|\__|  \__|
      $$ |                                                          
      $$ |                                                          
      \__|                                                          `)
	fmt.Println(version.ShortVersion())
	args := os.Args
	if len(args) > 1 && (args[1] == "-i" || args[1] == "--endpoint") {
		client.Execute(os.Args)
	} else {
		server.Execute(os.Args)
	}
}

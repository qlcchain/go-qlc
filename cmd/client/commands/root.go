/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
)

type Flag struct {
	Name  string
	Usage string
	Must  bool
	Value interface{}
}

var shell = ishell.NewWithConfig(&readline.Config{
	Prompt:      fmt.Sprintf("%c[1;0;32m%s%c[0m", 0x1B, ">> ", 0x1B),
	HistoryFile: "/tmp/readline.tmp",
	//AutoComplete:      completer,
	InterruptPrompt:   "^C",
	EOFPrompt:         "exit",
	HistorySearchFold: true,
	//FuncFilterInputRune: filterInput,
})

var (
	endpointP  string
	endpoint   Flag
	commonFlag []Flag
)

// set global variable
func init() {
	endpointP = "ws://0.0.0.0:9736"
	endpoint = Flag{
		Name:  "endpoint",
		Must:  false,
		Usage: "endpoint for client to connect to server",
		Value: endpointP,
	}
}

func Execute() {
	shell.Println("QLC Chain Client")

	//set common variable
	commonFlag = make([]Flag, 0)
	// commonFlag = append(commonFlag, p)

	// run shell
	shell.Run()
}

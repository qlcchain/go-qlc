/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"flag"
	"fmt"
	"github.com/qlcchain/go-qlc/config"
	"os"
)

func main() {
	var h bool
	var file string
	flag.BoolVar(&h, "h", false, "print help message")
	flag.StringVar(&file, "config", "", "config file")
	//flag.StringVar(&file, "c", "", "config file(shorthand)")
	flag.Parse()

	if len(file) == 0 {
		file = config.DefaultConfigFile()
	}
	fmt.Println("cfg file: ", file)

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stdout, `gqlc version: gqlc/1.10.0
Usage: gqlc [-config filename] [-c filename]

Options:
`)
		flag.PrintDefaults()
	}

	if h {
		flag.Usage()
	}
}

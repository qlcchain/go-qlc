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

var (
	version   string
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
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
		_, _ = fmt.Fprintf(os.Stdout, `gqlc version: %s-%s.%s
Usage: gqlc [-config filename] [-h]

Options:
`, version, sha1ver, buildTime)
		flag.PrintDefaults()
	}

	if h {
		flag.Usage()
	}
}

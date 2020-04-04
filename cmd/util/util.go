/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
	"github.com/sahilm/fuzzy"
)

var prefix = "--"

type Flag struct {
	Name  string
	Usage string
	Must  bool
	Value interface{}
}

func sliceIndex(array []string, s string) int {
	for i, a := range array {
		if a == s {
			return i
		}
	}
	return -1
}

func StringVar(rawArgs []string, args Flag) string {
	i := sliceIndex(rawArgs, prefix+args.Name)
	if i < 0 || i+1 >= len(rawArgs) {
		return args.Value.(string)
	}
	s := rawArgs[i+1]
	return s
}

func StringSliceVar(rawArgs []string, args Flag) []string {
	i := sliceIndex(rawArgs, prefix+args.Name)
	if i < 0 || i+1 >= len(rawArgs) {
		s, ok := args.Value.(string)
		if !ok {
			return nil
		}
		if len(s) <= 0 {
			return nil
		}
		return strings.Split(s, ",")
	}
	s := rawArgs[i+1]
	return strings.Split(s, ",")
}

func IntVar(rawArgs []string, args Flag) (int, error) {
	i := sliceIndex(rawArgs, prefix+args.Name)
	if i < 0 || i+1 >= len(rawArgs) {
		return args.Value.(int), nil
	}
	s := rawArgs[i+1]
	c, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}
	return c, nil
}

func BoolVar(rawArgs []string, args Flag) bool {
	i := sliceIndex(rawArgs, prefix+args.Name)
	if i < 0 || i+1 >= len(rawArgs) {
		return args.Value.(bool)
	}

	s := rawArgs[i+1]
	c, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}

	return c
}

func CheckArgs(c *ishell.Context, args []Flag) error {
	// help
	rawArgs := c.Args
	argsMap := make(map[string]interface{})

	//args = append(args, commonFlag...)
	for _, a := range args {
		argsMap[prefix+a.Name] = a.Value
		i := sliceIndex(rawArgs, prefix+a.Name)
		if i < 0 && a.Must {
			return fmt.Errorf("lack parameter: %s", a.Name)
		}
		if i >= 0 {
			if a.Value == true || a.Value == false {
				if len(rawArgs) <= i {
					return fmt.Errorf("lack value for parameter: %s", a.Name)
				}
			} else {
				if len(rawArgs) <= i+1 {
					return fmt.Errorf("lack value for parameter: %s", a.Name)
				}
			}
		}
	}

	index := 0
	for index < len(rawArgs) {
		if _, ok := argsMap[rawArgs[index]]; ok {
			if argsMap[rawArgs[index]] == true || argsMap[rawArgs[index]] == false {
				index = index + 1
			} else {
				index = index + 2
			}
		} else {
			return fmt.Errorf("'%s' is not a qlcc command. Please see 'COMMAND --help'", rawArgs[index])
		}
	}

	return nil
}

func HelpText(c *ishell.Context, args []Flag) bool {
	rawArgs := c.Args
	if len(rawArgs) == 1 && rawArgs[0] == prefix+"help" {
		fmt.Println(c.Cmd.HelpText())
		if len(args) == 0 {
			return true
		}
		fmt.Println("args:")
		for _, a := range args {
			var t string
			if a.Must {
				t = fmt.Sprintf("    %s%-12s %s", prefix, a.Name, a.Usage)
			} else {
				t = fmt.Sprintf("    %s%-12s (Optional) %s", prefix, a.Name, a.Usage)
			}
			fmt.Println(t)
		}
		fmt.Println()
		return true
	}
	return false
}

func infoPrefix() {
	color.Set(color.FgMagenta, color.Bold)
	defer color.Unset()
	fmt.Print("==> ")
}

func Info(a ...interface{}) {
	infoPrefix()
	fmt.Println(a...)
}

func warnPrefix() {
	color.Set(color.FgMagenta, color.Bold)
	defer color.Unset()
	fmt.Print("warn ")
}

func Warn(a ...interface{}) {
	warnPrefix()
	fmt.Println(a...)
}

// OptsCompleter creates completer function for ishell from list of opts
func OptsCompleter(args []Flag) func(prefix string, args []string) []string {
	opts := make([]string, 0)
	for _, a := range args {
		opts = append(opts, prefix+a.Name)
	}
	opts = append(opts, prefix+"help")

	return func(prefix string, args []string) []string {
		if prefix == "" {
			return opts
		}

		var completion []string

		matches := fuzzy.Find(prefix, opts)
		for _, match := range matches {
			completion = append(completion, match.Str)
		}

		return completion

	}
}

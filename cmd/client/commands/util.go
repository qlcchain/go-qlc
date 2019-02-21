package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
)

var prefix = "--"

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
	if i < 0 {
		return args.Value.(string)
	}
	s := rawArgs[i+1]
	return s
}

func StringSliceVar(rawArgs []string, args Flag) []string {
	i := sliceIndex(rawArgs, prefix+args.Name)
	if i < 0 {
		return args.Value.([]string)
	}
	s := rawArgs[i+1]
	return strings.Split(s, ",")
}

func IntVar(rawArgs []string, args Flag) (int, error) {
	i := sliceIndex(rawArgs, prefix+args.Name)
	if i < 0 {
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
	if i < 0 {
		return args.Value.(bool)
	}
	return true
}

func CheckArgs(c *ishell.Context, args []Flag) error {
	// help
	rawArgs := c.Args
	argsMap := make(map[string]interface{})

	args = append(args, commonFlag...)
	for _, a := range args {
		argsMap[prefix+a.Name] = a.Value
		i := sliceIndex(rawArgs, prefix+a.Name)
		if i < 0 && a.Must {
			return errors.New(fmt.Sprintf("lack parameter: %s", a.Name))
		}
		if i >= 0 {
			if a.Value == true || a.Value == false {
				if len(rawArgs) <= i {
					return errors.New(fmt.Sprintf("lack value for parameter: %s", a.Name))
				}
			} else {
				if len(rawArgs) <= i+1 {
					return errors.New(fmt.Sprintf("lack value for parameter: %s", a.Name))
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
			return errors.New(fmt.Sprintf("'%s' is not a qlcc command. Please see 'COMMAND --help'", rawArgs[index]))
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

//
//func blackPrefix() {
//	shell.Printf("    ")
//}
//
//func black(a ...interface{}) {
//	blackPrefix()
//	shell.Println(a...)
//}

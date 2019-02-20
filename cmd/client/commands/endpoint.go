package commands

import "github.com/abiosoft/ishell"

func init() {
	c := &ishell.Cmd{
		Name: "endpoint",
		Help: "set endpoint",
		Func: func(c *ishell.Context) {
			if HelpText(c, nil) {
				return
			}
			c.Print("endpoint: ")
			endpointP = c.ReadLine()
			endpoint.Value = endpointP
		},
	}
	shell.AddCmd(c)
}

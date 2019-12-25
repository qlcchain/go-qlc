package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addPublisherCmd() {
	if interactive {
		PublisherCmd := &ishell.Cmd{
			Name: "publisher",
			Help: "publisher commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(PublisherCmd)

		addPublishCmdByShell(PublisherCmd)
		addUnPublishCmdByShell(PublisherCmd)
	} else {
		var PublisherCmd = &cobra.Command{
			Use:   "publisher",
			Short: "publisher commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(PublisherCmd)
	}
}
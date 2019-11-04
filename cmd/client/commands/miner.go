package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addMinerCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "miner",
			Help: "miner contract commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addMinerRewardCmdByShell(cmd)
		addMinerRecvPendCmdByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "miner",
			Short: "miner contract commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)

		addMinerRewardCmdByCobra(cmd)
		addMinerRecvPendCmdByCobra(cmd)
	}
}

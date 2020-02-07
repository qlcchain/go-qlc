package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

// public key distribution cmd
func addPKDCmd() {
	if interactive {
		DPKICmd := &ishell.Cmd{
			Name: "dpki",
			Help: "decentralized public key infrastructure commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(DPKICmd)

		addVerifierRegisterCmdByShell(DPKICmd)
		addVerifierUnRegisterCmdByShell(DPKICmd)
		addPublishCmdByShell(DPKICmd)
		addUnPublishCmdByShell(DPKICmd)
		addOraclePublishCmdByShell(DPKICmd)
		addDpkiRewardInfoCmdByShell(DPKICmd)
		addDpkiRewardCmdByShell(DPKICmd)
		addDpkiGetVerifierStateListCmdByShell(DPKICmd)
		addDpkiGetPublishInfoCmdByShell(DPKICmd)
	} else {
		var DPKICmd = &cobra.Command{
			Use:   "dpki",
			Short: "decentralized public key infrastructure commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(DPKICmd)
	}
}

package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/common/types"
)

func addPovCmd() {
	if interactive {
		povCmd := &ishell.Cmd{
			Name: "pov",
			Help: "pov commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(povCmd)

		addPovMiningInfoCmdByShell(povCmd)
		addPovBlockInfoCmdByShell(povCmd)
		addPovBlockListCmdByShell(povCmd)
		addPovMinerInfoCmdByShell(povCmd)
		addPovAccountInfoCmdByShell(povCmd)
		addPovLastNHourInfoCmdByShell(povCmd)
		addPovTxInfoCmdByShell(povCmd)
		addPovRepInfoCmdByShell(povCmd)
		addPovRepStateListCmdByShell(povCmd)
	} else {
		var povCmd = &cobra.Command{
			Use:   "pov",
			Short: "pov commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(povCmd)

		addPovMiningInfoCmdByCobra(povCmd)
	}
}

func formatPovReward(amount types.Balance) string {
	n := float64(amount.Uint64())
	if n >= 1e8 {
		return fmt.Sprintf("%.2f", n/1e8)
	}
	return fmt.Sprintf("%.8f", n/1e8)
}

func formatPovDifficulty(diff float64) string {
	divisor := 1.0
	unit := ""

	if diff >= 1000000000000 {
		unit = "T"
		divisor = 1000000000000
	} else if diff >= 1000000000 {
		unit = "G"
		divisor = 1000000000
	} else if diff >= 1000000 {
		unit = "M"
		divisor = 1000000
	} else if diff >= 1000 {
		unit = "K"
		divisor = 1000
	}

	return fmt.Sprintf("%.2f%s", diff/divisor, unit)
}

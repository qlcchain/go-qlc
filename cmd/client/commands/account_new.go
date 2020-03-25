/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
)

func addGenerateAccountCmdByShell(parentCmd *ishell.Cmd) {
	count := util.Flag{
		Name:  "count",
		Must:  false,
		Usage: "account count",
		Value: 10,
	}
	seed := util.Flag{
		Name:  "seed",
		Must:  false,
		Usage: "account seed",
		Value: "",
	}
	args := []util.Flag{count, seed}
	c := &ishell.Cmd{
		Name:                "new",
		Help:                "generate account",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}
			countP, err := util.IntVar(c.Args, count)
			if err != nil {
				util.Warn(err)
				return
			}
			seedP := util.StringVar(c.Args, seed)
			if err := accountAction(countP, seedP, newShellPrinter()); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(c)
}

func addGenerateAccountCmdByCobra(parentCmd *cobra.Command) {
	var countP int
	var seedP string
	var accountCmd = &cobra.Command{
		Use:   "new",
		Short: "generate account",
		Run: func(cmd *cobra.Command, args []string) {
			err := accountAction(countP, seedP, newCmdPrinter(cmd))
			if err != nil {
				cmd.Println(err)
			}
		},
	}
	accountCmd.Flags().IntVar(&countP, "count", 10, "account count")
	accountCmd.Flags().StringVar(&seedP, "seed", "", "account seed")
	parentCmd.AddCommand(accountCmd)
}

func accountAction(countP int, seedP string, log printer) error {
	var s *types.Seed
	if len(seedP) > 0 {
		bytes, err := hex.DecodeString(seedP)
		if err != nil {
			return err
		}
		s, err = types.BytesToSeed(bytes)
		if err != nil {
			return err
		}
	} else {
		s, _ = types.NewSeed()
	}

	log.Info(fmt.Sprintf("%d accounts created from %s", countP, s.String()))
	for i := 0; i < countP; i++ {
		if a, err := s.Account(uint32(i)); err == nil {
			log.Info("Index:", i)
			log.Info("Address:", a.Address())
			log.Info("Private:", hex.EncodeToString(a.PrivateKey()))
		}
	}

	return nil
}

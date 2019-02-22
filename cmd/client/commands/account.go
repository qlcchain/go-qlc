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
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/spf13/cobra"
)

func account() {
	var countP int
	var seedP string

	if interactive {
		count := Flag{
			Name:  "count",
			Must:  false,
			Usage: "account count",
			Value: 10,
		}
		seed := Flag{
			Name:  "seed",
			Must:  false,
			Usage: "account seed",
			Value: "",
		}

		s := &ishell.Cmd{
			Name: "account",
			Help: "generate account",
			Func: func(c *ishell.Context) {
				args := []Flag{count, seed}
				if HelpText(c, args) {
					return
				}
				err := CheckArgs(c, args)
				if err != nil {
					Warn(err)
					return
				}
				countP, err = IntVar(c.Args, count)
				if err != nil {
					Warn(err)
					return
				}
				seedP = StringVar(c.Args, seed)
				if err := accountAction(countP, seedP); err != nil {
					Warn(err)
					return
				}
			},
		}
		shell.AddCmd(s)
	} else {
		var accountCmd = &cobra.Command{
			Use:   "account",
			Short: "generate account",
			Run: func(cmd *cobra.Command, args []string) {
				err := accountAction(countP, seedP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		accountCmd.Flags().IntVar(&countP, "count", 10, "account count")
		accountCmd.Flags().StringVar(&seedP, "seed", "", "account seed")
		rootCmd.AddCommand(accountCmd)
	}
}

func accountAction(countP int, seedP string) error {
	if len(seedP) > 0 {
		bytes, err := hex.DecodeString(seedP)
		if err != nil {
			return err
		}
		s, err := types.BytesToSeed(bytes)
		if err != nil {
			return err
		}
		a, err := s.Account(0)
		if err != nil {
			return err
		}
		if interactive {
			Info("account created:")
		}
		fmt.Println("Seed:", s.String())
		fmt.Println("Address:", a.Address())
		fmt.Println("Private:", hex.EncodeToString(a.PrivateKey()))
	} else {
		if interactive {
			Info(fmt.Sprintf("%d accounts created:", countP))
		}
		for i := 0; i < countP; i++ {
			seed, err := types.NewSeed()
			if err == nil {
				if a, err := seed.Account(0); err == nil {
					fmt.Println("Seed:", seed.String())
					fmt.Println("Address:", a.Address())
					fmt.Println("Private:", hex.EncodeToString(a.PrivateKey()))
				}
			}
		}
	}
	return nil
}

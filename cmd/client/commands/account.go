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
)

func init() {
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
			if err := CheckArgs(c, args); err != nil {
				Warn(err)
				return
			}
			countP, err := IntVar(c.Args, count)
			if err != nil {
				Warn(err)
				return
			}
			seedP := StringVar(c.Args, seed)
			if err := accountAction(countP, seedP); err != nil {
				Warn(err)
				return
			}
		},
	}
	shell.AddCmd(s)
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
		Info("account created:")
		fmt.Println("Seed:", s.String())
		fmt.Println("Address:", a.Address())
		fmt.Println("Private:", hex.EncodeToString(a.PrivateKey()))
	} else {
		Info(fmt.Sprintf("%d accounts created:", countP))
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

/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package db

import (
	"fmt"
	"sort"
	"testing"
)

type m1 struct {
}

func (m1) Migrate(txn StoreTxn) error {
	panic("implement me")
}

func (m1) StartVersion() int {
	return 1
}

func (m1) EndVersion() int {
	return 2
}

type m2 struct {
}

func (m2) Migrate(txn StoreTxn) error {
	panic("implement me")
}

func (m2) StartVersion() int {
	return 1
}

func (m2) EndVersion() int {
	return 3
}

type m3 struct {
}

func (m3) Migrate(txn StoreTxn) error {
	panic("implement me")
}

func (m3) StartVersion() int {
	return 3
}

func (m3) EndVersion() int {
	return 4
}

func TestSortMigration(t *testing.T) {
	m := []Migration{new(m3), new(m1), new(m2)}
	for _, mm := range m {
		fmt.Printf("%d->%d\n", mm.StartVersion(), mm.EndVersion())
	}
	fmt.Println("-------------")
	sort.Sort(Migrations(m))
	for _, mm := range m {
		fmt.Printf("%d->%d\n", mm.StartVersion(), mm.EndVersion())
	}
}

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package dpos

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
	"testing"
)

func TestTransaction(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		for _, n := range nodes {
			n.StopNodeAndRemoveDir()
		}
	}()

	for _, n := range nodes {
		n.RunNode()
		n.InitLedger()
	}

	n1 := nodes[0]
	n2 := nodes[1]

	n1.InitStatus()
	n2.InitStatus()

	toAcc := mock.Account()
	send := n1.GenerateSendBlock(testAccount, toAcc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockAndWaitConfirmed(send)
	n2.WaitBlockConfirmed(send.GetHash())
}

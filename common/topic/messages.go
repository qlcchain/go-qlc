/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package topic

import "github.com/qlcchain/go-qlc/common/types"

type EventPovRecvBlockMsg struct {
	Block   *types.PovBlock
	From    types.PovBlockFrom
	MsgPeer string
}

type EventRPCSyncCallMsg struct {
	Name string
	In   interface{}
	Out  interface{}
}

type EventPublishMsg struct {
	Block *types.StateBlock
	From  string
}

type EventConfirmReqMsg struct {
	Blocks []*types.StateBlock
	From   string
}

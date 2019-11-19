/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package p2p

import (
	"github.com/qlcchain/go-qlc/p2p/protos"
)

type EventSendMsgToSingleMsg struct {
	Type    MessageType
	Message interface{}
	PeerID  string
}

type EventAddP2PStreamMsg struct {
	PeerID string
}

type EventDeleteP2PStreamMsg struct {
	PeerID string
}

type EventPovPeerStatusMsg struct {
	Status *protos.PovStatus
	From   string
}

type EventPovBulkPullReqMsg struct {
	Req  *protos.PovBulkPullReq
	From string
}

type EventPovBulkPullRspMsg struct {
	Resp *protos.PovBulkPullRsp
	From string
}

type EventConfirmAckMsg struct {
	Block *protos.ConfirmAckBlock
	From  string
}

type EventBroadcastMsg struct {
	Type    MessageType
	Message interface{}
}

type EventFrontiersReqMsg struct {
	PeerID string
}

// +build  testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package p2p

import "time"

const (
	QlcProtocolID        = "qlc-testnet/1.0.0"
	QlcProtocolFOUND     = "/qlc-testnet/discovery/1.0.0"
	QlcMDnsFOUND         = "/qlc-testnet/MDns/1.0.0"
	discoveryConnTimeout = time.Second * 30
)

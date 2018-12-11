/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"github.com/qlcchain/go-qlc/config"
)

type QlcContext struct {
	config *config.Config
}

func New(cfg *config.Config) (*QlcContext, error) {
	return &QlcContext{config: cfg}, nil
}

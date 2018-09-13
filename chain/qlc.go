/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"fmt"

	"github.com/qlcchain/go-qlc/config"
)

type QlcContext struct {
	config *config.Config
}

func New(cfg *config.Config) (*QlcContext, error) {
	fmt.Println(cfg)
	return &QlcContext{}, nil
}

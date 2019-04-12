/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package profile

import (
	"time"

	"github.com/qlcchain/go-qlc/log"
)

var (
	logger = log.NewLogger("profile")
)

func Duration(invocation time.Time, name string) {
	elapsed := time.Since(invocation)
	logger.Debugf("%s cost %s", name, elapsed)
}

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package monitor

import "github.com/shirou/gopsutil/host"

func Host() (*host.InfoStat, error) {
	return host.Info()
}

func Users() ([]host.UserStat, error) {
	return host.Users()
}

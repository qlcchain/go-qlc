/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package metrics

import "github.com/shirou/gopsutil/host"

func Host() (*host.InfoStat, error) {
	return host.Info()
}

func Users() ([]host.UserStat, error) {
	return host.Users()
}

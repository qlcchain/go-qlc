/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"fmt"
	"time"
)

type timeSpan struct {
	years   int
	months  int
	days    int
	hours   int
	minutes int
	seconds int
}

func (ts *timeSpan) Calculate(t time.Time) time.Time {
	if duration, e := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", ts.hours, ts.minutes, ts.seconds)); e == nil {
		return t.UTC().AddDate(ts.years, ts.months, ts.days).Add(duration)
	} else {
		//fmt.Println(e)
	}
	return t.UTC().AddDate(ts.years, ts.months, ts.days)
}

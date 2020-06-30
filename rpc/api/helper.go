/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import "fmt"

func calculateRange(size, count int, offset *int) (start, end int, err error) {
	if size <= 0 {
		return 0, 0, fmt.Errorf("can not find any records, size=%d", size)
	}

	o := 0

	if count <= 0 {
		return 0, 0, fmt.Errorf("invalid count: %d", count)
	}

	if offset != nil && *offset >= 0 {
		o = *offset
	}

	if o >= size {
		return 0, 0, fmt.Errorf("overflow, max:%d,offset:%d", size, o)
	} else {
		start = o
		end = start + count
		if end > size {
			end = size
		}
		return
	}
}

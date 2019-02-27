/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

// SafeMul returns multiplication result and whether overflow occurred.
func SafeMul(x, y uint64) (uint64, bool) {
	if x == 0 || y == 0 {
		return 0, false
	}
	return x * y, y > MaxUint64/x
}

// SafeAdd returns the result and whether overflow occurred.
func SafeAdd(x, y uint64) (uint64, bool) {
	return x + y, y > MaxUint64-x
}

// SafeSub returns subtraction result and whether overflow occurred.
func SafeSub(x, y uint64) (uint64, bool) {
	return x - y, x < y
}

func Min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

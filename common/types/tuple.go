/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

type Tuple struct {
	First, Second interface{}
}

func NewTuple(first, second interface{}) *Tuple {
	return &Tuple{
		First:  first,
		Second: second,
	}
}

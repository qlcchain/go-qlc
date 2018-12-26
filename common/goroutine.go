/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"fmt"
	"github.com/pkg/errors"
)

func Go(fn func()) {
	go wrap(fn)
}

func wrap(fn func()) {
	defer catch()
	fn()
}

func catch() {
	if err := recover(); err != nil {
		var e error
		switch t := err.(type) {
		case error:
			e = errors.WithStack(t)
		case string:
			e = errors.New(t)
		default:
			e = errors.Errorf("unknown type: %v", err)
		}
		fmt.Println(e)
		panic(err)
	}
}

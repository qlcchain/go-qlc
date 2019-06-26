package common

import (
	"fmt"
	"log"

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
			e = errors.New("unknown type")
		}

		log.Println("panic", "err", err, "withstack", e)
		fmt.Printf("%+v", e)
		panic(err)
	}
}

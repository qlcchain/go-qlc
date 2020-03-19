package common

import (
	"errors"
	"log"
)

func Go(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				var e error
				switch t := err.(type) {
				case error:
					e = t
				case string:
					e = errors.New(t)
				default:
					e = errors.New("unknown type")
				}

				log.Printf("panic %s, %+v", err, e)
				//fmt.Printf("%+v", e)
				//panic(err)
			}
		}()
		fn()
	}()
}

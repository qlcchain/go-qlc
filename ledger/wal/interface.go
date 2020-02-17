package wal

type Location struct {
	Data   interface{}
	Type   byte
	Length int
}

const (
	block_wal byte = iota
)

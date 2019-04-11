package db

import "io"

type TableName byte

const (
	TableBlockHash TableName = iota
	TableBlockMessage
)

const LikeSign = "_like_"

func (t TableName) String() string {
	switch t {
	case TableBlockHash:
		return "blockhash"
	case TableBlockMessage:
		return "blockmessage"
	}
	return ""
}

type Column byte

const (
	ColumnHash Column = iota
	ColumnType
	ColumnAddress
	ColumnSender
	ColumnReceiver
	ColumnMessage
	ColumnTimestamp
	ColumnNoNeed
)

func (c Column) String() string {
	switch c {
	case ColumnHash:
		return "hash"
	case ColumnType:
		return "type"
	case ColumnAddress:
		return "address"
	case ColumnSender:
		return "sender"
	case ColumnReceiver:
		return "receiver"
	case ColumnMessage:
		return "message"
	case ColumnTimestamp:
		return "timestamp"
	}

	return ""
}

type DbStore interface {
	io.Closer
	Create(table TableName, condition map[Column]interface{}) error
	Read(table TableName, condition map[Column]interface{}, offset int, limit int, order Column, dest interface{}) error
	Update(table TableName, condition map[Column]interface{}) error
	Delete(table TableName, condition map[Column]interface{}) error
	Count(table TableName, dest interface{}) error
	Group(table TableName, column Column, dest interface{}) error
}

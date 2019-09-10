package main

import (
	"github.com/qlcchain/go-qlc/common/simplejson"
)

// StratumMsg is the basic message object from stratum.
type StratumMsg struct {
	ID     interface{}      `json:"id"`
	Method string           `json:"method,omitempty"`
	Params *simplejson.Json `json:"params,omitempty"`
	Error  *simplejson.Json `json:"error,omitempty"`
	Result *simplejson.Json `json:"result,omitempty"`

	RawMsg  []byte           `json:"-"`
	JsonMsg *simplejson.Json `json:"-"`
}

type StratumRsp struct {
	SessionID uint32
	MsgID     interface{}
	Result    interface{}
	ErrCode   int
	ErrMsg    string
}

// StratumNotify models the json from a mining.notify message.
type StratumNotify struct {
	JobID        string   // HEX(bytes)
	PrevHash     string   // HEX(LittleEndian), each uint32 is BigEndian
	Coinbase1    string   // HEX(bytes)
	Coinbase2    string   // HEX(bytes)
	MerkleBranch []string // HEX(LittleEndian)
	Version      string   // HEX(uint32, BE)
	NBits        string   // HEX(uint32, BE)
	NTime        string   // HEX(uint32, BE)
	CleanJobs    bool

	// follow fields are not exist in wire message
	Difficulty    float64
	CoinbaseExtra []byte
}

// StratumSubmit models a mining.submit message.
type StratumSubmit struct {
	SessionID uint32
	MsgID     interface{}

	WorkerName     string
	JobID          string // HEX(bytes)
	ExtraNonce2Hex string // HEX(uint64, BE)
	NTimeHex       string // HEX(uint32, BE)
	NonceHex       string // HEX(uint32, BE)

	// fields are used internal
	ExtraNonce2 uint64
	NTime       uint32
	Nonce       uint32

	// fields are not exist in wire message
	ExtraNonce1 uint32
}

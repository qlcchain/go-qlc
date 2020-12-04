package storage

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

type Store interface {
	Delete(k []byte) error
	Get([]byte) ([]byte, error)
	Put(k, v []byte) error
	Has(k []byte) (bool, error)
	Batch(bool) Batch
	PutBatch(Batch) error
	BatchWrite(bool, func(batch Batch) error) error
	//BatchView(func(batch Batch) error) error
	Iterator(prefix []byte, end []byte, f func(k, v []byte) error) error
	Count(prefix []byte) (uint64, error)
	Purge() error
	Drop(prefix []byte) error
	Upgrade(version int) error
	Action(at ActionType) (interface{}, error)
	Close() error
}

type Batch interface {
	Put(k []byte, v interface{}) error
	Delete(k []byte) error
	Get([]byte) (interface{}, error)
	Iterator(prefix []byte, end []byte, f func(k, v []byte) error) error
	Drop(prefix []byte) error
	Discard()
}

type KeyPrefix byte

const (
	KeyPrefixBlock KeyPrefix = iota
	KeyPrefixSmartContractBlock
	KeyPrefixUncheckedBlockPrevious
	KeyPrefixUncheckedBlockLink
	KeyPrefixAccount
	//KeyPrefixToken
	KeyPrefixFrontier
	KeyPrefixPending
	KeyPrefixRepresentation
	KeyPrefixPerformance
	KeyPrefixChild
	KeyPrefixVersion //10
	KeyPrefixStorage
	KeyPrefixToken    //discard
	KeyPrefixSender   //discard
	KeyPrefixReceiver //discard
	KeyPrefixMessage  //discard
	KeyPrefixMessageInfo
	KeyPrefixOnlineReps
	KeyPrefixPovHeader   // prefix + height + hash => header
	KeyPrefixPovBody     // prefix + height + hash => body
	KeyPrefixPovHeight   // prefix + hash => height (uint64) 20
	KeyPrefixPovTxLookup // prefix + txHash => TxLookup
	KeyPrefixPovBestHash // prefix + height => hash
	KeyPrefixPovTD       // prefix + height + hash => total difficulty (big int)
	KeyPrefixLink
	KeyPrefixBlockCache //block store this table before consensus complete
	KeyPrefixRepresentationCache
	KeyPrefixUncheckedTokenInfo
	KeyPrefixBlockCacheAccount
	KeyPrefixPovMinerStat    // prefix + day index => miners of best blocks per day
	KeyPrefixUnconfirmedSync //30
	KeyPrefixUncheckedSync
	KeyPrefixSyncCacheBlock
	KeyPrefixUncheckedPovHeight
	KeyPrefixPovLatestHeight  // prefix => height
	KeyPrefixPovTxlScanCursor // prefix => height
	KeyPrefixVoteHistory
	KeyPrefixPovDiffStat // prefix + dayIndex => average diff statistics per day
	KeyPrefixPeerInfo    //prefix+peerID => peerInfo
	KeyPrefixGapPublish
	KeyPrefixDPoS
	KeyPrefixAccountBlockHash
	KeyPrefixAccountPovHeight
	KeyPrefixPrivatePayload
	KeyPrefixGapDoDSettleState
	KeyPrefixGapPovHeight

	// Trie key space should be different
	KeyPrefixTrieVMStorage = 100 // Deprecated vm_store.go, idPrefixStorage
	KeyPrefixTrie          = 101 // 101 is used for trie intermediate node, trie.go, idPrefixTrie
	KeyPrefixTriePovState  = 102
	KeyPrefixContractValue = 103 // idPrefixContractValue ledger.go
	KeyPrefixVmLogs        = 104 // vm_store.go
	KeyPrefixVMStorage     = 105 // vm_store.go
	KeyPrefixTrieClean     = 106
	KeyPrefixPendingBackup = 253
	KeyPrefixGenericType   = 254
	KeyPrefixGenericTypeC  = 255
)

func GetKeyOfParts(t KeyPrefix, partList ...interface{}) ([]byte, error) {
	var buffer = []byte{byte(t)}
	for _, part := range partList {
		var src []byte
		switch v := part.(type) {
		case int:
			src = util.BE_Uint64ToBytes(uint64(v))
		case int32:
			src = util.BE_Uint64ToBytes(uint64(v))
		case uint32:
			src = util.BE_Uint64ToBytes(uint64(v))
		case int64:
			src = util.BE_Uint64ToBytes(uint64(v))
		case uint64:
			src = util.BE_Uint64ToBytes(v)
		case []byte:
			src = v
		case types.Hash:
			src = v[:]
		case *types.Hash:
			src = v[:]
		case types.Address:
			src = v[:]
		case types.Serializer:
			var err error
			src, err = v.Serialize()
			if err != nil {
				return nil, fmt.Errorf("key serialize: %s, %s", err, v)
			}
		default:
			return nil, errors.New("key contains of invalid part")
		}
		buffer = append(buffer, src...)
	}
	return buffer, nil
}

var KeyNotFound = errors.New("key not found")

type Cache interface {
	Get(key []byte) (interface{}, error)
	Put(key []byte, value interface{}) error
	Delete(key []byte) error
	Len() int64
}

type ActionType byte

const (
	Dump ActionType = iota
	GC
	Backup
	Restore
	Size
)

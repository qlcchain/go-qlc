package common

const (
	SyncCacheUnchecked byte = iota
	SyncCacheUnconfirmed
)

type SyncCacheWalkFunc func(kind byte, key []byte)

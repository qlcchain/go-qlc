package types

const (
	//	// Trie key space should be different
	//
	//	TriePrefixVMStorage = byte(100) // vm_store.go, idPrefixStorage
	//	TriePrefixTrie      = byte(101) // 101 is used for trie node, trie.go, idPrefixTrie
	TriePrefixPovState = byte(102) // 102 is not used in db key, just for trie key
)

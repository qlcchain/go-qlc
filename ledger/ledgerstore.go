/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package ledger

type Store interface {
	AccountStore
	SmartBlockStore
	BlockStore
	PendingStore
	FrontierStore
	RepresentationStore
	UncheckedBlockStore
	PeerInfoStore
	SyncStore
	DposStore
	PovStore
	Relation
	CacheStore
	LedgerStore
}

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/p2p"

	qcontext "github.com/qlcchain/go-qlc/chain/context"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

const (
	checkBlockCacheInterval = 60 * time.Second
	maxResendBlockCache     = 3
)

type ResendBlockService struct {
	common.ServiceLifecycle
	hashSet    *sync.Map
	id         string
	cfgFile    string
	blockCache chan *types.StateBlock
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *zap.SugaredLogger
}

type ResendTimes struct {
	hash        types.Hash
	resendTimes int
}

func NewResendBlockService(cfgFile string) *ResendBlockService {
	ctx, cancel := context.WithCancel(context.Background())
	return &ResendBlockService{
		hashSet:    new(sync.Map),
		cfgFile:    cfgFile,
		blockCache: make(chan *types.StateBlock, 100),
		ctx:        ctx,
		cancel:     cancel,
		logger:     log.NewLogger("resend_block_service"),
	}
}

func (rb *ResendBlockService) Init() error {
	if !rb.PreInit() {
		return errors.New("pre init fail")
	}
	defer rb.PostInit()
	cc := qcontext.NewChainContext(rb.cfgFile)
	rb.id, _ = cc.EventBus().Subscribe(common.EventAddBlockCache, func(blk *types.StateBlock) {
		if blk != nil {
			rb.blockCache <- blk
		}
	})
	return nil
}

func (rb *ResendBlockService) Start() error {
	if !rb.PreStart() {
		return errors.New("pre start fail")
	}
	defer rb.PostStart()

	cc := qcontext.NewChainContext(rb.cfgFile)
	ledgerService, _ := cc.Service(qcontext.LedgerService)
	for {
		//ledger service started
		if ledgerService != nil && ledgerService.Status() == int32(common.Started) {
			break
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
	l := ledgerService.(*LedgerService).Ledger
	verifier := process.NewLedgerVerifier(l)
	blocks := make([]*types.StateBlock, 0)
	accountMetaCaches := make([]*types.AccountMeta, 0)
	err := l.GetBlockCaches(func(block *types.StateBlock) error {
		blocks = append(blocks, block)
		return nil
	})
	if err != nil {
		rb.logger.Error("get block cache error")
	}
	err = l.GetAccountMetaCaches(func(am *types.AccountMeta) error {
		accountMetaCaches = append(accountMetaCaches, am)
		return nil
	})
	if err != nil {
		rb.logger.Error("get accountMeta cache error")
	}
	blockCaches := make([]*types.StateBlock, 0)
	for _, am := range accountMetaCaches {
		for _, v := range am.Tokens {
			blockCaches = blockCaches[:0:0]
			header := v.Header
			for {
				sb, _ := l.GetBlockCache(header)
				if sb != nil {
					blockCaches = append(blockCaches, sb)
				} else {
					break
				}
				header = sb.Previous
			}
			for i := len(blockCaches) - 1; i >= 0; i-- {
				rt := &ResendTimes{
					hash:        blockCaches[i].GetHash(),
					resendTimes: 0,
				}
				key := rb.getKey(blockCaches[i].Address, blockCaches[i].Token)
				if v, ok := rb.hashSet.Load(key); !ok {
					var hs []*ResendTimes
					hs = append(hs, rt)
					rb.hashSet.Store(key, hs)

				} else {
					hs := v.([]*ResendTimes)
					hs = append(hs, rt)
					rb.hashSet.Store(key, hs)
				}
			}
		}
	}

	go func() {
		ticker := time.NewTicker(checkBlockCacheInterval)
		var hs []*ResendTimes
		var hsTemp []*ResendTimes
		for {
			select {
			case <-rb.ctx.Done():
				return
			case <-ticker.C:
				rb.hashSet.Range(func(key, value interface{}) bool {
					h := key.(types.Hash)
					hs = value.([]*ResendTimes)
					if hs[0].resendTimes >= maxResendBlockCache {
						rb.hashSet.Delete(h)
						_ = verifier.RollbackCache(hs[0].hash)
					} else {
						for _, j := range hs {
							if b, _ := l.HasStateBlockConfirmed(j.hash); b {
								hsTemp = append(hsTemp, j)
							} else {
								if j.resendTimes >= maxResendBlockCache {
									rb.hashSet.Delete(h)
									_ = verifier.RollbackCache(j.hash)
									break
								} else {
									if sb, err := l.GetStateBlock(j.hash); err == nil {
										cc.EventBus().Publish(common.EventBroadcast, p2p.PublishReq, sb)
										cc.EventBus().Publish(common.EventGenerateBlock, process.Progress, sb)
									}
									j.resendTimes++
									break
								}
							}

						}
						if len(hsTemp) == len(hs) {
							rb.hashSet.Delete(h)
						} else {
							hsDiff := sliceDiff(hs, hsTemp)
							rb.hashSet.Store(h, hsDiff)
						}
					}
					return true
				})
			}
		}
	}()

	go func() {
		for {
			select {
			case <-rb.ctx.Done():
				if err := cc.EventBus().Unsubscribe(common.EventAddBlockCache, rb.id); err != nil {
					rb.logger.Error(err)
				}
				return
			case blk := <-rb.blockCache:
				rt := &ResendTimes{
					hash:        blk.GetHash(),
					resendTimes: 0,
				}
				key := rb.getKey(blk.Address, blk.Token)
				if v, ok := rb.hashSet.Load(key); !ok {
					var hs []*ResendTimes
					hs = append(hs, rt)
					rb.hashSet.Store(key, hs)

				} else {
					hs := v.([]*ResendTimes)
					hs = append(hs, rt)
					rb.hashSet.Store(key, hs)
				}
			}
		}
	}()

	return nil
}

func (rb *ResendBlockService) Stop() error {
	if !rb.PreStop() {
		return errors.New("pre stop fail")
	}
	defer rb.PostStop()

	rb.cancel()

	return nil
}

func (rb *ResendBlockService) Status() int32 {
	return rb.State()
}

func (rb *ResendBlockService) getKey(addr types.Address, token types.Hash) types.Hash {
	key := make([]byte, 0)
	key = append(key, addr.Bytes()...)
	key = append(key, token.Bytes()...)
	hash, _ := types.HashBytes(key)
	return hash
}

// InSliceIface checks given interface in interface slice.
func inSliceIface(v interface{}, sl []*ResendTimes) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// SliceDiff returns diff slice of slice1 - slice2.
func sliceDiff(slice1, slice2 []*ResendTimes) (diffSlice []*ResendTimes) {
	for _, v := range slice1 {
		if !inSliceIface(v, slice2) {
			diffSlice = append(diffSlice, v)
		}
	}
	return
}

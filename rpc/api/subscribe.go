package api

import (
	"context"
	"sync"

	rpc "github.com/qlcchain/jsonrpc2"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

const (
	MaxNotifyBlocks = 100
)

func createSubscription(ctx context.Context, fn func(notifier *rpc.Notifier, subscription *rpc.Subscription)) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, rpc.ErrNotificationsUnsupported
	}

	// by explicitly creating an subscription we make sure that the subscription id is send back to the client
	// before the first subscription.Notify is called.
	subscription := notifier.CreateSubscription()

	fn(notifier, subscription)
	return subscription, nil
}

// BlockSubscription subscript event from chain, and deliver to every connected websocket
type BlockSubscription struct {
	mu         *sync.Mutex
	eb         event.EventBus
	handlerIds map[common.TopicType]string // subscript event from chain
	chans      []chan struct{}             // deliver to every channel for each connected websocket
	blocks     []*types.StateBlock
	addrBlocks map[types.Address]*types.StateBlock
	newBlockCh chan struct{}
	stoped     chan bool
	ctx        context.Context
	logger     *zap.SugaredLogger
}

func NewBlockSubscription(eb event.EventBus, ctx context.Context) *BlockSubscription {
	bs := &BlockSubscription{
		eb:         eb,
		mu:         &sync.Mutex{},
		handlerIds: make(map[common.TopicType]string),
		chans:      make([]chan struct{}, 0),
		blocks:     make([]*types.StateBlock, 0, MaxNotifyBlocks),
		addrBlocks: make(map[types.Address]*types.StateBlock),
		newBlockCh: make(chan struct{}),
		stoped:     make(chan bool),
		ctx:        ctx,
		logger:     log.NewLogger("api_sub"),
	}
	bs.subscribeEvent()
	return bs
}

func (r *BlockSubscription) subscribeEvent() {
	if id, err := r.eb.Subscribe(common.EventAddRelation, r.setBlocks); err != nil {
		r.logger.Error("subscribe event error, ", err)
	} else {
		r.handlerIds[common.EventAddRelation] = id
	}
	if id, err := r.eb.Subscribe(common.EventAddSyncBlocks, r.setSyncBlocks); err != nil {
		r.logger.Error("subscribe event error, ", err)
	} else {
		r.handlerIds[common.EventAddSyncBlocks] = id
	}
	r.logger.Info("event subscribed ")
	go func() {
		select {
		case <-r.ctx.Done():
			r.unsubscribeEvent()
		}
	}()
}

func (r *BlockSubscription) unsubscribeEvent() {
	r.logger.Info("unsubscribe event")
	for k, v := range r.handlerIds {
		if err := r.eb.Unsubscribe(k, v); err != nil {
			r.logger.Error("unsubscribe event error, ", err)
		}
	}
	r.logger.Info("event unsubscribed ")
}

func (r *BlockSubscription) setBlocks(block *types.StateBlock) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.setBlockNoLock(block)
}

func (r *BlockSubscription) setSyncBlocks(block *types.StateBlock, done bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !done {
		r.setBlockNoLock(block)
	}
}

func (r *BlockSubscription) setBlockNoLock(block *types.StateBlock) {
	if len(r.chans) > 0 {
		if _, ok := r.addrBlocks[block.GetAddress()]; ok {
			r.addrBlocks[block.GetAddress()] = block
		}

		if len(r.blocks) >= cap(r.blocks) {
			copy(r.blocks, r.blocks[1:len(r.blocks)])
			r.blocks[cap(r.blocks)-1] = block
		} else {
			r.blocks = append(r.blocks, block)
		}

		select {
		case r.newBlockCh <- struct{}{}:
		default:
		}
	}
}

func (r *BlockSubscription) getBlocks() []*types.StateBlock {
	r.mu.Lock()
	defer r.mu.Unlock()

	retBlks := make([]*types.StateBlock, len(r.blocks))
	copy(retBlks, r.blocks)
	return retBlks
}

func (r *BlockSubscription) getAddressBlock(addr types.Address) *types.StateBlock {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.addrBlocks[addr]
}

func (r *BlockSubscription) addChan(addr types.Address, ch chan struct{}) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()
	r.addrBlocks[addr] = nil
	r.chans = append(r.chans, ch)
	r.logger.Infof("add chan %p to blockEvent, chan length %d", ch, len(r.chans))
	if len(r.chans) == 1 {
		go func() {
			for {
				select {
				case <-r.newBlockCh:
					for _, c := range r.chans { // broadcast event to every websocket channel
						select {
						case c <- struct{}{}:
						default:
						}
					}
				case <-r.stoped:
					r.logger.Info("broadcast subscription stopped")
					return
				}
			}
		}()
	}
}

func (r *BlockSubscription) removeChan(ch chan struct{}) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()
	for index, c := range r.chans {
		if c == ch {
			r.logger.Infof("remove chan %p ", c)
			r.chans = append(r.chans[:index], r.chans[index+1:]...)
			if len(r.chans) == 0 {
				r.stoped <- true
			}
			break
		}
	}
	if len(r.chans) == 0 {
		r.blocks = make([]*types.StateBlock, 0, MaxNotifyBlocks)
		r.addrBlocks = make(map[types.Address]*types.StateBlock)
	}
}

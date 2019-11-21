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
	allSubs    map[rpc.ID]*BlockSubscriber
	blocksCh   chan *types.StateBlock
	ctx        context.Context
	logger     *zap.SugaredLogger
}

type BlockSubscriber struct {
	notifyCh  chan struct{}
	address   types.Address
	blocks    []*types.StateBlock
	addrBlock *types.StateBlock
	batch     bool
}

func NewBlockSubscription(ctx context.Context, eb event.EventBus) *BlockSubscription {
	bs := &BlockSubscription{
		eb:         eb,
		mu:         &sync.Mutex{},
		handlerIds: make(map[common.TopicType]string),
		allSubs:    make(map[rpc.ID]*BlockSubscriber),
		blocksCh:   make(chan *types.StateBlock, MaxNotifyBlocks),
		ctx:        ctx,
		logger:     log.NewLogger("api_sub"),
	}
	bs.subscribeEvent()
	go bs.notifyLoop()
	return bs
}

func (r *BlockSubscription) subscribeEvent() {
	if id, err := r.eb.Subscribe(common.EventAddRelation, r.setBlocks); err != nil {
		r.logger.Error("subscribe EventAddRelation error, ", err)
	} else {
		r.handlerIds[common.EventAddRelation] = id
	}
	if id, err := r.eb.Subscribe(common.EventAddSyncBlocks, r.setSyncBlocks); err != nil {
		r.logger.Error("subscribe EventAddSyncBlocks error, ", err)
	} else {
		r.handlerIds[common.EventAddSyncBlocks] = id
	}
}

func (r *BlockSubscription) unsubscribeEvent() {
	r.logger.Info("unsubscribe event")
	for k, v := range r.handlerIds {
		if err := r.eb.Unsubscribe(k, v); err != nil {
			r.logger.Error("unsubscribe event error, ", err)
		}
	}
}

func (r *BlockSubscription) setBlocks(block *types.StateBlock) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.allSubs) == 0 {
		return
	}

	select {
	case r.blocksCh <- block:
	default:
	}
}

func (r *BlockSubscription) setSyncBlocks(block *types.StateBlock, done bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if done {
		return
	}

	if len(r.allSubs) == 0 {
		return
	}

	select {
	case r.blocksCh <- block:
	default:
	}
}

func (r *BlockSubscription) fetchBlocks(subID rpc.ID) []*types.StateBlock {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub := r.allSubs[subID]
	if sub == nil {
		return nil
	}
	if len(sub.blocks) == 0 {
		return nil
	}

	retBlks := sub.blocks
	sub.blocks = make([]*types.StateBlock, 0, MaxNotifyBlocks)
	return retBlks
}

func (r *BlockSubscription) fetchAddrBlock(subID rpc.ID) *types.StateBlock {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub := r.allSubs[subID]
	if sub == nil {
		return nil
	}

	retBlk := sub.addrBlock
	sub.addrBlock = nil
	return retBlk
}

func (r *BlockSubscription) addChan(subID rpc.ID, addr types.Address, batch bool, ch chan struct{}) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()

	sub := r.allSubs[subID]
	if sub != nil {
		r.logger.Errorf("chan %d exist already", subID)
		return
	}

	sub = new(BlockSubscriber)
	sub.notifyCh = ch
	sub.address = addr
	sub.batch = batch
	sub.blocks = make([]*types.StateBlock, 0, MaxNotifyBlocks)
	r.allSubs[subID] = sub
}

func (r *BlockSubscription) removeChan(subID rpc.ID) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()

	sub := r.allSubs[subID]
	if sub == nil {
		return
	}

	delete(r.allSubs, subID)
}

func (r *BlockSubscription) notifyLoop() {
	defer r.unsubscribeEvent()

	for {
		select {
		case block := <-r.blocksCh:
			r.notifyAllSubs(block)
		case <-r.ctx.Done():
			r.logger.Info("broadcast subscription stopped")
			return
		}
	}
}

func (r *BlockSubscription) notifyAllSubs(block *types.StateBlock) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, sub := range r.allSubs {
		if sub.address.IsZero() || (block.GetAddress() == sub.address && sub.batch) {
			if len(sub.blocks) >= cap(sub.blocks) {
				copy(sub.blocks, sub.blocks[1:len(sub.blocks)])
				sub.blocks[cap(sub.blocks)-1] = block
			} else {
				sub.blocks = append(sub.blocks, block)
			}

			select {
			case sub.notifyCh <- struct{}{}:
			default:
			}
		} else if block.GetAddress() == sub.address {
			sub.addrBlock = block

			select {
			case sub.notifyCh <- struct{}{}:
			default:
			}
		}
	}
}

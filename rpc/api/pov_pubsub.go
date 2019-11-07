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
	MaxNotifyPovBlocks = 100
)

func CreatePovSubscription(ctx context.Context,
	fn func(notifier *rpc.Notifier, subscription *rpc.Subscription)) (*rpc.Subscription, error) {
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

// PovSubscription subscript event from chain, and deliver to every connected websocket
type PovSubscription struct {
	mu         sync.Mutex
	eb         event.EventBus
	handlerIds map[common.TopicType]string // subscript event from chain
	allSubs    map[rpc.ID]*PovSubscriber
	blocksCh   chan *types.PovBlock
	ctx        context.Context
	logger     *zap.SugaredLogger
}

type PovSubscriber struct {
	notifyCh chan struct{}
	blocks   []*types.PovBlock
}

func NewPovSubscription(ctx context.Context, eb event.EventBus) *PovSubscription {
	be := &PovSubscription{
		eb:         eb,
		handlerIds: make(map[common.TopicType]string),
		allSubs:    make(map[rpc.ID]*PovSubscriber),
		blocksCh:   make(chan *types.PovBlock, MaxNotifyPovBlocks),
		ctx:        ctx,
		logger:     log.NewLogger("pov_pubsub"),
	}
	be.subscribeEvent()
	go be.notifyLoop()
	return be
}

func (r *PovSubscription) subscribeEvent() {
	if id, err := r.eb.Subscribe(common.EventPovConnectBestBlock, r.setBlocks); err != nil {
		r.logger.Error("subscribe EventPovConnectBestBlock error, ", err)
	} else {
		r.handlerIds[common.EventPovConnectBestBlock] = id
	}
}

func (r *PovSubscription) unsubscribeEvent() {
	for k, v := range r.handlerIds {
		if err := r.eb.Unsubscribe(k, v); err != nil {
			r.logger.Error("unsubscribe event error, ", err)
		}
	}
}

func (r *PovSubscription) setBlocks(block *types.PovBlock) {
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

func (r *PovSubscription) fetchBlocks(subID rpc.ID) []*types.PovBlock {
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
	sub.blocks = make([]*types.PovBlock, 0, MaxNotifyPovBlocks)
	return retBlks
}

func (r *PovSubscription) addChan(subID rpc.ID, ch chan struct{}) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()

	sub := r.allSubs[subID]
	if sub != nil {
		r.logger.Errorf("chan %d exist already", subID)
		return
	}

	sub = new(PovSubscriber)
	sub.notifyCh = ch
	sub.blocks = make([]*types.PovBlock, 0, MaxNotifyPovBlocks)
	r.allSubs[subID] = sub
}

func (r *PovSubscription) removeChan(subID rpc.ID) {
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

func (r *PovSubscription) notifyLoop() {
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

func (r *PovSubscription) notifyAllSubs(block *types.PovBlock) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, sub := range r.allSubs {
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
	}
}

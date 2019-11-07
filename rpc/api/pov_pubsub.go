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
	chans      []chan struct{}             // deliver to every channel for each connected websocket
	newBlockCh chan struct{}
	blocks     []*types.PovBlock
	stoped     chan bool
	ctx        context.Context
	logger     *zap.SugaredLogger
}

func NewPovSubscription(eb event.EventBus, ctx context.Context) *PovSubscription {
	be := &PovSubscription{
		eb:         eb,
		handlerIds: make(map[common.TopicType]string),
		chans:      make([]chan struct{}, 0),
		newBlockCh: make(chan struct{}, 1),
		blocks:     make([]*types.PovBlock, 0, MaxNotifyPovBlocks),
		stoped:     make(chan bool),
		ctx:        ctx,
		logger:     log.NewLogger("pov_pubsub"),
	}
	be.subscribeEvent()
	return be
}

func (r *PovSubscription) subscribeEvent() {
	if id, err := r.eb.Subscribe(common.EventPovConnectBestBlock, r.setBlocks); err != nil {
		r.logger.Error("subscribe EventPovConnectBestBlock error, ", err)
	} else {
		r.handlerIds[common.EventPovConnectBestBlock] = id
	}
	go func() {
		select {
		case <-r.ctx.Done():
			r.unsubscribeEvent()
		}
	}()
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

	if len(r.chans) > 0 {
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

func (r *PovSubscription) getBlocks() []*types.PovBlock {
	r.mu.Lock()
	defer r.mu.Unlock()

	retBlks := make([]*types.PovBlock, len(r.blocks))
	copy(retBlks, r.blocks)
	return retBlks
}

func (r *PovSubscription) addChan(ch chan struct{}) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()
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

func (r *PovSubscription) removeChan(ch chan struct{}) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()
	for index, c := range r.chans {
		if c == ch {
			r.logger.Infof("remove chan %p ", c)
			r.chans = append(r.chans[:index], r.chans[index+1:]...)
			if len(r.chans) == 0 { // when websocket all disconnected, should unsubscribe event from chain
				r.stoped <- true
			}
			break
		}
	}
	if len(r.chans) == 0 {
		r.blocks = make([]*types.PovBlock, 0, MaxNotifyPovBlocks)
	}
}

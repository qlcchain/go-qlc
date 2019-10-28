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
	mu         *sync.Mutex
	eb         event.EventBus
	handlerIds map[common.TopicType]string // subscript event from chain
	chans      []chan *types.PovBlock      // deliver to every channel for each connected websocket
	blocks     chan *types.PovBlock
	stoped     chan bool
	logger     *zap.SugaredLogger
}

func NewPovSubscription(eb event.EventBus) *PovSubscription {
	be := &PovSubscription{
		eb:         eb,
		mu:         &sync.Mutex{},
		handlerIds: make(map[common.TopicType]string),
		chans:      make([]chan *types.PovBlock, 0),
		blocks:     make(chan *types.PovBlock, 100),
		stoped:     make(chan bool),
		logger:     log.NewLogger("pov_pubsub"),
	}
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
	if len(r.chans) > 0 {
		r.blocks <- block
	}
}

func (r *PovSubscription) addChan(ch chan *types.PovBlock) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()
	r.chans = append(r.chans, ch)
	r.logger.Infof("add chan %p to blockEvent, chan length %d", ch, len(r.chans))
	if len(r.chans) == 1 {
		r.subscribeEvent() // only when first websocket to connect, should subscript event from chain
		go func() {
			for {
				select {
				case b := <-r.blocks:
					for _, c := range r.chans { // broadcast event to every websocket channel
						r.logger.Infof("broadcast (%s) to channel %p (%d, %d)", b.GetHash(), c, len(r.chans), len(r.blocks))
						c <- b
					}
				case <-r.stoped:
					r.logger.Info("broadcast subscription stopped")
					return
				}
			}
		}()
	}
}

func (r *PovSubscription) removeChan(ch chan *types.PovBlock) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
	}()
	for index, c := range r.chans {
		if c == ch {
			r.logger.Infof("remove chan %p ", c)
			r.chans = append(r.chans[:index], r.chans[index+1:]...)
			if len(r.chans) == 0 { // when websocket all disconnected, should unsubscribe event from chain
				r.unsubscribeEvent()
				r.stoped <- true
			}
			break
		}
	}
}

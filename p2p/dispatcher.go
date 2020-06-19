package p2p

import (
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/log"
)

// Dispatcher a message dispatcher service.
type Dispatcher struct {
	subscribersMap     *sync.Map
	quitCh             chan bool
	receivedMessageCh  chan *Message
	syncMessageCh      chan *Message
	dispatchedMessages *lru.Cache
	logger             *zap.SugaredLogger
}

// NewDispatcher create Dispatcher instance.
func NewDispatcher() *Dispatcher {
	dp := &Dispatcher{
		subscribersMap:    new(sync.Map),
		quitCh:            make(chan bool, 1),
		receivedMessageCh: make(chan *Message, common.P2PMsgChanSize),
		syncMessageCh:     make(chan *Message, common.P2PMsgChanSize),
		logger:            log.NewLogger("dispatcher"),
	}

	dp.dispatchedMessages, _ = lru.New(common.P2PMsgCacheSize)

	return dp
}

// Register register subscribers.
func (dp *Dispatcher) Register(subscribers ...*Subscriber) {
	for _, v := range subscribers {
		mt := v.MessageType()
		_, _ = dp.subscribersMap.LoadOrStore(mt, v)
	}
}

// Deregister deregister subscribers.
func (dp *Dispatcher) Deregister(subscribers *Subscriber) {
	mt := subscribers.MessageType()
	if _, ok := dp.subscribersMap.Load(mt); ok {
		dp.subscribersMap.Delete(mt)
	}
}

// Start start message dispatch goroutine.
func (dp *Dispatcher) Start() {
	go dp.loop()
}

func (dp *Dispatcher) loop() {
	dp.logger.Info("Started NewService Dispatcher.")

	for {
		select {
		case <-dp.quitCh:
			dp.logger.Info("Stoped Qlc Dispatcher.")
			return
		case msg := <-dp.receivedMessageCh:
			msgType := msg.MessageType()

			v, _ := dp.subscribersMap.Load(msgType)
			if v == nil {
				continue
			}

			select {
			case v.(*Subscriber).msgChan <- msg:
			default:
				dp.logger.Debug("timeout to dispatch message.")
			}
		case msg := <-dp.syncMessageCh:
			msgType := msg.MessageType()

			v, _ := dp.subscribersMap.Load(msgType)
			if v == nil {
				continue
			}

			select {
			case v.(*Subscriber).msgChan <- msg:
			default:
				dp.logger.Debug("timeout to dispatch message.")
			}
		}
	}
}

// Stop stop goroutine.
func (dp *Dispatcher) Stop() {
	//dp.logger.VInfo("Stopping QlcService Dispatcher...")

	dp.quitCh <- true
}

// PutMessage put new message to chan, then subscribers will be notified to process.
func (dp *Dispatcher) PutMessage(msg *Message) {
	select {
	case dp.receivedMessageCh <- msg:
	default:
		dp.logger.Debugf("dispatcher receive message chan expire")
	}
}

// PutMessage put new message to chan, then subscribers will be notified to process.
func (dp *Dispatcher) PutSyncMessage(msg *Message) {
	select {
	case dp.syncMessageCh <- msg:
	default:
		dp.logger.Debugf("dispatcher sync message chan expire")
	}
}

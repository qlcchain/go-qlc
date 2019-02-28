package p2p

import (
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"

	lru "github.com/hashicorp/golang-lru"
)

// Dispatcher a message dispatcher service.
type Dispatcher struct {
	subscribersMap     *sync.Map
	quitCh             chan bool
	receivedMessageCh  chan Message
	dispatchedMessages *lru.Cache
	filters            map[MessageType]bool
	logger             *zap.SugaredLogger
}

// NewDispatcher create Dispatcher instance.
func NewDispatcher() *Dispatcher {
	dp := &Dispatcher{
		subscribersMap:    new(sync.Map),
		quitCh:            make(chan bool, 1),
		receivedMessageCh: make(chan Message, 65536),
		filters:           make(map[MessageType]bool),
		logger:            log.NewLogger("dispatcher"),
	}

	dp.dispatchedMessages, _ = lru.New(51200)

	return dp
}

// Register register subscribers.
func (dp *Dispatcher) Register(subscribers ...*Subscriber) {
	for _, v := range subscribers {
		mt := v.MessageType()
		m, _ := dp.subscribersMap.LoadOrStore(mt, new(sync.Map))
		m.(*sync.Map).Store(v, true)
		dp.filters[mt] = v.DoFilter()
	}
}

// Deregister deregister subscribers.
func (dp *Dispatcher) Deregister(subscribers ...*Subscriber) {

	for _, v := range subscribers {
		mt := v.MessageType()
		m, _ := dp.subscribersMap.Load(mt)
		if m == nil {
			continue
		}
		m.(*sync.Map).Delete(v)
		delete(dp.filters, mt)
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
			m, _ := v.(*sync.Map)

			m.Range(func(key, value interface{}) bool {
				select {
				case key.(*Subscriber).msgChan <- msg:
				default:
					dp.logger.Debug("timeout to dispatch message.")
					time.Sleep(100 * time.Millisecond)
				}
				return true
			})
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Stop stop goroutine.
func (dp *Dispatcher) Stop() {
	//dp.logger.Info("Stopping QlcService Dispatcher...")

	dp.quitCh <- true
}

// PutMessage put new message to chan, then subscribers will be notified to process.
func (dp *Dispatcher) PutMessage(msg Message) {
	hash := msg.Hash()
	if dp.filters[msg.MessageType()] {
		if exist, _ := dp.dispatchedMessages.ContainsOrAdd(hash, hash); exist == true {
			return
		}
	}
	dp.receivedMessageCh <- msg
}

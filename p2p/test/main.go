package main

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/test/protos"
)

//  Message Type
const (
	TestHeadersRequest = "test" // test
)

type TestService struct {
	netService p2p.Service
	quitCh     chan bool
	messageCh  chan p2p.Message
}

// NewService return new Service.
func NewTestService(netService p2p.Service) *TestService {
	return &TestService{
		netService: netService,
		quitCh:     make(chan bool, 1),
		messageCh:  make(chan p2p.Message, 128),
	}
}

// Start start sync service.
func (ss *TestService) Start() {

	// register the network handler.
	netService := ss.netService
	netService.Register(p2p.NewSubscriber(ss, ss.messageCh, false, TestHeadersRequest))

	// start loop().
	go ss.startLoop()
}
func (ss *TestService) startLoop() {
	fmt.Println("Started Test Service.")

	for {
		select {
		case <-ss.quitCh:
			fmt.Println("Stopped Test Service.")
			return
		case message := <-ss.messageCh:
			switch message.MessageType() {
			case TestHeadersRequest:
				ss.onTestHeadersRequest(message)
			default:
				fmt.Println("Received unknown message.")
			}
		}
	}
}
func (ss *TestService) onTestHeadersRequest(message p2p.Message) {

	// handle TestHeadersRequest message.
	str := new(protos.Test)
	err := proto.Unmarshal(message.Data()[p2p.QlcMessageHeaderLength:], str)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("receive message :", str)
}

func main() {
	cfg := config.DefaultlConfig
	node, err := p2p.NewQlcService(cfg)
	if err != nil {
		fmt.Println(err)
	}
	node.Start()
	t := NewTestService(node)
	t.Start()

	msg := &protos.Test{
		StrTest: "I am ubuntu 2",
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println("Failed to marshal proto message.")
		return
	}
	for {
		node.Broadcast(TestHeadersRequest, data)
		time.Sleep(time.Duration(2) * time.Second)
	}

	select {}
}

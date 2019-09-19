package main

import (
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
	log "github.com/sirupsen/logrus"
	"time"
)

type NodeClient struct {
	client    *rpc.Client
	minerAddr string
	algoName  string
	lastWork  *api.PovApiGetWork

	eventChan chan Event
	quitCh    chan struct{}
}

func NewNodeClient(url string, minerAddr string, algoName string) *NodeClient {
	var err error

	nc := new(NodeClient)
	nc.minerAddr = minerAddr
	nc.algoName = algoName

	nc.client, err = rpc.Dial(url)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	nc.eventChan = make(chan Event, EventMaxChanSize)
	nc.quitCh = make(chan struct{})

	return nc
}

func (nc *NodeClient) Start() error {
	GetDefaultEventBus().Subscribe(EventJobSubmit, nc.eventChan)

	go nc.loop()

	return nil
}

func (nc *NodeClient) Stop() {
	close(nc.quitCh)
}

func (nc *NodeClient) loop() {
	log.Infof("node running loop, miner:%s, algo:%s", nc.minerAddr, nc.algoName)

	workTicker := time.NewTicker(5 * time.Second)
	defer workTicker.Stop()

	for {
		select {
		case <-nc.quitCh:
			return
		case <-workTicker.C:
			nc.getWork()
		case event := <-nc.eventChan:
			nc.consumeEvent(event)
		}
	}
}

func (nc *NodeClient) consumeEvent(event Event) {
	switch event.Topic {
	case EventJobSubmit:
		nc.consumeJobSubmit(event)
	}
}

func (nc *NodeClient) consumeJobSubmit(event Event) {
	work := event.Data.(*JobWork)
	if len(work.submits) <= 0 {
		log.Errorf("submit not exist for job work %s", work.JobHash)
		return
	}
	js := work.submits[len(work.submits)-1]

	apiSubmit := new(api.PovApiSubmitWork)

	apiSubmit.WorkHash = work.WorkHash
	apiSubmit.Nonce = js.Nonce
	apiSubmit.Timestamp = js.NTime
	apiSubmit.CoinbaseExtra = work.CoinbaseExtra
	apiSubmit.CoinbaseHash = work.CoinbaseHash
	apiSubmit.MerkleRoot = work.MerkleRoot
	apiSubmit.BlockHash = work.BlockHash

	nc.submitWork(apiSubmit)
}

func (nc *NodeClient) getWork() {
	getWorkRsp := new(api.PovApiGetWork)
	err := nc.client.Call(&getWorkRsp, "pov_getWork", nc.minerAddr, nc.algoName)
	if err != nil {
		log.Errorln(err)
		return
	}

	if nc.lastWork != nil && nc.lastWork.WorkHash == getWorkRsp.WorkHash {
		return
	}

	log.Infof("getWork response: %s", util.ToString(getWorkRsp))

	nc.lastWork = getWorkRsp

	GetDefaultEventBus().Publish(EventUpdateApiWork, getWorkRsp)
}

func (nc *NodeClient) submitWork(submitWorkReq *api.PovApiSubmitWork) {
	log.Infof("submitWork request: %s", util.ToString(submitWorkReq))
	err := nc.client.Call(nil, "pov_submitWork", &submitWorkReq)
	if err != nil {
		log.Errorln(err)
		return
	}
}

func (nc *NodeClient) getLatestHeader() *api.PovApiHeader {
	latestHeaderRsp := new(api.PovApiHeader)
	err := nc.client.Call(latestHeaderRsp, "pov_getLatestHeader")
	if err != nil {
		log.Errorln(err)
		return nil
	}
	return latestHeaderRsp
}

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common/simplejson"
	"github.com/willf/bitset"
)

const (
	MaxReqSize = 10 * 1024 * 1024
)

const (
	SessionStateConnected = iota
	SessionStateSubscribed
	SessionStateAuthenticated
)

type StratumConfig struct {
	ServerID uint16
	Host     string
	Port     uint32
	MaxConn  uint32
	Debug    bool
}

type StratumServer struct {
	config *StratumConfig

	sessionIDMng *bitset.BitSet
	sessAllocIdx uint
	sessionsMu   sync.RWMutex
	sessions     map[uint32]*StratumSession

	lastNotifyAll *StratumNotify

	eventChan chan Event
	quitCh    chan struct{}
}

type StratumSession struct {
	sync.Mutex
	server   *StratumServer
	conn     *net.TCPConn
	state    int
	reqMsgID int

	sessionID     uint32
	extraNonce1   uint32
	remoteIP      string
	clientVersion string
	username      string
	password      string
}

func NewStratumServer(config *StratumConfig) *StratumServer {
	for config.ServerID <= 0 {
		config.ServerID = uint16(rand.Intn(0xFFFF))
	}

	s := &StratumServer{config: config}

	s.sessionIDMng = bitset.New(0x0000FFFF)
	s.sessionIDMng.Set(0)

	for s.sessAllocIdx <= 0 {
		s.sessAllocIdx = uint(rand.Intn(255))
	}

	s.sessions = make(map[uint32]*StratumSession)

	s.eventChan = make(chan Event, EventMaxChanSize)
	s.quitCh = make(chan struct{})
	return s
}

func (s *StratumServer) Start() error {
	log.Infof("start stratum, serverID %d sessAllocIdx %d", s.config.ServerID, s.sessAllocIdx)

	GetDefaultEventBus().Subscribe(EventBCastJobWork, s.eventChan)
	GetDefaultEventBus().Subscribe(EventMinerSendRsp, s.eventChan)
	GetDefaultEventBus().Subscribe(EventStatisticsTicker, s.eventChan)

	go s.consumeLoop()

	go s.listenLoop()

	return nil
}

func (s *StratumServer) Stop() {
	close(s.quitCh)
}

func (s *StratumServer) consumeLoop() {
	log.Infoln("stratum running consume loop")

	statsTicker := time.NewTicker(5 * time.Minute)
	defer statsTicker.Stop()

	for {
		select {
		case <-s.quitCh:
			return
		case event := <-s.eventChan:
			s.consumeEvent(event)
		case <-statsTicker.C:
			s.onStatisticsTicker()
		}
	}
}

func (s *StratumServer) consumeEvent(event Event) {
	switch event.Topic {
	case EventBCastJobWork:
		s.consumeJobWork(event)
	case EventMinerSendRsp:
		s.consumeMinerSendRsp(event)
	case EventStatisticsTicker:
		s.consumeStatisticsTicker(event)
	}
}

func (s *StratumServer) consumeJobWork(event Event) {
	jobWork := event.Data.(*JobWork)

	nf := jobWork.ToNotify()
	s.sendMiningNotifyToAll(nf)

	s.lastNotifyAll = nf
}

func (s *StratumServer) consumeMinerSendRsp(event Event) {
	msgRsp := event.Data.(*StratumRsp)

	cs := s.sessions[msgRsp.SessionID]
	if cs == nil {
		log.Errorf("can't find session %08x", msgRsp.SessionID)
		return
	}

	var err error
	if msgRsp.ErrCode != 0 {
		err = cs.sendError(msgRsp.MsgID, msgRsp.ErrCode, msgRsp.ErrMsg)
	}
	if msgRsp.Result != nil {
		err = cs.sendResult(msgRsp.MsgID, msgRsp.Result)
	}
	if err != nil {
		log.Errorf("send miner rsp message err %s", err)
	}
}

func (s *StratumServer) consumeStatisticsTicker(event Event) {
	log.Infof("stratum clients: session:%d", len(s.sessions))
}

func (s *StratumServer) onStatisticsTicker() {
	GetDefaultEventBus().Publish(EventStatisticsTicker, time.Now())
}

func (s *StratumServer) listenLoop() {
	log.Infoln("stratum running listen loop")

	bindAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	addr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	server, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer server.Close()

	log.Infof("stratum listening on %s", bindAddr)
	accept := make(chan int, s.config.MaxConn)
	n := 0

	for {
		conn, err := server.AcceptTCP()
		if err != nil {
			continue
		}
		conn.SetKeepAlive(true)

		// new client session
		remoteIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		cs := &StratumSession{conn: conn, remoteIP: remoteIP, server: s, state: SessionStateConnected}
		cs.sessionID = uint32(s.allocSessionID())
		cs.extraNonce1 = cs.sessionID
		n += 1
		s.registerSession(cs)

		log.Infof("stratum accept client remoteIP %s session %08x", remoteIP, cs.sessionID)

		accept <- n
		go func() {
			s.handleClient(cs)
			<-accept
		}()
	}
}

func (s *StratumServer) setDeadline(conn *net.TCPConn) {
	conn.SetDeadline(time.Now().Add(15 * time.Minute))
}

//
//  SESSION ID: UINT32_T
//
//  0 bit or longer       8bit            24 bit or shorter
// -----------------    ---------    ----------------------------
// leading zero bits    server ID             session id
//     [000...]          [1, 255]    range: [0, 0x00FFFFFF]

func (s *StratumServer) allocSessionID() uint {
	if s.sessionIDMng.All() {
		return 0
	}

	for s.sessionIDMng.Test(s.sessAllocIdx) {
		s.sessAllocIdx++
	}
	s.sessionIDMng.Set(s.sessAllocIdx)

	id := (uint(s.config.ServerID) << 16) | s.sessAllocIdx
	return id
}

func (s *StratumServer) freeSessionID(id uint) {
	idx := id & 0x00FFFFFF
	s.sessionIDMng.Clear(idx)
}

func (s *StratumServer) setSessionID(id uint) {
	idx := id & 0x00FFFFFF
	s.sessionIDMng.Set(idx)
}

func (s *StratumServer) registerSession(cs *StratumSession) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	if s.sessions[cs.sessionID] != nil {
		log.Errorf("session id %08x exist", cs.sessionID)
	}
	s.sessions[cs.sessionID] = cs
}

func (s *StratumServer) removeSession(cs *StratumSession) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.freeSessionID(uint(cs.sessionID))
	delete(s.sessions, cs.sessionID)
}

func (s *StratumServer) isActive(cs *StratumSession) bool {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()
	_, exist := s.sessions[cs.sessionID]
	return exist
}

func (s *StratumServer) sendMiningNotifyToAll(nf *StratumNotify) {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	for _, sess := range s.sessions {
		if sess.state == SessionStateAuthenticated {
			_ = sess.sendMiningNotify(nf, nf.CleanJobs)
		}
	}
}

func (s *StratumServer) handleClient(cs *StratumSession) {
	conBuf := bufio.NewReaderSize(cs.conn, MaxReqSize)
	s.setDeadline(cs.conn)

	for {
		data, isPrefix, err := conBuf.ReadLine()
		if isPrefix {
			log.Errorln("socket flood detected from", cs.remoteIP)
			break
		} else if err == io.EOF {
			log.Errorln("client disconnected", cs.remoteIP)
			break
		} else if err != nil {
			log.Errorln("error reading:", err)
			break
		}

		// NOTICE: cpuminer-multi sends junk newlines, so we demand at least 1 byte for decode
		// NOTICE: Ns*CNMiner.exe will send malformed JSON on very low diff, not sure we should handle this
		if len(data) > 1 {
			reqMsg := cs.decodeMessage(data)
			if reqMsg == nil {
				log.Errorf("failed to decode message from %s", cs.remoteIP)
				continue
			}

			s.setDeadline(cs.conn)
			err = cs.handleMessage(reqMsg)
			if err != nil {
				//break
			}
		}
	}
	s.removeSession(cs)
	_ = cs.conn.Close()
}

func (cs *StratumSession) decodeMessage(rawMsg []byte) *StratumMsg {
	var err error

	msg := new(StratumMsg)
	msg.RawMsg = rawMsg
	msg.JsonMsg, err = simplejson.NewJson(rawMsg)
	if err != nil {
		log.Errorf("new json from raw msg, err %s", err)
		return nil
	}

	msg.ID = msg.JsonMsg.Get("id").Interface()
	msg.Method, _ = msg.JsonMsg.Get("method").String()
	msg.Params = msg.JsonMsg.Get("params")

	msg.Error = msg.JsonMsg.Get("error")
	msg.Result = msg.JsonMsg.Get("result")

	return msg
}

func (cs *StratumSession) handleMessage(msg *StratumMsg) error {
	if msg.ID == nil {
		err := fmt.Errorf("server disconnect request")
		log.Errorln(err)
		return err
	} else if msg.Params == nil {
		err := fmt.Errorf("server RPC request params")
		log.Errorln(err)
		return err
	}

	log.Debugf("handleMessage: %s", string(msg.RawMsg))

	// Handle RPC methods
	switch msg.Method {
	case "mining.subscribe":
		return cs.handleMiningSubscribe(msg)
	case "mining.authorize":
		return cs.handleMiningAuthorize(msg)
	case "mining.extranonce.subscribe":
		return cs.handleMiningExtraNonceSubscribe(msg)
	case "mining.submit":
		return cs.handleMiningSubmit(msg)
	case "mining.pong":
		return cs.handleMiningPong(msg)
	case "mining.configure":
		return cs.handleMiningConfigure(msg)
	default:
		log.Warnf("Unknown method %s", msg.Method)
		return cs.sendError(msg.ID, 20, "Unknown method")
	}

	return nil
}

func (cs *StratumSession) handleMiningAuthorize(reqMsg *StratumMsg) error {
	/*
		{"params": ["slush.miner1", "password"], "id": 2, "method": "mining.authorize"}\n
		{"error": null, "id": 2, "result": true}\n
	*/
	var err error

	if reqMsg.Params.ArrayLen() < 2 {
		return cs.sendError(reqMsg.ID, 27, "Illegal params")
	}

	cs.username, err = reqMsg.Params.GetIndex(0).String()
	if err != nil {
		return cs.sendError(reqMsg.ID, 27, "Illegal params[0]")
	}
	cs.password, err = reqMsg.Params.GetIndex(1).String()
	if err != nil {
		return cs.sendError(reqMsg.ID, 27, "Illegal params[0]")
	}

	log.Infof("handleMiningAuthorize: username: %s, password: %s", cs.username, cs.password)

	cs.state = SessionStateAuthenticated

	rspMsg := new(StratumMsg)
	rspMsg.ID = reqMsg.ID
	rspMsg.Result = simplejson.New()
	rspMsg.Result.SetPath(nil, true)

	return cs.sendMessage(rspMsg)
}

func (cs *StratumSession) handleMiningSubscribe(reqMsg *StratumMsg) error {
	/*
		On the beginning of the session, client subscribes current connection for receiving mining jobs:
		{"id": 1, "method": "mining.subscribe", "params": ["client version", "session id / ExtraNonce1 (hex)", miner's real IP (unit32)]}

		{"id": 1, "result": [ [ ["mining.set_difficulty", "subscription id 1"], ["mining.notify", "subscription id 2"]], "08000002", 4], "error": null}

		The result contains three items:
		Subscriptions details - 2-tuple with name of subscribed notification and subscription ID. Teoretically it may be used for unsubscribing, but obviously miners won't use it.
		Extranonce1 - Hex-encoded, per-connection unique string which will be used for coinbase serialization later. Keep it safe!
		Extranonce2_size - Represents expected length of extranonce2 which will be generated by the miner.

	*/
	if reqMsg.Params.ArrayLen() >= 1 {
		cs.clientVersion, _ = reqMsg.Params.GetIndex(0).String()
	}

	cs.state = SessionStateSubscribed

	sessIDHex := UInt32ToHexBE(cs.sessionID)

	subDetails := simplejson.NewArray(2)

	subDiff := simplejson.NewArray(2)
	subDiff.AddArray("mining.set_difficulty")
	subDiff.AddArray(sessIDHex)
	subDetails.AddArray(subDiff)

	subNf := simplejson.NewArray(2)
	subNf.AddArray("mining.notify")
	subNf.AddArray(sessIDHex)
	subDetails.AddArray(subNf)

	paramsSub := simplejson.NewArray(3)
	paramsSub.AddArray(subDetails)
	paramsSub.AddArray(UInt32ToHexLE(cs.extraNonce1))
	paramsSub.AddArray(8)

	rspMsg := new(StratumMsg)
	rspMsg.ID = reqMsg.ID
	rspMsg.Result = paramsSub

	err := cs.sendMessage(rspMsg)
	if err != nil {
		return err
	}

	if cs.server.lastNotifyAll != nil {
		err := cs.sendMiningNotify(cs.server.lastNotifyAll, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *StratumSession) handleMiningExtraNonceSubscribe(reqMsg *StratumMsg) error {
	return cs.sendResult(reqMsg.ID, true)
}

func (cs *StratumSession) handleMiningSubmit(reqMsg *StratumMsg) error {
	/*
		When miner find the job which meets requested difficulty, it can submit share to the server:

		{"params": ["slush.miner1", "bf", "00000001", "504e86ed", "b2957c02"], "id": 4, "method": "mining.submit"}
		{"error": null, "id": 4, "result": true}
		Values in particular order: worker_name (previously authorized!), job_id, extranonce2, ntime, nonce.
	*/

	if cs.state != SessionStateAuthenticated {
		log.Warnf("submit share but state %d is not authenticated", cs.state)
		cs.sendClientReconnect()
		return nil
	}

	if reqMsg.Params.ArrayLen() < 5 {
		return ErrInvalidParams
	}

	submit := new(StratumSubmit)
	submit.SessionID = cs.sessionID
	submit.MsgID = reqMsg.ID

	submit.WorkerName, _ = reqMsg.Params.GetIndex(0).String()
	submit.JobID, _ = reqMsg.Params.GetIndex(1).String()
	submit.ExtraNonce2Hex, _ = reqMsg.Params.GetIndex(2).String()
	submit.NTimeHex, _ = reqMsg.Params.GetIndex(3).String()
	submit.NonceHex, _ = reqMsg.Params.GetIndex(4).String()

	submit.NTime = HexBEToUInt32(submit.NTimeHex)
	submit.Nonce = HexBEToUInt32(submit.NonceHex)
	submit.ExtraNonce2 = HexLEToUInt64(submit.ExtraNonce2Hex)

	submit.ExtraNonce1 = cs.extraNonce1

	GetDefaultEventBus().Publish(EventMinerSubmit, submit)

	// job will send response msg to miner after process submit
	// return cs.sendResult(reqMsg.ID, true)
	return nil
}

func (cs *StratumSession) handleMiningPong(reqMsg *StratumMsg) error {
	return nil
}

func (cs *StratumSession) handleMiningConfigure(reqMsg *StratumMsg) error {
	//
	// {
	//   "method": "mining.configure",
	//   "id": 1,
	//   "params":
	//     [
	//       ["minimum-difficulty", "version-rolling"],
	//       {
	//         "minimum-difficulty.value": 2048,
	//         "version-rolling.mask": "1fffe000",
	//         "version-rolling.min-bit-count": 2
	//       }
	//     ]
	// }
	//
	// {
	//   "error": null,
	//   "id": 1,
	// 	 "result":
	//     {
	// 		  "version-rolling": true,
	// 		  "version-rolling.mask": "18000000",
	// 		  "minimum-difficulty": true
	// 	   }
	// }
	//
	if reqMsg.Params.ArrayLen() < 2 {
		return cs.sendError(reqMsg.ID, 27, "Illegal params")
	}

	extNames, err := reqMsg.Params.GetIndex(0).StringArray()
	if err != nil {
		return cs.sendError(reqMsg.ID, 27, "Illegal params[0]")
	}
	extOpts := reqMsg.Params.GetIndex(1)
	if extOpts == nil {
		return cs.sendError(reqMsg.ID, 27, "Illegal params[1]")
	}

	results := simplejson.New()
	for _, extName := range extNames {
		results.Set(extName, false)
	}

	return cs.sendResult(reqMsg.ID, results)
}

func (cs *StratumSession) sendMiningSetDifficulty(diff float64) error {
	/*
		Default share difficulty is 1 (big-endian target for difficulty 1 is 0x00000000ffff0000000000000000000000000000000000000000000000000000),
		but server can ask you anytime during the session to change it:
		{ "id": null, "method": "mining.set_difficulty", "params": [2]}
		This Means That Difficulty 2 Will Be Applied to Every Next Job Received From the Server.
	*/

	outMsg := new(StratumMsg)
	outMsg.Method = "mining.set_difficulty"
	outMsg.Params = simplejson.NewArray(1)
	outMsg.Params.AddArray(diff)
	return cs.sendMessage(outMsg)
}

func (cs *StratumSession) sendMiningSetExtraNonce(extraNonce1 string, extraNonce2Size int) error {
	outMsg := new(StratumMsg)
	outMsg.Method = "mining.set_extranonce"
	outMsg.Params = simplejson.NewArray(1)
	outMsg.Params.AddArray(extraNonce1)
	outMsg.Params.AddArray(extraNonce2Size)
	return cs.sendMessage(outMsg)
}

func (cs *StratumSession) sendMiningNotify(nf *StratumNotify, clean bool) error {
	cs.Lock()
	defer cs.Unlock()

	/*
		{"id": null, "method": "mining.notify", "params": ["bf", "4d16b6f85af6e2198f44ae2a6de67f78487ae5611b77c6c0440b921e00000000",
			"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff20020862062f503253482f04b8864e5008",
			"072f736c7573682f000000000100f2052a010000001976a914d23fcdf86f7e756a64a7a9688ef9903327048ed988ac00000000", [],
			"00000002", "1c2ac4af", "504e86b9", false]}
		job_id - ID of the job. Use this ID while submitting share generated from this job.
		prevhash - Hash of previous block.
		coinb1 - Initial part of coinbase transaction.
		coinb2 - Final part of coinbase transaction.
		merkle_branch - List of hashes, will be used for calculation of merkle root.
		                This is not a list of all transactions, it only contains prepared hashes of steps of merkle tree algorithm.
		version - Block version.
		nbits - Encoded current network difficulty
		ntime - Current ntime/
		clean_jobs - When true, server indicates that submitting shares from previous jobs don't have a sense and such shares will be rejected.
		             When this flag is set, miner should also drop all previous jobs, so job_ids can be eventually rotated.
	*/

	cs.sendMiningSetDifficulty(nf.Difficulty)

	paramsNf := simplejson.NewArray(9)
	paramsNf.AddArray(nf.JobID)
	paramsNf.AddArray(nf.PrevHash)
	paramsNf.AddArray(nf.Coinbase1)
	paramsNf.AddArray(nf.Coinbase2)
	paramsNf.AddArray(nf.MerkleBranch)
	paramsNf.AddArray(nf.Version)
	paramsNf.AddArray(nf.NBits)
	paramsNf.AddArray(nf.NTime)
	paramsNf.AddArray(clean)

	msg := new(StratumMsg)
	msg.ID = nil
	msg.Method = "mining.notify"
	msg.Params = paramsNf

	return cs.sendMessage(msg)
}

func (cs *StratumSession) sendClientReconnect() error {
	outMsg := new(StratumMsg)
	outMsg.Method = "client.reconnect"
	outMsg.Params = simplejson.NewArray(0)

	return cs.sendMessage(outMsg)
}

func (cs *StratumSession) sendMiningPing() error {
	cs.reqMsgID++

	outMsg := new(StratumMsg)
	outMsg.ID = cs.reqMsgID
	outMsg.Method = "mining.ping"
	outMsg.Params = simplejson.NewArray(0)

	return cs.sendMessage(outMsg)
}

func (cs *StratumSession) sendResult(id interface{}, result interface{}) error {
	outMsg := new(StratumMsg)
	outMsg.ID = id
	outMsg.Result = simplejson.New()
	outMsg.Result.SetPath(nil, result)

	return cs.sendMessage(outMsg)
}

func (cs *StratumSession) sendError(id interface{}, errCode int, errMsg string) error {
	outMsg := new(StratumMsg)
	outMsg.ID = id
	outMsg.Error = simplejson.NewArray(3)
	outMsg.Error.AddArray(errCode)
	outMsg.Error.AddArray(errMsg)
	outMsg.Error.AddArray(nil)

	return cs.sendMessage(outMsg)
}

func (cs *StratumSession) SendMessageWithLock(msg *StratumMsg) error {
	cs.Lock()
	defer cs.Unlock()

	return cs.sendMessage(msg)
}

func (cs *StratumSession) sendMessage(msg *StratumMsg) error {
	rawMsg, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("json marshal msg err: %s", err)
		return err
	}

	log.Debugf("sendMessage: %s", string(rawMsg))

	msgBuf := new(bytes.Buffer)
	msgBuf.Write(rawMsg)
	msgBuf.WriteByte('\n')

	_, err = cs.conn.Write(msgBuf.Bytes())
	if err != nil {
		log.Errorf("write message err %s", err)
	}

	return err
}

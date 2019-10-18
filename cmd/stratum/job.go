package main

import (
	"bytes"
	"encoding/hex"
	"time"

	"github.com/qlcchain/go-qlc/cmd/stratum/txscript"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	log "github.com/sirupsen/logrus"
)

const (
	MaxWorkExpireSec = 10 * 60
)

type JobRepository struct {
	prevWorks      map[types.Hash][]*JobWork
	allWorks       map[types.Hash]*JobWork
	lastJobWork    *JobWork
	acceptedSubmit int
	rejectedSubmit int

	eventChan chan Event
	quitCh    chan struct{}
}

func NewJobRepository() *JobRepository {
	r := new(JobRepository)
	r.prevWorks = make(map[types.Hash][]*JobWork)
	r.allWorks = make(map[types.Hash]*JobWork)
	r.eventChan = make(chan Event, EventMaxChanSize)
	r.quitCh = make(chan struct{})
	return r
}

func (r *JobRepository) Start() error {
	GetDefaultEventBus().Subscribe(EventUpdateApiWork, r.eventChan)
	GetDefaultEventBus().Subscribe(EventMinerSubmit, r.eventChan)
	GetDefaultEventBus().Subscribe(EventStatisticsTicker, r.eventChan)

	go r.consumeLoop()

	return nil
}

func (r *JobRepository) Stop() {
	close(r.quitCh)
}

func (r *JobRepository) consumeLoop() {
	log.Infof("job running consume loop")

	checkJobTicker := time.NewTicker(10 * time.Second)

	defer checkJobTicker.Stop()

	for {
		select {
		case <-r.quitCh:
			return
		case event := <-r.eventChan:
			r.consumeEvent(event)
		case <-checkJobTicker.C:
			r.OnCheckJobTicker()
		}
	}
}

func (r *JobRepository) consumeEvent(event Event) {
	switch event.Topic {
	case EventUpdateApiWork:
		r.consumeApiWork(event)
	case EventMinerSubmit:
		r.consumeMinerSubmit(event)
	case EventStatisticsTicker:
		r.consumeStatisticsTicker(event)
	}
}

func (r *JobRepository) consumeApiWork(event Event) {
	apiWork := event.Data.(*api.PovApiGetWork)

	jobHash := apiWork.WorkHash
	if r.allWorks[jobHash] != nil {
		return
	}

	w := new(JobWork)
	w.JobHash = jobHash
	w.CreateAt = time.Now()
	w.CheckAt = w.CreateAt.Add(60 * time.Second)

	w.WorkHash = apiWork.WorkHash
	w.Version = apiWork.Version
	w.PrevHash = apiWork.Previous
	w.Bits = apiWork.Bits
	w.Height = apiWork.Height
	w.MinTime = apiWork.MinTime
	w.MerkleBranch = apiWork.MerkleBranch
	w.CoinBaseData1 = apiWork.CoinBaseData1
	w.CoinBaseData2 = apiWork.CoinBaseData2

	w.PoolTime = uint32(time.Now().Unix())

	target := types.CompactToBig(w.Bits)
	w.Difficulty = TargetToDiff(target, nil)

	txsb := txscript.NewScriptBuilder()
	txsb.AddInt64(int64(w.Height))
	txsb.AddInt64(int64(w.PoolTime))
	scriptBytes, _ := txsb.Script()

	cbExtBuf := new(bytes.Buffer)
	cbExtBuf.Write(scriptBytes)
	cbExtBuf.WriteString("/QLC Solo Pool/")
	w.CBPoolInfo = cbExtBuf.Bytes()

	w.CBExtraLen = len(w.CBPoolInfo) + 4 + 8

	if len(r.prevWorks[w.PrevHash]) > 0 {
		w.CleanJobs = false
	} else {
		w.CleanJobs = true
	}

	r.allWorks[w.JobHash] = w
	r.lastJobWork = w
	// Blocks sorted by PrevHash, so it's easy to drop them on work update
	r.prevWorks[w.PrevHash] = append(r.prevWorks[w.PrevHash], w)

	// drop works of obsolete blocks
	for ph, pws := range r.prevWorks {
		if ph != w.PrevHash {
			delete(r.prevWorks, ph)
			if len(pws) > 0 {
				log.Infof("clean obsolete works %d for PrevHash %s", len(pws), ph)
				for _, w := range pws {
					delete(r.allWorks, w.JobHash)
				}
			}
		}
	}

	log.Infof("new work %s for PrevHash %s", w.JobHash, w.PrevHash)

	GetDefaultEventBus().Publish(EventBCastJobWork, w)
}

func (r *JobRepository) consumeMinerSubmit(event Event) {
	submitMsg := event.Data.(*StratumSubmit)

	// if can't find job, could do nothing
	workHash, _ := types.NewHash(submitMsg.JobID)
	work := r.allWorks[workHash]
	if work == nil {
		log.Errorf("work not exist for job %s", submitMsg.JobID)
		r.rejectMinerSubmit(submitMsg, 21, "Job not found")
		return
	}

	if submitMsg.NTime < work.MinTime {
		log.Errorf("NTime %d less than MinTime %d", submitMsg.NTime, work.MinTime)
		r.rejectMinerSubmit(submitMsg, 20, "Invalid time")
		return
	}

	js := new(JobSubmit)
	js.WorkerName = submitMsg.WorkerName
	js.ExtraNonce1 = submitMsg.ExtraNonce1
	js.ExtraNonce2 = submitMsg.ExtraNonce2
	js.Nonce = submitMsg.Nonce
	js.NTime = submitMsg.NTime

	checkDup := false
	for _, oldJs := range work.submits {
		if oldJs.ExtraNonce1 == js.ExtraNonce1 &&
			oldJs.ExtraNonce2 == js.ExtraNonce2 &&
			oldJs.Nonce == js.Nonce &&
			oldJs.NTime == js.NTime {
			checkDup = true
			break
		}
	}
	if checkDup {
		log.Errorf("duplicate submit for work %s", work.WorkHash)
		r.rejectMinerSubmit(submitMsg, 22, "Duplicate share")
		return
	}
	work.submits = append(work.submits, js)

	log.Infof("new submit for work %s, [%08x, %016x, %08x, %08x]", work.WorkHash, js.ExtraNonce1, js.ExtraNonce2, js.Nonce, js.NTime)

	//extra = coinbase info + extraNonce1 + extraNonce2
	extraNonce1Bytes := util.LE_Uint32ToBytes(submitMsg.ExtraNonce1)
	extraNonce2Bytes := util.LE_Uint64ToBytes(submitMsg.ExtraNonce2)
	cbExtraBuf := new(bytes.Buffer)
	cbExtraBuf.Write(work.CBPoolInfo)
	cbExtraBuf.Write(extraNonce1Bytes)
	cbExtraBuf.Write(extraNonce2Bytes)
	work.CoinbaseExtra = cbExtraBuf.Bytes()

	cbTxDataBuf := new(bytes.Buffer)
	cbTxDataBuf.Write(work.CoinBaseData1)
	cbTxDataBuf.Write(util.LE_EncodeVarInt(uint64(work.CBExtraLen)))
	cbTxDataBuf.Write(work.CoinbaseExtra)
	cbTxDataBuf.Write(work.CoinBaseData2)
	cbTxHash, _ := types.Sha256D_HashBytes(cbTxDataBuf.Bytes())
	work.CoinbaseHash = cbTxHash

	work.MerkleRoot = merkle.CalcCoinbaseMerkleRoot(&cbTxHash, work.MerkleBranch)

	log.Debugf("CoinbaseExtra:%s", hex.EncodeToString(work.CoinbaseExtra))
	log.Debugf("CoinbaseHash:%s, MerkleRoot:%s", work.CoinbaseHash, work.MerkleRoot)

	hdr := work.BuildBlockHeader(js)
	work.BlockHash = hdr.ComputeHash()

	powHash := hdr.ComputePowHash()

	log.Debugf("BlockHash:%s, PowHash:%s", work.BlockHash, powHash)

	// check proof-of-work difficulty
	powInt := types.HashToBig(&powHash)
	targetInt := hdr.GetAlgoTargetInt()
	if powInt.Cmp(targetInt) > 0 {
		log.Warnf("PowHash(%064x) greater than Bits(%064x)", powInt, targetInt)
		r.rejectMinerSubmit(submitMsg, 23, "Low difficulty share")
		return
	}

	minerRsp := new(StratumRsp)
	minerRsp.SessionID = submitMsg.SessionID
	minerRsp.MsgID = submitMsg.MsgID
	minerRsp.Result = true
	GetDefaultEventBus().Publish(EventMinerSendRsp, minerRsp)

	r.acceptedSubmit++

	GetDefaultEventBus().Publish(EventJobSubmit, work)
}

func (r *JobRepository) rejectMinerSubmit(submitMsg *StratumSubmit, errCode int, errMsg string) {
	r.rejectedSubmit++

	minerRsp := new(StratumRsp)
	minerRsp.SessionID = submitMsg.SessionID
	minerRsp.MsgID = submitMsg.MsgID
	minerRsp.ErrCode = errCode
	minerRsp.ErrMsg = errMsg
	GetDefaultEventBus().Publish(EventMinerSendRsp, minerRsp)
}

func (r *JobRepository) consumeStatisticsTicker(event Event) {
	log.Infof("job works: allWorks:%d, prevWorks:%d", len(r.allWorks), len(r.prevWorks))
	log.Infof("job submits: accepted:%d, rejected:%d", r.acceptedSubmit, r.rejectedSubmit)
}

func (r *JobRepository) OnCheckJobTicker() {
	r.tryCleanExpiredJobs()

	r.checkAndSendMiningNotify()
}

func (r *JobRepository) checkAndSendMiningNotify() {}

func (r *JobRepository) tryCleanExpiredJobs() {
	nowSec := uint32(time.Now().Unix())

	for ph, pws := range r.prevWorks {
		if len(pws) <= 0 {
			delete(r.prevWorks, ph)
			continue
		}

		// check latest pending work is expired or not
		lw := pws[len(pws)-1]
		if nowSec < (lw.PoolTime + MaxWorkExpireSec) {
			continue
		}

		// clean all expired work for the same previous
		log.Warnf("clean expired works %d for PrevHash %s", len(pws), ph)
		for _, w := range pws {
			delete(r.allWorks, w.JobHash)
			if w.JobHash == r.lastJobWork.JobHash {
				r.lastJobWork = nil
			}
		}
		delete(r.prevWorks, ph)
	}
}

type JobSubmit struct {
	WorkerName  string
	ExtraNonce1 uint32
	ExtraNonce2 uint64
	NTime       uint32
	Nonce       uint32
}

type JobWork struct {
	JobHash  types.Hash
	CreateAt time.Time
	CheckAt  time.Time

	// fields from node's getWork api
	WorkHash      types.Hash
	Version       uint32
	PrevHash      types.Hash
	Bits          uint32
	MerkleBranch  []*types.Hash
	CoinBaseData1 []byte
	CoinBaseData2 []byte
	Height        uint64
	MinTime       uint32

	// all miner's submit
	submits []*JobSubmit

	// fileds for pool internal using
	PoolTime   uint32
	Difficulty float64
	CBPoolInfo []byte
	CBExtraLen int
	CleanJobs  bool

	// fileds for node's submit work
	CoinbaseExtra []byte
	CoinbaseHash  types.Hash
	MerkleRoot    types.Hash
	BlockHash     types.Hash
}

func (w *JobWork) BuildBlockHeader(js *JobSubmit) *types.PovHeader {
	header := types.NewPovHeader()
	header.BasHdr.Version = w.Version
	header.BasHdr.Previous = w.PrevHash
	header.BasHdr.MerkleRoot = w.MerkleRoot
	header.BasHdr.Timestamp = js.NTime
	header.BasHdr.Bits = w.Bits
	header.BasHdr.Nonce = js.Nonce

	if log.GetLevel() >= log.DebugLevel {
		data := header.BuildHashData()
		log.Debugf("HashData:%s", hex.EncodeToString(data))
	}

	return header
}

func (w *JobWork) ToNotify() *StratumNotify {
	nf := new(StratumNotify)
	nf.JobID = w.JobHash.String()
	// API PrevHash is LE, Miner PrevHash is Special LE, so LE -> BE -> SLE
	nf.PrevHash = w.PrevHash.ReverseByte().ReverseEndian().String()

	nfCB1Buf := new(bytes.Buffer)
	nfCB1Buf.Write(w.CoinBaseData1)
	nfCB1Buf.Write(util.LE_EncodeVarInt(uint64(w.CBExtraLen)))
	nfCB1Buf.Write(w.CBPoolInfo)

	nf.Coinbase1 = hex.EncodeToString(nfCB1Buf.Bytes())
	nf.Coinbase2 = hex.EncodeToString(w.CoinBaseData2)

	nf.MerkleBranch = make([]string, 0)
	for _, mbHash := range w.MerkleBranch {
		mbHashHex := mbHash.String()
		nf.MerkleBranch = append(nf.MerkleBranch, mbHashHex)
	}

	nf.Version = UInt32ToHexBE(w.Version)
	nf.NBits = UInt32ToHexBE(w.Bits)
	nf.NTime = UInt32ToHexBE(w.PoolTime)
	nf.CleanJobs = w.CleanJobs

	nf.Difficulty = w.Difficulty

	return nf
}

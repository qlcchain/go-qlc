package privacy

import (
	"errors"
	"runtime"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/bluele/gcache"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
)

const (
	ptmNodeUnknown = iota
	ptmNodeRunning
	ptmNodeOffline
)

type PTM struct {
	cfg        *config.Config
	logger     *zap.SugaredLogger
	clientPool sync.Pool
	mainClient *Client
	cache      gcache.Cache
	status     atomic.Int32
	quitCh     chan struct{}

	statRspOk  atomic.Uint64
	statRspErr atomic.Uint64

	fakeMode bool
}

var (
	ErrPtmNodeNotRunning = errors.New("ptm node not running")
	ErrPtmAPIFailed      = errors.New("ptm api failed")
	ErrPtmClientNull     = errors.New("ptm client null")
)

func NewPTM(cfg *config.Config) *PTM {
	m := &PTM{cfg: cfg}
	m.clientPool.New = func() interface{} {
		return NewClient(m.cfg.Privacy.PtmNode)
	}
	m.status.Store(ptmNodeUnknown)
	return m
}

func (m *PTM) Init() error {
	m.cache = gcache.New(common.DPoSMaxBlocks).LRU().Build()
	m.logger = log.NewLogger("privacy_ptm")

	m.mainClient = NewClient(m.cfg.Privacy.PtmNode)
	if m.mainClient == nil {
		return errors.New("invalid ptm node")
	}

	nCPU := runtime.NumCPU()
	for i := 0; i < nCPU; i++ {
		c := NewClient(m.cfg.Privacy.PtmNode)
		if c == nil {
			return errors.New("invalid ptm node")
		}
		m.clientPool.Put(c)
	}

	return nil
}

func (m *PTM) Start() error {
	m.quitCh = make(chan struct{})

	m.logger.Info("ptm node at ", m.cfg.Privacy.PtmNode)

	common.Go(m.mainLoop)

	return nil
}

func (m *PTM) Stop() error {
	close(m.quitCh)
	return nil
}

func (m *PTM) Send(data []byte, from string, to []string) (out []byte, err error) {
	if m.status.Load() != ptmNodeRunning {
		return nil, ErrPtmNodeNotRunning
	}

	cli := m.acquireClient()
	if cli == nil {
		return nil, ErrPtmClientNull
	}
	defer m.releaseClient(cli)

	out, err = cli.SendPayload(data, from, to)
	if err != nil {
		m.statRspErr.Inc()
		m.logger.Infof("failed to send payload, %s", err)
		return nil, ErrPtmAPIFailed
	}
	m.statRspOk.Inc()

	_ = m.cache.Set(string(out), data)
	return out, nil
}

func (m *PTM) Receive(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	if m.status.Load() != ptmNodeRunning {
		return nil, ErrPtmNodeNotRunning
	}

	dataStr := string(data)
	x, err := m.cache.Get(dataStr)
	if err == nil {
		if x == nil {
			return nil, nil
		}
		return x.([]byte), nil
	}

	// Ignore this error since not being a recipient of
	// a payload isn't an error.
	// TODO: Return an error if it's anything OTHER than
	// 'you are not a recipient.'

	cli := m.acquireClient()
	if cli == nil {
		return nil, ErrPtmClientNull
	}
	defer m.releaseClient(cli)

	pl, err := cli.ReceivePayload(data)
	if err != nil {
		m.statRspErr.Inc()
		m.logger.Infof("failed to recv payload, %s", err)
		return nil, ErrPtmAPIFailed
	}
	m.statRspOk.Inc()
	if len(pl) == 0 {
		pl = nil
	}

	_ = m.cache.Set(dataStr, pl)

	return pl, nil
}

func (m *PTM) mainLoop() {
	upChkTicker := time.NewTicker(10 * time.Second)
	defer upChkTicker.Stop()

	m.onUpCheckTicker()

	for {
		select {
		case <-m.quitCh:
			return
		case <-upChkTicker.C:
			m.onUpCheckTicker()
		}
	}
}

func (m *PTM) onUpCheckTicker() {
	oldStatus := m.status.Load()
	newStatus := int32(ptmNodeUnknown)

	chkOk, chkErr := m.mainClient.Upcheck()
	if chkOk {
		newStatus = int32(ptmNodeRunning)
	} else {
		newStatus = int32(ptmNodeOffline)
	}

	if oldStatus == newStatus {
		return
	}

	if newStatus == int32(ptmNodeRunning) {
		m.logger.Infof("ptm node is online, url %s", m.cfg.Privacy.PtmNode)
	} else {
		m.logger.Errorf("ptm node is offline, url [%s], err [%s]", m.cfg.Privacy.PtmNode, chkErr)
	}

	m.status.Store(newStatus)
}

func (m *PTM) acquireClient() *Client {
	if m.fakeMode {
		return m.mainClient
	}
	v := m.clientPool.Get()
	if v != nil {
		return v.(*Client)
	}
	return nil
}

func (m *PTM) releaseClient(c *Client) {
	if m.fakeMode {
		return
	}
	m.clientPool.Put(c)
}

func (m *PTM) GetDebugInfo() map[string]interface{} {
	info := make(map[string]interface{})
	info["status"] = m.status.Load()
	info["cache"] = m.cache.Len(false)
	info["statRspOk"] = m.statRspOk.Load()
	info["statRspErr"] = m.statRspErr.Load()
	return info
}

func (m *PTM) SetFakeMode(mode bool) {
	m.fakeMode = mode
	m.mainClient.transport.SetFakeMode(mode)
}

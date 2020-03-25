package privacy

import (
	"errors"
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
	cfg    *config.Config
	logger *zap.SugaredLogger
	client *Client
	cache  gcache.Cache
	status atomic.Int32
	quitCh chan struct{}
}

var (
	errPtmNodeNotRunning = errors.New("PTM node not running")
)

func NewPTM(cfg *config.Config) *PTM {
	m := &PTM{cfg: cfg}
	m.status.Store(ptmNodeUnknown)
	return m
}

func (m *PTM) Init() error {
	m.cache = gcache.New(10000).LRU().Expiration(5 * time.Minute).Build()
	m.logger = log.NewLogger("privacy_ptm")
	return nil
}

func (m *PTM) Start() error {
	m.client = NewClient(m.cfg.Privacy.PtmNode)
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
		return nil, errPtmNodeNotRunning
	}
	out, err = m.client.SendPayload(data, from, to)
	if err != nil {
		return nil, err
	}
	m.cache.Set(string(out), data)
	return out, nil
}

func (m *PTM) SendSignedTx(data []byte, to []string) (out []byte, err error) {
	if m.status.Load() != ptmNodeRunning {
		return nil, errPtmNodeNotRunning
	}
	out, err = m.client.SendSignedPayload(data, to)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *PTM) Receive(data []byte) ([]byte, error) {
	if m.status.Load() != ptmNodeRunning {
		return nil, errPtmNodeNotRunning
	}
	if len(data) == 0 {
		return data, nil
	}
	// Ignore this error since not being a recipient of
	// a payload isn't an error.
	// TODO: Return an error if it's anything OTHER than
	// 'you are not a recipient.'
	dataStr := string(data)
	x, err := m.cache.Get(dataStr)
	if err != nil {
		return x.([]byte), nil
	}
	pl, _ := m.client.ReceivePayload(data)
	m.cache.Set(dataStr, pl)
	return pl, nil
}

func (m *PTM) mainLoop() {
	upChkTicker := time.NewTicker(time.Second)
	defer upChkTicker.Stop()

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

	chkOk, chkErr := m.client.Upcheck()
	if chkOk {
		newStatus = int32(ptmNodeRunning)
	} else {
		newStatus = int32(ptmNodeOffline)
	}

	if oldStatus == newStatus {
		return
	}

	if newStatus == int32(ptmNodeRunning) {
		m.logger.Info("ptm node is online")
	} else {
		m.logger.Error("ptm node is offline, err", chkErr)
	}

	m.status.Store(newStatus)
}

package privacy

import (
	"errors"
	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/config"
	"time"
)

type PTM struct {
	cfg         *config.Config
	client      *Client
	cache       gcache.Cache
	isNotUsePTM bool
}

var (
	errPTMNotUsed = errors.New("PTM not in use")
)

func (m *PTM) Init() error {
	m.cache = gcache.New(10000).LRU().Expiration(5 * time.Minute).Build()
	return nil
}

func (m *PTM) Start() error {
	err := RunNode(m.cfg.Privacy.PtmNode)
	if err != nil {
		return nil
	}

	m.client, err = NewClient(m.cfg.Privacy.PtmNode)
	if err != nil {
		return err
	}

	return nil
}

func (g *PTM) Stop() error {
	return nil
}

func (g *PTM) Send(data []byte, from string, to []string) (out []byte, err error) {
	if g.isNotUsePTM {
		return nil, errPTMNotUsed
	}
	out, err = g.client.SendPayload(data, from, to)
	if err != nil {
		return nil, err
	}
	g.cache.Set(string(out), data)
	return out, nil
}

func (g *PTM) SendSignedTx(data []byte, to []string) (out []byte, err error) {
	if g.isNotUsePTM {
		return nil, errPTMNotUsed
	}
	out, err = g.client.SendSignedPayload(data, to)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (g *PTM) Receive(data []byte) ([]byte, error) {
	if g.isNotUsePTM {
		return nil, nil
	}
	if len(data) == 0 {
		return data, nil
	}
	// Ignore this error since not being a recipient of
	// a payload isn't an error.
	// TODO: Return an error if it's anything OTHER than
	// 'you are not a recipient.'
	dataStr := string(data)
	x, err := g.cache.Get(dataStr)
	if err != nil {
		return x.([]byte), nil
	}
	pl, _ := g.client.ReceivePayload(data)
	g.cache.Set(dataStr, pl)
	return pl, nil
}

func NewPTM(cfg *config.Config) *PTM {
	return &PTM{cfg: cfg}
}

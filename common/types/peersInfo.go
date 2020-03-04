package types

//go:generate msgp
type PeerInfo struct {
	PeerID         string  `json:"peerid"`
	Address        string  `json:"address"`
	Version        string  `json:"version"`
	Rtt            float64 `json:"rtt"`
	LastUpdateTime string  `json:"lastUpdateTime"`
}

func (p *PeerInfo) Serialize() ([]byte, error) {
	return p.MarshalMsg(nil)
}

func (p *PeerInfo) Deserialize(text []byte) error {
	_, err := p.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

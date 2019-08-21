package types

//go:generate msgp

type PovBtcHeader struct {
	Version    uint32 `msg:"v" json:"version"`
	Previous   Hash   `msg:"p,extension" json:"previous"`
	MerkleRoot Hash   `msg:"mr,extension" json:"merkleRoot"`
	Timestamp  uint32 `msg:"ts" json:"timestamp"`
	Bits       uint32 `msg:"b" json:"bits"`
	Nonce      uint32 `msg:"n" json:"nonce"`
}

func (h *PovBtcHeader) Serialize() ([]byte, error) {
	return h.MarshalMsg(nil)
}

func (h *PovBtcHeader) Deserialize(text []byte) error {
	_, err := h.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovBtcOutPoint struct {
	Hash  Hash   `msg:"h,extension" json:"hash"`
	Index uint32 `msg:"i" json:"index"`
}

func (p *PovBtcOutPoint) Serialize() ([]byte, error) {
	return p.MarshalMsg(nil)
}

func (p *PovBtcOutPoint) Deserialize(text []byte) error {
	_, err := p.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovBtcTxIn struct {
	PreviousOutPoint PovBtcOutPoint `msg:"pop" json:"previousOutPoint"`
	SignatureScript  []byte         `msg:"ss" json:"signatureScript"`
	Sequence         uint32         `msg:"s" json:"sequence"`
}

func (ti *PovBtcTxIn) Serialize() ([]byte, error) {
	return ti.MarshalMsg(nil)
}

func (ti *PovBtcTxIn) Deserialize(text []byte) error {
	_, err := ti.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovBtcTxOut struct {
	Value    int64  `msg:"v" json:"value"`
	PkScript []byte `msg:"pks" json:"pkScript"`
}

func (to *PovBtcTxOut) Serialize() ([]byte, error) {
	return to.MarshalMsg(nil)
}

func (to *PovBtcTxOut) Deserialize(text []byte) error {
	_, err := to.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovBtcTx struct {
	Version  int32          `msg:"v" json:"version"`
	TxIn     []*PovBtcTxIn  `msg:"ti" json:"txIn"`
	TxOut    []*PovBtcTxOut `msg:"to" json:"txOut"`
	LockTime uint32         `msg:"lt" json:"lockTime"`
}

func (tx *PovBtcTx) Serialize() ([]byte, error) {
	return tx.MarshalMsg(nil)
}

func (tx *PovBtcTx) Deserialize(text []byte) error {
	_, err := tx.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

package types

import (
	"bytes"
	"github.com/qlcchain/go-qlc/common/util"
)

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

func (h *PovBtcHeader) BuildHashData() []byte {
	buf := new(bytes.Buffer)
	buf.Write(util.LE_Uint32ToBytes(h.Version))
	buf.Write(h.Previous.Bytes())
	buf.Write(h.MerkleRoot.Bytes())
	buf.Write(util.LE_Uint32ToBytes(h.Timestamp))
	buf.Write(util.LE_Uint32ToBytes(h.Bits))
	buf.Write(util.LE_Uint32ToBytes(h.Nonce))
	return buf.Bytes()
}

func (h *PovBtcHeader) ComputePowHash(algo PovAlgoType) Hash {
	d := h.BuildHashData()

	switch algo {
	case ALGO_SHA256D:
		powHash := Sha256D_HashData(d)
		return powHash
	case ALGO_SCRYPT:
		powHash := Scrypt_HashData(d)
		return powHash
	case ALGO_X11:
		powHash := X11_HashData(d)
		return powHash
	}

	return FFFFHash
}

func (h *PovBtcHeader) ComputeHash() Hash {
	d := h.BuildHashData()
	powHash := Sha256D_HashData(d)
	return powHash
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

func (ti *PovBtcTxIn) BuildHashData() []byte {
	buf := new(bytes.Buffer)
	buf.Write(ti.PreviousOutPoint.Hash.Bytes())
	buf.Write(util.LE_Uint32ToBytes(ti.PreviousOutPoint.Index))
	buf.Write(util.LE_EncodeVarInt(uint64(len(ti.SignatureScript))))
	buf.Write(ti.SignatureScript)
	buf.Write(util.LE_Uint32ToBytes(ti.Sequence))
	return buf.Bytes()
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

func (to *PovBtcTxOut) BuildHashData() []byte {
	buf := new(bytes.Buffer)
	buf.Write(util.LE_Uint64ToBytes(uint64(to.Value)))
	buf.Write(util.LE_EncodeVarInt(uint64(len(to.PkScript))))
	buf.Write(to.PkScript)
	return buf.Bytes()
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

func (tx *PovBtcTx) BuildHashData() []byte {
	buf := new(bytes.Buffer)
	buf.Write(util.LE_Uint32ToBytes(uint32(tx.Version)))

	buf.Write(util.LE_EncodeVarInt(uint64(len(tx.TxIn))))
	for _, ti := range tx.TxIn {
		buf.Write(ti.BuildHashData())
	}

	buf.Write(util.LE_EncodeVarInt(uint64(len(tx.TxOut))))
	for _, to := range tx.TxOut {
		buf.Write(to.BuildHashData())
	}

	buf.Write(util.LE_Uint32ToBytes(tx.LockTime))
	return buf.Bytes()
}

func (tx *PovBtcTx) ComputeHash() Hash {
	d := tx.BuildHashData()
	powHash := Sha256D_HashData(d)
	return powHash
}

func NewPovBtcTx(txIn []*PovBtcTxIn, txOut []*PovBtcTxOut) *PovBtcTx {
	tx := &PovBtcTx{
		Version:  1,
		TxIn:     make([]*PovBtcTxIn, 0),
		TxOut:    make([]*PovBtcTxOut, 0),
		LockTime: 0,
	}

	tx.TxIn = append(tx.TxIn, txIn...)
	tx.TxOut = append(tx.TxOut, txOut...)

	return tx
}

package relation

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/relation/db"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type Relation struct {
	store  db.DbStore
	logger *zap.SugaredLogger
}

type blocksHash struct {
	Id        int64
	Hash      string
	Type      string
	Address   string
	Timestamp int64
}

type blocksMessage struct {
	ID        int64
	Hash      string
	Sender    string
	Receiver  string
	Message   string
	Timestamp int64
}

func NewRelation(dir string) (*Relation, error) {
	s, err := db.NewSQLDB(dir)
	if err != nil {
		return nil, err
	}
	return &Relation{store: s, logger: log.NewLogger("relation")}, nil
}

func (r *Relation) Close() error {
	panic("implement me")
}

func (r *Relation) AccountBlocks(address types.Address, limit int, offset int) ([]types.Hash, error) {
	condition := make(map[db.Column]interface{})
	condition[db.ColumnAddress] = address.String()
	var h []blocksHash
	err := r.store.Read(db.TableBlockHash, condition, offset, limit, db.ColumnTimestamp, &h)
	if err != nil {
		return nil, err
	}
	return blockHash(h)
}

func (r *Relation) BlocksCount() (uint64, error) {
	var count uint64
	err := r.store.Count(db.TableBlockHash, &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

type blocksType struct {
	Type  string
	Count uint64
}

func (r *Relation) BlocksCountByType() (map[string]uint64, error) {
	var t []blocksType
	err := r.store.Group(db.TableBlockHash, db.ColumnType, &t)
	if err != nil {
		return nil, err
	}
	return blockType(t), nil
}

func (r *Relation) Blocks(limit int, offset int) ([]types.Hash, error) {
	var h []blocksHash
	err := r.store.Read(db.TableBlockHash, nil, offset, limit, db.ColumnTimestamp, &h)
	if err != nil {
		return nil, err
	}
	return blockHash(h)
}

func (r *Relation) PhoneBlocks(phone []byte) ([]types.Hash, error) {
	condition := make(map[db.Column]interface{})
	condition[db.ColumnSender] = byteToString(phone)
	condition[db.ColumnReceiver] = byteToString(phone)
	var h []blocksMessage
	err := r.store.Read(db.TableBlockHash, condition, -1, -1, db.ColumnNoNeed, &h)
	if err != nil {
		return nil, err
	}
	return blockMessage(h)
}

func (r *Relation) MessageBlocks(hash types.Hash) ([]types.Hash, error) {
	condition := make(map[db.Column]interface{})
	condition[db.ColumnMessage] = hash.String()
	var h []blocksMessage
	err := r.store.Read(db.TableBlockHash, condition, -1, -1, db.ColumnNoNeed, &h)
	if err != nil {
		return nil, err
	}
	return blockMessage(h)
}

func (r *Relation) AddBlock(block *types.StateBlock) error {
	conHash := make(map[db.Column]interface{})
	conHash[db.ColumnHash] = block.GetHash().String()
	conHash[db.ColumnTimestamp] = block.Timestamp
	conHash[db.ColumnType] = block.GetType().String()
	conHash[db.ColumnAddress] = block.GetAddress().String()
	if err := r.store.Create(db.TableBlockHash, conHash); err != nil {
		return err
	}
	message := block.GetMessage()
	if block.GetSender() != nil || block.GetReceiver() != nil || !message.IsZero() {
		conMessage := make(map[db.Column]interface{})
		conMessage[db.ColumnHash] = block.GetHash().String()
		conMessage[db.ColumnMessage] = message.String()
		conMessage[db.ColumnSender] = byteToString(block.GetSender())
		conMessage[db.ColumnReceiver] = byteToString(block.GetReceiver())
		conHash[db.ColumnTimestamp] = block.Timestamp
		if err := r.store.Create(db.TableBlockMessage, conMessage); err != nil {
			return err
		}
	}
	return nil
}

func (r *Relation) DeleteBlock(hash types.Hash) error {
	condition := make(map[db.Column]interface{})
	condition[db.ColumnHash] = hash.String()
	err := r.store.Delete(db.TableBlockHash, condition)
	if err != nil {
		return err
	}
	return r.store.Delete(db.TableBlockMessage, condition)
}

func byteToString(b []byte) string {
	if b == nil {
		return ""
	}
	return string(b)
}

func blockHash(bs []blocksHash) ([]types.Hash, error) {
	hs := make([]types.Hash, 0)
	for _, b := range bs {
		var h types.Hash
		if err := h.Of(b.Hash); err != nil {
			return nil, err
		}
		hs = append(hs, h)
	}
	return hs, nil
}

func blockMessage(bs []blocksMessage) ([]types.Hash, error) {
	hs := make([]types.Hash, 0)
	for _, b := range bs {
		var h types.Hash
		if err := h.Of(b.Hash); err != nil {
			return nil, err
		}
		hs = append(hs, h)
	}
	return hs, nil
}

func blockType(bs []blocksType) map[string]uint64 {
	t := make(map[string]uint64)
	for _, b := range bs {
		t[b.Type] = b.Count
	}
	return t
}

package relation

import (
	"fmt"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
)

type schema struct {
	tableName string
	create    string
	insert    string
	delete    string
}

type Table interface {
	TableID() string
	DeleteKey() string
}

type BlockHash struct {
	Id        int64
	Hash      string `db:"hash" typ:"char(64)"`
	Type      string `db:"type"  typ:"varchar(15)"`
	Address   string `db:"address" typ:"char(64)"`
	Timestamp int64  `db:"timestamp" typ:"integer"`
}

func (s *BlockHash) TableID() string {
	return "blockhash"
}

func (s *BlockHash) DeleteKey() string {
	return fmt.Sprintf("DELETE FROM BlockHash WHERE Hash = '%s'", s.Hash)
}

func convertSchemaType(db string, typ string) string {
	switch db {
	case "sqlite", "sqlite3":
		return typ
	case "mysql":
		// TODO mysql type may different
		return typ
	default:
		return "varchar(100)"
	}
}

func create(tableName string, columns map[string]string, key string) string {
	cs := ""
	for k, v := range columns {
		cs = cs + fmt.Sprintf(" %s %s ,", k, v)
	}
	cs = strings.TrimRight(cs, ",")
	if key == "" {
		return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
		( id integer PRIMARY KEY AUTOINCREMENT, %s)`, tableName, cs)
	} else {
		//TODO set with primary key
		return ""
	}
}

func insert(tableName string, columns []string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (:%s)", tableName, strings.Join(columns, ","), strings.Join(columns, ", :"))
}

func TableConvert(obj interface{}) Table {
	switch obj.(type) {
	case *types.StateBlock:
		blk := obj.(*types.StateBlock)
		return &BlockHash{
			Type:      blk.Type.String(),
			Address:   blk.Address.String(),
			Timestamp: blk.Timestamp,
			Hash:      blk.GetHash().String(),
		}
	default:
		return obj.(Table)
	}
}
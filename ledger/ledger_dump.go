package ledger

import (
	"encoding/hex"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Block struct {
	Type           types.BlockType `msg:"type" json:"type"`
	Token          types.Hash      `msg:"token,extension" json:"token"`
	Address        types.Address   `msg:"address,extension" json:"address"`
	Balance        types.Balance   `msg:"balance,extension" json:"balance"`
	Vote           types.Balance   `msg:"vote,extension" json:"vote"`
	Network        types.Balance   `msg:"network,extension" json:"network"`
	Storage        types.Balance   `msg:"storage,extension" json:"storage"`
	Oracle         types.Balance   `msg:"oracle,extension" json:"oracle"`
	Previous       types.Hash      `msg:"previous,extension" json:"previous"`
	Link           types.Hash      `msg:"link,extension" json:"link"`
	Sender         []byte          `msg:"sender" json:"sender,omitempty"`
	Receiver       []byte          `msg:"receiver" json:"receiver,omitempty"`
	Message        types.Hash      `msg:"message,extension" json:"message,omitempty"`
	Data           []byte          `msg:"data" json:"data,omitempty"`
	PoVHeight      uint64          `msg:"povHeight" json:"povHeight"`
	Timestamp      int64           `msg:"timestamp" json:"timestamp"`
	Extra          types.Hash      `msg:"extra,extension" json:"extra,omitempty"`
	Representative types.Address   `msg:"representative,extension" json:"representative"`
	Work           types.Work      `msg:"work,extension" json:"work"`
	Signature      types.Signature `msg:"signature,extension" json:"signature"`
	Hash           types.Hash      `msg:"hash" json:"hash"`
	BlockHeight    int             `msg:"blockHeight" json:"blockHeight"` // 0（block is isolated）
	InChain        int             `msg:"inChain" json:"inChain"`         // 0（false）, 1（true）。
	Reason         string          `msg:"reason" json:"reason"`
}

type Account struct {
	Address        types.Address `msg:"address,extension" json:"address"`
	Token          types.Hash    `msg:"token,extension" json:"token"`
	Header         types.Hash    `msg:"header,extension" json:"header"`
	Representative types.Address `msg:"rep,extension" json:"representative"`
	OpenBlock      types.Hash    `msg:"open,extension" json:"open"`
	Balance        types.Balance `msg:"balance,extension" json:"balance"`
	Modified       int64         `msg:"modified" json:"modified"`
	BlockCount     int64         `msg:"blockCount," json:"blockCount"`
	Certain        int           `msg:"certain" json:"certain"`         // 0（false）, 1（true）。
	Consecutive    int           `msg:"consecutive" json:"consecutive"` // 0（false）, 1（true）。
}

type BlockLink struct {
	Hash  types.Hash      `msg:"hash" json:"hash"`
	Type  types.BlockType `msg:"type" json:"type"`
	Child types.Hash      `msg:"child" json:"child"`
	Link  types.Hash      `msg:"link" json:"link"`
}

func (l *Ledger) Dump(t int) (string, error) {
	p := path.Join(l.dir, "relation")
	if err := util.CreateDirIfNotExist(p); err != nil {
		return "", err
	}
	dir := path.Join(p, "dump.db")
	dirPro := path.Join(p, "dump.processing")

	if t == 1 {
		if _, err := os.Stat(dirPro); err == nil {
			return "", nil
		} else {
			return "done", nil
		}
	}

	if _, err := os.Stat(dir); err == nil {
		if err := os.Remove(dir); err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(dirPro); err == nil {
		if err := os.Remove(dirPro); err != nil {
			return "", err
		}
	}

	file, err := os.Create(dirPro)
	if err != nil {
		return "", err
	}

	db, err := sqlx.Connect("sqlite3", dir)
	if err != nil {
		return "", fmt.Errorf("connect sqlite error: %s", err)
	}
	go func() {
		defer func() {
			if err := file.Close(); err != nil {
				l.logger.Error(err)
			}
			if err := os.Remove(dirPro); err != nil {
				l.logger.Error(err)
			}
		}()
		if err := l.dump(db); err != nil {
			l.logger.Error(err)
		}
		fmt.Println("dump done")
	}()
	return dir, nil
}

func (l *Ledger) dump(db *sqlx.DB) error {
	if err := initDumpTables(db); err != nil {
		return fmt.Errorf("init dump tables error: %s", err)
	}
	if err := l.dumpBlock(db); err != nil {
		return fmt.Errorf("dump block error: %s", err)
	}
	if err := l.dumpBlockCache(db); err != nil {
		return fmt.Errorf("dump block cache error: %s", err)
	}
	if err := l.dumpBlockLink(db); err != nil {
		return fmt.Errorf("dump block link error: %s", err)
	}
	if err := l.dumpBlockUnchecked(db); err != nil {
		return fmt.Errorf("dump block unchecked error: %s", err)
	}
	return nil
}

func (l *Ledger) dumpBlock(db *sqlx.DB) error {
	batchAccounts := make([]*types.AccountMeta, 0)
	err := l.GetAccountMetas(func(am *types.AccountMeta) error {
		batchAccounts = append(batchAccounts, am)
		if len(batchAccounts) > 200 {
			if err := l.dumpBlockByAccount(batchAccounts, db); err != nil {
				return err
			}
			batchAccounts = batchAccounts[:0]
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(batchAccounts) > 0 {
		if err := l.dumpBlockByAccount(batchAccounts, db); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) dumpBlockByAccount(batchAccounts []*types.AccountMeta, db *sqlx.DB) error {
	allBlocks := make(map[string][]*Block, 0)
	allAccounts := make([]*Account, 0)

	batchAccountsMap := make(map[types.Address]int, 0)
	for _, a := range batchAccounts {
		batchAccountsMap[a.Address] = 1
	}

	if err := l.GetStateBlocksConfirmed(func(block *types.StateBlock) error {
		if _, ok := batchAccountsMap[block.Address]; ok {
			hash, _ := types.HashBytes(block.Address[:], block.Token[:])
			allBlocks[hash.String()] = append(allBlocks[hash.String()], convertToDumpBlock(block))
		}
		return nil
	}); err != nil {
		return err
	}

	for _, acc := range batchAccounts {
		for _, tm := range acc.Tokens {
			a := convertToDumpAccount(tm)
			tokenBlocksMap := make(map[types.Hash]*Block)
			tokenOffChainBlocks := make([]types.Hash, 0)

			hash, _ := types.HashBytes(tm.BelongTo[:], tm.Type[:])
			for _, blk := range allBlocks[hash.String()] {
				tokenBlocksMap[blk.Hash] = blk
			}

			curHash := tm.Header
			curBlock := tokenBlocksMap[curHash]
			index := 1
			for {
				curBlock = tokenBlocksMap[curHash]
				if curBlock == nil {
					l.logger.Error("block is not found: ", curHash.String())
					break
				} else {
					curBlock.BlockHeight = index
					curBlock.InChain = 1

					curHash = curBlock.Previous
					if curHash == types.ZeroHash {
						if curBlock.Hash == tm.OpenBlock {
							a.Consecutive = 1
						}
						break
					}
					index = index + 1
				}
			}

			for h, b := range tokenBlocksMap {
				if b.InChain == 0 {
					tokenOffChainBlocks = append(tokenOffChainBlocks, h)
				}
			}

			if len(tokenOffChainBlocks) > 0 {
				for _, hash := range tokenOffChainBlocks {
					if !tokenBlocksMap[hash].Previous.IsZero() {
						if tokenBlocksMap[tokenBlocksMap[hash].Previous].InChain == 1 {
							tokenBlocksMap[hash].BlockHeight = tokenBlocksMap[tokenBlocksMap[hash].Previous].BlockHeight - 1

							chash := hash
							for {
								findPre := false
								for _, b := range tokenOffChainBlocks {
									if tokenBlocksMap[hash].Previous == chash {
										tokenBlocksMap[b].BlockHeight = tokenBlocksMap[chash].BlockHeight - 1
										chash = b
										findPre = true
										break
									}
								}
								if !findPre {
									break
								}
							}
						}
					}
				}
			}

			totalCount := index + 1
			for _, b := range tokenBlocksMap {
				b.BlockHeight = totalCount - b.BlockHeight
			}

			if err := dumpBlocks(tokenBlocksMap, tableBlock, db); err != nil {
				return err
			}

			if len(tokenOffChainBlocks) > 0 {
				a.Certain = 0
			} else {
				a.Certain = 1
			}
			allAccounts = append(allAccounts, a)
		}
	}
	if err := dumpAccounts(allAccounts, tableBlock, db); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) dumpBlockCache(db *sqlx.DB) error {
	blockCacheMap := make(map[types.Hash]*Block)
	err := l.GetBlockCaches(func(block *types.StateBlock) error {
		b := convertToDumpBlock(block)
		blockCacheMap[block.GetHash()] = b
		return nil
	})
	if err != nil {
		return err
	}

	if len(blockCacheMap) > 0 {
		for _, blk := range blockCacheMap {
			confirmed, _ := l.HasStateBlockConfirmed(blk.Previous)
			if confirmed {
				blockCacheMap[blk.Hash].BlockHeight = 1
				cBlk := blk
				for {
					findPre := false
					for _, b := range blockCacheMap {
						if b.Previous == cBlk.Hash {
							blockCacheMap[b.Hash].BlockHeight = blockCacheMap[cBlk.Hash].BlockHeight + 1
							cBlk = b
							findPre = true
							break
						}
					}
					if !findPre {
						break
					}
				}
			}
		}
		if err := dumpBlocks(blockCacheMap, tableBlockCache, db); err != nil {
			return err
		}
	}

	if err := l.dumpAccountCache(blockCacheMap, db); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) dumpAccountCache(blockCacheMap map[types.Hash]*Block, db *sqlx.DB) error {
	accountDumps := make([]*Account, 0)
	accountBlockMap := make(map[types.Hash][]*Block)
	for _, blk := range blockCacheMap {
		acc, _ := types.HashBytes(blk.Address[:], blk.Token[:])
		if _, ok := accountBlockMap[acc]; ok {
			accountBlockMap[acc] = append(accountBlockMap[acc], blk)
		} else {
			accountBlockMap[acc] = make([]*Block, 0)
			accountBlockMap[acc] = append(accountBlockMap[acc], blk)
		}
	}
	for _, abs := range accountBlockMap {
		accountCache, _ := l.GetAccountMeteCache(abs[0].Address)
		if accountCache != nil {
			tokenCache := accountCache.Token(abs[0].Token)
			if tokenCache != nil {
				accountDump := convertToDumpAccount(tokenCache)
				sort.Slice(abs, func(i, j int) bool {
					return abs[i].BlockHeight < abs[j].BlockHeight
				})
				certain := true
				for i := 0; i < len(abs)-1; i++ {
					if abs[i].BlockHeight == abs[i+1].BlockHeight {
						certain = false
					}
				}
				if certain == true {
					accountDump.Certain = 1
				} else {
					accountDump.Certain = 0
				}
				accountDumps = append(accountDumps, accountDump)
			}
		}
	}

	err := l.GetAccountMetaCaches(func(am *types.AccountMeta) error {
		for _, tm := range am.Tokens {
			acc, _ := types.HashBytes(tm.BelongTo[:], tm.Type[:])
			if _, ok := accountBlockMap[acc]; !ok {
				accountDump := convertToDumpAccount(tm)
				accountDump.Certain = 0
				accountDumps = append(accountDumps, accountDump)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := dumpAccounts(accountDumps, tableBlockCache, db); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) dumpBlockUnchecked(db *sqlx.DB) error {
	blockUncheckedMap := make(map[types.Hash]*Block)
	err := l.GetUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
		b := convertToDumpBlock(block)
		reason := ""
		switch unCheckType {
		case types.UncheckedKindPrevious:
			reason = "GapPrevious"
		case types.UncheckedKindLink:
			reason = "GapLink"
		case types.UncheckedKindTokenInfo:
			reason = "GapTokenInfo"
		}
		b.Reason = reason
		blockUncheckedMap[block.GetHash()] = b
		return nil
	})
	if err != nil {
		return err
	}
	if err := dumpBlocks(blockUncheckedMap, tableBlockUnchecked, db); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) dumpBlockLink(db *sqlx.DB) error {
	blkLinks := make([]*BlockLink, 0)
	err := l.GetStateBlocksConfirmed(func(block *types.StateBlock) error {
		hash := block.GetHash()
		blkLink := &BlockLink{
			Hash: hash,
			Type: block.GetType(),
		}
		child, _ := l.GetBlockChild(hash)
		blkLink.Child = child
		if block.IsSendBlock() {
			link, _ := l.GetBlockLink(hash)
			blkLink.Link = link
		} else if block.IsReceiveBlock() {
			blkLink.Link = block.GetLink()
		}
		blkLinks = append(blkLinks, blkLink)
		return nil
	})
	if err != nil {
		return err
	}
	if err := dumpBlockLinks(blkLinks, db); err != nil {
		return err
	}
	return nil
}

const (
	tableBlock byte = iota
	tableBlockCache
	tableBlockUnchecked
)

func dumpBlocks(blocks map[types.Hash]*Block, table byte, db *sqlx.DB) error {
	if len(blocks) == 0 {
		return nil
	}
	block := new(Block)
	rv := reflect.ValueOf(*block)
	rt := reflect.TypeOf(*block)
	var tableName string
	switch table {
	case tableBlock:
		tableName = strings.ToLower(rt.Name())
	case tableBlockCache:
		tableName = "blockcache"
	case tableBlockUnchecked:
		tableName = "blockunchecked"
	}
	cols := make([]string, 0)
	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).CanInterface() { //判断是否为可导出字段
			cols = append(cols, rt.Field(i).Name)
		}
	}

	rows := make([][]interface{}, 0)
	for _, block := range blocks {
		cs := make([]interface{}, 0)
		rv := reflect.ValueOf(*block)

		for _, c := range cols {
			cs = append(cs, convertField(rv.FieldByName(c).Interface()))
		}
		rows = append(rows, cs)
	}

	if err := execSql(tableName, cols, rows, db); err != nil {
		return err
	}
	return nil
}

func dumpAccounts(accounts []*Account, table byte, db *sqlx.DB) error {
	if len(accounts) == 0 {
		return nil
	}
	account := new(Account)
	rv := reflect.ValueOf(*account)
	rt := reflect.TypeOf(*account)
	var tableName string
	switch table {
	case tableBlock:
		tableName = strings.ToLower(rt.Name())
	case tableBlockCache:
		tableName = "accountcache"
	}
	cols := make([]string, 0)
	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).CanInterface() { //判断是否为可导出字段
			cols = append(cols, rt.Field(i).Name)
		}
	}

	rows := make([][]interface{}, 0)
	for _, acc := range accounts {
		cs := make([]interface{}, 0)
		rv := reflect.ValueOf(*acc)

		for _, c := range cols {
			cs = append(cs, convertField(rv.FieldByName(c).Interface()))
		}
		rows = append(rows, cs)
	}

	if err := execSql(tableName, cols, rows, db); err != nil {
		return err
	}
	return nil
}

func dumpBlockLinks(bLinks []*BlockLink, db *sqlx.DB) error {
	if len(bLinks) == 0 {
		return nil
	}
	blockLink := new(BlockLink)
	rv := reflect.ValueOf(*blockLink)
	rt := reflect.TypeOf(*blockLink)
	tableName := strings.ToLower(rt.Name())

	cols := make([]string, 0)
	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).CanInterface() { //判断是否为可导出字段
			cols = append(cols, rt.Field(i).Name)
		}
	}

	rows := make([][]interface{}, 0)
	for _, b := range bLinks {
		cs := make([]interface{}, 0)
		rv := reflect.ValueOf(*b)

		for _, c := range cols {
			cs = append(cs, convertField(rv.FieldByName(c).Interface()))
		}
		rows = append(rows, cs)
	}

	if err := execSql(tableName, cols, rows, db); err != nil {
		return err
	}
	return nil
}

// create tables schema
func createTables(obj interface{}, tableName string, db *sqlx.DB) error {
	rv := reflect.ValueOf(obj)
	rt := reflect.TypeOf(obj)
	fields := getTableField(rv, rt)
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id integer PRIMARY KEY AUTOINCREMENT,  %s ) ", tableName, strings.Join(fields, ","))
	if _, err := db.Exec(sql); err != nil {
		fmt.Printf("exec create error, sql: %s, err: %s \n", sql, err.Error())
		return err
	}
	return nil
}

func initDumpTables(db *sqlx.DB) error {
	if err := createTables(Block{}, "block", db); err != nil {
		return err
	}
	if err := createTables(Account{}, "account", db); err != nil {
		return err
	}
	if err := createTables(BlockLink{}, "blocklink", db); err != nil {
		return err
	}
	if err := createTables(Block{}, "blockcache", db); err != nil {
		return err
	}
	if err := createTables(Account{}, "accountcache", db); err != nil {
		return err
	}
	if err := createTables(Block{}, "blockunchecked", db); err != nil {
		return err
	}
	return nil
}

func getTableField(rv reflect.Value, rt reflect.Type) []string {
	fields := make([]string, 0)
	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).CanInterface() {
			name := strings.ToLower(rt.Field(i).Name)
			typ := rv.Field(i).Interface()
			switch typ.(type) {
			case int64, uint64, int:
				fields = append(fields, fmt.Sprintf("%s %s", name, "integer"))
			default:
				fields = append(fields, fmt.Sprintf("%s %s", name, "text"))
			}
		}
	}
	return fields
}

// exec sql
func execSql(table string, cols []string, rows [][]interface{}, db *sqlx.DB) error {
	MaxProcess := 100

	exVals := make([][]interface{}, 0)
	for _, v := range rows {
		exVals = append(exVals, v)
		if len(exVals) == MaxProcess {
			if err := batchSql(table, cols, exVals, db); err != nil {
				return err
			}
			exVals = make([][]interface{}, 0)
		}
	}
	if len(exVals) > 0 {
		if err := batchSql(table, cols, exVals, db); err != nil {
			return err
		}
	}
	return nil
}

func batchSql(table string, cols []string, rows [][]interface{}, db *sqlx.DB) error {
	var keys []string
	var values []string

	for _, col := range cols {
		keys = append(keys, strings.ToLower(col))
	}
	for _, val := range rows {
		var t []string
		for _, v := range val {
			switch v.(type) {
			case string:
				t = append(t, fmt.Sprintf("'%s'", v.(string)))
			case int64:
				t = append(t, strconv.FormatInt(v.(int64), 10))
			case uint64:
				t = append(t, strconv.FormatUint(v.(uint64), 10))
			case int:
				t = append(t, strconv.Itoa(v.(int)))
			}
		}
		values = append(values, fmt.Sprintf("(%s)", strings.Join(t, ",")))
	}
	sql := fmt.Sprintf("insert into %s (%s) values %s", string(table), strings.Join(keys, ","), strings.Join(values, ","))
	if _, err := db.Exec(sql); err != nil {
		return fmt.Errorf("exec insert error, sql: %s, err: %s \n", sql, err.Error())
	}
	return nil
}

func convertField(field interface{}) interface{} {
	switch v := field.(type) {
	case []byte:
		return hex.EncodeToString(v)
	case types.BlockType:
		return v.String()
	case types.Balance:
		return v.String()
	case types.Hash:
		return v.String()
	case types.Work:
		return v.String()
	case types.Signature:
		return v.String()
	case types.Address:
		return v.String()
	case uint64:
		return v
	case int64:
		return v
	case int:
		return v
	case string:
		return v
	default:
		return ""
	}
}

func convertToDumpBlock(blk *types.StateBlock) *Block {
	return &Block{
		Type:           blk.GetType(),
		Token:          blk.GetToken(),
		Address:        blk.GetAddress(),
		Balance:        blk.GetBalance(),
		Vote:           blk.GetVote(),
		Network:        blk.GetNetwork(),
		Storage:        blk.GetStorage(),
		Oracle:         blk.GetOracle(),
		Previous:       blk.GetPrevious(),
		Link:           blk.GetLink(),
		Sender:         blk.GetSender(),
		Receiver:       blk.GetReceiver(),
		Message:        blk.GetMessage(),
		Data:           blk.GetData(),
		PoVHeight:      blk.PoVHeight,
		Timestamp:      blk.GetTimestamp(),
		Extra:          blk.GetExtra(),
		Representative: blk.GetRepresentative(),
		Work:           blk.GetWork(),
		Signature:      blk.GetSignature(),
		Hash:           blk.GetHash(),
	}
}

func convertToDumpAccount(tm *types.TokenMeta) *Account {
	return &Account{
		Address:        tm.BelongTo,
		Token:          tm.Type,
		Header:         tm.Header,
		Representative: tm.Representative,
		OpenBlock:      tm.OpenBlock,
		Modified:       tm.Modified,
		BlockCount:     tm.BlockCount,
		Balance:        tm.Balance,
	}
}

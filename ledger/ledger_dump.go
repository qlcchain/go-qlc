package ledger

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
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
	BlockHeight    int64           `msg:"blockHeight" json:"blockHeight"`
	InChain        int             `msg:"inChain" json:"inChain"`
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
	Certain        int           `msg:"certain" json:"certain"` // 0（false）, 1（true）。
}

type BlockLink struct {
	Hash  types.Hash      `msg:"hash" json:"hash"`
	Type  types.BlockType `msg:"type" json:"type"`
	Child types.Hash      `msg:"child" json:"child"`
	Link  types.Hash      `msg:"link" json:"link"`
}

func (l *Ledger) Dump() (string, error) {
	dir := path.Join(l.dir, "relation")
	os.Remove(dir)
	if err := util.CreateDirIfNotExist(dir); err != nil {
		return "", err
	}
	dir = path.Join(dir, "dump.db")
	db, err := sqlx.Connect("sqlite3", dir)
	if err != nil {
		return "", fmt.Errorf("connect sqlite error: %s", err)
	}
	if err := initDumpTables(db); err != nil {
		return "", fmt.Errorf("init dump tables error: %s", err)
	}
	if err := l.dumpBlock(db); err != nil {
		return "", fmt.Errorf("dump block error: %s", err)
	}
	if err := l.dumpBlockCache(db); err != nil {
		return "", fmt.Errorf("dump block cache error: %s", err)
	}
	if err := l.dumpBlockLink(db); err != nil {
		return "", fmt.Errorf("dump block link error: %s", err)
	}
	return dir, nil
}

func (l *Ledger) dumpBlock(db *sqlx.DB) error {
	inChainBlocksMap := make(map[types.Hash]*Block)
	err := l.GetAccountMetas(func(am *types.AccountMeta) error {
		for _, tm := range am.Tokens {
			blkHash := tm.OpenBlock
			var h int64
			h = 1
			for {
				blk, err := l.GetStateBlockConfirmed(blkHash)
				if err != nil {
					break
				}
				inBlock := convertToDumpBlock(blk)
				inBlock.InChain = 1
				inBlock.BlockHeight = h

				inChainBlocksMap[blk.GetHash()] = inBlock
				h++
				if blkHash, err = l.GetChild(blkHash); err != nil {
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	offChainBlocksMap := make(map[types.Hash]*Block, 0)
	err = l.GetStateBlocks(func(block *types.StateBlock) error {
		if _, ok := inChainBlocksMap[block.GetHash()]; !ok {
			b := convertToDumpBlock(block)
			b.InChain = 0
			offChainBlocksMap[block.GetHash()] = b
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(offChainBlocksMap) > 0 {
		for _, blk := range offChainBlocksMap {
			if b, ok := inChainBlocksMap[blk.Previous]; ok {
				offChainBlocksMap[blk.Hash].BlockHeight = b.BlockHeight + 1

				cBlk := blk
				for {
					findPre := false
					for _, b := range offChainBlocksMap {
						if b.Previous == cBlk.Hash {
							offChainBlocksMap[b.Hash].BlockHeight = offChainBlocksMap[cBlk.Hash].BlockHeight + 1
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
	}
	if err := dumpBlocks(inChainBlocksMap, false, db); err != nil {
		return err
	}
	if err := dumpBlocks(offChainBlocksMap, false, db); err != nil {
		return err
	}
	if err := l.dumpAccount(offChainBlocksMap, db); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) dumpAccount(offChainBlocks map[types.Hash]*Block, db *sqlx.DB) error {
	unCertains := make(map[types.Hash]int, 0)
	for _, blk := range offChainBlocks {
		accountHash, _ := types.HashBytes(blk.Address[:], blk.Token[:])
		unCertains[accountHash] = 0
	}

	certainAccounts := make([]*Account, 0)
	unCertainAccounts := make([]*Account, 0)
	err := l.GetAccountMetas(func(am *types.AccountMeta) error {
		for _, tm := range am.Tokens {
			accountHash, _ := types.HashBytes(tm.BelongTo[:], tm.Type[:])
			if _, ok := unCertains[accountHash]; ok {
				a := convertToDumpAccount(tm)
				a.Certain = 0
				unCertainAccounts = append(unCertainAccounts, a)
			} else {
				a := convertToDumpAccount(tm)
				a.Certain = 1
				certainAccounts = append(certainAccounts, a)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := dumpAccounts(certainAccounts, false, db); err != nil {
		return err
	}
	if err := dumpAccounts(unCertainAccounts, false, db); err != nil {
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
		if err := dumpBlocks(blockCacheMap, true, db); err != nil {
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
		accountCache, _ := l.GetAccountMetaCache(abs[0].Address)
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
	if err := dumpAccounts(accountDumps, true, db); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) dumpBlockLink(db *sqlx.DB) error {
	blkLinks := make([]*BlockLink, 0)
	err := l.GetStateBlocks(func(block *types.StateBlock) error {
		hash := block.GetHash()
		blkLink := &BlockLink{
			Hash: hash,
			Type: block.GetType(),
		}
		child, _ := l.GetChild(hash)
		blkLink.Child = child
		if block.IsSendBlock() {
			link, _ := l.GetLinkBlock(hash)
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

func dumpBlocks(blocks map[types.Hash]*Block, cache bool, db *sqlx.DB) error {
	if len(blocks) == 0 {
		return nil
	}
	block := new(Block)
	rv := reflect.ValueOf(*block)
	rt := reflect.TypeOf(*block)
	var tableName string
	if cache {
		tableName = "blockcache"
	} else {
		tableName = strings.ToLower(rt.Name())
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

func dumpAccounts(accounts []*Account, cache bool, db *sqlx.DB) error {
	if len(accounts) == 0 {
		return nil
	}
	account := new(Account)
	rv := reflect.ValueOf(*account)
	rt := reflect.TypeOf(*account)
	var tableName string
	if cache {
		tableName = "accountcache"
	} else {
		tableName = strings.ToLower(rt.Name())
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

func initDumpTables(db *sqlx.DB) error {
	block := new(Block)
	rv := reflect.ValueOf(*block)
	rt := reflect.TypeOf(*block)
	fields := getTableField(rv, rt)
	tableName := strings.ToLower(rt.Name())
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id integer PRIMARY KEY AUTOINCREMENT,  %s ) ", tableName, strings.Join(fields, ","))
	if _, err := db.Exec(sql); err != nil {
		fmt.Printf("exec error, sql: %s, err: %s \n", sql, err.Error())
		return err
	}

	account := new(Account)
	rv2 := reflect.ValueOf(*account)
	rt2 := reflect.TypeOf(*account)
	fields2 := getTableField(rv2, rt2)
	tableName2 := strings.ToLower(rt2.Name())
	sql2 := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id integer PRIMARY KEY AUTOINCREMENT,  %s ) ", tableName2, strings.Join(fields2, ","))
	if _, err := db.Exec(sql2); err != nil {
		fmt.Printf("exec error, sql: %s, err: %s \n", sql2, err.Error())
		return err
	}

	blkLink := new(BlockLink)
	rv3 := reflect.ValueOf(*blkLink)
	rt3 := reflect.TypeOf(*blkLink)
	fields3 := getTableField(rv3, rt3)
	tableName3 := strings.ToLower(rt3.Name())
	sql3 := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id integer PRIMARY KEY AUTOINCREMENT,  %s ) ", tableName3, strings.Join(fields3, ","))
	if _, err := db.Exec(sql3); err != nil {
		fmt.Printf("exec error, sql: %s, err: %s \n", sql3, err.Error())
		return err
	}

	blockcache := new(Block)
	rv4 := reflect.ValueOf(*blockcache)
	rt4 := reflect.TypeOf(*blockcache)
	fields4 := getTableField(rv4, rt4)
	tableName4 := "blockcache"
	sql4 := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id integer PRIMARY KEY AUTOINCREMENT,  %s ) ", tableName4, strings.Join(fields4, ","))
	if _, err := db.Exec(sql4); err != nil {
		fmt.Printf("exec error, sql: %s, err: %s \n", sql4, err.Error())
		return err
	}

	accountcache := new(Account)
	rv5 := reflect.ValueOf(*accountcache)
	rt5 := reflect.TypeOf(*accountcache)
	fields5 := getTableField(rv5, rt5)
	tableName5 := "accountcache"
	sql5 := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id integer PRIMARY KEY AUTOINCREMENT,  %s ) ", tableName5, strings.Join(fields5, ","))
	if _, err := db.Exec(sql5); err != nil {
		fmt.Printf("exec error, sql: %s, err: %s \n", sql5, err.Error())
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
	if err := batchSql(table, cols, exVals, db); err != nil {
		return err
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
		return fmt.Errorf("exec error, sql: %s, err: %s \n", sql, err.Error())
	}
	return nil
}

func convertField(field interface{}) interface{} {
	switch field.(type) {
	case types.BlockType:
		return field.(types.BlockType).String()
	case types.Balance:
		return field.(types.Balance).String()
	case types.Hash:
		return field.(types.Hash).String()
	case []byte:
		return hex.EncodeToString(field.([]byte))
	case types.Work:
		return field.(types.Work).String()
	case types.Signature:
		return field.(types.Signature).String()
	case types.Address:
		return field.(types.Address).String()
	case uint64:
		return field.(uint64)
	case int64:
		return field.(int64)
	case int:
		return field.(int)
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
	}
}

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/version"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
)

type ChainApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

func NewChainApi(l *ledger.Ledger) *ChainApi {
	return &ChainApi{ledger: l, logger: log.NewLogger("api_chain")}
}

//func (c *ChainApi) LedgerSize() (map[string]int64, error) {
//	lsm, vlog := c.ledger.Size()
//	r := make(map[string]int64)
//	r["lsm"] = lsm
//	r["vlog"] = vlog
//	r["total"] = lsm + vlog
//	return r, nil
//}

func (c *ChainApi) Version() (map[string]string, error) {
	r := make(map[string]string)
	r["build time"] = version.BuildTime
	r["version"] = version.Version
	r["hash"] = version.GitRev
	r["mode"] = version.Mode
	return r, nil
}

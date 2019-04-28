package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type PovApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

type PovApiBlock struct {
	*types.PovBlock
}

func NewPovApi(ledger *ledger.Ledger) *PovApi {
	return &PovApi{ledger: ledger, logger: log.NewLogger("rpc/pov")}
}

func (api *PovApi) GetBlockByHeight(height uint64) (*PovApiBlock, error) {
	block, err := api.ledger.GetPovBlockByHeight(height)
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}

	return apiBlock, err
}

func (api *PovApi) GetBlockByHash(hash types.Hash) (*PovApiBlock, error) {
	block, err := api.ledger.GetPovBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}

	return apiBlock, err
}

func (api *PovApi) GetLatestBlock() (*PovApiBlock, error) {
	block, err := api.ledger.GetLatestPovBlock()
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}

	return apiBlock, err
}
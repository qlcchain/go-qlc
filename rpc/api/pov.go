package api

import (
	"errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"
)

type PovApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

type PovApiBlock struct {
	*types.PovBlock
}

type PovApiState struct {
	*types.PovAccountState
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

func (api *PovApi) GetAccountState(stateHash types.Hash, address types.Address) (*PovApiState, error) {
	db := api.ledger.Store
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	asBytes := stateTrie.GetValue(address.Bytes())
	if len(asBytes) <= 0 {
		return nil, errors.New("account value not exist")
	}

	as := new(types.PovAccountState)
	err := as.Deserialize(asBytes)
	if err != nil {
		return nil, err
	}

	apiState := &PovApiState{
		PovAccountState: as,
	}

	return apiState, err
}
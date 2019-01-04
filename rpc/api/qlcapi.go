package api

import (
	"strconv"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/test/mock"
)

var logger = log.NewLogger("rpcapi")

type QlcApi struct{}

func NewQlcApi() *QlcApi {
	return &QlcApi{}
}

func (l *QlcApi) TestAdd(index int, str string) string {
	//ledger := newLedger()
	//return ledger.AddBlock(block)
	logger.Info(index)
	return strconv.Itoa(index) + " " + str
}

func (l *QlcApi) TestAccounts(address []types.Address) error {
	logger.Info("address", address)
	return nil
}

func (l *QlcApi) TestGetBlock() (types.Block, error) {
	return mock.StateBlock(), nil
}

func (l *QlcApi) TestAddBlock(block types.StateBlock) error {
	logger.Info("addblock,", block)
	return nil
}

func (l *QlcApi) AccountsBalances(address []types.Address) (map[types.Hash]types.Balance, error) {
	logger.Info("address", address)
	return nil, nil
}

func (l *QlcApi) AccountsFrontiers(address []types.Address) ([]types.Frontier, error) {
	return nil, nil
}

func (l *QlcApi) AccountsPending(address []types.Address) ([]types.PendingInfo, error) {
	return nil, nil
}

func (l *QlcApi) RepresentativesOnline() []types.Address {
	return nil
}

func (l *QlcApi) BlocksInfo([]types.Hash) ([]types.Block, error) {
	return nil, nil
}

func (l *QlcApi) Process(block types.Block) error {
	return nil
}

func (l *QlcApi) AccountHistoryTopn(address types.Address) ([]types.Block, error) {
	return nil, nil
}

func (l *QlcApi) AccountInfo(address types.Address) (types.AccountMeta, error) {
	return *new(types.AccountMeta), nil
}

func (l *QlcApi) ValidateAccountNumber() (int, error) {
	return 0, nil
}

func (l *QlcApi) Pending(address types.Address) (types.PendingInfo, error) {
	return *new(types.PendingInfo), nil
}

func (l *QlcApi) Tokens() ([]types.Hash, error) {
	return nil, nil
}

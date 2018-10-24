package genesis

import (
	"github.com/qlcchain/go-qlc/common/types"
)

type Genesis struct {
	Block         types.StateBlock
	Balance       types.Balance
	WorkThreshold uint64
}

var (
	Live = Genesis{}

	Beta = Genesis{}

	Test = Genesis{}
)

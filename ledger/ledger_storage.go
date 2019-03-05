/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/abi/contract"
)

//TODO: implement
func (l *Ledger) GetStorage(addr *types.Address, key []byte) []byte {
	return nil
}

func (l *Ledger) SetStorage(key []byte, value []byte) error {
	return nil
}

func (l *Ledger) ListTokens() []*types.TokenInfo {
	return nil
}

func (l *Ledger) GetTokenById(tokenId types.Hash) (types.TokenInfo, error) {
	return types.TokenInfo{}, nil
}

func (l *Ledger) GetTokenByName(tokenName string) (*types.TokenInfo, error) {
	return &types.TokenInfo{}, nil
}

func (l *Ledger) GetGenesis() []*types.StateBlock {
	return nil
}

func (l *Ledger) testUnpack() {
	block := types.StateBlock{}
	_, err := cabi.ParseTokenInfo(block.Data)
	if err != nil {
		l.logger.Error(err)
	}
}

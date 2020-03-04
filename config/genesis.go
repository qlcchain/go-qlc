/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	genesisInfos []*GenesisInfo

	genesisAddress      types.Address
	chainToken          types.Hash
	genesisMintageBlock types.StateBlock
	genesisMintageHash  types.Hash
	genesisBlock        types.StateBlock
	genesisBlockHash    types.Hash

	gasAddress      types.Address
	gasToken        types.Hash
	gasMintageBlock types.StateBlock
	gasMintageHash  types.Hash
	gasBlock        types.StateBlock
	gasBlockHash    types.Hash
)

func GenesisInfos() []*GenesisInfo {
	return genesisInfos
}

func GenesisAddress() types.Address {
	return genesisAddress
}

func GasAddress() types.Address {
	return gasAddress
}

func ChainToken() types.Hash {
	return chainToken
}

func GasToken() types.Hash {
	return gasToken
}

func GenesisMintageBlock() types.StateBlock {
	return genesisMintageBlock
}

func GenesisMintageHash() types.Hash {
	return genesisMintageHash
}

func GenesisBlock() types.StateBlock {
	return genesisBlock
}

func GenesisBlockHash() types.Hash {
	return genesisBlockHash
}

func GasBlockHash() types.Hash {
	return gasBlockHash
}

func GasMintageBlock() types.StateBlock {
	return gasMintageBlock
}

func GasBlock() types.StateBlock {
	return gasBlock
}

// IsGenesis check block is chain token genesis
func IsGenesisBlock(block *types.StateBlock) bool {
	hash := block.GetHash()
	return hash == genesisMintageHash || hash == genesisBlockHash || hash == gasMintageHash || hash == gasBlockHash
}

// IsGenesis check token is chain token genesis
func IsGenesisToken(hash types.Hash) bool {
	return hash == chainToken || hash == gasToken
}

func AllGenesisBlocks() []types.StateBlock {
	return []types.StateBlock{genesisMintageBlock, genesisBlock, gasMintageBlock, gasBlock}
}

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

var genesisInfos []*GenesisInfo

func GenesisInfos() []*GenesisInfo {
	return genesisInfos
}

func GenesisAddress() types.Address {
	for _, v := range genesisInfos {
		if v.ChainToken {
			return v.Genesis.Address
		}
	}
	return types.ZeroAddress
}

func ChainToken() types.Hash {
	for _, v := range genesisInfos {
		if v.ChainToken {
			return v.Genesis.Token
		}
	}
	return types.ZeroHash
}

func GasToken() types.Hash {
	for _, v := range genesisInfos {
		if v.GasToken {
			return v.Genesis.Token
		}
	}
	return types.ZeroHash
}

func GenesisMintageBlock() types.StateBlock {
	for _, v := range genesisInfos {
		if v.ChainToken {
			return v.Mintage
		}
	}
	return types.StateBlock{}
}

func GenesisMintageHash() types.Hash {
	for _, v := range genesisInfos {
		if v.ChainToken {
			return v.Mintage.GetHash()
		}
	}
	return types.ZeroHash
}

func GenesisBlock() types.StateBlock {
	for _, v := range genesisInfos {
		if v.ChainToken {
			return v.Genesis
		}
	}
	return types.StateBlock{}
}

func GenesisBlockHash() types.Hash {
	for _, v := range genesisInfos {
		if v.ChainToken {
			return v.Genesis.GetHash()
		}
	}
	return types.ZeroHash
}

func GasBlockHash() types.Hash {
	for _, v := range genesisInfos {
		if v.GasToken {
			return v.Genesis.GetHash()
		}
	}
	return types.ZeroHash
}

func GasMintageBlock() types.StateBlock {
	for _, v := range genesisInfos {
		if v.GasToken {
			return v.Mintage
		}
	}
	return types.StateBlock{}
}

func GasBlock() types.StateBlock {
	for _, v := range genesisInfos {
		if v.GasToken {
			return v.Genesis
		}
	}
	return types.StateBlock{}
}

// IsGenesis check block is chain token genesis
func IsGenesisBlock(block *types.StateBlock) bool {
	hash := block.GetHash()
	for _, v := range genesisInfos {
		if hash == v.Genesis.GetHash() || hash == v.Mintage.GetHash() {
			return true
		}
	}
	return false
}

// IsGenesis check token is chain token genesis
func IsGenesisToken(hash types.Hash) bool {
	for _, v := range genesisInfos {
		if hash == v.Mintage.Token {
			return true
		}
	}
	return false
}

func AllGenesisBlocks() []types.StateBlock {
	var blocks []types.StateBlock
	for _, v := range genesisInfos {
		blocks = append(blocks, v.Mintage)
		blocks = append(blocks, v.Genesis)
	}
	return blocks
}

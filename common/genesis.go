package common

import (
	"github.com/qlcchain/go-qlc/common/types"
)

type GenesisInfo struct {
	ChainToken          bool
	GasToken            bool
	GenesisMintageBlock types.StateBlock
	GenesisBlock        types.StateBlock
}

var GenesisInfos []*GenesisInfo

func GenesisAddress() types.Address {
	for _, v := range GenesisInfos {
		if v.ChainToken {
			return v.GenesisBlock.Address
		}
	}
	return types.ZeroAddress
}

func ChainToken() types.Hash {
	for _, v := range GenesisInfos {
		if v.ChainToken {
			return v.GenesisBlock.Token
		}
	}
	return types.ZeroHash
}

func GasToken() types.Hash {
	for _, v := range GenesisInfos {
		if v.GasToken {
			return v.GenesisBlock.Token
		}
	}
	return types.ZeroHash
}

func GenesisMintageBlock() types.StateBlock {
	for _, v := range GenesisInfos {
		if v.ChainToken {
			return v.GenesisMintageBlock
		}
	}
	return types.StateBlock{}
}

func GenesisMintageHash() types.Hash {
	for _, v := range GenesisInfos {
		if v.ChainToken {
			return v.GenesisMintageBlock.GetHash()
		}
	}
	return types.ZeroHash
}

func GenesisBlock() types.StateBlock {
	for _, v := range GenesisInfos {
		if v.ChainToken {
			return v.GenesisBlock
		}
	}
	return types.StateBlock{}
}

func GenesisBlockHash() types.Hash {
	for _, v := range GenesisInfos {
		if v.ChainToken {
			return v.GenesisBlock.GetHash()
		}
	}
	return types.ZeroHash
}

func GasBlockHash() types.Hash {
	for _, v := range GenesisInfos {
		if v.GasToken {
			return v.GenesisBlock.GetHash()
		}
	}
	return types.ZeroHash
}

// IsGenesis check block is chain token genesis
func IsGenesisBlock(block *types.StateBlock) bool {
	hash := block.GetHash()
	for _, v := range GenesisInfos {
		if hash.Cmp(v.GenesisBlock.GetHash()) == 0 || hash.Cmp(v.GenesisMintageBlock.GetHash()) == 0 {
			return true
		}
	}
	return false
}

// IsGenesis check token is chain token genesis
func IsGenesisToken(hash types.Hash) bool {
	for _, v := range GenesisInfos {
		if hash.Cmp(v.GenesisMintageBlock.Token) == 0 {
			return true
		}
	}
	return false
}

func AllGenesisBlocks() []types.StateBlock {
	var blocks []types.StateBlock
	for _, v := range GenesisInfos {
		blocks = append(blocks, v.GenesisMintageBlock)
		blocks = append(blocks, v.GenesisBlock)
	}
	return blocks
}

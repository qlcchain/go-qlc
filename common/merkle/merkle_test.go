package merkle

import (
	"math/rand"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
)

func TestCalcCoinbaseMerkleRoot_1(t *testing.T) {
	txHashes := make([]*types.Hash, 0)
	txData := make([]byte, 300)
	_ = random.Bytes(txData)
	txHash := types.Sha256D_HashData(txData)
	txHashes = append(txHashes, &txHash)

	mr := CalcMerkleTreeRootHash(txHashes)

	if mr.Cmp(txHash) != 0 {
		t.Fatal("mr != txHash", mr, txHash)
	}
}

func TestCalcCoinbaseMerkleRoot_2(t *testing.T) {
	txNum := int(rand.Int31n(30000))
	txHashes := make([]*types.Hash, 0, txNum)
	for txCnt := 0; txCnt < txNum; txCnt++ {
		txData := make([]byte, 300)
		_ = random.Bytes(txData)
		txHash := types.Sha256D_HashData(txData)
		txHashes = append(txHashes, &txHash)
	}

	cbTxData := make([]byte, 200)
	_ = random.Bytes(cbTxData)
	cbTxHash := types.Sha256D_HashData(cbTxData)

	allTxHash := make([]*types.Hash, 0, txNum+1)
	allTxHash = append(allTxHash, &cbTxHash)
	allTxHash = append(allTxHash, txHashes...)

	mr1 := CalcMerkleTreeRootHash(allTxHash)

	b1 := BuildCoinbaseMerkleBranch(txHashes)
	mr2 := CalcCoinbaseMerkleRoot(&cbTxHash, b1)

	if mr1.Cmp(mr2) != 0 {
		t.Fatal("mr1 != mr2", mr1, mr2)
	}
}

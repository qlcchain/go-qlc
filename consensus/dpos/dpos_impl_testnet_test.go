// +build testnet

package dpos

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/mock"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

func TestDispatchAckedBlock(t *testing.T) {
	dps := getTestDpos()

	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = 100
	err := dps.ledger.AddPovBlock(povBlk, povTd)
	if err != nil {
		t.Fatal(err)
	}

	err = dps.ledger.AddPovBestHash(povBlk.GetHeight(), povBlk.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = dps.ledger.SetPovLatestHeight(povBlk.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	blk1 := mock.StateBlockWithoutWork()
	blk1.Type = types.ContractSend
	blk1.Link = contractaddress.DoDSettlementAddress.ToHash()
	blk1.PoVHeight = 100

	param := new(cabi.DoDSettleUpdateOrderInfoParam)
	blk1.Data, _ = param.ToABI()
	dps.dispatchAckedBlock(blk1, mock.Hash(), 0)

	blk2 := mock.StateBlockWithoutWork()
	blk2.Type = types.ContractSend
	blk2.Link = contractaddress.DoDSettlementAddress.ToHash()
	blk2.PoVHeight = 100

	param2 := new(cabi.DoDSettleCreateOrderParam)
	blk2.Data, _ = param2.ToABI()
	err = dps.ledger.AddStateBlock(blk2)
	if err != nil {
		t.Fatal(err)
	}

	blk3 := mock.StateBlockWithoutWork()
	blk3.Type = types.ContractReward
	blk3.Link = blk2.GetHash()
	blk3.PoVHeight = 100
	dps.dispatchAckedBlock(blk3, mock.Hash(), 0)
}

package contract

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"testing"
)

var (
	jsonBlk = `
		{
			"type": "ContractSend",
			"token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
			"address": "qlc_1oq9r6w9bc7x4j6j79qxk15uxf7wtzorko6qegrmwrsctukf65dz3jhz3pfu",
			"balance": "0",
			"vote": "900000000",
			"network": "0",
			"storage": "0",
			"oracle": "0",
			"previous": "25204c0d5f28b1d9a54d0a9be87c8348014fe79b7d66ce7f0013e244234cf81c",
			"link": "b7902600dfc79387b2601edc347b854d55d6b31142e324a4e54ff00a4c519c91",
			"message": "0000000000000000000000000000000000000000000000000000000000000000",
			"data": "I9Wy11bnwTh0qL0USRKe/ZAHvrS81+uJVJdjsT5jKtbk0g1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAF/uOzIwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEA3MmQwZTZlMmM2MDVjNTM1ZjE0NWNhN2MyM2JkNWE4ZGNjYzJlM2ViMjk1ZGM4ZTBiMDY5MmZmNmE1MjllZGNl",
			"povHeight": 0,
			"timestamp": 1561209257,
			"extra": "0000000000000000000000000000000000000000000000000000000000000000",
			"representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
			"work": "fe4555a7d5e8d9c6",
			"signature": "047aee2d5c2c93fe1a0e2da1a814b3d8bb243967c10fdae4ecf27c64ac04d7035555f8f52de70b999528a258871e809bbace0b9e859d7bf2638e50fef48a8e01",
			"tokenName": "QLC",
			"amount": "6592300000000",
			"hash": "fe81ccf92bc9734855fcdf5f71906649d07276558fa818f8fffd1595a0d0bd4e",
			"povConfirmHeight": 2,
			"povConfirmCount": 69184
		}
	`
)

func TestNep5Pledge_GetTargetReceiver(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	blk := new(types.StateBlock)
	err := json.Unmarshal([]byte(jsonBlk), blk)
	if err != nil {
		t.Fatal(err)
	}

	nep5 := new(WithdrawNep5Pledge)
	ctx := vmstore.NewVMContext(l)
	target, _ := nep5.GetTargetReceiver(ctx, blk)
	t.Log(target[:])
}

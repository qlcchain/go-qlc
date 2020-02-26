package contract

import (
	"encoding/json"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestAirdropRewords_GetTargetReceiver(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	var jsonSend = `
		{
			"type": "ContractSend",
			"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
			"address": "qlc_35dyxo7wwg5rzwmipiu19u99ayh3qcrnjwnnekij7oijby466qnqy9hbs5td",
			"balance": "8516298185573",
			"vote": "0",
			"network": "0",
			"storage": "0",
			"oracle": "0",
			"previous": "9963e0b7d9798bb137e7d40cfc7bec27615d77e29d978aa481dfb97cbc51fc61",
			"link": "d614bb9d5e20ad063316ce091148e77c99136c6194d55c7ecc7ffa9620dbcaeb",
			"message": "0000000000000000000000000000000000000000000000000000000000000000",
			"data": "jEWRCk8mS05D3Rpl6xkMdhRem9+8Z7uPTpSWcBASbunxRgwbr/oqPs02onf1Trq9RZvN54hRiZXRLKWgXdqVC27HClSZY+C32XmLsTfn1Az8e+wnYV134p2XiqSB37l8vFH8YeuK0HJelB0uj4sC8AMydR9du/iygkHV9jsfJg0oTx7BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABooo+Sr8OJAEL4pKJGZjKCG+iGdOAADYohd7rD9aqAA/r56m4wa4wZ8lnrXYc5Of/bPDtF0dki0fRB7U0pcs76ZnUsA",
			"povHeight": 0,
			"timestamp": 1565395308,
			"extra": "0000000000000000000000000000000000000000000000000000000000000000",
			"representative": "qlc_1111111111111111111111111111111111111111111111111111hifc8npp",
			"work": "0000000000000000",
			"signature": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			"tokenName": "QGAS",
			"amount": "438871012",
			"hash": "8cbacbe054e57dd44e69f7b2f5b87291eb50c253ae94f7c75cb3c9eda1b6de4d",
			"povConfirmHeight": 6,
			"povConfirmCount": 69180
		}
	`
	blk := new(types.StateBlock)
	_ = json.Unmarshal([]byte(jsonSend), blk)
	ar := new(AirdropRewords)
	ctx := vmstore.NewVMContext(l)
	addr, err := ar.GetTargetReceiver(ctx, blk)
	if err != nil || addr.String() != "qlc_3dzt7azetfo4gztnxgoxapfwuswac86sdnbenpi7upno3fqeg4knnnb7d94x" {
		t.Fatal(err)
	}
}

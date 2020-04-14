/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"bytes"
	"testing"

	"github.com/qlcchain/go-qlc/crypto/random"
)

func TestContractValue_Serialize(t *testing.T) {
	h1 := Hash{}
	_ = random.Bytes(h1[:])

	h2 := Hash{}
	_ = random.Bytes(h2[:])
	type fields struct {
		Previous Hash
		Root     *Hash
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"1", fields{
			Previous: h1,
			Root:     nil,
		}, false},
		{"2", fields{
			Previous: h1,
			Root:     &h2,
		}, false},
	}
	for idx, test := range tests {
		tt := test
		i := idx

		t.Run(tt.name, func(t *testing.T) {
			z := &ContractValue{
				BlockHash: tt.fields.Previous,
			}
			t.Log(z.String())
			got, err := z.Serialize()
			if (err != nil) != tt.wantErr {
				t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := &ContractValue{}
			if err = want.Deserialize(got); err != nil {
				t.Fatalf("%s[%d]: Deserialize failed, %s", tt.name, i, err)
			} else {
				if !bytes.EqualFold(want.BlockHash[:], tt.fields.Previous[:]) {
					t.Fatalf("exp: %s, act: %s", tt.fields.Previous, want.BlockHash)
				}

				//if tt.fields.Root != nil {
				//	if !bytes.EqualFold(want.Root[:], tt.fields.Root[:]) {
				//		t.Fatalf("exp: %s, act: %s", tt.fields.Root, want.Root)
				//	}
				//} else {
				//	if want.Root != nil {
				//		t.Fatalf("suffix should be nil, %v", want.Root)
				//	}
				//}
			}
		})
	}
}

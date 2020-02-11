/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"encoding/hex"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestGetChainContractName(t *testing.T) {
	data, err := hex.DecodeString("d357c55f83b3437265617465436f6e7472616374506172616d8da27061c7206397bca9f6647d38783d38b9f491fc983d28b99f6ffa139e2de4e72c7c464e5bd7a2616ea26331a27062c720638871d173a94e9a2765ba4b5d018bc97bf9e23545e2bf6498bc1a2b003127d05aa2626ea26332a3707265c7206500ec3ac071a4d6e24eb5f5afd6af99082125536c2ed2f09e12ef67a855b0b26ca26964d94030656636376433623939343932663965363131363464343934316136643730633830613538623366653638626433333066363435336636373261383339636166a36d636301a36d6e6302a17464a17502a163a3555344a27431d25e428830a27361c7406799cc8b0dd84c463a513f0c4c10520d794b5b3423532853e6d6b945a046e9ce7b6e3739259e912c164969b7b985c907afa3ce34ec89d169738b3cc9a80cafe501a27432d25e428830a27362c74067a781f995bb847e2cff9bc1e5c7ecc1c9bce9ebf27f44a5f0eccfadcbd636bdf2bb97134e9683364da01cbbd178dca8ebd6aa88016a6f7786906df91bff5f8006")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		addr           types.Address
		methodSelector []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   bool
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				addr:           types.SettlementAddress,
				methodSelector: data,
			},
			want:    "CreateContract",
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetChainContractName(tt.args.addr, tt.args.methodSelector)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChainContractName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetChainContractName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetChainContractName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

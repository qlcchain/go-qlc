/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/mock"
)

type mockDataContractApi struct {
	l  ledger.Store
	cc *qctx.ChainContext
	eb event.EventBus
}

func setupTestCaseContractApi(t *testing.T) (func(t *testing.T), *mockDataContractApi) {
	md := new(mockDataContractApi)

	dir := filepath.Join(config.QlcTestDataDir(), "rewards", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()

	md.l = ledger.NewLedger(cm.ConfigFile)

	md.cc = qctx.NewChainContext(cm.ConfigFile)
	md.cc.Init(nil)

	md.eb = md.cc.EventBus()

	return func(t *testing.T) {
		_ = md.eb.Close()
		err := md.l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestNewContractApi(t *testing.T) {
	tearDown, md := setupTestCaseContractApi(t)
	defer tearDown(t)

	api := NewContractApi(md.cc, md.l)
	if abi, err := api.GetAbiByContractAddress(contractaddress.BlackHoleAddress); err != nil {
		t.Fatal(err)
	} else if len(abi) == 0 {
		t.Fatal("invalid abi")
	} else {
		if data, err := api.PackContractData(abi, "Destroy", []string{mock.Address().String(),
			mock.Hash().String(), mock.Hash().String(), "111", types.ZeroSignature.String()}); err != nil {
			t.Fatal(err)
		} else if len(data) == 0 {
			t.Fatal("invalid data")
		} else {
			t.Log(hex.EncodeToString(data))
		}
	}

	if addressList := api.ContractAddressList(); len(addressList) == 0 {
		t.Fatal("can not find any on-chain contract")
	}
}

func TestContractApi_PackContractData(t *testing.T) {
	logger := log.NewLogger("TestContractApi_PackContractData")

	type fields struct {
		logger *zap.SugaredLogger
	}
	type args struct {
		abiStr     string
		methodName string
		params     []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "address[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "address[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`["qlc_38nm8t5rimw6h6j7wyokbs8jiygzs7baoha4pqzhfw1k79npyr1km8w6y7r8","qlc_38nm8t5rimw6h6j7wyokbs8jiygzs7baoha4pqzhfw1k79npyr1km8w6y7r8"]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "hash[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "hash[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`["2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E","2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "tokenId[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "tokenId[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`["2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E","2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "string[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "string[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`["2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E","2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "signature[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "signature[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`["bb5c2e6d3b30f2edd749669452a447c0dfd45538edc09d32e407c22ba4c728f7945aec3dd253405b2b7eb81543e081a91edccf10362a8bbe722b75021305d901","bb5c2e6d3b30f2edd749669452a447c0dfd45538edc09d32e407c22ba4c728f7945aec3dd253405b2b7eb81543e081a91edccf10362a8bbe722b75021305d901"]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "int8",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "int8"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"100"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "int16",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "int16"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"100"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "int32",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "int32"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"100"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "int64",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "int64"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"100"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "int128",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "int128"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"1011111110"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint8",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint8"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"100"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint16",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint16"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"100"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint32",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint32"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"100"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint64",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint64"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"100"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint128",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint128"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"1011111110"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "bytes32",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "bytes32"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "bytes",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "bytes"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{"2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "bool[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "bool[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[true, false]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "bool",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "bool"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`false`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "int8[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "int8[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[11, 22]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "int16[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "int16[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[111, 123]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "int32[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "int32[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[111, 123]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "int64[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "int64[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[111, 123]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint8[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint8[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[11, 22]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint16[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint16[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[111, 123]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint32[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint32[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[111, 123]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint64[]",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint64[]"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[111, 123]`},
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "uint64_error",
			fields: fields{
				logger: logger,
			},
			args: args{
				abiStr: `[
  {
    "type": "function",
    "name": "Destroy",
    "inputs": [
      {
        "name": "value",
        "type": "uint64"
      }
    ]
  }
]
`,
				methodName: "Destroy",
				params:     []string{`[111, 123]`},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ContractApi{
				logger: tt.fields.logger,
			}
			_, err := c.PackContractData(tt.args.abiStr, tt.args.methodName, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("PackContractData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("PackContractData() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestT(t *testing.T) {
	a := []bool{true, false}
	t.Log(util.ToString(a))
}

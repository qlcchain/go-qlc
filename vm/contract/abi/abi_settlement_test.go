/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"bytes"
	"encoding/json"
	"math"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/util"

	"github.com/qlcchain/go-qlc/crypto/random"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/mock"
)

var (
	createContractParam = CreateContractParam{
		PartyA: Contractor{
			Address: mock.Address(),
			Name:    "PCCWG",
		},
		PartyB: Contractor{
			Address: mock.Address(),
			Name:    "HKTCSL",
		},
		Previous: mock.Hash(),
		Services: []ContractService{{
			ServiceId:   mock.Hash().String(),
			Mcc:         1,
			Mnc:         2,
			TotalAmount: 100,
			UnitPrice:   2,
			Currency:    "USD",
		}, {
			ServiceId:   mock.Hash().String(),
			Mcc:         22,
			Mnc:         1,
			TotalAmount: 300,
			UnitPrice:   4,
			Currency:    "USD",
		}},
		SignDate:  time.Now().AddDate(0, 0, -5).Unix(),
		StartDate: time.Now().AddDate(0, 0, -2).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 2).Unix(),
	}

	cdrParam = CDRParam{
		Index:       1,
		SmsDt:       time.Now().Unix(),
		Sender:      "PCCWG",
		Destination: "85257***343",
		//DstCountry:    "Hong Kong",
		//DstOperator:   "HKTCSL",
		//DstMcc:        454,
		//DstMnc:        0,
		//SellPrice:     1,
		//SellCurrency:  "USD",
		//CustomerName:  "Tencent",
		//CustomerID:    "11667",
		SendingStatus: SendingStatusSent,
		DlrStatus:     DLRStatusDelivered,
	}

	cdrs = `[
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457880,
        "smsDt": 1578315202,
        "sender": "WeChat",
        "destination": "85251***214",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457913,
        "smsDt": 1578316536,
        "sender": "WeChat",
        "destination": "85265***827",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457914,
        "smsDt": 1578316560,
        "sender": "WeChat",
        "destination": "85251***906",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457828,
        "smsDt": 1578313123,
        "sender": "WeChat",
        "destination": "85263***704",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457828,
        "smsDt": 1578313122,
        "sender": "WeChat",
        "destination": "85263***704",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457801,
        "smsDt": 1578312041,
        "sender": "WeChat",
        "destination": "85293***089",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457824,
        "smsDt": 1578312984,
        "sender": "WeChat",
        "destination": "85263***013",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457955,
        "smsDt": 1578318205,
        "sender": "WeChat",
        "destination": "85268***993",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457679,
        "smsDt": 1578307189,
        "sender": "WeChat",
        "destination": "85254***787",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457679,
        "smsDt": 1578307188,
        "sender": "WeChat",
        "destination": "85254***787",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457833,
        "smsDt": 1578313358,
        "sender": "WeChat",
        "destination": "85296***080",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457970,
        "smsDt": 1578318838,
        "sender": "WeChat",
        "destination": "85261***580",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457713,
        "smsDt": 1578308525,
        "sender": "WeChat",
        "destination": "85265***196",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457713,
        "smsDt": 1578308523,
        "sender": "WeChat",
        "destination": "85265***196",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457682,
        "smsDt": 1578307311,
        "sender": "WeChat",
        "destination": "85257***552",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457682,
        "smsDt": 1578307311,
        "sender": "WeChat",
        "destination": "85257***552",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457797,
        "smsDt": 1578311919,
        "sender": "WeChat",
        "destination": "85260***148",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457797,
        "smsDt": 1578311918,
        "sender": "WeChat",
        "destination": "85260***148",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457965,
        "smsDt": 1578318639,
        "sender": "WeChat",
        "destination": "85265***725",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457965,
        "smsDt": 1578318638,
        "sender": "WeChat",
        "destination": "85265***725",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457711,
        "smsDt": 1578308459,
        "sender": "WeChat",
        "destination": "85262***421",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457711,
        "smsDt": 1578308458,
        "sender": "WeChat",
        "destination": "85262***421",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457739,
        "smsDt": 1578309560,
        "sender": "WeChat",
        "destination": "85290***155",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457801,
        "smsDt": 1578312063,
        "sender": "WeChat",
        "destination": "85257***142",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457801,
        "smsDt": 1578312062,
        "sender": "WeChat",
        "destination": "85257***142",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457938,
        "smsDt": 1578317547,
        "sender": "WeChat",
        "destination": "85291***514",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457938,
        "smsDt": 1578317546,
        "sender": "WeChat",
        "destination": "85291***514",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457688,
        "smsDt": 1578307525,
        "sender": "WeChat",
        "destination": "85296***614",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457688,
        "smsDt": 1578307524,
        "sender": "WeChat",
        "destination": "85296***614",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457649,
        "smsDt": 1578305968,
        "sender": "WeChat",
        "destination": "85269***618",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457649,
        "smsDt": 1578305967,
        "sender": "WeChat",
        "destination": "85269***618",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457909,
        "smsDt": 1578316388,
        "sender": "WeChat",
        "destination": "85263***043",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457909,
        "smsDt": 1578316387,
        "sender": "WeChat",
        "destination": "85263***043",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457836,
        "smsDt": 1578313465,
        "sender": "WeChat",
        "destination": "85260***641",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457951,
        "smsDt": 1578318077,
        "sender": "WeChat",
        "destination": "85267***920",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457951,
        "smsDt": 1578318076,
        "sender": "WeChat",
        "destination": "85267***920",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457827,
        "smsDt": 1578313096,
        "sender": "WeChat",
        "destination": "85263***093",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457827,
        "smsDt": 1578313093,
        "sender": "WeChat",
        "destination": "85263***093",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457850,
        "smsDt": 1578314020,
        "sender": "WeChat",
        "destination": "85265***187",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457850,
        "smsDt": 1578314019,
        "sender": "WeChat",
        "destination": "85265***187",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457948,
        "smsDt": 1578317923,
        "sender": "WeChat",
        "destination": "85295***966",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457948,
        "smsDt": 1578317922,
        "sender": "WeChat",
        "destination": "85295***966",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457807,
        "smsDt": 1578312304,
        "sender": "WeChat",
        "destination": "85262***774",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457807,
        "smsDt": 1578312303,
        "sender": "WeChat",
        "destination": "85262***774",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457840,
        "smsDt": 1578313612,
        "sender": "WeChat",
        "destination": "85296***939",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457840,
        "smsDt": 1578313610,
        "sender": "WeChat",
        "destination": "85296***939",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457956,
        "smsDt": 1578318251,
        "sender": "WeChat",
        "destination": "85296***332",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457956,
        "smsDt": 1578318250,
        "sender": "WeChat",
        "destination": "85296***332",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457827,
        "smsDt": 1578313087,
        "sender": "WeChat",
        "destination": "85255***712",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457827,
        "smsDt": 1578313085,
        "sender": "WeChat",
        "destination": "85255***712",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457762,
        "smsDt": 1578310488,
        "sender": "WeChat",
        "destination": "85251***281",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457762,
        "smsDt": 1578310487,
        "sender": "WeChat",
        "destination": "85251***281",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457801,
        "smsDt": 1578312058,
        "sender": "WeChat",
        "destination": "85260***512",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457801,
        "smsDt": 1578312057,
        "sender": "WeChat",
        "destination": "85260***512",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457943,
        "smsDt": 1578317738,
        "sender": "WeChat",
        "destination": "85257***784",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457943,
        "smsDt": 1578317738,
        "sender": "WeChat",
        "destination": "85257***784",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457922,
        "smsDt": 1578316913,
        "sender": "WeChat",
        "destination": "85295***254",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457922,
        "smsDt": 1578316913,
        "sender": "WeChat",
        "destination": "85295***254",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457976,
        "smsDt": 1578319060,
        "sender": "WeChat",
        "destination": "85257***327",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457976,
        "smsDt": 1578319059,
        "sender": "WeChat",
        "destination": "85257***327",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457824,
        "smsDt": 1578312998,
        "sender": "WeChat",
        "destination": "85267***973",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457692,
        "smsDt": 1578307691,
        "sender": "WeChat",
        "destination": "85252***321",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457692,
        "smsDt": 1578307690,
        "sender": "WeChat",
        "destination": "85252***321",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457825,
        "smsDt": 1578313000,
        "sender": "WeChat",
        "destination": "85267***973",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457943,
        "smsDt": 1578317746,
        "sender": "WeChat",
        "destination": "85261***752",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457943,
        "smsDt": 1578317745,
        "sender": "WeChat",
        "destination": "85261***752",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457715,
        "smsDt": 1578308606,
        "sender": "WeChat",
        "destination": "85297***851",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457715,
        "smsDt": 1578308603,
        "sender": "WeChat",
        "destination": "85297***851",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457889,
        "smsDt": 1578315569,
        "sender": "WeChat",
        "destination": "85257***215",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457889,
        "smsDt": 1578315568,
        "sender": "WeChat",
        "destination": "85257***215",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457711,
        "smsDt": 1578308469,
        "sender": "WeChat",
        "destination": "85257***624",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457711,
        "smsDt": 1578308467,
        "sender": "WeChat",
        "destination": "85257***624",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457888,
        "smsDt": 1578315542,
        "sender": "WeChat",
        "destination": "85265***143",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457888,
        "smsDt": 1578315541,
        "sender": "WeChat",
        "destination": "85265***143",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457670,
        "smsDt": 1578306808,
        "sender": "WeChat",
        "destination": "85252***387",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457670,
        "smsDt": 1578306807,
        "sender": "WeChat",
        "destination": "85252***387",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457900,
        "smsDt": 1578316011,
        "sender": "WeChat",
        "destination": "85253***433",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457713,
        "smsDt": 1578308531,
        "sender": "WeChat",
        "destination": "85261***554",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457713,
        "smsDt": 1578308530,
        "sender": "WeChat",
        "destination": "85261***554",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457848,
        "smsDt": 1578313948,
        "sender": "WeChat",
        "destination": "85257***733",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457848,
        "smsDt": 1578313947,
        "sender": "WeChat",
        "destination": "85257***733",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457972,
        "smsDt": 1578318887,
        "sender": "WeChat",
        "destination": "85265***368",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457972,
        "smsDt": 1578318887,
        "sender": "WeChat",
        "destination": "85265***368",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457920,
        "smsDt": 1578316834,
        "sender": "WeChat",
        "destination": "85261***722",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457920,
        "smsDt": 1578316833,
        "sender": "WeChat",
        "destination": "85261***722",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457921,
        "smsDt": 1578316865,
        "sender": "WeChat",
        "destination": "85265***693",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457824,
        "smsDt": 1578312981,
        "sender": "WeChat",
        "destination": "85260***641",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457708,
        "smsDt": 1578308346,
        "sender": "WeChat",
        "destination": "85257***796",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457708,
        "smsDt": 1578308345,
        "sender": "WeChat",
        "destination": "85257***796",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457955,
        "smsDt": 1578318226,
        "sender": "WeChat",
        "destination": "85265***476",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457955,
        "smsDt": 1578318225,
        "sender": "WeChat",
        "destination": "85265***476",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457722,
        "smsDt": 1578308902,
        "sender": "WeChat",
        "destination": "85251***100",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457722,
        "smsDt": 1578308901,
        "sender": "WeChat",
        "destination": "85251***100",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457913,
        "smsDt": 1578316543,
        "sender": "WeChat",
        "destination": "85256***659",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457913,
        "smsDt": 1578316542,
        "sender": "WeChat",
        "destination": "85256***659",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457877,
        "smsDt": 1578315112,
        "sender": "WeChat",
        "destination": "85261***351",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457877,
        "smsDt": 1578315109,
        "sender": "WeChat",
        "destination": "85261***351",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457877,
        "smsDt": 1578315114,
        "sender": "WeChat",
        "destination": "85265***577",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457877,
        "smsDt": 1578315112,
        "sender": "WeChat",
        "destination": "85265***577",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457905,
        "smsDt": 1578316201,
        "sender": "WeChat",
        "destination": "85269***977",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457905,
        "smsDt": 1578316201,
        "sender": "WeChat",
        "destination": "85269***977",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457885,
        "smsDt": 1578315419,
        "sender": "WeChat",
        "destination": "85257***310",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457885,
        "smsDt": 1578315418,
        "sender": "WeChat",
        "destination": "85257***310",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457913,
        "smsDt": 1578316537,
        "sender": "WeChat",
        "destination": "85256***423",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457913,
        "smsDt": 1578316534,
        "sender": "WeChat",
        "destination": "85256***423",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457965,
        "smsDt": 1578318621,
        "sender": "WeChat",
        "destination": "85295***587",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457965,
        "smsDt": 1578318619,
        "sender": "WeChat",
        "destination": "85295***587",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457886,
        "smsDt": 1578315455,
        "sender": "WeChat",
        "destination": "85265***219",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457886,
        "smsDt": 1578315454,
        "sender": "WeChat",
        "destination": "85265***219",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457648,
        "smsDt": 1578305956,
        "sender": "WeChat",
        "destination": "85255***809",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457786,
        "smsDt": 1578311444,
        "sender": "WeChat",
        "destination": "85267***181",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457709,
        "smsDt": 1578308362,
        "sender": "WeChat",
        "destination": "85293***845",
        "sendingStatus": "Sent",
        "dlrStatus": "Rejected",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457709,
        "smsDt": 1578308361,
        "sender": "WeChat",
        "destination": "85293***845",
        "sendingStatus": "Sent",
        "dlrStatus": "Rejected",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457868,
        "smsDt": 1578314751,
        "sender": "WeChat",
        "destination": "85263***041",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457868,
        "smsDt": 1578314749,
        "sender": "WeChat",
        "destination": "85263***041",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457829,
        "smsDt": 1578313171,
        "sender": "WeChat",
        "destination": "85262***373",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457829,
        "smsDt": 1578313168,
        "sender": "WeChat",
        "destination": "85262***373",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457846,
        "smsDt": 1578313840,
        "sender": "WeChat",
        "destination": "85262***404",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457738,
        "smsDt": 1578309559,
        "sender": "WeChat",
        "destination": "85290***155",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457876,
        "smsDt": 1578315057,
        "sender": "WeChat",
        "destination": "85263***093",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457876,
        "smsDt": 1578315057,
        "sender": "WeChat",
        "destination": "85263***093",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457710,
        "smsDt": 1578308437,
        "sender": "WeChat",
        "destination": "85257***343",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457710,
        "smsDt": 1578308437,
        "sender": "WeChat",
        "destination": "85257***343",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457853,
        "smsDt": 1578314137,
        "sender": "WeChat",
        "destination": "85257***274",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457853,
        "smsDt": 1578314136,
        "sender": "WeChat",
        "destination": "85257***274",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457708,
        "smsDt": 1578308349,
        "sender": "WeChat",
        "destination": "85257***094",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457708,
        "smsDt": 1578308347,
        "sender": "WeChat",
        "destination": "85257***094",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457757,
        "smsDt": 1578310318,
        "sender": "WeChat",
        "destination": "85261***405",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457757,
        "smsDt": 1578310317,
        "sender": "WeChat",
        "destination": "85261***405",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457705,
        "smsDt": 1578308226,
        "sender": "WeChat",
        "destination": "85257***324",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457705,
        "smsDt": 1578308225,
        "sender": "WeChat",
        "destination": "85257***324",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457957,
        "smsDt": 1578318317,
        "sender": "WeChat",
        "destination": "85266***784",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457957,
        "smsDt": 1578318317,
        "sender": "WeChat",
        "destination": "85266***784",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457872,
        "smsDt": 1578314901,
        "sender": "WeChat",
        "destination": "85266***391",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457872,
        "smsDt": 1578314900,
        "sender": "WeChat",
        "destination": "85266***391",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457911,
        "smsDt": 1578316448,
        "sender": "WeChat",
        "destination": "85261***977",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457911,
        "smsDt": 1578316448,
        "sender": "WeChat",
        "destination": "85261***977",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457912,
        "smsDt": 1578316509,
        "sender": "WeChat",
        "destination": "85251***769",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457882,
        "smsDt": 1578315289,
        "sender": "WeChat",
        "destination": "85265***193",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457882,
        "smsDt": 1578315288,
        "sender": "WeChat",
        "destination": "85265***193",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457753,
        "smsDt": 1578310124,
        "sender": "WeChat",
        "destination": "85263***353",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457753,
        "smsDt": 1578310123,
        "sender": "WeChat",
        "destination": "85263***353",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457717,
        "smsDt": 1578308682,
        "sender": "WeChat",
        "destination": "85260***931",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457717,
        "smsDt": 1578308681,
        "sender": "WeChat",
        "destination": "85260***931",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457759,
        "smsDt": 1578310389,
        "sender": "WeChat",
        "destination": "85262***560",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457864,
        "smsDt": 1578314560,
        "sender": "WeChat",
        "destination": "85267***259",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457949,
        "smsDt": 1578317992,
        "sender": "WeChat",
        "destination": "85294***497",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457949,
        "smsDt": 1578317991,
        "sender": "WeChat",
        "destination": "85294***497",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457833,
        "smsDt": 1578313336,
        "sender": "WeChat",
        "destination": "85266***024",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457823,
        "smsDt": 1578312958,
        "sender": "WeChat",
        "destination": "85263***649",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457935,
        "smsDt": 1578317424,
        "sender": "WeChat",
        "destination": "85257***986",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457935,
        "smsDt": 1578317423,
        "sender": "WeChat",
        "destination": "85257***986",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457621,
        "smsDt": 1578304872,
        "sender": "WeChat",
        "destination": "85257***294",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457621,
        "smsDt": 1578304870,
        "sender": "WeChat",
        "destination": "85257***294",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457942,
        "smsDt": 1578317718,
        "sender": "WeChat",
        "destination": "85257***499",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457942,
        "smsDt": 1578317717,
        "sender": "WeChat",
        "destination": "85257***499",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457816,
        "smsDt": 1578312664,
        "sender": "WeChat",
        "destination": "85294***727",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457816,
        "smsDt": 1578312664,
        "sender": "WeChat",
        "destination": "85294***727",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457664,
        "smsDt": 1578306573,
        "sender": "WeChat",
        "destination": "85254***867",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457664,
        "smsDt": 1578306572,
        "sender": "WeChat",
        "destination": "85254***867",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457978,
        "smsDt": 1578319159,
        "sender": "WeChat",
        "destination": "85263***824",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457874,
        "smsDt": 1578314971,
        "sender": "WeChat",
        "destination": "85297***493",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457874,
        "smsDt": 1578314971,
        "sender": "WeChat",
        "destination": "85297***493",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457798,
        "smsDt": 1578311932,
        "sender": "WeChat",
        "destination": "85263***034",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457798,
        "smsDt": 1578311931,
        "sender": "WeChat",
        "destination": "85263***034",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457840,
        "smsDt": 1578313633,
        "sender": "WeChat",
        "destination": "85262***893",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457741,
        "smsDt": 1578309646,
        "sender": "WeChat",
        "destination": "85263***522",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457741,
        "smsDt": 1578309645,
        "sender": "WeChat",
        "destination": "85263***522",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457794,
        "smsDt": 1578311792,
        "sender": "WeChat",
        "destination": "85260***187",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457794,
        "smsDt": 1578311790,
        "sender": "WeChat",
        "destination": "85260***187",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457940,
        "smsDt": 1578317607,
        "sender": "WeChat",
        "destination": "85265***347",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457950,
        "smsDt": 1578318021,
        "sender": "WeChat",
        "destination": "85264***805",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457872,
        "smsDt": 1578314893,
        "sender": "WeChat",
        "destination": "85296***822",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457975,
        "smsDt": 1578319032,
        "sender": "WeChat",
        "destination": "85257***495",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457975,
        "smsDt": 1578319030,
        "sender": "WeChat",
        "destination": "85257***495",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457813,
        "smsDt": 1578312543,
        "sender": "WeChat",
        "destination": "85264***958",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457813,
        "smsDt": 1578312542,
        "sender": "WeChat",
        "destination": "85264***958",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457896,
        "smsDt": 1578315858,
        "sender": "WeChat",
        "destination": "85298***684",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457896,
        "smsDt": 1578315856,
        "sender": "WeChat",
        "destination": "85298***684",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457772,
        "smsDt": 1578310893,
        "sender": "WeChat",
        "destination": "85262***633",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457865,
        "smsDt": 1578314638,
        "sender": "WeChat",
        "destination": "85257***679",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457865,
        "smsDt": 1578314638,
        "sender": "WeChat",
        "destination": "85257***679",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457622,
        "smsDt": 1578304888,
        "sender": "WeChat",
        "destination": "85257***443",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457622,
        "smsDt": 1578304885,
        "sender": "WeChat",
        "destination": "85257***443",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457888,
        "smsDt": 1578315557,
        "sender": "WeChat",
        "destination": "85257***054",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457888,
        "smsDt": 1578315556,
        "sender": "WeChat",
        "destination": "85257***054",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457934,
        "smsDt": 1578317385,
        "sender": "WeChat",
        "destination": "85267***986",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457659,
        "smsDt": 1578306381,
        "sender": "WeChat",
        "destination": "85265***586",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457724,
        "smsDt": 1578308973,
        "sender": "WeChat",
        "destination": "85251***298",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457894,
        "smsDt": 1578315780,
        "sender": "WeChat",
        "destination": "85257***114",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457894,
        "smsDt": 1578315779,
        "sender": "WeChat",
        "destination": "85257***114",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457620,
        "smsDt": 1578304828,
        "sender": "WeChat",
        "destination": "85257***584",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457620,
        "smsDt": 1578304826,
        "sender": "WeChat",
        "destination": "85257***584",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457882,
        "smsDt": 1578315318,
        "sender": "WeChat",
        "destination": "85257***468",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457882,
        "smsDt": 1578315316,
        "sender": "WeChat",
        "destination": "85257***468",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457860,
        "smsDt": 1578314415,
        "sender": "WeChat",
        "destination": "85297***968",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457713,
        "smsDt": 1578308526,
        "sender": "WeChat",
        "destination": "85257***447",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457713,
        "smsDt": 1578308526,
        "sender": "WeChat",
        "destination": "85257***447",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457893,
        "smsDt": 1578315738,
        "sender": "WeChat",
        "destination": "85297***455",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457893,
        "smsDt": 1578315737,
        "sender": "WeChat",
        "destination": "85297***455",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457794,
        "smsDt": 1578311790,
        "sender": "WeChat",
        "destination": "85261***958",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457794,
        "smsDt": 1578311789,
        "sender": "WeChat",
        "destination": "85261***958",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457706,
        "smsDt": 1578308256,
        "sender": "WeChat",
        "destination": "85257***709",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457706,
        "smsDt": 1578308256,
        "sender": "WeChat",
        "destination": "85257***709",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457978,
        "smsDt": 1578319120,
        "sender": "WeChat",
        "destination": "85257***122",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457960,
        "smsDt": 1578318420,
        "sender": "WeChat",
        "destination": "85253***639",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457960,
        "smsDt": 1578318419,
        "sender": "WeChat",
        "destination": "85253***639",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457803,
        "smsDt": 1578312153,
        "sender": "WeChat",
        "destination": "85257***344",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457803,
        "smsDt": 1578312152,
        "sender": "WeChat",
        "destination": "85257***344",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457648,
        "smsDt": 1578305947,
        "sender": "WeChat",
        "destination": "85263***774",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457648,
        "smsDt": 1578305945,
        "sender": "WeChat",
        "destination": "85263***774",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457929,
        "smsDt": 1578317169,
        "sender": "WeChat",
        "destination": "85267***835",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457929,
        "smsDt": 1578317168,
        "sender": "WeChat",
        "destination": "85267***835",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457895,
        "smsDt": 1578315823,
        "sender": "WeChat",
        "destination": "85261***191",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457885,
        "smsDt": 1578315435,
        "sender": "WeChat",
        "destination": "85265***293",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457885,
        "smsDt": 1578315433,
        "sender": "WeChat",
        "destination": "85265***293",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457823,
        "smsDt": 1578312949,
        "sender": "WeChat",
        "destination": "85263***013",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457644,
        "smsDt": 1578305796,
        "sender": "WeChat",
        "destination": "85290***504",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457644,
        "smsDt": 1578305795,
        "sender": "WeChat",
        "destination": "85290***504",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457879,
        "smsDt": 1578315181,
        "sender": "WeChat",
        "destination": "85251***450",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457676,
        "smsDt": 1578307077,
        "sender": "WeChat",
        "destination": "85294***274",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457676,
        "smsDt": 1578307076,
        "sender": "WeChat",
        "destination": "85294***274",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457937,
        "smsDt": 1578317488,
        "sender": "WeChat",
        "destination": "85296***319",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457946,
        "smsDt": 1578317863,
        "sender": "WeChat",
        "destination": "85261***613",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457946,
        "smsDt": 1578317862,
        "sender": "WeChat",
        "destination": "85261***613",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457956,
        "smsDt": 1578318244,
        "sender": "WeChat",
        "destination": "85265***726",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457956,
        "smsDt": 1578318243,
        "sender": "WeChat",
        "destination": "85265***726",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457635,
        "smsDt": 1578305400,
        "sender": "WeChat",
        "destination": "85256***429",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457952,
        "smsDt": 1578318084,
        "sender": "WeChat",
        "destination": "85256***767",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457952,
        "smsDt": 1578318083,
        "sender": "WeChat",
        "destination": "85256***767",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457918,
        "smsDt": 1578316736,
        "sender": "WeChat",
        "destination": "85265***416",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457764,
        "smsDt": 1578310596,
        "sender": "WeChat",
        "destination": "85267***204",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457764,
        "smsDt": 1578310594,
        "sender": "WeChat",
        "destination": "85267***204",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457759,
        "smsDt": 1578310366,
        "sender": "WeChat",
        "destination": "85298***516",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457759,
        "smsDt": 1578310365,
        "sender": "WeChat",
        "destination": "85298***516",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457977,
        "smsDt": 1578319119,
        "sender": "WeChat",
        "destination": "85257***122",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457884,
        "smsDt": 1578315399,
        "sender": "WeChat",
        "destination": "85291***809",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457970,
        "smsDt": 1578318814,
        "sender": "WeChat",
        "destination": "85290***442",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457970,
        "smsDt": 1578318813,
        "sender": "WeChat",
        "destination": "85290***442",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457976,
        "smsDt": 1578319042,
        "sender": "WeChat",
        "destination": "85265***910",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457976,
        "smsDt": 1578319042,
        "sender": "WeChat",
        "destination": "85265***910",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457814,
        "smsDt": 1578312589,
        "sender": "WeChat",
        "destination": "85267***761",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457814,
        "smsDt": 1578312589,
        "sender": "WeChat",
        "destination": "85267***761",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457959,
        "smsDt": 1578318383,
        "sender": "WeChat",
        "destination": "85297***188",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457967,
        "smsDt": 1578318687,
        "sender": "WeChat",
        "destination": "85265***779",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457896,
        "smsDt": 1578315845,
        "sender": "WeChat",
        "destination": "85253***299",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457896,
        "smsDt": 1578315845,
        "sender": "WeChat",
        "destination": "85253***299",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457955,
        "smsDt": 1578318229,
        "sender": "WeChat",
        "destination": "85292***563",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457955,
        "smsDt": 1578318228,
        "sender": "WeChat",
        "destination": "85292***563",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457794,
        "smsDt": 1578311786,
        "sender": "WeChat",
        "destination": "85292***276",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457971,
        "smsDt": 1578318840,
        "sender": "WeChat",
        "destination": "85261***580",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457949,
        "smsDt": 1578317993,
        "sender": "WeChat",
        "destination": "85261***824",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457949,
        "smsDt": 1578317992,
        "sender": "WeChat",
        "destination": "85261***824",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457888,
        "smsDt": 1578315532,
        "sender": "WeChat",
        "destination": "85267***008",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457632,
        "smsDt": 1578305288,
        "sender": "WeChat",
        "destination": "85263***880",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457705,
        "smsDt": 1578308220,
        "sender": "WeChat",
        "destination": "85260***438",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457705,
        "smsDt": 1578308219,
        "sender": "WeChat",
        "destination": "85260***438",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457699,
        "smsDt": 1578307992,
        "sender": "WeChat",
        "destination": "85269***482",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457699,
        "smsDt": 1578307990,
        "sender": "WeChat",
        "destination": "85269***482",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457777,
        "smsDt": 1578311109,
        "sender": "WeChat",
        "destination": "85262***080",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457721,
        "smsDt": 1578308871,
        "sender": "WeChat",
        "destination": "85265***403",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457729,
        "smsDt": 1578309198,
        "sender": "WeChat",
        "destination": "85265***403",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457669,
        "smsDt": 1578306780,
        "sender": "WeChat",
        "destination": "85257***427",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457669,
        "smsDt": 1578306778,
        "sender": "WeChat",
        "destination": "85257***427",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457820,
        "smsDt": 1578312814,
        "sender": "WeChat",
        "destination": "85295***974",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457820,
        "smsDt": 1578312813,
        "sender": "WeChat",
        "destination": "85295***974",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457933,
        "smsDt": 1578317338,
        "sender": "WeChat",
        "destination": "85291***792",
        "sendingStatus": "Sent",
        "dlrStatus": "Rejected",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457933,
        "smsDt": 1578317337,
        "sender": "WeChat",
        "destination": "85291***792",
        "sendingStatus": "Sent",
        "dlrStatus": "Rejected",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457770,
        "smsDt": 1578310826,
        "sender": "WeChat",
        "destination": "85267***465",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457770,
        "smsDt": 1578310824,
        "sender": "WeChat",
        "destination": "85267***465",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457910,
        "smsDt": 1578316404,
        "sender": "WeChat",
        "destination": "85262***786",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457881,
        "smsDt": 1578315265,
        "sender": "WeChat",
        "destination": "85266***253",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457881,
        "smsDt": 1578315263,
        "sender": "WeChat",
        "destination": "85266***253",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457885,
        "smsDt": 1578315402,
        "sender": "WeChat",
        "destination": "85291***809",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457842,
        "smsDt": 1578313702,
        "sender": "WeChat",
        "destination": "85252***385",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457842,
        "smsDt": 1578313700,
        "sender": "WeChat",
        "destination": "85252***385",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457620,
        "smsDt": 1578304831,
        "sender": "WeChat",
        "destination": "85257***235",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457620,
        "smsDt": 1578304830,
        "sender": "WeChat",
        "destination": "85257***235",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457796,
        "smsDt": 1578311868,
        "sender": "WeChat",
        "destination": "85257***331",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457796,
        "smsDt": 1578311867,
        "sender": "WeChat",
        "destination": "85257***331",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457916,
        "smsDt": 1578316655,
        "sender": "WeChat",
        "destination": "85251***682",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457916,
        "smsDt": 1578316654,
        "sender": "WeChat",
        "destination": "85251***682",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457936,
        "smsDt": 1578317454,
        "sender": "WeChat",
        "destination": "85267***387",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457936,
        "smsDt": 1578317453,
        "sender": "WeChat",
        "destination": "85267***387",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457753,
        "smsDt": 1578310141,
        "sender": "WeChat",
        "destination": "85257***588",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457753,
        "smsDt": 1578310140,
        "sender": "WeChat",
        "destination": "85257***588",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457760,
        "smsDt": 1578310407,
        "sender": "WeChat",
        "destination": "85263***740",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457760,
        "smsDt": 1578310405,
        "sender": "WeChat",
        "destination": "85263***740",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457886,
        "smsDt": 1578315455,
        "sender": "WeChat",
        "destination": "85294***430",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457886,
        "smsDt": 1578315454,
        "sender": "WeChat",
        "destination": "85294***430",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457726,
        "smsDt": 1578309060,
        "sender": "WeChat",
        "destination": "85255***210",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457726,
        "smsDt": 1578309059,
        "sender": "WeChat",
        "destination": "85255***210",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457827,
        "smsDt": 1578313099,
        "sender": "WeChat",
        "destination": "85263***297",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457827,
        "smsDt": 1578313096,
        "sender": "WeChat",
        "destination": "85263***297",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457701,
        "smsDt": 1578308041,
        "sender": "WeChat",
        "destination": "85253***553",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457701,
        "smsDt": 1578308040,
        "sender": "WeChat",
        "destination": "85253***553",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457708,
        "smsDt": 1578308332,
        "sender": "WeChat",
        "destination": "85257***844",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457708,
        "smsDt": 1578308331,
        "sender": "WeChat",
        "destination": "85257***844",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457919,
        "smsDt": 1578316775,
        "sender": "WeChat",
        "destination": "85269***448",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457820,
        "smsDt": 1578312829,
        "sender": "WeChat",
        "destination": "85291***577",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457932,
        "smsDt": 1578317289,
        "sender": "WeChat",
        "destination": "85263***448",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457932,
        "smsDt": 1578317287,
        "sender": "WeChat",
        "destination": "85263***448",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457900,
        "smsDt": 1578316034,
        "sender": "WeChat",
        "destination": "85263***423",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457900,
        "smsDt": 1578316032,
        "sender": "WeChat",
        "destination": "85263***423",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457928,
        "smsDt": 1578317127,
        "sender": "WeChat",
        "destination": "85262***166",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457878,
        "smsDt": 1578315133,
        "sender": "WeChat",
        "destination": "85263***034",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457878,
        "smsDt": 1578315132,
        "sender": "WeChat",
        "destination": "85263***034",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457833,
        "smsDt": 1578313336,
        "sender": "WeChat",
        "destination": "85262***047",
        "sendingStatus": "Sent",
        "dlrStatus": "Undelivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457833,
        "smsDt": 1578313333,
        "sender": "WeChat",
        "destination": "85262***047",
        "sendingStatus": "Sent",
        "dlrStatus": "Unknown",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457805,
        "smsDt": 1578312207,
        "sender": "WeChat",
        "destination": "85292***259",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457678,
        "smsDt": 1578307137,
        "sender": "WeChat",
        "destination": "85290***563",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457678,
        "smsDt": 1578307137,
        "sender": "WeChat",
        "destination": "85290***563",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457944,
        "smsDt": 1578317781,
        "sender": "WeChat",
        "destination": "85259***321",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457944,
        "smsDt": 1578317779,
        "sender": "WeChat",
        "destination": "85259***321",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457786,
        "smsDt": 1578311456,
        "sender": "WeChat",
        "destination": "85297***587",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457786,
        "smsDt": 1578311455,
        "sender": "WeChat",
        "destination": "85297***587",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457841,
        "smsDt": 1578313660,
        "sender": "WeChat",
        "destination": "85292***124",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457841,
        "smsDt": 1578313658,
        "sender": "WeChat",
        "destination": "85292***124",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457824,
        "smsDt": 1578312986,
        "sender": "WeChat",
        "destination": "85263***649",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457870,
        "smsDt": 1578314809,
        "sender": "WeChat",
        "destination": "85259***080",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457870,
        "smsDt": 1578314808,
        "sender": "WeChat",
        "destination": "85259***080",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457650,
        "smsDt": 1578306015,
        "sender": "WeChat",
        "destination": "85255***316",
        "sendingStatus": "Sent",
        "dlrStatus": "Rejected",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457650,
        "smsDt": 1578306014,
        "sender": "WeChat",
        "destination": "85255***316",
        "sendingStatus": "Sent",
        "dlrStatus": "Rejected",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "failure"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457977,
        "smsDt": 1578319090,
        "sender": "WeChat",
        "destination": "85257***641",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457977,
        "smsDt": 1578319089,
        "sender": "WeChat",
        "destination": "85257***641",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457742,
        "smsDt": 1578309710,
        "sender": "WeChat",
        "destination": "85265***403",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      }
    ],
    "status": "stage1"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457800,
        "smsDt": 1578312033,
        "sender": "WeChat",
        "destination": "85257***489",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457800,
        "smsDt": 1578312032,
        "sender": "WeChat",
        "destination": "85257***489",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457907,
        "smsDt": 1578316304,
        "sender": "WeChat",
        "destination": "85266***253",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457907,
        "smsDt": 1578316303,
        "sender": "WeChat",
        "destination": "85266***253",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457877,
        "smsDt": 1578315094,
        "sender": "WeChat",
        "destination": "85253***714",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457877,
        "smsDt": 1578315093,
        "sender": "WeChat",
        "destination": "85253***714",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457716,
        "smsDt": 1578308670,
        "sender": "WeChat",
        "destination": "85254***010",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457716,
        "smsDt": 1578308669,
        "sender": "WeChat",
        "destination": "85254***010",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457913,
        "smsDt": 1578316551,
        "sender": "WeChat",
        "destination": "85257***350",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457913,
        "smsDt": 1578316550,
        "sender": "WeChat",
        "destination": "85257***350",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457979,
        "smsDt": 1578319168,
        "sender": "WeChat",
        "destination": "85260***276",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457979,
        "smsDt": 1578319167,
        "sender": "WeChat",
        "destination": "85260***276",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  },
  {
    "params": [
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457757,
        "smsDt": 1578310292,
        "sender": "WeChat",
        "destination": "85298***190",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "A2P_PCCWG",
        "nextStop": "",
        "from": "qlc_3exbms47d63ywggnhb9iko9twphsnsx563qf6faufp33167o5dqfoawa8gtj"
      },
      {
        "contractAddress": "qlc_3671s4ij7nmpm1r39ag395zcqfr8qhoincn8fygorknquuq47bg8bj36sw4k",
        "index": 39457757,
        "smsDt": 1578310292,
        "sender": "WeChat",
        "destination": "85298***190",
        "sendingStatus": "Sent",
        "dlrStatus": "Delivered",
        "preStop": "",
        "nextStop": "CSL Hong Kong @ 3397",
        "from": "qlc_3giz1uwgsmq46xzspo9mbutade6foqh5fuja4m9rwfiuyzp4x8zu5hkorq4z"
      }
    ],
    "status": "success"
  }
]
`
)

func buildContractParam() (param *ContractParam) {
	cp := createContractParam
	cp.PartyA.Address = mock.Address()
	cp.PartyB.Address = mock.Address()

	cd := time.Now().Unix()
	param = &ContractParam{
		CreateContractParam: cp,
		PreStops:            []string{"PCCWG", "test1"},
		NextStops:           []string{"HKTCSL", "test2"},
		ConfirmDate:         cd,
		Status:              ContractStatusActivated,
	}
	return
}

func TestGetContractsByAddress(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	var contracts []*ContractParam

	for i := 0; i < 4; i++ {
		param := buildContractParam()
		contracts = append(contracts, param)
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	//if err := ctx.Iterator(types.SettlementAddress[:], func(key []byte, value []byte) error {
	//	t.Log(hex.EncodeToString(key), " >>> ", hex.EncodeToString(value))
	//	return nil
	//}); err != nil {
	//	t.Fatal(err)
	//}

	if contracts == nil || len(contracts) != 4 {
		t.Fatalf("invalid mock contract data, exp: 4, act: %d", len(contracts))
	}

	a := contracts[0].PartyA.Address

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "1st",
			args: args{
				ctx:  ctx,
				addr: &a,
			},
			want:    []*ContractParam{contracts[0]},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsByAddress(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContractsByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("GetContractsByAddress() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("GetContractsByAddress() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func mockContractData(size int) []*ContractParam {
	var contracts []*ContractParam
	accounts := []types.Address{mock.Address(), mock.Address()}

	for i := 0; i < size; i++ {
		cp := createContractParam

		var a1 types.Address
		var a2 types.Address
		if i%2 == 0 {
			a1 = accounts[0]
			a2 = accounts[1]
		} else {
			a1 = accounts[1]
			a2 = accounts[0]
		}
		cp.PartyA.Address = a1
		cp.PartyB.Address = a2

		for _, s := range cp.Services {
			s.Mcc = s.Mcc + uint64(i)
			s.Mnc = s.Mnc + uint64(i)
		}

		cd := time.Now().Unix()
		param := &ContractParam{
			CreateContractParam: cp,
			ConfirmDate:         cd,
		}
		contracts = append(contracts, param)
	}
	return contracts
}

func TestParseContractParam(t *testing.T) {
	param := buildContractParam()
	abi, err := param.ToABI()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		v []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				v: abi,
			},
			want:    param,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseContractParam(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContractParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseContractParam() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCDRStatus(t *testing.T) {
	param1 := cdrParam

	if data, err := param1.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		cp := &CDRParam{}
		if err := cp.FromABI(data); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(cp, &param1) {
				t.Fatal("invalid param")
			} else {
				t.Log(cp.String())
			}
		}
	}

	param2 := param1
	param2.Sender = "PCCWG"

	a1 := mock.Address().String()
	a2 := mock.Address().String()
	s := CDRStatus{
		Params: map[string][]CDRParam{a1: {param1}, a2: {param2}},
		Status: 0,
	}

	if data, err := s.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		s1 := &CDRStatus{}
		if err := s1.FromABI(data); err != nil {
			t.Fatal(err)
		} else {
			if len(s1.Params) != 2 {
				t.Fatalf("invalid param size, exp: 2, act:%d", len(s1.Params))
			}
			if !reflect.DeepEqual(param1, s1.Params[a1][0]) {
				t.Fatalf("invalid csl data, exp: %s, act: %s", util.ToIndentString(param1), util.ToIndentString(s1.Params[a1][0]))
			}
			if !reflect.DeepEqual(param2, s1.Params[a2][0]) {
				t.Fatal("invalid pccwg data")
			}
			t.Log(s1.String())
		}
	}
}

func TestGetAllSettlementContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	for i := 0; i < 4; i++ {
		param := buildContractParam()
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  *vmstore.VMContext
		size int
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx:  ctx,
				size: 4,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllSettlementContract(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllSettlementContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.args.size {
				t.Fatalf("invalid size: exp: %d, act: %d", tt.args.size, len(got))
			}
			for _, c := range got {
				t.Log(c.String())
			}
		})
	}
}

func TestCDRParam_Verify(t *testing.T) {
	type fields struct {
		Index       uint64
		SmsDt       int64
		Sender      string
		Destination string
		//DstCountry    string
		//DstOperator   string
		//DstMcc        uint64
		//DstMnc        uint64
		//SellPrice     float64
		//SellCurrency  string
		//CustomerName  string
		//CustomerID    string
		SendingStatus SendingStatus
		DlrStatus     DLRStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Index:       1,
				SmsDt:       time.Now().Unix(),
				Sender:      "PCCWG",
				Destination: "85257***343",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: false,
		},
		{
			name: "f1",
			fields: fields{
				Index:       0,
				SmsDt:       time.Now().Unix(),
				Sender:      "PCCWG",
				Destination: "85257***343",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: true,
		},
		{
			name: "f2",
			fields: fields{
				Index:       1,
				SmsDt:       0,
				Sender:      "PCCWG",
				Destination: "85257***343",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: true,
		},
		{
			name: "f3",
			fields: fields{
				Index:       1,
				SmsDt:       time.Now().Unix(),
				Sender:      "",
				Destination: "85257***343",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: true,
		},
		{
			name: "f4",
			fields: fields{
				Index:       1,
				SmsDt:       time.Now().Unix(),
				Sender:      "PCCWG",
				Destination: "",
				//DstCountry:    "Hong Kong",
				//DstOperator:   "HKTCSL",
				//DstMcc:        454,
				//DstMnc:        0,
				//SellPrice:     1,
				//SellCurrency:  "USD",
				//CustomerName:  "Tencent",
				//CustomerID:    "11667",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
			},
			wantErr: true,
		},
		//{
		//	name: "f5",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		//DstCountry:    "",
		//		//DstOperator:   "HKTCSL",
		//		//DstMcc:        454,
		//		//DstMnc:        0,
		//		//SellPrice:     1,
		//		//SellCurrency:  "USD",
		//		//CustomerName:  "Tencent",
		//		//CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//}, {
		//	name: "f6",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		//DstCountry:    "Hong Kong",
		//		//DstOperator:   "",
		//		//DstMcc:        454,
		//		//DstMnc:        0,
		//		//SellPrice:     1,
		//		//SellCurrency:  "USD",
		//		//CustomerName:  "Tencent",
		//		//CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//},
		////{
		////	name: "f7",
		////	fields: fields{
		////		Index:         1,
		////		SmsDt:         time.Now().Unix(),
		////		Sender:        "PCCWG",
		////		Destination:   "85257***343",
		////		DstCountry:    "Hong Kong",
		////		DstOperator:   "HKTCSL",
		////		DstMcc:        0,
		////		DstMnc:        0,
		////		SellPrice:     1,
		////		SellCurrency:  "USD",
		////		CustomerName:  "Tencent",
		////		CustomerID:    "11667",
		////		SendingStatus: "Send",
		////		DlrStatus:     "Delivered",
		////	},
		////	wantErr: true,
		////}, {
		////	name: "f8",
		////	fields: fields{
		////		Index:         1,
		////		SmsDt:         time.Now().Unix(),
		////		Sender:        "PCCWG",
		////		Destination:   "85257***343",
		////		DstCountry:    "Hong Kong",
		////		DstOperator:   "HKTCSL",
		////		DstMcc:        454,
		////		DstMnc:        0,
		////		SellPrice:     1,
		////		SellCurrency:  "USD",
		////		CustomerName:  "Tencent",
		////		CustomerID:    "11667",
		////		SendingStatus: "Send",
		////		DlrStatus:     "Delivered",
		////	},
		////	wantErr: true,
		////},
		//{
		//	name: "f9",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		DstCountry:    "Hong Kong",
		//		DstOperator:   "HKTCSL",
		//		DstMcc:        454,
		//		DstMnc:        0,
		//		SellPrice:     0,
		//		SellCurrency:  "USD",
		//		CustomerName:  "Tencent",
		//		CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//}, {
		//	name: "f10",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		DstCountry:    "Hong Kong",
		//		DstOperator:   "HKTCSL",
		//		DstMcc:        454,
		//		DstMnc:        0,
		//		SellPrice:     1,
		//		SellCurrency:  "",
		//		CustomerName:  "Tencent",
		//		CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//}, {
		//	name: "f11",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		DstCountry:    "Hong Kong",
		//		DstOperator:   "HKTCSL",
		//		DstMcc:        454,
		//		DstMnc:        0,
		//		SellPrice:     1,
		//		SellCurrency:  "USD",
		//		CustomerName:  "",
		//		CustomerID:    "11667",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//}, {
		//	name: "f12",
		//	fields: fields{
		//		Index:         1,
		//		SmsDt:         time.Now().Unix(),
		//		Sender:        "PCCWG",
		//		Destination:   "85257***343",
		//		DstCountry:    "Hong Kong",
		//		DstOperator:   "HKTCSL",
		//		DstMcc:        454,
		//		DstMnc:        0,
		//		SellPrice:     1,
		//		SellCurrency:  "USD",
		//		CustomerName:  "Tencent",
		//		CustomerID:    "",
		//		SendingStatus: SendingStatusSend,
		//		DlrStatus:     DLRStatusDelivered,
		//	},
		//	wantErr: true,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRParam{
				Index:       tt.fields.Index,
				SmsDt:       tt.fields.SmsDt,
				Sender:      tt.fields.Sender,
				Destination: tt.fields.Destination,
				//DstCountry:    tt.fields.DstCountry,
				//DstOperator:   tt.fields.DstOperator,
				//DstMcc:        tt.fields.DstMcc,
				//DstMnc:        tt.fields.DstMnc,
				//SellPrice:     tt.fields.SellPrice,
				//SellCurrency:  tt.fields.SellCurrency,
				//CustomerName:  tt.fields.CustomerName,
				//CustomerID:    tt.fields.CustomerID,
				SendingStatus: tt.fields.SendingStatus,
				DlrStatus:     tt.fields.DlrStatus,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseCDRStatus(t *testing.T) {
	status := &CDRStatus{
		Params: map[string][]CDRParam{mock.Address().String(): {cdrParam}},
		Status: SettlementStatusStage1,
	}

	if abi, err := status.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		if s2, err := ParseCDRStatus(abi); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(status, s2) {
				t.Fatalf("invalid cdr status %v, %v", status, s2)
			}
		}
	}
}

func TestCDRParam_ToHash(t *testing.T) {
	if h, err := cdrParam.ToHash(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(h)
	}
}

func TestGetContracts(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	ctx := vmstore.NewVMContext(l)
	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		} else {
			//t.Log(hex.EncodeToString(abi))
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	for _, d := range data {
		if addr, err := d.Address(); err != nil {
			t.Fatal(err)
		} else {
			if contract, err := GetSettlementContract(ctx, &addr); err != nil {
				t.Fatal(err)
			} else {
				if !reflect.DeepEqual(contract, d) {
					t.Fatalf("invalid %v, %v", contract, d)
				}
			}
		}
	}
}

func TestCDRStatus_DoSettlement(t *testing.T) {
	addr1 := mock.Address()
	smsDt := time.Now().Unix()

	type fields struct {
		Params map[string][]CDRParam
		Status SettlementStatus
	}

	type args struct {
		cdr    SettlementCDR
		status SettlementStatus
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "1 records",
			fields: fields{
				Params: nil,
				Status: SettlementStatusStage1,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         time.Now().Unix(),
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
					From: mock.Address(),
				},
				status: SettlementStatusStage1,
			},
			wantErr: false,
		},
		{
			name: "2 records",
			fields: fields{
				Params: map[string][]CDRParam{mock.Address().String(): {
					{
						Index:       1,
						SmsDt:       time.Now().Unix(),
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
				}},
				Status: 0,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         time.Now().Unix(),
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
					From: mock.Address(),
				},
				status: SettlementStatusSuccess,
			},
			wantErr: false,
		}, {
			name: "2 records failed",
			fields: fields{
				Params: map[string][]CDRParam{mock.Address().String(): {
					{
						Index:       1,
						SmsDt:       time.Now().Unix(),
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusRejected,
					},
				}},
				Status: 0,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         time.Now().Unix(),
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusRejected,
					},
					From: mock.Address(),
				},
				status: SettlementStatusFailure,
			},
			wantErr: false,
		}, {
			name: "2 duplicate records of one account",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {
					{
						Index:       1,
						SmsDt:       smsDt,
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusUndelivered,
					},
				}},
				Status: SettlementStatusFailure,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         smsDt,
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
					From: addr1,
				},
				status: SettlementStatusStage1,
			},
			wantErr: false,
		},
		{
			name: "2 duplicate records",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {
					{
						Index:       1,
						SmsDt:       smsDt,
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusUndelivered,
					},
				}, mock.Address().String(): {
					{
						Index:       1,
						SmsDt:       smsDt,
						Sender:      "PCCWG",
						Destination: "85257***343",
						//DstCountry:    "Hong Kong",
						//DstOperator:   "HKTCSL",
						//DstMcc:        454,
						//DstMnc:        0,
						//SellPrice:     1,
						//SellCurrency:  "USD",
						//CustomerName:  "Tencent",
						//CustomerID:    "11668",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusUndelivered,
					},
				},
				},
				Status: SettlementStatusFailure,
			},
			args: args{
				cdr: SettlementCDR{
					CDRParam: CDRParam{
						Index:         1,
						SmsDt:         smsDt,
						Sender:        "HKTCSL",
						Destination:   "85257***343",
						SendingStatus: SendingStatusSent,
						DlrStatus:     DLRStatusDelivered,
					},
					From: addr1,
				},
				status: SettlementStatusDuplicate,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			if err := z.DoSettlement(tt.args.cdr); (err != nil) != tt.wantErr {
				t.Errorf("DoSettlement() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if z.Status != tt.args.status {
					t.Fatalf("invalid status, exp: %s, act: %s", tt.args.status.String(), z.Status.String())
				}
			}
		})
	}
}

func TestContractParam_FromABI(t *testing.T) {
	cp := buildContractParam()

	abi, err := cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				data: abi,
			},
			wantErr: false,
		}, {
			name: "failed",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				data: abi[:10],
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
			}
			if err := z.FromABI(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("FromABI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContractService_Balance(t *testing.T) {
	type fields struct {
		ServiceId   string
		Mcc         uint64
		Mnc         uint64
		TotalAmount uint64
		UnitPrice   float64
		Currency    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    types.Balance
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				ServiceId:   mock.Hash().String(),
				Mcc:         22,
				Mnc:         1,
				TotalAmount: 100,
				UnitPrice:   0.04,
				Currency:    "USD",
			},
			want:    types.Balance{Int: new(big.Int).Mul(new(big.Int).SetUint64(100), new(big.Int).SetUint64(0.04*1e8))},
			wantErr: false,
		}, {
			name: "overflow",
			fields: fields{
				ServiceId:   mock.Hash().String(),
				Mcc:         22,
				Mnc:         1,
				TotalAmount: math.MaxUint64,
				UnitPrice:   0.04,
				Currency:    "USD",
			},
			want:    types.ZeroBalance,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractService{
				ServiceId:   tt.fields.ServiceId,
				Mcc:         tt.fields.Mcc,
				Mnc:         tt.fields.Mnc,
				TotalAmount: tt.fields.TotalAmount,
				UnitPrice:   tt.fields.UnitPrice,
				Currency:    tt.fields.Currency,
			}
			got, err := z.Balance()
			if (err != nil) != tt.wantErr {
				t.Errorf("Balance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Balance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateContractParam_Balance(t *testing.T) {
	type fields struct {
		PartyA    Contractor
		PartyB    Contractor
		Previous  types.Hash
		Services  []ContractService
		SignDate  int64
		StartDate int64
		EndData   int64
	}
	tests := []struct {
		name    string
		fields  fields
		want    types.Balance
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				PartyA: Contractor{
					Address: mock.Address(),
					Name:    "PCCWG",
				},
				PartyB: Contractor{
					Address: mock.Address(),
					Name:    "HKTCSL",
				},
				Previous: mock.Hash(),
				Services: []ContractService{{
					ServiceId:   mock.Hash().String(),
					Mcc:         1,
					Mnc:         2,
					TotalAmount: 100,
					UnitPrice:   2,
					Currency:    "USD",
				}, {
					ServiceId:   mock.Hash().String(),
					Mcc:         22,
					Mnc:         1,
					TotalAmount: 300,
					UnitPrice:   4,
					Currency:    "USD",
				}},
				SignDate:  time.Now().Unix(),
				StartDate: time.Now().AddDate(0, 0, 2).Unix(),
				EndData:   time.Now().AddDate(1, 0, 2).Unix(),
			},
			want:    types.Balance{Int: new(big.Int).Mul(big.NewInt(100), big.NewInt(2*1e8))}.Add(types.Balance{Int: new(big.Int).Mul(big.NewInt(300), big.NewInt(4*1e8))}),
			wantErr: false,
		},
		{
			name: "overflow",
			fields: fields{
				PartyA: Contractor{
					Address: mock.Address(),
					Name:    "PCCWG",
				},
				PartyB: Contractor{
					Address: mock.Address(),
					Name:    "HKTCSL",
				},
				Previous: mock.Hash(),
				Services: []ContractService{{
					ServiceId:   mock.Hash().String(),
					Mcc:         1,
					Mnc:         2,
					TotalAmount: 100,
					UnitPrice:   2,
					Currency:    "USD",
				}, {
					ServiceId:   mock.Hash().String(),
					Mcc:         22,
					Mnc:         1,
					TotalAmount: math.MaxUint64,
					UnitPrice:   4,
					Currency:    "USD",
				}},
				SignDate:  time.Now().Unix(),
				StartDate: time.Now().AddDate(0, 0, 2).Unix(),
				EndData:   time.Now().AddDate(1, 0, 2).Unix(),
			},
			want:    types.ZeroBalance,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:    tt.fields.PartyA,
				PartyB:    tt.fields.PartyB,
				Previous:  tt.fields.Previous,
				Services:  tt.fields.Services,
				SignDate:  tt.fields.SignDate,
				StartDate: tt.fields.StartDate,
				EndDate:   tt.fields.EndData,
			}
			got, err := z.Balance()
			if (err != nil) != tt.wantErr {
				t.Errorf("Balance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Balance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateContractParam_ToABI(t *testing.T) {
	cp := createContractParam

	abi, err := cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	cp2 := &CreateContractParam{}
	if err = cp2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		a1, err := cp.Address()
		if err != nil {
			t.Fatal(err)
		}
		a2, err := cp2.Address()
		if err != nil {
			t.Fatal(err)
		}
		if a1 != a2 {
			t.Fatalf("invalid create contract params, %v, %v", cp, cp2)
		} else {
			t.Log(cp2.String())
		}
	}
}

func TestCDRParam_FromABI(t *testing.T) {
	cdr := cdrParam
	abi, err := cdr.ToABI()
	if err != nil {
		t.Fatal(err)
	}
	type fields struct {
		Index         uint64
		SmsDt         int64
		Sender        string
		Destination   string
		DstCountry    string
		DstOperator   string
		DstMcc        uint64
		DstMnc        uint64
		SellPrice     float64
		SellCurrency  string
		CustomerName  string
		CustomerID    string
		SendingStatus SendingStatus
		DlrStatus     DLRStatus
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Index:       cdr.Index,
				SmsDt:       cdr.SmsDt,
				Sender:      cdr.Sender,
				Destination: cdr.Destination,
				//DstCountry:    cdr.DstCountry,
				//DstOperator:   cdr.DstOperator,
				//DstMcc:        cdr.DstMcc,
				//DstMnc:        cdr.DstMnc,
				//SellPrice:     cdr.SellPrice,
				//SellCurrency:  cdr.SellCurrency,
				//CustomerName:  cdr.CustomerName,
				//CustomerID:    cdr.CustomerID,
				SendingStatus: cdr.SendingStatus,
				DlrStatus:     cdr.DlrStatus,
			},
			args: args{
				data: abi,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRParam{
				Index:       tt.fields.Index,
				SmsDt:       tt.fields.SmsDt,
				Sender:      tt.fields.Sender,
				Destination: tt.fields.Destination,
				//DstCountry:    tt.fields.DstCountry,
				//DstOperator:   tt.fields.DstOperator,
				//DstMcc:        tt.fields.DstMcc,
				//DstMnc:        tt.fields.DstMnc,
				//SellPrice:     tt.fields.SellPrice,
				//SellCurrency:  tt.fields.SellCurrency,
				//CustomerName:  tt.fields.CustomerName,
				//CustomerID:    tt.fields.CustomerID,
				SendingStatus: tt.fields.SendingStatus,
				DlrStatus:     tt.fields.DlrStatus,
			}
			if err := z.FromABI(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("FromABI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCDRParam_String(t *testing.T) {
	s := cdrParam.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}
}

func TestCDRStatus_FromABI(t *testing.T) {
	status := CDRStatus{
		Params: map[string][]CDRParam{
			mock.Address().String(): {cdrParam},
		},
		Status: SettlementStatusSuccess,
	}

	abi, err := status.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	s2 := &CDRStatus{}
	if err = s2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&status, s2) {
			t.Fatalf("invalid cdr status data, %v, %v", &status, s2)
		}
	}
}

func TestCDRStatus_String(t *testing.T) {
	status := CDRStatus{
		Params: map[string][]CDRParam{
			mock.Address().String(): {cdrParam},
		},
		Status: SettlementStatusSuccess,
	}
	s := status.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}

}

func TestContractParam_Equal(t *testing.T) {
	param := buildContractParam()
	cp := param.CreateContractParam

	type fields struct {
		CreateContractParam CreateContractParam
		ConfirmDate         int64
		SignatureB          *types.Signature
	}
	type args struct {
		cp *CreateContractParam
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "equal",
			fields: fields{
				CreateContractParam: cp,
				ConfirmDate:         param.ConfirmDate,
			},
			args: args{
				cp: &cp,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "nil",
			fields: fields{
				CreateContractParam: cp,
				ConfirmDate:         param.ConfirmDate,
			},
			args: args{
				cp: nil,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				ConfirmDate:         tt.fields.ConfirmDate,
			}
			got, err := z.Equal(tt.args.cp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Equal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Equal() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContractParam_String(t *testing.T) {
	param := buildContractParam()
	s := param.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}
}

func TestContractParam_ToABI(t *testing.T) {
	param := buildContractParam()
	abi, err := param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	p2 := &ContractParam{}
	if err = p2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		if param.String() != p2.String() {
			t.Fatalf("invalid contract param data, %s, %s", param.String(), p2.String())
		}
	}
}

func TestContractService_ToABI(t *testing.T) {
	param := ContractService{
		ServiceId:   mock.Hash().String(),
		Mcc:         1,
		Mnc:         2,
		TotalAmount: 100,
		UnitPrice:   2,
		Currency:    "USD",
	}
	abi, err := param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	p2 := &ContractService{}
	if err = p2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&param, p2) {
			t.Fatalf("invalid contract service data, %v, %v", &param, p2)
		}
	}
}

func TestContractor_FromABI(t *testing.T) {
	c := createContractParam.PartyA
	abi, err := c.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	c2 := &Contractor{}
	if err = c2.FromABI(abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&c, c2) {
			t.Fatalf("invalid contractor, %v, %v", &c, c2)
		}
	}
}

func TestCreateContractParam_Address(t *testing.T) {
	cp := createContractParam
	if address, err := cp.Address(); err != nil {
		t.Fatal(err)
	} else {
		cp2 := createContractParam
		if address2, err := cp2.Address(); err != nil {
			t.Fatal(err)
		} else {
			if address != address2 {
				t.Fatalf("invalid address, %v, %v", address, address2)
			}
		}
	}
}

func TestCreateContractParam_String(t *testing.T) {
	param := buildContractParam()
	s := param.String()
	if len(s) == 0 {
		t.Fatal("invalid string")
	}
}

func TestCreateContractParam_ToContractParam(t *testing.T) {
	cp := createContractParam
	type fields struct {
		PartyA    Contractor
		PartyB    Contractor
		Previous  types.Hash
		Services  []ContractService
		SignDate  int64
		StartDate int64
		EndDate   int64
	}
	tests := []struct {
		name   string
		fields fields
		want   *ContractParam
	}{
		{
			name: "OK",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want: &ContractParam{
				CreateContractParam: cp,
				PreStops:            nil,
				NextStops:           nil,
				ConfirmDate:         0,
				Status:              0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:    tt.fields.PartyA,
				PartyB:    tt.fields.PartyB,
				Previous:  tt.fields.Previous,
				Services:  tt.fields.Services,
				SignDate:  tt.fields.SignDate,
				StartDate: tt.fields.StartDate,
				EndDate:   cp.EndDate,
			}
			if got := z.ToContractParam(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToContractParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateContractParam_Verify(t *testing.T) {
	cp := createContractParam

	type fields struct {
		PartyA    Contractor
		PartyB    Contractor
		Previous  types.Hash
		Services  []ContractService
		SignDate  int64
		StartDate int64
		EndDate   int64
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "f1",
			fields: fields{
				PartyA: Contractor{
					Address: types.ZeroAddress,
					Name:    "CC",
				},
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				PartyA: Contractor{
					Address: mock.Address(),
					Name:    "",
				},
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				PartyA: cp.PartyA,
				PartyB: Contractor{
					Address: mock.Address(),
					Name:    "",
				},
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f4",
			fields: fields{
				PartyA: cp.PartyA,
				PartyB: Contractor{
					Address: types.ZeroAddress,
					Name:    "CC",
				},
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f5",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  types.ZeroHash,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f6",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  nil,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f7",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  0,
				StartDate: cp.StartDate,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f8",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: 0,
				EndDate:   cp.EndDate,
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f9",
			fields: fields{
				PartyA:    cp.PartyA,
				PartyB:    cp.PartyB,
				Previous:  cp.Previous,
				Services:  cp.Services,
				SignDate:  cp.SignDate,
				StartDate: cp.StartDate,
				EndDate:   0,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CreateContractParam{
				PartyA:    tt.fields.PartyA,
				PartyB:    tt.fields.PartyB,
				Previous:  tt.fields.Previous,
				Services:  tt.fields.Services,
				SignDate:  tt.fields.SignDate,
				StartDate: tt.fields.StartDate,
				EndDate:   tt.fields.EndDate,
			}
			got, err := z.Verify()
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Verify() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetContractsIDByAddressAsPartyA(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)
	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	a1 := data[0].PartyA.Address

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &a1,
			},
			want:    []*ContractParam{data[0]},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsIDByAddressAsPartyA(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContractsIDByAddressAsPartyA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("GetContractsIDByAddressAsPartyA() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("GetContractsIDByAddressAsPartyA() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetContractsIDByAddressAsPartyB(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)
	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	a2 := data[0].PartyB.Address

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &a2,
			},
			want:    []*ContractParam{data[0]},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsIDByAddressAsPartyB(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetContractsIDByAddressAsPartyB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("TestGetContractsIDByAddressAsPartyB() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("TestGetContractsIDByAddressAsPartyB() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIsContractAvailable(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)

	var contracts []*ContractParam

	for i := 0; i < 2; i++ {
		param := buildContractParam()

		if i%2 == 1 {
			param.Status = ContractStatusActiveStage1
		}

		contracts = append(contracts, param)
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}

		if storage, err := ctx.GetStorage(types.SettlementAddress[:], a[:]); err == nil {
			if !bytes.Equal(storage, abi) {
				t.Fatalf("invalid saved contract, exp: %v, act: %v", abi, storage)
			} else {
				if p, err := ParseContractParam(storage); err == nil {
					t.Log(a.String(), ": ", p.String())
				} else {
					t.Fatal(err)
				}
			}
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	if contracts == nil || len(contracts) != 2 {
		t.Fatal("invalid mock contract data")
	}

	a1, _ := contracts[0].Address()
	a2, _ := contracts[1].Address()

	t.Log(a1.String(), " >>> ", a2.String())

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &a1,
			},
			want: true,
		},
		{
			name: "fail",
			args: args{
				ctx:  ctx,
				addr: &a2,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContractAvailable(tt.args.ctx, tt.args.addr); got != tt.want {
				t.Errorf("IsContractAvailable() of %s = %v, want %v", tt.args.addr.String(), got, tt.want)
			}
		})
	}
}

func TestSignContractParam_ToABI(t *testing.T) {
	sc := &SignContractParam{
		ContractAddress: mock.Address(),
		ConfirmDate:     time.Now().Unix(),
	}

	if abi, err := sc.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		sc2 := &SignContractParam{}
		if err := sc2.FromABI(abi); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(sc, sc2) {
				t.Errorf("ToABI() got = %v, want %v", sc2, sc)
			}
		}
	}
}

func TestGetSettlementContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)

	var contracts []*ContractParam

	for i := 0; i < 2; i++ {
		param := buildContractParam()

		if i%2 == 1 {
			param.Status = ContractStatusActiveStage1
		}

		contracts = append(contracts, param)
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], a[:], abi[:]); err != nil {
			t.Fatal(err)
		}

		//if storage, err := ctx.GetStorage(types.SettlementAddress[:], a[:]); err == nil {
		//	if !bytes.Equal(storage, abi) {
		//		t.Fatalf("invalid saved contract, exp: %v, act: %v", abi, storage)
		//	} else {
		//		if p, err := ParseContractParam(storage); err == nil {
		//			t.Log(a.String(), ": ", p.String())
		//		} else {
		//			t.Fatal(err)
		//		}
		//	}
		//}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	if contracts == nil || len(contracts) != 2 {
		t.Fatal("invalid mock contract data")
	}
	a1 := contracts[0].PartyA.Address
	cdr := cdrParam
	cdr.NextStop = "HKTCSL"

	if c, err := FindSettlementContract(ctx, &a1, &cdr); err != nil {
		t.Fatal(err)
	} else {
		t.Log(c)
	}
}

func TestStopParam_Verify(t *testing.T) {
	type fields struct {
		StopName string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				StopName: "test1",
			},
			wantErr: false,
		}, {
			name: "false",
			fields: fields{
				StopName: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &StopParam{
				StopName: tt.fields.StopName,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateStopParam_Verify(t *testing.T) {
	type fields struct {
		StopName string
		NewName  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				StopName: "test1",
				NewName:  "test2",
			},
			wantErr: false,
		}, {
			name: "false",
			fields: fields{
				StopName: "",
				NewName:  "222",
			},
			wantErr: true,
		}, {
			name: "false2",
			fields: fields{
				StopName: "111",
				NewName:  "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &UpdateStopParam{
				StopName: tt.fields.StopName,
				New:      tt.fields.NewName,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSignContractParam_Verify(t *testing.T) {
	type fields struct {
		ContractAddress types.Address
		ConfirmDate     int64
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				ContractAddress: mock.Address(),
				ConfirmDate:     time.Now().Unix(),
			},
			want:    true,
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				ContractAddress: types.ZeroAddress,
				ConfirmDate:     time.Now().Unix(),
			},
			want:    false,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				ContractAddress: mock.Address(),
				ConfirmDate:     0,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &SignContractParam{
				ContractAddress: tt.fields.ContractAddress,
				ConfirmDate:     tt.fields.ConfirmDate,
			}
			got, err := z.Verify()
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Verify() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContractParam_IsPreStop(t *testing.T) {
	param := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
	}
	type args struct {
		n string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "ok",
			fields: fields{
				CreateContractParam: param.CreateContractParam,
				PreStops:            []string{"PCCWG", "111"},
				NextStops:           nil,
				ConfirmDate:         param.ConfirmDate,
				Status:              param.Status,
			},
			args: args{
				n: "PCCWG",
			},
			want: true,
		}, {
			name: "ok",
			fields: fields{
				CreateContractParam: param.CreateContractParam,
				PreStops:            []string{"PCCWG", "111"},
				NextStops:           nil,
				ConfirmDate:         param.ConfirmDate,
				Status:              param.Status,
			},
			args: args{
				n: "PCCWG1",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
			}
			if got := z.IsPreStop(tt.args.n); got != tt.want {
				t.Errorf("IsPreStop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCDRParam_Status(t *testing.T) {
	type fields struct {
		Index         uint64
		SmsDt         int64
		Sender        string
		Destination   string
		SendingStatus SendingStatus
		DlrStatus     DLRStatus
		PreStop       string
		NextStop      string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "ok",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusDelivered,
				PreStop:       "",
				NextStop:      "",
			},
			want: true,
		}, {
			name: "f1",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusUnknown,
				PreStop:       "",
				NextStop:      "",
			},
			want: true,
		}, {
			name: "f2",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusUndelivered,
				PreStop:       "",
				NextStop:      "",
			},
			want: false,
		}, {
			name: "f4",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusSent,
				DlrStatus:     DLRStatusEmpty,
				PreStop:       "",
				NextStop:      "",
			},
			want: true,
		}, {
			name: "f5",
			fields: fields{
				Index:         0,
				SmsDt:         0,
				Sender:        "",
				Destination:   "",
				SendingStatus: SendingStatusError,
				DlrStatus:     DLRStatusEmpty,
				PreStop:       "",
				NextStop:      "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRParam{
				Index:         tt.fields.Index,
				SmsDt:         tt.fields.SmsDt,
				Sender:        tt.fields.Sender,
				Destination:   tt.fields.Destination,
				SendingStatus: tt.fields.SendingStatus,
				DlrStatus:     tt.fields.DlrStatus,
				PreStop:       tt.fields.PreStop,
				NextStop:      tt.fields.NextStop,
			}
			if got := z.Status(); got != tt.want {
				t.Errorf("Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStopParam_ToABI(t *testing.T) {
	param := StopParam{
		StopName: "test",
	}
	abi, err := param.ToABI(MethodNameAddNextStop)
	if err != nil {
		t.Fatal(err)
	}

	p2 := &StopParam{}
	if err = p2.FromABI(MethodNameAddNextStop, abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&param, p2) {
			t.Fatalf("invalid param, %v, %v", &param, p2)
		} else {
			if err := p2.Verify(); err != nil {
				t.Fatalf("verify failed, %s", err)
			}
		}
	}
}

func TestUpdateStopParam_ToABI(t *testing.T) {
	param := UpdateStopParam{
		StopName: "test",
		New:      "hahah",
	}
	abi, err := param.ToABI(MethodNameUpdateNextStop)
	if err != nil {
		t.Fatal(err)
	}

	p2 := &UpdateStopParam{}
	if err = p2.FromABI(MethodNameUpdateNextStop, abi); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&param, p2) {
			t.Fatalf("invalid param, %v, %v", &param, p2)
		} else {
			if err := p2.Verify(); err != nil {
				t.Fatalf("verify failed, %s", err)
			}
		}
	}
}

func TestTerminateParam_ToABI(t *testing.T) {
	param := &TerminateParam{ContractAddress: mock.Address()}
	if abi, err := param.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		p2 := &TerminateParam{}
		if err := p2.FromABI(abi); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(param, p2) {
				t.Fatalf("invalid param, %v, %v", param, p2)
			} else {
				t.Log(param.String())
			}
		}
	}
}

func TestContractParam_IsContractor(t *testing.T) {
	cp := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
		Terminator          *types.Address
	}
	type args struct {
		addr types.Address
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "partyA",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				addr: cp.PartyA.Address,
			},
			want: true,
		}, {
			name: "partyB",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				addr: cp.PartyB.Address,
			},
			want: true,
		}, {
			name: "failed",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              cp.Status,
			},
			args: args{
				addr: mock.Address(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
				Terminator:          tt.fields.Terminator,
			}
			if got := z.IsContractor(tt.args.addr); got != tt.want {
				t.Errorf("IsContractor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContractParam_DoActive(t *testing.T) {
	cp := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
		Terminator          *types.Address
	}
	type args struct {
		operator types.Address
		status   ContractStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActiveStage1,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusActivated,
			},
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActiveStage1,
			},
			args: args{
				operator: cp.PartyA.Address,
				status:   ContractStatusActiveStage1,
			},
			wantErr: true,
		},
		{
			name: "f2",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActivated,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusActiveStage1,
			},
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyed,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusActiveStage1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
				Terminator:          tt.fields.Terminator,
			}
			if err := z.DoActive(tt.args.operator); (err != nil) != tt.wantErr {
				t.Errorf("DoActive() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if err == nil && tt.args.status != z.Status {
					t.Errorf("DoActive() status = %s, want = %s", z.Status.String(), tt.args.status.String())
				}
			}
		})
	}
}

func TestContractParam_DoTerminate(t *testing.T) {
	cp := buildContractParam()

	type fields struct {
		CreateContractParam CreateContractParam
		PreStops            []string
		NextStops           []string
		ConfirmDate         int64
		Status              ContractStatus
		Terminator          *types.Address
	}
	type args struct {
		operator types.Address
		status   ContractStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "partyA_destroy_active_stage1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActiveStage1,
			},
			args: args{
				operator: cp.PartyA.Address,
				status:   ContractStatusDestroyed,
			},
			wantErr: false,
		}, {
			name: "partyB_destroy_active_stage1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActiveStage1,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusRejected,
			},
			wantErr: false,
		}, {
			name: "partyA_destroy_stage1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusActivated,
			},
			args: args{
				operator: cp.PartyA.Address,
				status:   ContractStatusDestroyStage1,
			},
			wantErr: false,
		}, {
			name: "partyB_destroy",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyStage1,
				Terminator:          &cp.PartyA.Address,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusDestroyed,
			},
			wantErr: false,
		},
		{
			name: "f1",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyStage1,
				Terminator:          &cp.PartyA.Address,
			},
			args: args{
				operator: cp.PartyA.Address,
				status:   ContractStatusDestroyed,
			},
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyStage1,
				Terminator:          &cp.PartyA.Address,
			},
			args: args{
				operator: mock.Address(),
				status:   ContractStatusDestroyed,
			},
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				CreateContractParam: cp.CreateContractParam,
				PreStops:            cp.PreStops,
				NextStops:           cp.NextStops,
				ConfirmDate:         cp.ConfirmDate,
				Status:              ContractStatusDestroyed,
				Terminator:          &cp.PartyA.Address,
			},
			args: args{
				operator: cp.PartyB.Address,
				status:   ContractStatusDestroyed,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractParam{
				CreateContractParam: tt.fields.CreateContractParam,
				PreStops:            tt.fields.PreStops,
				NextStops:           tt.fields.NextStops,
				ConfirmDate:         tt.fields.ConfirmDate,
				Status:              tt.fields.Status,
				Terminator:          tt.fields.Terminator,
			}
			if err := z.DoTerminate(tt.args.operator); (err != nil) != tt.wantErr {
				t.Errorf("DoTerminate() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if err == nil && tt.args.status != z.Status {
					t.Errorf("DoTerminate() status = %s, want = %s", z.Status.String(), tt.args.status.String())
				}
			}
		})
	}
}

func buildCDRStatus() *CDRStatus {
	cdr1 := cdrParam
	i, _ := random.Intn(10000)
	cdr1.Index = uint64(i)

	status := &CDRStatus{
		Params: map[string][]CDRParam{
			mock.Address().String(): {cdr1},
		},
		Status: SettlementStatusSuccess,
	}

	return status
}

func TestGetAllCDRStatus(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	contractAddr := mock.Address()

	var data []*CDRStatus
	for i := 0; i < 4; i++ {
		s := buildCDRStatus()
		if h, err := s.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}
			if abi, err := s.ToABI(); err != nil {
				t.Fatal(err)
			} else {
				if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
					t.Fatal(err)
				} else {
					data = append(data, s)
				}
			}
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
		size int
	}
	tests := []struct {
		name    string
		args    args
		want    []*CDRStatus
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &contractAddr,
				size: len(data),
			},
			want:    data,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllCDRStatus(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllCDRStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(data) {
				t.Errorf("GetAllCDRStatus() got = %d, want %d", len(got), len(data))
			}
			//for i, s := range tt.want {
			//	for k, v := range s.Params {
			//		g := got[i].Params[k]
			//		if !reflect.DeepEqual(g, v) {
			//			t.Errorf("GetAllCDRStatus() got = %v, want %v", g, s)
			//		}
			//	}
			//}
		})
	}
}

func TestGetCDRStatus(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	contractAddr := mock.Address()
	contractAddr2 := mock.Address()

	s := buildCDRStatus()
	h, err := s.ToHash()
	if err != nil {
		t.Fatal(err)
	} else {
		if abi, err := s.ToABI(); err != nil {
			t.Fatal(err)
		} else {
			if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  *vmstore.VMContext
		addr *types.Address
		hash types.Hash
	}
	tests := []struct {
		name    string
		args    args
		want    *CDRStatus
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  ctx,
				addr: &contractAddr,
				hash: h,
			},
			want:    s,
			wantErr: false,
		}, {
			name: "fail",
			args: args{
				ctx:  ctx,
				addr: &contractAddr,
				hash: mock.Hash(),
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "f2",
			args: args{
				ctx:  ctx,
				addr: &contractAddr2,
				hash: h,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCDRStatus(tt.args.ctx, tt.args.addr, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCDRStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCDRStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCDRStatus_IsInCycle(t *testing.T) {
	param := cdrParam
	param.SmsDt = 1582001974

	type fields struct {
		Params map[string][]CDRParam
		Status SettlementStatus
	}
	type args struct {
		start int64
		end   int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "ok1",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {cdrParam},
				},
				Status: 0,
			},
			args: args{
				start: 0,
				end:   100,
			},
			want: true,
		}, {
			name: "ok2",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {param},
				},
				Status: 0,
			},
			args: args{
				start: 1581829174,
				end:   1582433974,
			},
			want: true,
		}, {
			name: "ok3",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {param, param},
				},
				Status: 0,
			},
			args: args{
				start: 1581829174,
				end:   1582433974,
			},
			want: true,
		}, {
			name: "f1",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {param},
				},
				Status: 0,
			},
			args: args{
				start: 1582433974,
				end:   1582434974,
			},
			want: false,
		}, {
			name: "f2",
			fields: fields{
				Params: map[string][]CDRParam{
					mock.Address().String(): {cdrParam, cdrParam},
				},
				Status: 0,
			},
			args: args{
				start: 1582433974,
				end:   1582434974,
			},
			want: false,
		}, {
			name: "f3",
			fields: fields{
				Params: nil,
				Status: 0,
			},
			args: args{
				start: 1582433974,
				end:   1582434974,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			if got := z.IsInCycle(tt.args.start, tt.args.end); got != tt.want {
				t.Errorf("IsInCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerminateParam_Verify(t *testing.T) {
	type fields struct {
		ContractAddress types.Address
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				ContractAddress: mock.Address(),
			},
			wantErr: false,
		}, {
			name: "fail",
			fields: fields{
				ContractAddress: types.ZeroAddress,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &TerminateParam{
				ContractAddress: tt.fields.ContractAddress,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSummaryRecord_DoCalculate(t *testing.T) {
	r := &SummaryRecord{
		Total:   0,
		Success: 10,
		Fail:    2,
		Result:  0,
	}
	r.DoCalculate()
	if r.Total != 12 {
		t.Fail()
	}
	t.Log(r.String())
}

func TestNewSummaryResult(t *testing.T) {
	r := NewSummaryResult()

	for i := 0; i < 20; i++ {
		r.UpdatePartyState("WeChat", true, i%3 == 0)
		r.UpdatePartyState("WeChat", false, i%2 == 0)
		r.UpdatePartyState("Slack", true, i%2 == 0)
		r.UpdatePartyState("Slack", false, i%3 == 0)
	}

	for i := 0; i < 5; i++ {
		r.UpdateGlobalState("WeChat", true, i%2 == 0)
		r.UpdateGlobalState("Slack", false, i%2 == 0)
	}

	r.DoCalculate()
	t.Log(r.String())
}

func TestCDRStatus_State(t *testing.T) {
	addr1 := mock.Address()
	addr2 := mock.Address()

	type fields struct {
		Params map[string][]CDRParam
		Status SettlementStatus
	}
	type args struct {
		addr *types.Address
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
	}{
		{
			name: "ok",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdrParam}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr1,
			},
			want:  "PCCWG",
			want1: true,
		}, {
			name: "f1",
			fields: fields{
				Params: nil,
				Status: 0,
			},
			args: args{
				addr: &addr1,
			},
			want:  "",
			want1: false,
		}, {
			name: "f2",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdrParam}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr2,
			},
			want:  "",
			want1: false,
		}, {
			name: "f3",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdrParam, cdrParam}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr1,
			},
			want:  "PCCWG",
			want1: false,
		}, {
			name: "f4",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {}},
				Status: SettlementStatusSuccess,
			},
			args: args{
				addr: &addr1,
			},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			got, got1 := z.State(tt.args.addr)
			if got != tt.want {
				t.Errorf("State() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("State() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestCDRStatus_ExtractID(t *testing.T) {
	addr1 := mock.Address()
	//addr2 := mock.Address()

	type fields struct {
		Params map[string][]CDRParam
		Status SettlementStatus
	}
	tests := []struct {
		name            string
		fields          fields
		wantDt          int64
		wantSender      string
		wantDestination string
		wantErr         bool
	}{
		{
			name: "ok",
			fields: fields{
				Params: map[string][]CDRParam{addr1.String(): {cdrParam}},
				Status: SettlementStatusSuccess,
			},
			wantDt:          cdrParam.SmsDt,
			wantSender:      cdrParam.Sender,
			wantDestination: cdrParam.Destination,
			wantErr:         false,
		}, {
			name: "fail",
			fields: fields{
				Params: nil,
				Status: SettlementStatusSuccess,
			},
			wantDt:          0,
			wantSender:      "",
			wantDestination: "",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &CDRStatus{
				Params: tt.fields.Params,
				Status: tt.fields.Status,
			}
			gotDt, gotSender, gotDestination, err := z.ExtractID()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDt != tt.wantDt {
				t.Errorf("ExtractID() gotDt = %v, want %v", gotDt, tt.wantDt)
			}
			if gotSender != tt.wantSender {
				t.Errorf("ExtractID() gotSender = %v, want %v", gotSender, tt.wantSender)
			}
			if gotDestination != tt.wantDestination {
				t.Errorf("ExtractID() gotDestination = %v, want %v", gotDestination, tt.wantDestination)
			}
		})
	}
}

type mockCDR struct {
	Params []SettlementCDR  `json:"params"`
	Status SettlementStatus `json:"status"`
}

func TestGetSummaryReport(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	// mock settlement contract
	ac1 := mock.Account()
	ac2 := mock.Account()
	a1 := ac1.Address()
	a2 := ac2.Address()

	param := buildContractParam()
	param.PartyA.Address = a1
	param.PartyB.Address = a2
	param.NextStops = []string{"CSL Hong Kong @ 3397"}
	param.PreStops = []string{"A2P_PCCWG"}

	contractAddr, _ := param.Address()
	abi, _ := param.ToABI()
	if err := ctx.SetStorage(types.SettlementAddress[:], contractAddr[:], abi[:]); err != nil {
		t.Fatal(err)
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	var cdr []*mockCDR
	if err := json.Unmarshal([]byte(cdrs), &cdr); err != nil {
		t.Fatal(err)
	} else {
		for _, c := range cdr {
			params := make(map[string][]CDRParam, 0)
			param1 := c.Params[0].CDRParam
			param1.ContractAddress = contractAddr
			params[a1.String()] = []CDRParam{param1}
			if len(c.Params) > 1 {
				param2 := c.Params[1].CDRParam
				param2.ContractAddress = contractAddr
				params[a2.String()] = []CDRParam{param2}
			}

			s := &CDRStatus{
				Params: params,
				Status: c.Status,
			}
			//t.Log(s.String())
			if h, err := s.ToHash(); err != nil {
				t.Fatal(err)
			} else {
				if h.IsZero() {
					t.Fatal("invalid hash")
				}
				if abi, err := s.ToABI(); err != nil {
					t.Fatal(err)
				} else {
					if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
		if err := ctx.SaveStorage(); err != nil {
			t.Fatal(err)
		}
	}

	if report, err := GetSummaryReport(ctx, &contractAddr, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		t.Log(report)
	}

	if invoices, err := GenerateInvoices(ctx, &a1, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if len(invoices) == 0 {
			t.Fatal("invalid invoice")
		}
		for _, i := range invoices {
			t.Log(util.ToIndentString(i))
		}
	}

	if invoices, err := GenerateInvoicesByContract(ctx, &contractAddr, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if len(invoices) == 0 {
			t.Fatal("invalid invoice")
		}
		t.Log(util.ToIndentString(invoices))
	}
}

func TestGetStopNames(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	// mock settlement contract
	ac1 := mock.Account()
	ac2 := mock.Account()
	a1 := ac1.Address()
	a2 := ac2.Address()

	var contracts []*ContractParam

	param := buildContractParam()
	param.PartyA.Address = a1
	param.PartyB.Address = a2
	param.NextStops = []string{"CSL Hong Kong @ 3397"}
	param.PreStops = []string{"A2P_PCCWG"}
	contracts = append(contracts, param)

	param2 := buildContractParam()
	param2.PartyA.Address = a2
	param2.PartyB.Address = a1
	param2.PreStops = []string{"CSL Hong Kong @ 33971"}
	param2.NextStops = []string{"A2P_PCCWG2"}

	contracts = append(contracts, param2)
	for _, c := range contracts {
		contractAddr, _ := c.Address()
		abi, _ := c.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], contractAddr[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}

	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	if names, err := GetPreStopNames(ctx, &a1); err != nil {
		t.Fatal(err)
	} else {
		if len(names) != 1 {
			t.Fatalf("invalid len %d", len(names))
		}

		if names[0] != "CSL Hong Kong @ 33971" {
			t.Fatal(names)
		}
	}

	if names, err := GetNextStopNames(ctx, &a1); err != nil {
		t.Fatal(err)
	} else {
		if len(names) != 1 {
			t.Fatalf("invalid len %d", len(names))
		}

		if names[0] != "CSL Hong Kong @ 3397" {
			t.Fatal(names)
		}
	}
}

func TestContractAddressList_Append(t *testing.T) {
	a1 := mock.Address()
	a2 := mock.Address()
	cl := newContractAddressList(&a1)

	type fields struct {
		AddressList []*types.Address
	}
	type args struct {
		address *types.Address
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "ok",
			fields: fields{
				AddressList: cl.AddressList,
			},
			args: args{
				address: &a2,
			},
			want: true,
		}, {
			name: "exist",
			fields: fields{
				AddressList: cl.AddressList,
			},
			args: args{
				address: &a1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &ContractAddressList{
				AddressList: tt.fields.AddressList,
			}
			if got := z.Append(tt.args.address); got != tt.want {
				t.Errorf("Append() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContractAddressList_ToABI(t *testing.T) {
	a1 := mock.Address()
	cl := newContractAddressList(&a1)

	if abi, err := cl.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		cl2 := &ContractAddressList{}
		if err := cl2.FromABI(abi); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(cl, cl2) {
				t.Fatalf("invalid %v,%v", cl, cl2)
			} else {
				t.Log(cl.String())
			}
		}
	}
}

func TestSaveCDRStatus(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	a1 := mock.Address()
	cdr := buildCDRStatus()

	h, err := cdr.ToHash()
	if err != nil {
		t.Fatal(err)
	}
	if err = SaveCDRStatus(ctx, &a1, &h, cdr); err != nil {
		t.Fatal(err)
	}

	if s, err := GetCDRStatus(ctx, &a1, h); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(cdr, s) {
			t.Fatalf("invalid cdr, act: %v, exp: %v", s, cdr)
		} else {
			if addresses, err := GetCDRMapping(ctx, &h); err != nil {
				t.Fatal(err)
			} else {
				if len(addresses) != 1 {
					t.Fatalf("invalid address len: %d", len(addresses))
				}

				if !reflect.DeepEqual(addresses[0], &a1) {
					t.Fatalf("invalid address, act: %s, exp: %s", addresses[0].String(), a1.String())
				}
			}
		}
	}
}

func TestGetCDRMapping(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	a1 := mock.Address()
	a2 := mock.Address()
	h := mock.Hash()

	exp := []*types.Address{&a1, &a2}

	if err := saveCDRMapping(ctx, &a1, &h); err != nil {
		t.Fatal(err)
	}
	if err := saveCDRMapping(ctx, &a2, &h); err != nil {
		t.Fatal(err)
	}

	if addressList, err := GetCDRMapping(ctx, &h); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(addressList, exp) {
			t.Fatalf("invalid address, act: %v, exp:%v", addressList, exp)
		}
	}
}

func TestGenerateMultiPartyInvoice(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	a1 := mock.Address()
	a2 := mock.Address()
	a3 := mock.Address()

	//prepare two contracts
	var contracts []*ContractParam

	// Montnets-PCCWG
	param1 := buildContractParam()
	param1.PartyA.Address = a1
	param1.PartyB.Address = a2
	param1.NextStops = []string{"A2P_PCCWG"}
	param1.PreStops = []string{"MONTNETS"}
	contracts = append(contracts, param1)

	// PCCWG-CSL
	param2 := buildContractParam()
	param2.PartyA.Address = a2
	param2.PartyB.Address = a3
	param2.NextStops = []string{"CSL Hong Kong @ 3397"}
	param2.PreStops = []string{"A2P_PCCWG"}

	contracts = append(contracts, param2)
	for _, c := range contracts {
		contractAddr, _ := c.Address()
		abi, _ := c.ToABI()
		if err := ctx.SetStorage(types.SettlementAddress[:], contractAddr[:], abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	ca1, _ := contracts[0].Address()
	ca2, _ := contracts[1].Address()

	// upload CDR
	template := cdrParam
	template.SmsDt = time.Now().Unix()
	template.Sender = "WeChat"

	p1 := template
	p1.NextStop = "A2P_PCCWG"

	cdr1 := &CDRStatus{
		Params: map[string][]CDRParam{
			a1.String(): {p1},
		},
		Status: SettlementStatusSuccess,
	}

	if h, err := cdr1.ToHash(); err != nil {
		t.Fatal(err)
	} else {
		if h.IsZero() {
			t.Fatal("invalid hash")
		}

		t.Log("p1", h.String())
		if err := SaveCDRStatus(ctx, &ca1, &h, cdr1); err != nil {
			t.Fatal(err)
		}
	}

	p2 := template
	p2.PreStop = "MONTNETS"

	cdr2 := &CDRStatus{
		Params: map[string][]CDRParam{
			a2.String(): {p2},
		},
		Status: SettlementStatusSuccess,
	}

	if h, err := cdr2.ToHash(); err != nil {
		t.Fatal(err)
	} else {
		if h.IsZero() {
			t.Fatal("invalid hash")
		}
		t.Log("p2", h.String())
		if err := SaveCDRStatus(ctx, &ca1, &h, cdr2); err != nil {
			t.Fatal(err)
		}
	}

	// upload CDR to PCCWG-CSL
	p3 := template
	p3.NextStop = "CSL Hong Kong @ 3397"

	cdr3 := &CDRStatus{
		Params: map[string][]CDRParam{
			a2.String(): {p3},
		},
		Status: SettlementStatusSuccess,
	}

	if h, err := cdr3.ToHash(); err != nil {
		t.Fatal(err)
	} else {
		if h.IsZero() {
			t.Fatal("invalid hash")
		}

		t.Log("p3", h.String())
		if err := SaveCDRStatus(ctx, &ca2, &h, cdr3); err != nil {
			t.Fatal(err)
		}
	}

	p4 := template
	p4.PreStop = "A2P_PCCWG"

	cdr4 := &CDRStatus{
		Params: map[string][]CDRParam{
			a2.String(): {p4},
		},
		Status: SettlementStatusSuccess,
	}

	if h, err := cdr4.ToHash(); err != nil {
		t.Fatal(err)
	} else {
		if h.IsZero() {
			t.Fatal("invalid hash")
		}

		t.Log("p4", h.String())
		if err := SaveCDRStatus(ctx, &ca1, &h, cdr4); err != nil {
			t.Fatal(err)
		}
	}

	// save to db
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	// generate invoice
	if invoice, err := GenerateMultiPartyInvoice(ctx, &ca1, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if len(invoice) == 0 {
			t.Fatal("invalid invoice")
		}
		t.Log(util.ToIndentString(invoice))
	}
}

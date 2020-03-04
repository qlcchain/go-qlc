/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"reflect"
	"testing"
)

func TestPeerInfo_Serialize(t *testing.T) {
	info := &PeerInfo{
		PeerID:         "ddafdadafdasfd",
		Address:        "qlc_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3",
		Version:        "1",
		Rtt:            1,
		LastUpdateTime: "",
	}

	if data, err := info.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		info2 := &PeerInfo{}
		if err := info2.Deserialize(data); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(info, info2) {
				t.Fatalf("exp: %v, act: %v", info, info2)
			}
		}
	}
}

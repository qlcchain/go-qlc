/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"github.com/google/uuid"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"
)

func TestRegisterServices(t *testing.T) {
	cfgFile2 := filepath.Join(config.QlcTestDataDir(), "config1", uuid.New().String(), config.QlcConfigFile)
	cm := config.NewCfgManagerWithName(filepath.Dir(cfgFile2), filepath.Base(cfgFile2))
	cc := context.NewChainContext(cm.ConfigFile)

	if err := RegisterServices(cc); err != nil {
		t.Fatal(err)
	}

	if services, err := cc.AllServices(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(len(services))
	}
}

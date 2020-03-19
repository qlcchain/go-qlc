/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"
)

func TestRegisterServices(t *testing.T) {
	cfgFile := filepath.Join(config.QlcTestDataDir(), "config", uuid.New().String(), config.QlcConfigFile)
	cm := config.NewCfgManagerWithName(filepath.Dir(cfgFile), filepath.Base(cfgFile))
	cc := context.NewChainContext(cm.ConfigFile)

	defer func() {
		os.RemoveAll(cfgFile)
	}()

	if err := RegisterServices(cc); err != nil {
		t.Fatal(err)
	}

	if services, err := cc.AllServices(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(len(services))
	}
}

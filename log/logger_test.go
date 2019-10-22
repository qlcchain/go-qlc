/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package log

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/config"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger("test1")
	log.Debug("debug1")
	log.Warn("warning message")

	rawJSON := []byte(`{
		"level": "info",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"encoding": "json",
		"encoderConfig": {
			"messageKey": "message",
			"levelKey": "level",
			"levelEncoder": "lowercase"
		}
	}`)
	var config zap.Config
	if err := json.Unmarshal(rawJSON, &config); err != nil {
		t.Fatal(err)
	}
	//config.DisableStacktrace = false
	config.EncoderConfig = zap.NewProductionEncoderConfig()
	//t.Log(config)
	logger, _ := config.Build()
	logger.Sugar().Named("rrrrr").Warn("xxxxx")
}

func TestInit(t *testing.T) {
	cfg, err := config.DefaultConfig(config.DefaultDataDir())
	if err != nil {
		t.Fatal(err)
	}

	err = InitLog(cfg)
	if err != nil {
		t.Fatal(err)
	}

	logger := NewLogger("test2")
	logger.Warn("xxxxxxxxxxxxxxxxxxxxxx")
}

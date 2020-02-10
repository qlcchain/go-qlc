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
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		t.Fatal(err)
	}
	//cfg.DisableStacktrace = false
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	//t.Log(cfg)
	logger, _ := cfg.Build()
	logger.Sugar().Named("rrrrr").Warn("xxxxx")
}

func TestInit(t *testing.T) {
	cfg, err := config.DefaultConfig(config.DefaultDataDir())
	if err != nil {
		t.Fatal(err)
	}

	err = Setup(cfg)
	if err != nil {
		t.Fatal(err)
	}

	logger := NewLogger("test2")
	logger.Warn("xxxxxxxxxxxxxxxxxxxxxx")
}

//
//func TestDynamicLevel(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	var i int
//	go func(ctx context.Context, i int) {
//		select {
//		case <-ctx.Done():
//			return
//		default:
//			i++
//			Root.VInfo(i)
//		}
//	}(ctx, i)
//	time.Sleep(3 * time.Second)
//	Root.Desugar()
//	time.Sleep(3 * time.Second)
//	cancel()
//}

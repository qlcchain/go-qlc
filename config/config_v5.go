/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import "github.com/qlcchain/go-qlc/common/util"

const tokenLength = 32

type ConfigV5 struct {
	ConfigV4 `mapstructure:",squash"`
	Metrics  *MetricsConfig `json:"metrics"`
	Manager  *Manager       `json:"manager"`
}

type MetricsConfig struct {
	Enable         bool    `json:"enable"`
	SampleInterval int     `json:"sampleInterval" validate:"min=1"`
	Influx         *Influx `json:"influx"`
}

type Influx struct {
	Enable   bool   `json:"enable"`
	URL      string `json:"url" validate:"nonzero"`
	Database string `json:"database" validate:"nonzero"`
	User     string `json:"user" validate:"nonzero"`
	Password string `json:"password"`
	Interval int    `json:"interval" validate:"min=1"`
}

type Manager struct {
	AdminToken string `json:"adminToken"`
}

func DefaultConfigV5(dir string) (*ConfigV5, error) {
	var cfg ConfigV5
	cfg4, _ := DefaultConfigV4(dir)
	cfg.ConfigV4 = *cfg4
	cfg.Version = configVersion
	cfg.PoV.PovEnabled = true
	cfg.Metrics = defaultMetrics()
	cfg.Manager = &Manager{AdminToken: util.RandomFixedString(tokenLength)}

	return &cfg, nil
}

func defaultMetrics() *MetricsConfig {
	return &MetricsConfig{
		Enable:         false,
		SampleInterval: 60,
		Influx: &Influx{
			Enable:   false,
			URL:      "http://localhost:8086",
			Database: "qlcchain",
			User:     "qlcchain",
			Password: "",
			Interval: 10,
		},
	}
}

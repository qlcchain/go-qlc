/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"context"
	"errors"
	"time"

	"github.com/qlcchain/go-qlc/monitor/influxdb"

	ctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/monitor"
)

func NewMetricsService(cfgFile string) *MetricsService {
	cc := ctx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	ctx2, cancel := context.WithCancel(context.Background())
	return &MetricsService{
		cfg:    cfg,
		ctx:    ctx2,
		cancel: cancel,
	}
}

type MetricsService struct {
	common.ServiceLifecycle
	cfg    *config.Config
	ctx    context.Context
	cancel context.CancelFunc
}

func (m *MetricsService) Init() error {
	if !m.PreInit() {
		return errors.New("pre init fail")
	}
	defer m.PostInit()

	return nil
}

func (m *MetricsService) Start() error {
	if !m.PreStart() {
		return errors.New("pre start fail")
	}
	defer m.PostStart()

	if m.cfg.Metrics.Enable {
		d := time.Second * time.Duration(m.cfg.Metrics.SampleInterval)
		go monitor.CaptureRuntimeCPUStats(m.ctx, d)
		go monitor.CaptureRuntimeDiskStats(m.ctx, d)
		go monitor.CaptureRuntimeNetStats(m.ctx, d)
	}

	influx := m.cfg.Metrics.Influx
	if influx != nil && influx.Enable && len(influx.URL) > 0 && len(influx.Database) > 0 {
		go influxdb.InfluxDB(m.ctx,
			monitor.SystemRegistry,                     // metrics registry
			time.Second*time.Duration(influx.Interval), // interval
			influx.URL,      // the InfluxDB url
			influx.Database, // your InfluxDB database
			influx.User,     // your InfluxDB user
			influx.Password, // your InfluxDB password
		)
	}

	return nil
}

func (m *MetricsService) Stop() error {
	if !m.PreStop() {
		return errors.New("pre stop fail")
	}
	defer m.PostStop()

	m.cancel()

	return nil
}

func (m *MetricsService) Status() int32 {
	return m.State()
}

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

type HttpService struct {
	common.ServiceLifecycle
	logger *zap.SugaredLogger
	cfg    *config.Config
}

func NewHttpService(cfgFile string) *HttpService {
	cc := context.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	return &HttpService{cfg: cfg, logger: log.NewLogger("http_service")}
}

func (hs *HttpService) Init() error {
	if !hs.PreInit() {
		return errors.New("pre init fail")
	}
	defer hs.PostInit()
	return nil
}

func (hs *HttpService) Start() error {
	if !hs.PreStart() {
		return errors.New("pre start fail")
	}
	defer hs.PostStart()

	http.HandleFunc("/bootNode", func(w http.ResponseWriter, r *http.Request) {
		bootNode := hs.cfg.P2P.Listen + "/p2p/" + hs.cfg.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe(hs.cfg.P2P.BootNodeHttpServer, nil); err != nil {
			hs.logger.Error(err)
		}
	}()
	return nil
}

func (hs *HttpService) Stop() error {
	if !hs.PreStop() {
		return errors.New("pre stop fail")
	}
	defer hs.PostStop()
	return nil
}

func (hs *HttpService) Status() int32 {
	return hs.State()
}

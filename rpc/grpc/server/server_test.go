package grpcServer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGRPCServer(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "grpc", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	cfgFile := cm.ConfigFile
	cc := chainctx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	d, _ := json.Marshal(cfg.RPC)
	fmt.Println(string(d))

	server, err := Start(cfgFile, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	server.Stop()
}

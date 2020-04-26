package chain

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	rpc "github.com/qlcchain/jsonrpc2"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type PermissionService struct {
	common.ServiceLifecycle
	logger     *zap.SugaredLogger
	cfgFile    string
	cfg        *config.Config
	cc         *context.ChainContext
	subscriber *event.ActorSubscriber
	work       chan struct{}
	exit       chan struct{}
	account    *types.Account
	client     *rpc.Client
	vmCtx      *vmstore.VMContext
	checkTime  uint8
	nodes      []*api.NodeParam
	blkHash    []types.Hash
}

func NewPermissionService(cfgFile string) *PermissionService {
	cm := config.NewCfgManagerWithFile(cfgFile)
	cfg, _ := cm.Config()
	l := ledger.NewLedger(cfgFile)

	ps := &PermissionService{
		cfgFile: cfgFile,
		logger:  log.NewLogger("permission service"),
		cfg:     cfg,
		cc:      context.NewChainContext(cfgFile),
		work:    make(chan struct{}, 1),
		exit:    make(chan struct{}, 1),
		vmCtx:   vmstore.NewVMContext(l),
		nodes:   make([]*api.NodeParam, 0),
	}

	return ps
}

func (ps *PermissionService) Init() error {
	ps.PreInit()

	admin, err := abi.PermissionGetAdmin(ps.vmCtx)
	if err != nil {
		return err
	}

	for _, acc := range ps.cc.Accounts() {
		for _, adm := range admin {
			if acc.Address() == adm.Account {
				ps.account = acc
				break
			}
		}
	}

	if ps.account == nil {
		return nil
	}

	// check if the nodes already exist in ledger
	for _, n := range ps.cfg.WhiteList.WhiteListInfos {
		ni, _ := abi.PermissionGetNode(ps.vmCtx, n.PeerId)
		if ni == nil {
			node := &api.NodeParam{
				Admin:   ps.account.Address(),
				NodeId:  n.PeerId,
				NodeUrl: n.Addr,
				Comment: n.Comment,
			}
			ps.nodes = append(ps.nodes, node)
		}
	}

	ps.subscriber = event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case topic.SyncState:
			ps.onPovSyncState(msg)
		}
	}), ps.cc.EventBus())

	if err := ps.subscriber.Subscribe(topic.EventPovSyncState); err != nil {
		return err
	}

	ps.PostInit()
	return nil
}

func (ps *PermissionService) Start() error {
	ps.PreStart()

	if len(ps.nodes) > 0 {
		rs, err := ps.cc.Service(context.RPCService)
		if err != nil {
			ps.logger.Error(err)
			return err
		}

		ps.client, err = rs.(*RPCService).RPC().Attach()
		if err != nil {
			ps.logger.Error(err)
			return err
		}

		go ps.addPermissionNode()
	}

	ps.PostStart()
	return nil
}

func (ps *PermissionService) Stop() error {
	ps.PreStop()

	if ps.subscriber != nil {
		if err := ps.subscriber.UnsubscribeAll(); err != nil {
			return err
		}
	}

	ps.exit <- struct{}{}

	ps.PostStop()
	return nil
}

func (ps *PermissionService) Status() int32 {
	return ps.State()
}

func (ps *PermissionService) onPovSyncState(state topic.SyncState) {
	if state == topic.SyncDone {
		ps.work <- struct{}{}
	}
}

func (ps *PermissionService) addPermissionNode() {
	checkInterval := time.NewTicker(5 * time.Second)
	syncDone := false

	for {
		select {
		case <-ps.exit:
			ps.logger.Info("permission service exit")
			return
		case <-checkInterval.C:
			// if can't get the node after 10 minutes, do it again
			if !syncDone {
				continue
			}

			ps.checkTime++

			nodes := make([]*api.NodeParam, 0)
			for i, h := range ps.blkHash {
				if has, _ := ps.vmCtx.Ledger.HasStateBlockConfirmed(h); !has {
					nodes = append(nodes, ps.nodes[i])
				}

				if len(nodes) > 0 {
					if ps.checkTime >= 120 {
						ps.nodes = nodes
						ps.work <- struct{}{}
						ps.checkTime = 0
					}
				} else {
					ps.logger.Info("added all nodes, permission service exit")
					return
				}
			}
		case <-ps.work:
			syncDone = true

			if !ps.cc.IsPoVDone() {
				ps.work <- struct{}{}
				time.Sleep(time.Second)
				continue
			}

			for _, param := range ps.nodes {
				blk := new(types.StateBlock)
				err := ps.client.Call(&blk, "permission_getNodeUpdateBlock", &param)
				if err != nil {
					ps.logger.Error(err)
					continue
				}

				var w types.Work
				worker, _ := types.NewWorker(w, blk.Root())
				blk.Work = worker.NewWork()

				hash := blk.GetHash()
				blk.Signature = ps.account.Sign(hash)
				ps.blkHash = append(ps.blkHash, hash)

				var h types.Hash
				err = ps.client.Call(&h, "ledger_process", &blk)
				if err != nil {
					ps.logger.Error(err)
				}
			}
		}
	}
}

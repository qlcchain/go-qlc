package consensus

import (
	"errors"
	"github.com/json-iterator/go"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"

	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/wallet"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p"
)

var logger = log.NewLogger("consensus")

type DposService struct {
	common.ServiceLifecycle
	ns       p2p.Service
	ledger   *ledger.Ledger
	eventMsg map[p2p.EventType]p2p.EventSubscriber
	quitCh   chan bool
	bp       *BlockProcessor
	wallet   *wallet.WalletStore
	actrx    *ActiveTrx
	account  types.Address
	password string
}

func (dps *DposService) Init() error {
	if !dps.PreInit() {
		return errors.New("pre init fail")
	}
	defer dps.PostInit()

	return nil
}

func (dps *DposService) Start() error {
	if !dps.PreStart() {
		return errors.New("pre start fail")
	}
	defer dps.PostStart()
	dps.setEvent()
	logger.Info("start dpos service")
	go dps.bp.Start()
	go dps.actrx.start()

	return nil
}

func (dps *DposService) Stop() error {
	if !dps.PreStop() {
		return errors.New("pre stop fail")
	}
	defer dps.PostStop()
	dps.quitCh <- true
	dps.bp.quitCh <- true
	dps.actrx.quitCh <- true
	for i, j := range dps.eventMsg {
		dps.ns.MessageEvent().GetEvent("consensus").UnSubscribe(i, j)
	}

	return nil
}

func (dps *DposService) Status() int32 {
	panic("implement me")
}

func NewDposService(cfg *config.Config, netService p2p.Service, account types.Address, password string) (*DposService, error) {
	bp := NewBlockProcessor()
	actrx := NewActiveTrx()
	l := ledger.NewLedger(cfg.LedgerDir())

	dps := &DposService{
		ns:       netService,
		ledger:   l,
		eventMsg: make(map[p2p.EventType]p2p.EventSubscriber),
		quitCh:   make(chan bool, 1),
		bp:       bp,
		actrx:    actrx,
		wallet:   wallet.NewWalletStore(cfg),
		account:  account,
		password: password,
	}
	//test begin...
	//dps.accounts = append(dps.accounts, ac.Address())
	//test end...
	dps.bp.SetDpos(dps)
	dps.actrx.SetDposService(dps)
	return dps, nil
}

func (dps *DposService) SetWalletStore(wallet *wallet.WalletStore) {
	dps.wallet = wallet
}

/*func (dps *DposService) SetAccounts() error {
	accounts, err := dps.walletsession.GetAccounts()
	if err != nil {
		return err
	}
	dps.accounts = accounts
	return nil
}*/

func (dps *DposService) setEvent() {
	event1 := dps.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventPublish, dps.ReceivePublish)
	event2 := dps.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmReq, dps.ReceiveConfirmReq)
	event3 := dps.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmAck, dps.ReceiveConfirmAck)
	event4 := dps.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventSyncBlock, dps.ReceiveSyncBlock)
	dps.eventMsg[p2p.EventPublish] = event1
	dps.eventMsg[p2p.EventConfirmReq] = event2
	dps.eventMsg[p2p.EventConfirmAck] = event3
	dps.eventMsg[p2p.EventSyncBlock] = event4
}

func (dps *DposService) ReceivePublish(v interface{}) {
	logger.Info("Publish Event")
	dps.bp.blocks <- v.(types.Block)
}

func (dps *DposService) ReceiveConfirmReq(v interface{}) {
	logger.Info("ConfirmReq Event")
	dps.bp.blocks <- v.(types.Block)
	dps.onReceiveConfirmReq(v.(types.Block))
}

func (dps *DposService) onReceiveConfirmReq(block types.Block) {
	accounts := dps.getAccounts()
	for _, k := range accounts {
		isRep := dps.isThisAccountRepresentation(k)
		if isRep == true {
			logger.Infof("send confirm ack for hash %s,previous hash is %s", block.GetHash(), block.Root())
			dps.sendConfirmAck(block, k)
		} else {
			logger.Infof("send confirm req for hash %s,previous hash is %s", block.GetHash(), block.Root())
			dps.sendConfirmReq(block)
		}
	}
	//if len(accounts) == 0 {
	//	logger.Info("this is just a node,not a wallet")
	//	dps.sendConfirmReq(block)
	//}
}

func (dps *DposService) ReceiveConfirmAck(v interface{}) {
	logger.Info("ConfirmAck Event")
	vote := v.(*protos.ConfirmAckBlock)
	dps.bp.blocks <- vote.Blk
	dps.onReceiveConfirmAck(vote)
}

func (dps *DposService) onReceiveConfirmAck(va *protos.ConfirmAckBlock) {
	valid := IsAckSignValidate(va)
	if valid != true {
		return
	}
	if v, ok := dps.actrx.roots[va.Blk.Root()]; ok {
		result, vt := v.vote.voteExit(va.Account)
		if result == true {
			if vt.Sequence < va.Sequence {
				v.vote.repVotes[va.Account] = va
			}
		} else {
			ta := v.vote.voteStatus(va)
			if ta == confirm {
				currentvote := v.vote.repVotes[va.Account]
				if currentvote.Sequence < va.Sequence {
					v.vote.repVotes[va.Account] = va
				}
			}
		}
		dps.actrx.vote(va)
	} else {
		exit, err := dps.ledger.HasBlock(va.Blk.GetHash())
		if err != nil {
			return
		}
		if exit {
			accounts := dps.getAccounts()
			for _, k := range accounts {
				isRep := dps.isThisAccountRepresentation(k)
				if isRep == true {
					dps.sendConfirmAck(va.Blk, k)
				} else {
					data, err := protos.ConfirmAckBlockToProto(va)
					if err != nil {
						logger.Error("vote to proto error")
					}
					dps.ns.Broadcast(p2p.ConfirmAck, data)
				}
			}
			//if len(accounts) == 0 {
			//	logger.Info("this is just a node,not a wallet")
			//	data, err := protos.ConfirmAckBlockToProto(va)
			//	if err != nil {
			//		logger.Error("vote to proto error")
			//	}
			//	dps.ns.Broadcast(p2p.ConfirmAck, data)
			//}
		}
	}
}

func (dps *DposService) ReceiveSyncBlock(v interface{}) {
	logger.Info("Sync Event")
	dps.bp.blocks <- v.(types.Block)
}

func (dps *DposService) sendConfirmReq(block types.Block) error {
	packet := &protos.ConfirmReqBlock{
		Blk: block,
	}
	data, err := protos.ConfirmReqBlockToProto(packet)
	if err != nil {
		logger.Error("ConfirmReq Block to Proto error")
		return err
	}
	dps.ns.Broadcast(p2p.ConfirmReq, data)
	return nil
}

func (dps *DposService) sendConfirmAck(block types.Block, account types.Address) error {
	va, err := dps.voteGenerate(block, account)
	if err != nil {
		logger.Error("vote generate error")
		return err
	}
	data, err := protos.ConfirmAckBlockToProto(va)
	if err != nil {
		logger.Error("vote to proto error")
	}
	dps.ns.Broadcast(p2p.ConfirmAck, data)
	return nil
}

func (dps *DposService) voteGenerate(block types.Block, account types.Address) (*protos.ConfirmAckBlock, error) {
	var va protos.ConfirmAckBlock
	prv, err := dps.GetAccountPrv(account)
	if err != nil {
		logger.Error("Get prv error")
		return nil, err
	}
	if v, ok := dps.actrx.roots[block.Root()]; ok {
		result, vt := v.vote.voteExit(account)
		if result == true {
			va.Sequence = vt.Sequence + 1
			va.Blk = block
			va.Account = account
			va.Signature = prv.Sign(block.GetHash())
		} else {
			va.Sequence = 0
			va.Blk = block
			va.Account = account
			va.Signature = prv.Sign(block.GetHash())
		}
		v.vote.voteStatus(&va)
	} else {
		va.Sequence = 0
		va.Blk = block
		va.Account = account
		va.Signature = prv.Sign(block.GetHash())
	}
	return &va, nil
}

func (dps *DposService) GetAccountPrv(account types.Address) (*types.Account, error) {
	session := dps.wallet.NewSession(dps.account)
	return session.GetRawKey(account)
}

func (dps *DposService) isThisAccountRepresentation(address types.Address) bool {
	_, err := dps.ledger.GetRepresentation(address)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (dps *DposService) getAccounts() []types.Address {
	session := dps.wallet.NewSession(dps.account)
	if verify, err := session.VerifyPassword(dps.password); verify && err == nil {
		if a, err := session.GetAccounts(); err == nil {
			if len(a) == 0 {
				if addresses, e := dps.wallet.WalletIds(); e == nil {
					logger.Debug(jsoniter.MarshalToString(&addresses))
				}
			}

			return a
		} else {
			logger.Error(err)
		}
	} else {
		logger.Debugf("verify password[%s] faild", dps.password)
	}

	return []types.Address{}
}

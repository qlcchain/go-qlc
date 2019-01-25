package consensus

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/util"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/wallet"
)

//var logger = log.NewLogger("consensus")

const (
	findOnlineRepresentativesIntervalms = 60 * time.Second
)

type DposService struct {
	common.ServiceLifecycle
	ns                 p2p.Service
	ledger             *ledger.Ledger
	eventMsg           map[p2p.EventType]p2p.EventSubscriber
	quitCh             chan bool
	bp                 *BlockProcessor
	wallet             *wallet.WalletStore
	actrx              *ActiveTrx
	account            types.Address
	password           string
	onlineRepAddresses []types.Address
	logger             *zap.SugaredLogger
	priInfos           *sync.Map
	session            *wallet.Session
}

func (dps *DposService) GetP2PService() p2p.Service {
	return dps.ns
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
	dps.logger.Info("start dpos service")
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
		logger:   log.NewLogger("consensus"),
		priInfos: new(sync.Map),
	}
	dps.session = dps.wallet.NewSession(account)
	//test begin...
	//dps.accounts = append(dps.accounts, ac.Address())
	//test end...
	err := dps.setPriInfo()
	if err != nil {
		dps.logger.Error(err)
		return nil, err
	}
	dps.bp.SetDpos(dps)
	dps.actrx.SetDposService(dps)
	return dps, nil
}

func (dps *DposService) SetWalletStore(wallet *wallet.WalletStore) {
	dps.wallet = wallet
}

func (dps *DposService) setPriInfo() error {
	err := dps.getPriInfo(dps.session)
	if err != nil {
		return err
	}
	return nil
}

func (dps *DposService) getPriInfo(session *wallet.Session) error {
	if dps.account != types.ZeroAddress {
		if verify, err := session.VerifyPassword(dps.password); verify && err == nil {
			if a, err := session.GetAccounts(); err == nil {
				for i := 0; i < len(a); i++ {
					acc, err := session.GetRawKey(a[i])
					if err != nil {
						continue
					}
					dps.priInfos.LoadOrStore(a[i], acc)
				}
			} else {
				return err
			}
		} else {
			return errors.New("invalid password")
		}
	}
	return nil
}

func (dps *DposService) refreshPriInfo() error {
	dps.priInfos.Range(func(key, value interface{}) bool {
		dps.priInfos.Delete(key.(types.Address))
		return true
	})
	err := dps.setPriInfo()
	if err != nil {
		return err
	}
	return nil
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
	dps.logger.Info("Publish Event")
	dps.bp.blocks <- v.(types.Block)
}

func (dps *DposService) ReceiveConfirmReq(v interface{}) {
	dps.logger.Info("ConfirmReq Event")
	dps.bp.blocks <- v.(types.Block)
	dps.onReceiveConfirmReq(v.(types.Block))
}

func (dps *DposService) onReceiveConfirmReq(block types.Block) {
	dps.priInfos.Range(func(key, value interface{}) bool {
		isRep := dps.isThisAccountRepresentation(key.(types.Address))
		if isRep {
			dps.logger.Infof("send confirm ack for hash %s,previous hash is %s", block.GetHash(), block.Root())
			dps.putRepresentativesToOnline(key.(types.Address))
			dps.sendConfirmAck(block, key.(types.Address), value.(*types.Account))
		}
		return true
	})
	//if len(accounts) == 0 {
	//	logger.Info("this is just a node,not a wallet")
	//	dps.sendConfirmReq(block)
	//}
}

func (dps *DposService) ReceiveConfirmAck(v interface{}) {
	dps.logger.Info("ConfirmAck Event")
	vote := v.(*protos.ConfirmAckBlock)
	dps.bp.blocks <- vote.Blk
	dps.onReceiveConfirmAck(vote)
}

func (dps *DposService) onReceiveConfirmAck(va *protos.ConfirmAckBlock) {
	valid := IsAckSignValidate(va)
	if !valid {
		return
	}
	dps.putRepresentativesToOnline(va.Account)
	if v, ok := dps.actrx.roots.Load(va.Blk.Root()); ok {
		result, vt := v.(*Election).vote.voteExit(va.Account)
		if result {
			if vt.Sequence < va.Sequence {
				v.(*Election).vote.repVotes[va.Account] = va
			}
		} else {
			ta := v.(*Election).vote.voteStatus(va)
			if ta == confirm {
				currentvote := v.(*Election).vote.repVotes[va.Account]
				if currentvote.Sequence < va.Sequence {
					v.(*Election).vote.repVotes[va.Account] = va
				}
			}
		}
		dps.actrx.vote(va)
	} else {
		exit, err := dps.ledger.HasStateBlock(va.Blk.GetHash())
		if err != nil {
			return
		}
		if !exit {
			dps.bp.blocks <- va.Blk
		}
	}
}

func (dps *DposService) ReceiveSyncBlock(v interface{}) {
	dps.logger.Info("Sync Event")
	dps.bp.blocks <- v.(types.Block)
}

func (dps *DposService) sendConfirmReq(block types.Block) error {
	packet := &protos.ConfirmReqBlock{
		Blk: block,
	}
	data, err := protos.ConfirmReqBlockToProto(packet)
	if err != nil {
		dps.logger.Error("ConfirmReq Block to Proto error")
		return err
	}
	dps.ns.Broadcast(p2p.ConfirmReq, data)
	return nil
}

func (dps *DposService) sendConfirmAck(block types.Block, account types.Address, acc *types.Account) error {
	va, err := dps.voteGenerate(block, account, acc)
	if err != nil {
		dps.logger.Error("vote generate error")
		return err
	}
	data, err := protos.ConfirmAckBlockToProto(va)
	if err != nil {
		dps.logger.Error("vote to proto error")
	}
	dps.ns.Broadcast(p2p.ConfirmAck, data)
	return nil
}

func (dps *DposService) voteGenerate(block types.Block, account types.Address, acc *types.Account) (*protos.ConfirmAckBlock, error) {
	var va protos.ConfirmAckBlock
	if v, ok := dps.actrx.roots.Load(block.Root()); ok {
		result, vt := v.(*Election).vote.voteExit(account)
		if result {
			va.Sequence = vt.Sequence + 1
			va.Blk = block
			va.Account = account
			va.Signature = acc.Sign(block.GetHash())
		} else {
			va.Sequence = 0
			va.Blk = block
			va.Account = account
			va.Signature = acc.Sign(block.GetHash())
		}
		v.(*Election).vote.voteStatus(&va)
	} else {
		va.Sequence = 0
		va.Blk = block
		va.Account = account
		va.Signature = acc.Sign(block.GetHash())
	}
	return &va, nil
}

func (dps *DposService) GetAccountPrv(account types.Address) (*types.Account, error) {
	session := dps.wallet.NewSession(dps.account)
	if b, err := session.VerifyPassword(dps.password); b && err == nil {
		return session.GetRawKey(account)
	} else {
		return nil, fmt.Errorf("invalid password")
	}
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
					dps.logger.Debug(util.ToString(&addresses))
				}
			}

			return a
		} else {
			dps.logger.Error(err)
		}
	} else {
		dps.logger.Debugf("verify password[%s] failed", dps.password)
	}

	return []types.Address{}
}

func (dps *DposService) putRepresentativesToOnline(addr types.Address) {
	if len(dps.onlineRepAddresses) == 0 {
		dps.onlineRepAddresses = append(dps.onlineRepAddresses, addr)
	} else {
		for i, v := range dps.onlineRepAddresses {
			if v == addr {
				break
			}
			if i == (len(dps.onlineRepAddresses) - 1) {
				dps.onlineRepAddresses = append(dps.onlineRepAddresses, addr)
			}
		}
	}
}

func (dps *DposService) GetOnlineRepresentatives() []types.Address {
	return dps.onlineRepAddresses
}

func (dps *DposService) findOnlineRepresentatives() error {
	dps.priInfos.Range(func(key, value interface{}) bool {
		isRep := dps.isThisAccountRepresentation(key.(types.Address))
		if isRep {
			dps.putRepresentativesToOnline(key.(types.Address))
		}
		return true
	})
	blk, err := dps.ledger.GetRandomStateBlock()
	if err != nil {
		return err
	}
	err = dps.sendConfirmReq(blk)
	if err != nil {
		return err
	}
	return nil
}

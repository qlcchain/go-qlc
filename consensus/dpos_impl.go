package consensus

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/wallet"
	"go.uber.org/zap"
)

//var logger = log.NewLogger("consensus")

const (
	msgCacheSize                       = 65536
	msgCacheExpirationTime             = 30 * time.Minute
	findOnlineRepresentativesIntervals = 2 * time.Minute
)

type DposService struct {
	common.ServiceLifecycle
	ns                 p2p.Service
	ledger             *ledger.Ledger
	eventMsg           map[p2p.EventType]p2p.EventSubscriber
	bp                 *BlockProcessor
	wallet             *wallet.WalletStore
	acTrx              *ActiveTrx
	account            types.Address
	password           string
	onlineRepAddresses []types.Address
	logger             *zap.SugaredLogger
	priInfos           *sync.Map
	session            *wallet.Session
	cache              gcache.Cache
	cfg                *config.Config
}

func (dps *DposService) GetP2PService() p2p.Service {
	return dps.ns
}

func (dps *DposService) Init() error {
	if !dps.PreInit() {
		return errors.New("pre init fail")
	}
	defer dps.PostInit()
	dps.session = dps.wallet.NewSession(dps.account)
	err := dps.setPriInfo()
	if err != nil {
		dps.logger.Error(err)
		return err
	}
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
	go dps.acTrx.start()

	return nil
}

func (dps *DposService) Stop() error {
	if !dps.PreStop() {
		return errors.New("pre stop fail")
	}
	defer dps.PostStop()
	dps.bp.quitCh <- true
	dps.acTrx.quitCh <- true
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
	acTrx := NewActiveTrx()
	l := ledger.NewLedger(cfg.LedgerDir())

	dps := &DposService{
		ns:       netService,
		ledger:   l,
		eventMsg: make(map[p2p.EventType]p2p.EventSubscriber),
		bp:       bp,
		acTrx:    acTrx,
		wallet:   wallet.NewWalletStore(cfg),
		account:  account,
		password: password,
		logger:   log.NewLogger("consensus"),
		priInfos: new(sync.Map),
		cache:    gcache.New(msgCacheSize).LRU().Expiration(msgCacheExpirationTime).Build(),
		cfg:      cfg,
	}
	dps.bp.SetDpos(dps)
	dps.acTrx.SetDposService(dps)
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
	e := v.(p2p.Message)
	p, err := protos.PublishBlockFromProto(e.Data())
	if err != nil {
		dps.logger.Info(err)
		return
	}
	bs := blockSource{
		block:     p.Blk,
		blockFrom: types.UnSynchronized,
	}
	dps.bp.blocks <- bs
	dps.onReceivePublish(e, p.Blk)
}

func (dps *DposService) onReceivePublish(e p2p.Message, blk *types.StateBlock) {
	if !dps.cache.Has(e.Hash()) {
		dps.ns.SendMessageToPeers(p2p.PublishReq, blk, e.MessageFrom())
		err := dps.cache.Set(e.Hash(), "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with publish message", err, blk.GetHash())
		}
	}
}

func (dps *DposService) ReceiveConfirmReq(v interface{}) {
	dps.logger.Info("ConfirmReq Event")
	e := v.(p2p.Message)
	r, err := protos.ConfirmReqBlockFromProto(e.Data())
	if err != nil {
		dps.logger.Error(err)
		return
	}
	dps.onReceiveConfirmReq(e, r.Blk)
}

func (dps *DposService) onReceiveConfirmReq(e p2p.Message, blk *types.StateBlock) {
	bs := blockSource{
		block:     blk,
		blockFrom: types.UnSynchronized,
	}
	if !dps.cache.Has(e.Hash()) {
		var isRep bool
		dps.priInfos.Range(func(key, value interface{}) bool {
			isRep = dps.isThisAccountRepresentation(key.(types.Address))
			if isRep {
				dps.putRepresentativesToOnline(key.(types.Address))
				result, _ := dps.ledger.Process(bs.block)
				if result == ledger.Old {
					dps.logger.Infof("send confirm ack for hash %s,previous hash is %s", bs.block.GetHash(), bs.block.Root())
					dps.sendConfirmAck(bs.block, key.(types.Address), value.(*types.Account))
				}
				dps.bp.processResult(result, bs)
			}
			return true
		})
		if !isRep {
			dps.bp.blocks <- bs
		}
		dps.ns.SendMessageToPeers(p2p.ConfirmReq, blk, e.MessageFrom())
		err := dps.cache.Set(e.Hash(), "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with confirmReq message", err, blk.GetHash())
		}
	}
}

func (dps *DposService) ReceiveConfirmAck(v interface{}) {
	dps.logger.Info("ConfirmAck Event")
	e := v.(p2p.Message)
	ack, err := protos.ConfirmAckBlockFromProto(e.Data())
	if err != nil {
		dps.logger.Info(err)
		return
	}
	dps.onReceiveConfirmAck(e, ack)
}

func (dps *DposService) onReceiveConfirmAck(e p2p.Message, ack *protos.ConfirmAckBlock) {
	bs := blockSource{
		block:     ack.Blk,
		blockFrom: types.UnSynchronized,
	}
	valid := IsAckSignValidate(ack)
	if !valid {
		return
	}
	dps.acTrx.vote(ack)
	if !dps.cache.Has(e.Hash()) {
		var isRep bool
		dps.putRepresentativesToOnline(ack.Account)
		dps.priInfos.Range(func(key, value interface{}) bool {
			isRep = dps.isThisAccountRepresentation(key.(types.Address))
			if isRep {
				dps.putRepresentativesToOnline(key.(types.Address))
				result, _ := dps.ledger.Process(bs.block)
				if result == ledger.Old {
					dps.logger.Infof("send confirm ack for hash %s,previous hash is %s", bs.block.GetHash(), bs.block.Root())
					dps.sendConfirmAck(bs.block, key.(types.Address), value.(*types.Account))
				}
				dps.bp.processResult(result, bs)
				if result == ledger.Progress {
					dps.acTrx.vote(ack)
				}
			}
			return true
		})
		if !isRep {
			dps.bp.blocks <- bs
		}

		dps.ns.SendMessageToPeers(p2p.ConfirmAck, ack, e.MessageFrom())
		err := dps.cache.Set(e.Hash(), "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with confirmAck message", err, ack.Blk.GetHash())
		}
	}
}

func (dps *DposService) ReceiveSyncBlock(v interface{}) {
	dps.logger.Info("Sync Event")
	bs := blockSource{
		block:     v.(*types.StateBlock),
		blockFrom: types.Synchronized,
	}
	dps.bp.blocks <- bs
}

func (dps *DposService) sendConfirmAck(block *types.StateBlock, account types.Address, acc *types.Account) error {
	va, err := dps.voteGenerate(block, account, acc)
	if err != nil {
		dps.logger.Error("vote generate error")
		return err
	}
	dps.ns.Broadcast(p2p.ConfirmAck, va)
	return nil
}

func (dps *DposService) voteGenerate(block *types.StateBlock, account types.Address, acc *types.Account) (*protos.ConfirmAckBlock, error) {
	va := &protos.ConfirmAckBlock{
		Sequence:  0,
		Blk:       block,
		Account:   account,
		Signature: acc.Sign(block.GetHash()),
	}
	return va, nil
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
	dps.ns.Broadcast(p2p.ConfirmReq, blk)
	return nil
}

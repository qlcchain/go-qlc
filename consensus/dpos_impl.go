package consensus

import (
	"encoding/hex"
	"fmt"

	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/wallet"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p"
)

var logger = log.NewLogger("consensus")
var seed = "DB68096C0E2D2954F59DA5DAAE112B7B6F72BE35FC96327FE0D81FD0CE5794A9"

type DposService struct {
	ns       p2p.Service
	ledger   *ledger.Ledger
	eventMsg map[p2p.EventType]p2p.EventSubscriber
	quitCh   chan bool
	bp       *BlockProcessor
	wallet   *wallet.WalletStore
	accounts []types.Address
	actrx    *ActiveTrx
}

func NewDposService(netService p2p.Service, ledger *ledger.Ledger) (*DposService, error) {
	//test begin...
	s, err := hex.DecodeString(seed)
	seed, err := types.BytesToSeed(s)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	ac, err := seed.Account(0)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//test end...
	bp := NewBlockProcessor()
	actrx := NewActiveTrx()
	dps := &DposService{
		ns:       netService,
		ledger:   ledger,
		eventMsg: make(map[p2p.EventType]p2p.EventSubscriber),
		quitCh:   make(chan bool, 1),
		bp:       bp,
		actrx:    actrx,
	}
	//test begin...
	dps.accounts = append(dps.accounts, ac.Address())
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
func (dps *DposService) Start() {
	dps.setEvent()
	logger.Info("start dpos service")
	go dps.bp.Start()
	go dps.actrx.start()
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
	for _, k := range dps.accounts {
		isRep := dps.isThisAccountRepresentation(k)
		if isRep == true {
			logger.Infof("send confirm ack for hash %s,previous hash is %s", block.GetHash(), block.Root())
			dps.sendConfirmAck(block, k)
		} else {
			logger.Infof("send confirm req for hash %s,previous hash is %s", block.GetHash(), block.Root())
			dps.sendConfirmReq(block)
		}
	}
	if len(dps.accounts) == 0 {
		logger.Info("this is just a node,not a wallet")
		dps.sendConfirmReq(block)
	}
}
func (dps *DposService) ReceiveConfirmAck(v interface{}) {
	logger.Info("ConfirmAck Event")
	vote := v.(*protos.ConfirmAckBlock)
	dps.bp.blocks <- vote.Blk
	dps.onReceiveConfirmAck(vote)
}
func (dps *DposService) onReceiveConfirmAck(vote_a *protos.ConfirmAckBlock) {
	valid := IsAckSignValidate(vote_a)
	if valid != true {
		return
	}
	if v, ok := dps.actrx.roots[vote_a.Blk.Root()]; ok {
		result, vt := v.vote.voteExit(vote_a.Account)
		if result == true {
			if vt.Sequence < vote_a.Sequence {
				v.vote.rep_votes[vote_a.Account] = vote_a
			}
		} else {
			ta := v.vote.voteStatus(vote_a)
			if ta == confirm {
				currentvote := v.vote.rep_votes[vote_a.Account]
				if currentvote.Sequence < vote_a.Sequence {
					v.vote.rep_votes[vote_a.Account] = vote_a
				}
			}
		}
		dps.actrx.vote(vote_a)
	} else {
		exit, err := dps.ledger.HasBlock(vote_a.Blk.GetHash())
		if err != nil {
			return
		}
		if exit == true {
			for _, k := range dps.accounts {
				isRep := dps.isThisAccountRepresentation(k)
				if isRep == true {
					dps.sendConfirmAck(vote_a.Blk, k)
				} else {
					data, err := protos.ConfirmAckBlockToProto(vote_a)
					if err != nil {
						logger.Error("vote to proto error")
					}
					dps.ns.Broadcast(p2p.ConfirmAck, data)
				}
			}
			if len(dps.accounts) == 0 {
				logger.Info("this is just a node,not a wallet")
				data, err := protos.ConfirmAckBlockToProto(vote_a)
				if err != nil {
					logger.Error("vote to proto error")
				}
				dps.ns.Broadcast(p2p.ConfirmAck, data)
			}
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
	var password string
	vote_a, err := dps.voteGenerate(block, account, password)
	if err != nil {
		logger.Error("vote generate error")
		return err
	}
	data, err := protos.ConfirmAckBlockToProto(vote_a)
	if err != nil {
		logger.Error("vote to proto error")
	}
	dps.ns.Broadcast(p2p.ConfirmAck, data)
	return nil
}
func (dps *DposService) voteGenerate(block types.Block, account types.Address, password string) (*protos.ConfirmAckBlock, error) {
	var vote_a protos.ConfirmAckBlock
	prv, err := dps.GetAccountPrv(account, password)
	if err != nil {
		logger.Error("Get prv error")
		return nil, err
	}
	if v, ok := dps.actrx.roots[block.Root()]; ok {
		result, vt := v.vote.voteExit(account)
		if result == true {
			vote_a.Sequence = vt.Sequence + 1
			vote_a.Blk = block
			vote_a.Account = account
			vote_a.Signature = prv.Sign(block.GetHash())
		} else {
			vote_a.Sequence = 0
			vote_a.Blk = block
			vote_a.Account = account
			vote_a.Signature = prv.Sign(block.GetHash())
		}
		v.vote.voteStatus(&vote_a)
	} else {
		vote_a.Sequence = 0
		vote_a.Blk = block
		vote_a.Account = account
		vote_a.Signature = prv.Sign(block.GetHash())
	}
	return &vote_a, nil
}
func (dps *DposService) GetAccountPrv(account types.Address, password string) (*types.Account, error) {
	/*	session := dps.wallet.NewSession(account)
		session.EnterPassword(password)
		return session.GetRawKey(account)*/
	//test begin......
	s, err := hex.DecodeString(seed)
	seed, err := types.BytesToSeed(s)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	ac, err := seed.Account(0)
	return ac, nil
	//test end......
}
func (dps *DposService) isThisAccountRepresentation(address types.Address) bool {
	_, err := dps.ledger.GetRepresentation(address)
	if err != nil {
		return false
	} else {
		return true
	}
}
func (dp *DposService) Stop() {
	dp.quitCh <- true
	dp.bp.quitCh <- true
	dp.actrx.quitCh <- true
	for i, j := range dp.eventMsg {
		dp.ns.MessageEvent().GetEvent("consensus").UnSubscribe(i, j)
	}
}

package dpos

import (
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
)

const (
	waitingVoteMaxTime = 180
)

type voteKey [types.HashSize]byte

type ActiveTrx struct {
	dps    *DPoS
	roots  *sync.Map
	quitCh chan bool
	exited chan struct{}
}

func newActiveTrx() *ActiveTrx {
	return &ActiveTrx{
		roots:  new(sync.Map),
		quitCh: make(chan bool, 1),
		exited: make(chan struct{}, 1),
	}
}

func (act *ActiveTrx) setDPoSService(dps *DPoS) {
	act.dps = dps
}

func (act *ActiveTrx) start() {
	timerCheckVotes := time.NewTicker(time.Second)

	for {
		select {
		case <-act.quitCh:
			act.dps.logger.Info("act stopped")
			act.exited <- struct{}{}
			return
		case <-timerCheckVotes.C:
			act.checkVotes()
		}
	}
}

func (act *ActiveTrx) stop() {
	act.quitCh <- true
	<-act.exited
}

func getVoteKey(block *types.StateBlock) voteKey {
	var key voteKey

	if block.IsOpen() {
		hash, _ := types.HashBytes(block.Address[:], block.Token[:])
		copy(key[:], hash[:])
	} else {
		copy(key[:], block.Previous[:])
	}

	return key
}

func (act *ActiveTrx) addToRoots(block *types.StateBlock) bool {
	vk := getVoteKey(block)

	if _, ok := act.roots.Load(vk); !ok {
		el := newElection(act.dps, block)
		act.roots.Store(vk, el)
		return true
	} else {
		act.dps.logger.Debugf("block :%s already exist in roots", block.GetHash())
		return false
	}
}

func (act *ActiveTrx) cleanFrontierVotes() {
	dps := act.dps

	dps.frontiersStatus.Range(func(k, v interface{}) bool {
		fHash := k.(types.Hash)
		dps.hash2el.Delete(fHash)

		act.roots.Range(func(key, value interface{}) bool {
			el := value.(*Election)
			if _, ok := el.blocks.Load(fHash); ok {
				el.blocks.Delete(fHash)

				num := 0
				el.blocks.Range(func(kk, vv interface{}) bool {
					num++
					return true
				})

				if num == 0 {
					act.roots.Delete(el.vote.id)
				}

				return false
			}
			return true
		})

		return true
	})
}

func (act *ActiveTrx) checkVotes() {
	nowTime := time.Now().Unix()
	dps := act.dps

	act.roots.Range(func(key, value interface{}) bool {
		el := value.(*Election)
		if nowTime-el.lastTime < waitingVoteMaxTime {
			return true
		}

		block := el.status.winner
		hash := block.GetHash()

		if !el.ifValidAndSetInvalid() {
			return true
		}

		act.roots.Delete(el.vote.id)
		el.cleanBlockInfo()
		act.dps.lv.RollbackUnchecked(hash)

		if dps.isReceivedFrontier(hash) {
			dps.logger.Warnf("frontier[%s] was not confirmed in %d seconds", hash, waitingVoteMaxTime)
		} else {
			dps.logger.Warnf("block[%s] was not confirmed in %d seconds", hash, waitingVoteMaxTime)
		}

		return true
	})
}

func (act *ActiveTrx) addWinner2Ledger(block *types.StateBlock) {
	hash := block.GetHash()
	dps := act.dps
	dps.logger.Debugf("block[%s] confirmed", hash)

	if exist, err := dps.ledger.HasStateBlockConfirmed(hash); !exist && err == nil {
		err := dps.lv.BlockProcess(block)
		if err != nil {
			dps.logger.Error(err)
		} else {
			dps.confirmedBlockInc(hash)
			dps.statBlockInc()
			dps.logger.Debugf("save block[%s]", hash)
		}
	} else {
		dps.logger.Debugf("%s, %v", hash, err)
	}
}

func (act *ActiveTrx) addSyncBlock2Ledger(block *types.StateBlock) {
	hash := block.GetHash()
	dps := act.dps
	dps.logger.Infof("sync block[%s] confirmed", hash)

	if exist, err := dps.ledger.HasStateBlockConfirmed(hash); !exist && err == nil {
		err := dps.lv.BlockProcess(block)
		if err != nil {
			dps.logger.Error(err)
		} else {
			dps.confirmedBlockInc(hash)
			dps.statBlockInc()
			dps.logger.Debugf("save block[%s]", hash)
		}
	} else {
		dps.logger.Debugf("%s, %v", hash, err)
	}

	dps.chainFinished(hash)
}

func (act *ActiveTrx) rollBack(blocks []*types.StateBlock) {
	dps := act.dps

	for _, v := range blocks {
		hash := v.GetHash()
		dps.logger.Info("loser hash is:", hash)

		has, err := dps.ledger.HasStateBlock(hash)
		if err != nil {
			dps.logger.Errorf("error [%s] when run HasStateBlock func ", err)
			continue
		}

		if has {
			err = dps.lv.Rollback(hash)
			if err != nil {
				dps.logger.Errorf("error [%s] when rollback hash [%s]", err, hash)
			}
		}
	}
}

func (act *ActiveTrx) vote(vi *voteInfo) {
	if v, ok := act.dps.hash2el.Load(vi.hash); ok {
		el := v.(*Election)
		el.voteAction(vi)
	}
}

func (act *ActiveTrx) voteFrontier(vi *voteInfo) (confirmed bool) {
	if v, ok := act.dps.hash2el.Load(vi.hash); ok {
		el := v.(*Election)
		return el.voteFrontier(vi)
	}
	return false
}

func (act *ActiveTrx) isVoting(block *types.StateBlock) bool {
	vk := getVoteKey(block)

	if _, ok := act.roots.Load(vk); ok {
		return true
	}

	return false
}

func (act *ActiveTrx) getVoteInfo(block *types.StateBlock) *Election {
	vk := getVoteKey(block)

	if v, ok := act.roots.Load(vk); ok {
		return v.(*Election)
	}

	return nil
}

func (act *ActiveTrx) setVoteHash(block *types.StateBlock) {
	vk := getVoteKey(block)

	if v, ok := act.roots.Load(vk); ok {
		v.(*Election).voteHash = block.GetHash()
	}
}

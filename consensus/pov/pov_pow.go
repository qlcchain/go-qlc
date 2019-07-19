package pov

import (
	"errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
	"sync"
)

type ConsensusPow struct {
	chainR        PovConsensusChainReader
	logger        *zap.SugaredLogger
	mineWorkerNum int
}

func NewConsensusPow(chainR PovConsensusChainReader) *ConsensusPow {
	consPow := &ConsensusPow{chainR: chainR}
	consPow.logger = log.NewLogger("pov_cs_pow")
	return consPow
}

func (c *ConsensusPow) Init() error {
	/*
		cpuNum := runtime.NumCPU()
		if cpuNum >= 4 {
			c.mineWorkerNum = cpuNum / 2
		} else {
			c.mineWorkerNum = 1
		}
	*/
	return nil
}

func (c *ConsensusPow) Start() error {
	return nil
}

func (c *ConsensusPow) Stop() error {
	return nil
}

func (c *ConsensusPow) VerifyHeader(header *types.PovHeader) error {
	var err error

	err = c.verifyTarget(header)
	if err != nil {
		return err
	}

	err = c.verifyProducer(header)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConsensusPow) verifyProducer(header *types.PovHeader) error {
	if header.GetHeight() < common.PovMinerVerifyHeightStart {
		return nil
	}

	prevHeader := c.chainR.GetHeaderByHash(header.GetPrevious())
	if prevHeader == nil {
		return errors.New("failed to get previous header")
	}

	prevStateHash := prevHeader.GetStateHash()
	prevTrie := c.chainR.GetStateTrie(&prevStateHash)
	if prevTrie == nil {
		return errors.New("failed to get previous state tire")
	}

	asBytes := prevTrie.GetValue(header.GetCoinbase().Bytes())
	if len(asBytes) <= 0 {
		return errors.New("failed to get account state value")
	}

	as := new(types.PovAccountState)
	err := as.Deserialize(asBytes)
	if err != nil {
		return errors.New("failed to deserialize account state value")
	}

	if as.RepState == nil {
		return errors.New("account rep state is nil")
	}
	rs := as.RepState

	if rs.Vote.Compare(common.PovMinerPledgeAmountMin) == types.BalanceCompSmaller {
		return errors.New("coinbase pledge amount not enough")
	}

	return nil
}

func (c *ConsensusPow) verifyTarget(header *types.PovHeader) error {
	prevHeader := c.chainR.GetHeaderByHash(header.GetPrevious())
	if prevHeader == nil {
		return errors.New("failed to get previous header")
	}

	expectedTarget, err := c.chainR.CalcNextRequiredTarget(prevHeader)
	if err != nil {
		return err
	}
	if expectedTarget != header.GetTarget() {
		return errors.New("target not equal next required target")
	}

	voteHash := header.ComputeVoteHash()
	voteSig := header.GetVoteSignature()

	isVerified := header.GetCoinbase().Verify(voteHash.Bytes(), voteSig.Bytes())
	if !isVerified {
		return errors.New("bad vote signature")
	}

	voteSigInt := voteSig.ToBigInt()

	targetSig := header.GetTarget()
	targetInt := targetSig.ToBigInt()

	if voteSigInt.Cmp(targetInt) > 0 {
		return errors.New("target greater than vote signature")
	}

	return nil
}

func (c *ConsensusPow) SealHeader(header *types.PovHeader, cbAccount *types.Account, quitCh chan struct{}, resultCh chan<- *types.PovHeader) error {
	var wgMine sync.WaitGroup

	abortCh := make(chan struct{})
	localCh := make(chan *types.PovHeader)

	for id := 1; id <= c.mineWorkerNum; id++ {
		wgMine.Add(1)
		go func(id int, gap uint64) {
			defer wgMine.Done()
			c.mineWorker(id, gap, header, cbAccount, abortCh, localCh)
		}(id, uint64(c.mineWorkerNum))
	}

	go func() {
		select {
		case <-quitCh:
			close(abortCh)
		case result := <-localCh:
			select {
			case resultCh <- result:
			default:
				// there's no hash
				c.logger.Warnf("failed to send sealing result to miner")
			}
			close(abortCh)
		}
		wgMine.Wait()
	}()

	return nil
}

func (c *ConsensusPow) mineWorker(id int, gap uint64, header *types.PovHeader, cbAccount *types.Account, abortCh chan struct{}, localCh chan *types.PovHeader) {
	copyHdr := header.Copy()
	targetInt := copyHdr.Target.ToBigInt()

	tryCnt := 0
	for nonce := gap; nonce < common.PovMaxNonce; nonce += gap {
		tryCnt++
		if tryCnt >= 100 {
			tryCnt = 0
			select {
			case <-abortCh:
				c.logger.Debugf("mine worker %d abort search nonce", id)
				localCh <- nil
				return
			default:
				//Non-blocking select to fall through
			}
		}

		copyHdr.Nonce = nonce

		voteHash := copyHdr.ComputeVoteHash()
		voteSignature := cbAccount.Sign(voteHash)
		voteSigInt := voteSignature.ToBigInt()
		if voteSigInt.Cmp(targetInt) <= 0 {
			c.logger.Debugf("mine worker %d found nonce %d", id, nonce)
			copyHdr.VoteSignature = voteSignature
			localCh <- copyHdr
			return
		}
	}

	c.logger.Debugf("mine worker %d exhaust nonce", id)
	localCh <- nil
}
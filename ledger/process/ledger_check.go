package process

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

// check block base info
type baseInfoCheck interface {
	baseInfo(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error)
}

type blockBaseInfoCheck struct {
}

func (blockBaseInfoCheck) baseInfo(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	hash := block.GetHash()
	address := block.GetAddress()
	checkWork := true
	checkSign := true

	lv.logger.Debug("check block ", hash)
	blockExist, err := lv.l.HasStateBlockConfirmed(hash)
	if err != nil {
		return errorP(block, Other, err)
	}

	if blockExist {
		return errorP(block, Old, nil)
	}

	if block.GetType() == types.ContractSend {
		// check private tx
		if block.IsPrivate() && !block.IsRecipient() {
			return Progress, nil
		}

		if c, ok, err := contract.GetChainContract(types.Address(block.GetLink()), block.GetPayload()); ok && err == nil {
			checkWork = c.GetDescribe().WithWork()
			checkSign = c.GetDescribe().WithSignature()
		}
	}

	if block.GetType() == types.ContractReward {
		// check private tx
		if block.IsPrivate() && !block.IsRecipient() {
			return Progress, nil
		}

		linkBlk, err := lv.l.GetStateBlockConfirmed(block.GetLink())
		if err != nil {
			return errorP(block, GapSource, nil)
		}

		if c, ok, err := contract.GetChainContract(types.Address(linkBlk.GetLink()), block.GetPayload()); ok && err == nil {
			checkWork = c.GetDescribe().WithWork()
			checkSign = c.GetDescribe().WithSignature()
		}
	}

	if checkWork && !block.IsValid() {
		return errorP(block, BadWork, nil)
	}

	if checkSign {
		signature := block.GetSignature()
		if !address.Verify(hash[:], signature[:]) {
			return errorP(block, BadSignature, nil)
		}
	}

	return Progress, nil
}

type cacheBlockBaseInfoCheck struct {
}

func (cacheBlockBaseInfoCheck) baseInfo(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	hash := block.GetHash()
	address := block.GetAddress()
	checkWork := true
	checkSign := true

	lv.logger.Debug("check cache block base info: ", hash)
	if pr := checkReceiveBlockRepeat(lv, block); pr != Progress {
		return errorP(block, pr, nil)
	}
	blockExist, err := lv.l.HasStateBlock(hash)
	if err != nil {
		return errorP(block, Other, err)
	}

	if blockExist {
		return errorP(block, Old, nil)
	}

	if block.GetType() == types.ContractSend {
		// check private tx
		if block.IsPrivate() && !block.IsRecipient() {
			return Progress, nil
		}

		if c, ok, err := contract.GetChainContract(types.Address(block.GetLink()), block.GetPayload()); ok && err == nil {
			checkWork = c.GetDescribe().WithWork()
			checkSign = c.GetDescribe().WithSignature()
		}
	}

	if block.GetType() == types.ContractReward {
		// check private tx
		if block.IsPrivate() && !block.IsRecipient() {
			return Progress, nil
		}

		linkBlk, err := lv.l.GetStateBlock(block.GetLink())
		if err != nil {
			return errorP(block, GapSource, nil)
		}

		// check private tx
		if linkBlk.IsPrivate() && !linkBlk.IsRecipient() {
			return Progress, nil
		}

		if c, ok, err := contract.GetChainContract(types.Address(linkBlk.GetLink()), linkBlk.GetPayload()); ok && err == nil {
			checkWork = c.GetDescribe().WithWork()
			checkSign = c.GetDescribe().WithSignature()
		}
	}

	if checkWork && !block.IsValid() {
		return errorP(block, BadWork, nil)
	}

	if checkSign {
		signature := block.GetSignature()
		if !address.Verify(hash[:], signature[:]) {
			return errorP(block, BadSignature, nil)
		}
	}

	return Progress, nil
}

func checkReceiveBlockRepeat(lv *LedgerVerifier, block *types.StateBlock) ProcessResult {
	r := Progress
	if block.IsReceiveBlock() {
		var repeatedFound error
		err := lv.l.GetBlockCaches(func(b *types.StateBlock) error {
			if block.GetLink() == b.GetLink() && block.GetHash() != b.GetHash() {
				r = ReceiveRepeated
				return repeatedFound
			}
			return nil
		})
		if err != nil && err != repeatedFound {
			return Other
		}
	}
	return r
}

// check block fork
type forkCheck interface {
	fork(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error)
}

type blockForkCheck struct {
}

func (blockForkCheck) fork(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Send, types.Receive, types.Change, types.Online, types.ContractSend:
		if previous, err := lv.l.GetStateBlockConfirmed(block.GetPrevious()); err != nil {
			return errorP(block, GapPrevious, nil)
		} else {
			if tm, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken()); err == nil && previous.GetHash() != tm.Header {
				return errorP(block, Fork, fmt.Errorf("previous is %s, but token header is %s", previous.GetHash().String(), tm.Header.String()))
			}
		}
		return Progress, nil
	case types.Open:
		//check fork
		if _, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken()); err == nil {
			return errorP(block, Fork, fmt.Errorf("open block, but token already exist"))
		}
		return Progress, nil
	case types.ContractReward:
		// check previous
		if !block.IsOpen() {
			// check previous
			if previous, err := lv.l.GetStateBlockConfirmed(block.GetPrevious()); err != nil {
				return errorP(block, GapPrevious, nil)
			} else {
				//check fork
				if tm, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken()); err == nil && previous.GetHash() != tm.Header {
					return errorP(block, Fork, fmt.Errorf("previous is %s, but token header is %s", previous.GetHash().String(), tm.Header.String()))
				}
			}
		} else {
			//check fork
			if _, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken()); err == nil {
				return errorP(block, Fork, fmt.Errorf("open block, but token already exist"))
			}
		}
		return Progress, nil
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

type cacheBlockForkCheck struct {
}

func (cacheBlockForkCheck) fork(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Send, types.Receive, types.Change, types.Online, types.ContractSend:
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			return errorP(block, GapPrevious, nil)
		} else {
			//check fork
			if tm, err := lv.l.GetTokenMeta(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
				return errorP(block, Fork, fmt.Errorf("previous is %s, token header is %s", previous.GetHash().String(), tm.Header.String()))
			}
		}
		return Progress, nil
	case types.Open:
		if _, err := lv.l.GetTokenMeta(block.Address, block.Token); err == nil {
			return errorP(block, Fork, fmt.Errorf("open block, token already exist"))
		}
		return Progress, nil
	case types.ContractReward:
		if !block.IsOpen() {
			if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
				return errorP(block, GapPrevious, nil)
			} else {
				if tm, err := lv.l.GetTokenMeta(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
					return errorP(block, Fork, fmt.Errorf("previous is %s, token header is %s", previous.GetHash().String(), tm.Header.String()))
				}
			}
		} else {
			if _, err := lv.l.GetTokenMeta(block.Address, block.Token); err == nil {
				return errorP(block, Fork, fmt.Errorf("open block, token already exist"))
			}
		}
		return Progress, nil
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

// check block balance
type balanceCheck interface {
	balance(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error)
}

type blockBalanceCheck struct {
}

func (blockBalanceCheck) balance(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Send:
		if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
			return errorP(block, GapPrevious, nil)
		} else {
			//check balance
			if !(previous.Balance.Compare(block.Balance) == types.BalanceCompBigger) {
				return errorP(block, BalanceMismatch, fmt.Errorf("previous: %s, block: %s", previous.Balance.String(), block.Balance.String()))
			}
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
				return errorP(block, BalanceMismatch, fmt.Errorf("weight not equal"))
			}
		}
		return Progress, nil
	case types.ContractSend:
		if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
			return errorP(block, GapPrevious, nil)
		} else {
			//check totalBalance
			if previous.TotalBalance().Compare(block.TotalBalance()) == types.BalanceCompSmaller {
				return errorP(block, BalanceMismatch, fmt.Errorf("previous: %s, block: %s", previous.TotalBalance().String(), block.TotalBalance().String()))
			}
		}
		return Progress, nil
	case types.Change, types.Online:
		// check previous
		if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
			return errorP(block, GapPrevious, nil)
		} else {
			//check balance
			if block.Balance.Compare(previous.Balance) != types.BalanceCompEqual {
				return errorP(block, BalanceMismatch, fmt.Errorf("previous: %s, block: %s", previous.Balance.String(), block.Balance.String()))
			}
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
				return errorP(block, BalanceMismatch, fmt.Errorf("weight not equal"))
			}
		}
		return Progress, nil
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

type cacheBlockBalanceCheck struct {
}

func (cacheBlockBalanceCheck) balance(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Send:
		// check previous
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			return errorP(block, GapPrevious, nil)
		} else {
			if !(previous.Balance.Compare(block.Balance) == types.BalanceCompBigger) {
				return errorP(block, BalanceMismatch, fmt.Errorf("previous: %s, block: %s", previous.Balance.String(), block.Balance.String()))
			}
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
				return errorP(block, BalanceMismatch, fmt.Errorf("weight not equal"))
			}
		}
		return Progress, nil
	case types.Change, types.Online:
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			return errorP(block, GapPrevious, nil)
		} else {
			//check balance
			if block.Balance.Compare(previous.Balance) != types.BalanceCompEqual {
				return errorP(block, BalanceMismatch, fmt.Errorf("previous: %s, block: %s", previous.Balance.String(), block.Balance.String()))
			}
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
				return errorP(block, BalanceMismatch, fmt.Errorf("weight not equal"))
			}
		}
		return Progress, nil
	case types.ContractSend:
		// check previous
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			return errorP(block, GapPrevious, nil)
		} else {
			//check totalBalance
			if previous.TotalBalance().Compare(block.TotalBalance()) == types.BalanceCompSmaller {
				return errorP(block, BalanceMismatch, fmt.Errorf("previous: %s, block: %s", previous.TotalBalance().String(), block.TotalBalance().String()))
			}
		}
		return Progress, nil
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

// check block pending
type pendingCheck interface {
	pending(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error)
}

type blockPendingCheck struct {
}

func (blockPendingCheck) pending(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Receive:
		// check previous
		if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
			return errorP(block, GapPrevious, nil)
		} else {
			pendingKey := types.PendingKey{
				Address: block.Address,
				Hash:    block.Link,
			}
			//check pending
			if pending, err := lv.l.GetPending(&pendingKey); err == nil {
				if tm, err := lv.l.GetTokenMetaConfirmed(block.Address, block.Token); err == nil {
					transferAmount := block.GetBalance().Sub(tm.Balance)
					if !pending.Amount.Equal(transferAmount) || pending.Type != block.Token {
						return errorP(block, BalanceMismatch, fmt.Errorf("pending: %s, transfer: %s", pending.Amount.String(), transferAmount.String()))
					}
					//check vote,network,storage,oracle
					if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
						previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
						previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
						previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
						return errorP(block, BalanceMismatch, fmt.Errorf("weight not equal"))
					}
				} else {
					return errorP(block, Other, err)
				}
			} else if err == ledger.ErrPendingNotFound {
				return errorP(block, UnReceivable, fmt.Errorf("pending key: %s", pendingKey))
			} else {
				return errorP(block, Other, err)
			}
		}
		return Progress, nil
	case types.Open:
		//check link
		pendingKey := types.PendingKey{
			Address: block.Address,
			Hash:    block.Link,
		}
		//check pending
		if pending, err := lv.l.GetPending(&pendingKey); err == nil {
			if !pending.Amount.Equal(block.Balance) || pending.Type != block.Token {
				return errorP(block, BalanceMismatch, fmt.Errorf("pending: %s, block: %s", pending.Amount.String(), block.Balance.String()))
			}
			//check vote,network,storage,oracle
			vote := block.GetVote()
			network := block.GetNetwork()
			storage := block.GetStorage()
			oracle := block.GetOracle()
			if !vote.IsZero() || !network.IsZero() ||
				!storage.IsZero() || !oracle.IsZero() {
				return errorP(block, BalanceMismatch, fmt.Errorf("weight is not zero"))
			}
		} else if err == ledger.ErrPendingNotFound {
			return errorP(block, UnReceivable, fmt.Errorf("pending key: %s", pendingKey))
		} else {
			return errorP(block, Other, err)
		}
		return Progress, nil
	case types.ContractReward:
		return checkContractPending(lv, block)
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

type cacheBlockPendingCheck struct {
}

func (cacheBlockPendingCheck) pending(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Receive:
		// check previous
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			pendingKey := types.PendingKey{
				Address: block.Address,
				Hash:    block.Link,
			}

			//check pending
			if pending, err := lv.l.GetPending(&pendingKey); err == nil {
				if tm, err := lv.l.GetTokenMeta(block.Address, block.Token); err == nil {
					transferAmount := block.GetBalance().Sub(tm.Balance)
					if !pending.Amount.Equal(transferAmount) || pending.Type != block.Token {
						return errorP(block, BalanceMismatch, fmt.Errorf("pending: %s, transfer: %s", pending.Amount.String(), transferAmount.String()))
					}
					//check vote,network,storage,oracle
					if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
						previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
						previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
						previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
						return errorP(block, BalanceMismatch, fmt.Errorf("weight not equal"))
					}
				} else {
					return errorP(block, Other, fmt.Errorf("pending check err: %s", err))
				}
			} else if err == ledger.ErrPendingNotFound {
				return errorP(block, UnReceivable, fmt.Errorf("pending key: %s", pendingKey))
			} else {
				return errorP(block, Other, err)
			}
		}
		return Progress, nil
	case types.Open:
		pendingKey := types.PendingKey{
			Address: block.Address,
			Hash:    block.Link,
		}
		//check pending
		if pending, err := lv.l.GetPending(&pendingKey); err == nil {
			if !pending.Amount.Equal(block.Balance) || pending.Type != block.Token {
				return errorP(block, BalanceMismatch, fmt.Errorf("pending: %s, block: %s", pending.Amount.String(), block.Balance.String()))
			}
			//check vote,network,storage,oracle
			vote := block.GetVote()
			network := block.GetNetwork()
			storage := block.GetStorage()
			oracle := block.GetOracle()
			if !vote.IsZero() || !network.IsZero() ||
				!storage.IsZero() || !oracle.IsZero() {
				return errorP(block, BalanceMismatch, fmt.Errorf("weight is not zero"))
			}
		} else if err == ledger.ErrPendingNotFound {
			return errorP(block, UnReceivable, fmt.Errorf("pending key: %s", pendingKey))
		} else {
			return errorP(block, Other, err)
		}
		return Progress, nil
	case types.ContractReward:
		return checkContractPending(lv, block)
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

func checkContractPending(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	input, err := lv.l.GetStateBlockConfirmed(block.GetLink())
	if err != nil {
		lv.logger.Errorf("send block must be confirmed: %s", err)
		return errorP(block, GapSource, nil)
	}

	// check private tx
	if input.IsPrivate() && !input.IsRecipient() {
		return Progress, nil
	}

	pendingKey := types.PendingKey{
		Address: block.GetAddress(),
		Hash:    block.GetLink(),
	}

	// check pending
	if c, ok, err := contract.GetChainContract(types.Address(input.Link), input.GetPayload()); ok && err == nil {
		d := c.GetDescribe()
		if d.WithPending() {
			if _, err := lv.l.GetPending(&pendingKey); err == nil {
				return Progress, nil
			} else if err == ledger.ErrPendingNotFound {
				return errorP(block, UnReceivable, fmt.Errorf("pending key is: %s", pendingKey))
			} else {
				return errorP(block, Other, fmt.Errorf("get contract pending: %s", err))
			}
		}
	} else {
		return errorP(block, Other, fmt.Errorf("can not find chain contract: %s", input.GetLink().String()))
	}
	return Progress, nil
}

// check block source
type sourceCheck interface {
	source(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error)
}

type blockSourceCheck struct {
}

func (blockSourceCheck) source(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Receive, types.Open:
		if b, err := lv.l.HasStateBlockConfirmed(block.Link); !b && err == nil {
			return errorP(block, GapSource, nil)
		}
		return Progress, nil
	case types.ContractReward:
		// implement in contract data check
		return Progress, nil
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

type cacheBlockSourceCheck struct {
}

func (cacheBlockSourceCheck) source(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Receive, types.Open:
		if b, err := lv.l.HasStateBlock(block.Link); !b && err == nil {
			return errorP(block, GapSource, nil)
		}
		return Progress, nil
	case types.ContractReward:
		// implement in contract data check
		return Progress, nil
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

// check block contract
type contractCheck interface {
	contract(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error)
}

type blockContractCheck struct {
}

func (blockContractCheck) contract(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	// check private tx
	if block.IsPrivate() && !block.IsRecipient() {
		return Progress, nil
	}

	switch block.GetType() {
	case types.ContractSend:
		return checkContractSendBlock(lv, block)
	case types.ContractReward:
		// check receive link
		input, err := lv.l.GetStateBlockConfirmed(block.GetLink())
		if err != nil {
			return errorP(block, GapSource, nil)
		}

		// check private tx
		if input.IsPrivate() && !input.IsRecipient() {
			return Progress, nil
		}

		address := types.Address(input.GetLink())
		if c, ok, err := contract.GetChainContract(address, input.GetPayload()); ok && err == nil {
			clone := block.Clone()
			vmCtx := vmstore.NewVMContextWithBlock(lv.l, block)
			if vmCtx == nil {
				return errorP(block, Other, errors.New("check contract: can not get vm context"))
			}
			if g, e := c.DoReceive(vmCtx, clone, input); e == nil {
				if len(g) > 0 {
					amount, err := lv.l.CalculateAmount(block)
					if err != nil {
						return errorP(block, Other, fmt.Errorf("calculate amount error: %s", err))
					}
					if bytes.EqualFold(g[0].Block.Data, block.Data) && g[0].Token == block.Token &&
						g[0].Amount.Compare(amount) == types.BalanceCompEqual && g[0].ToAddress == block.Address {
						return Progress, nil
					} else {
						errInfo := fmt.Errorf("data from contract, %s, %s, %s, %s, data from block, %s, %s, %s, %s",
							g[0].Block.Data, g[0].Token, g[0].Amount, g[0].ToAddress, block.Data, block.Token, amount, block.Address)
						lv.logger.Error(errInfo)
						return errorP(block, InvalidData, errInfo)
					}
				} else {
					return errorP(block, Other, fmt.Errorf("can not generate receive block"))
				}
			} else {
				if address == contractaddress.MintageAddress && e == vmstore.ErrStorageNotFound {
					return errorP(block, GapTokenInfo, nil)
				} else {
					return errorP(block, Other, fmt.Errorf("DoReceive err: %s ", e))
				}
			}
		} else {
			//call vm.Run();
			return errorP(block, Other, fmt.Errorf("can not find chain contract %s", address.String()))
		}
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

type cacheBlockContractCheck struct {
}

func (cacheBlockContractCheck) contract(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	// check private tx
	if block.IsPrivate() && !block.IsRecipient() {
		return Progress, nil
	}

	switch block.GetType() {
	case types.ContractSend:
		return checkContractSendBlock(lv, block)
	case types.ContractReward:
		//check smart c exist
		input, err := lv.l.GetStateBlock(block.GetLink())
		if err != nil {
			return errorP(block, GapSource, nil)
		}

		// check private tx
		if input.IsPrivate() && !input.IsRecipient() {
			return Progress, nil
		}

		address := types.Address(input.GetLink())
		if c, ok, err := contract.GetChainContract(address, input.GetPayload()); ok && err == nil {
			clone := block.Clone()
			vmCtx := vmstore.NewVMContextWithBlock(lv.l, block)
			if vmCtx == nil {
				return errorP(block, Other, errors.New("contract: can not get vm context"))
			}
			if g, e := c.DoReceive(vmCtx, clone, input); e == nil {
				if len(g) > 0 {
					amount, err := lv.l.CalculateAmount(block)
					if err != nil {
						return errorP(block, Other, fmt.Errorf("calculate amount error: %s", err))
					}
					if bytes.EqualFold(g[0].Block.Data, block.Data) && g[0].Token == block.Token &&
						g[0].Amount.Compare(amount) == types.BalanceCompEqual && g[0].ToAddress == block.Address {
						return Progress, nil
					} else {
						errorInfo := fmt.Errorf("data from contract, %s, %s, %s, %s, data from block, %s, %s, %s, %s",
							g[0].Block.Data, g[0].Token, g[0].Amount, g[0].ToAddress, block.Data, block.Token, amount, block.Address)
						lv.logger.Error(errorInfo)
						return errorP(block, InvalidData, errorInfo)
					}
				} else {
					return errorP(block, Other, fmt.Errorf("can not generate receive block"))
				}
			} else {
				if address == contractaddress.MintageAddress && e == vmstore.ErrStorageNotFound {
					return errorP(block, GapTokenInfo, nil)
				} else {
					return errorP(block, Other, fmt.Errorf("DoReceive error : %s ", e))
				}
			}
		} else {
			//call vm.Run();
			return errorP(block, Other, fmt.Errorf("can not find chain contract %s", address.String()))
		}
	default:
		return errorP(block, Other, invalidBlockType(block.GetType().String()))
	}
}

func checkContractSendBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	// check smart c exist
	address := types.Address(block.GetLink())

	if !contract.IsChainContract(address) {
		if b, err := lv.l.HasSmartContractBlock(address.ToHash()); !b && err == nil {
			return errorP(block, GapSmartContract, nil)
		}
	}

	// check private tx
	if block.IsPrivate() && !block.IsRecipient() {
		return Progress, nil
	}

	// verify data
	if c, ok, err := contract.GetChainContract(address, block.GetPayload()); ok && err == nil {
		clone := block.Clone()
		vmCtx := vmstore.NewVMContextWithBlock(lv.l, block)
		if vmCtx == nil {
			return errorP(block, Other, errors.New("can not get vm context"))
		}
		d := c.GetDescribe()
		switch d.GetVersion() {
		case contract.SpecVer1:
			if err := c.DoSend(vmCtx, clone); err == nil {
				if bytes.EqualFold(block.Data, clone.Data) {
					return Progress, nil
				} else {
					lv.logger.Errorf("data not equal: %s, %s", block.Data, clone.Data)
					return errorP(block, InvalidData, fmt.Errorf("data not equal: %s, %s", block.Data, clone.Data))
				}
			} else {
				return errorP(block, Other, fmt.Errorf("v1 ProcessSend error, err: %s", err))
			}
		case contract.SpecVer2:
			if gapResult, _, err := c.DoGap(vmCtx, clone); err == nil {
				switch gapResult {
				case common.ContractRewardGapPov:
					return errorP(block, GapPovHeight, nil)
				case common.ContractDPKIGapPublish:
					return errorP(block, GapPublish, nil)
				case common.ContractDoDOrderState:
					return errorP(block, GapDoDSettleState, nil)
				}
			} else {
				return errorP(block, Other, fmt.Errorf("do gap error: %s", err))
			}

			if _, _, err := c.ProcessSend(vmCtx, clone); err == nil {
				if bytes.EqualFold(block.Data, clone.Data) {
					return Progress, nil
				} else {
					lv.logger.Errorf("data not equal: %v, %v", block.Data, clone.Data)
					return errorP(block, InvalidData, fmt.Errorf("data not equal %s, %s", block.Data, clone.Data))
				}
			} else {
				return errorP(block, Other, fmt.Errorf("v2 ProcessSend error,  err: %s", err))
			}
		default:
			return errorP(block, Other, fmt.Errorf("unsupported chain contract version %d", d.GetVersion()))
		}
	} else {
		// call vm.Run();
		return errorP(block, Other, fmt.Errorf("can not find chain contract %s", address.String()))
	}
}

func invalidBlockType(typ string) error {
	return fmt.Errorf("invalid process block type, %s", typ)
}

type blockCheck interface {
	Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error)
}

func errorP(block *types.StateBlock, result ProcessResult, detail error) (ProcessResult, error) {
	switch result {
	case Other:
		return Other, fmt.Errorf("%s[%s],%s", block.GetHash(), block.GetType(), detail)
	case GapPrevious:
		return GapPrevious, fmt.Errorf("%s[%s]: gap previous: %s", block.GetHash(), block.GetType(), block.GetPrevious().String())
	case GapSource:
		return GapSource, fmt.Errorf("%s[%s]: gap source: %s", block.GetHash(), block.GetType(), block.GetLink().String())
	default:
		if detail == nil {
			return result, fmt.Errorf("%s[%s]:%s", block.GetHash(), block.GetType(), result.String())
		} else {
			return result, fmt.Errorf("%s[%s]:%s, %s", block.GetHash(), block.GetType(), result.String(), detail)
		}
	}
}

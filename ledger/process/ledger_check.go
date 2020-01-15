package process

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
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

	lv.logger.Debug("check block ", hash)
	blockExist, err := lv.l.HasStateBlockConfirmed(hash)
	if err != nil {
		return Other, err
	}

	if blockExist {
		return Old, nil
	}

	if block.GetType() == types.ContractSend {
		if types.IsNoSignContractAddress(types.Address(block.GetLink())) {
			return Progress, nil
		}
	}
	if block.GetType() == types.ContractReward {
		linkBlk, err := lv.l.GetStateBlockConfirmed(block.GetLink())
		if err != nil {
			return GapSource, nil
		}
		if types.IsNoSignContractAddress(types.Address(linkBlk.GetLink())) {
			return Progress, nil
		}
	}

	if !block.IsValid() {
		return BadWork, errors.New("bad work")
	}

	signature := block.GetSignature()
	if !address.Verify(hash[:], signature[:]) {
		return BadSignature, errors.New("bad signature")
	}

	return Progress, nil
}

type cacheBlockBaseInfoCheck struct {
}

func (cacheBlockBaseInfoCheck) baseInfo(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	hash := block.GetHash()
	address := block.GetAddress()

	lv.logger.Debug("check block ", hash)
	if err := checkReceiveBlockRepeat(lv, block); err != Progress {
		return err, nil
	}
	blockExist, err := lv.l.HasStateBlock(hash)
	if err != nil {
		return Other, err
	}

	if blockExist {
		return Old, nil
	}

	if block.GetType() == types.ContractSend {
		if types.IsNoSignContractAddress(types.Address(block.GetLink())) {
			return Progress, nil
		}
	}
	if block.GetType() == types.ContractReward {
		linkBlk, err := lv.l.GetStateBlock(block.GetLink())
		if err != nil {
			return GapSource, nil
		}
		if types.IsNoSignContractAddress(types.Address(linkBlk.GetLink())) {
			return Progress, nil
		}
	}

	if !block.IsValid() {
		return BadWork, errors.New("bad work")
	}

	signature := block.GetSignature()
	if !address.Verify(hash[:], signature[:]) {
		return BadSignature, errors.New("bad signature")
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

type syncBlockBaseInfoCheck struct {
}

func (syncBlockBaseInfoCheck) baseInfo(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	hash := block.GetHash()
	address := block.GetAddress()

	lv.logger.Debug("check block ", hash)
	blockExist, err := lv.l.HasStateBlockConfirmed(hash)
	if err != nil {
		return Other, err
	}

	if blockExist {
		return Old, nil
	}

	if block.GetType() == types.ContractSend {
		if block.GetLink() == types.Hash(types.RewardsAddress) {
			return Progress, nil
		}
	}
	if block.GetType() == types.ContractReward {
		//linkBlk, err := lv.l.GetStateBlockConfirmed(block.GetLink())
		//if err != nil {
		//	return GapSource, nil
		//}
		//if linkBlk.GetLink() == types.Hash(types.RewardsAddress) {
		//	return Progress, nil
		//}
		return Progress, nil
	}

	if !block.IsValid() {
		return BadWork, errors.New("bad work")
	}

	signature := block.GetSignature()
	if !address.Verify(hash[:], signature[:]) {
		return BadSignature, errors.New("bad signature")
	}

	return Progress, nil
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
			return GapPrevious, nil
		} else {
			if tm, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken()); err == nil && previous.GetHash() != tm.Header {
				return Fork, nil
			}
		}
		return Progress, nil
	case types.Open:
		//check fork
		if _, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken()); err == nil {
			return Fork, nil
		}
		return Progress, nil
	case types.ContractReward:
		// check previous
		if !block.IsOpen() {
			// check previous
			if previous, err := lv.l.GetStateBlockConfirmed(block.GetPrevious()); err != nil {
				return GapPrevious, nil
			} else {
				//check fork
				if tm, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken()); err == nil && previous.GetHash() != tm.Header {
					return Fork, nil
				}
			}
		} else {
			//check fork
			if _, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken()); err == nil {
				return Fork, nil
			}
		}
		return Progress, nil
	default:
		return Other, invalidBlockType(block.GetType().String())
	}
}

type cacheBlockForkCheck struct {
}

func (cacheBlockForkCheck) fork(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Send, types.Receive, types.Change, types.Online, types.ContractSend:
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			//check fork
			if tm, err := lv.l.GetTokenMeta(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
				return Fork, nil
			}
		}
		return Progress, nil
	case types.Open:
		if _, err := lv.l.GetTokenMeta(block.Address, block.Token); err == nil {
			return Fork, nil
		}
		return Progress, nil
	case types.ContractReward:
		if !block.IsOpen() {
			if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
				return GapPrevious, nil
			} else {
				if tm, err := lv.l.GetTokenMeta(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
					return Fork, nil
				}
			}
		} else {
			if _, err := lv.l.GetTokenMeta(block.Address, block.Token); err == nil {
				return Fork, nil
			}
		}
		return Progress, nil
	default:
		return Other, invalidBlockType(block.GetType().String())
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
			return GapPrevious, nil
		} else {
			//check balance
			if !(previous.Balance.Compare(block.Balance) == types.BalanceCompBigger) {
				return BalanceMismatch, nil
			}
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
				return BalanceMismatch, nil
			}
		}
		return Progress, nil
	case types.ContractSend:
		if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			//check totalBalance
			if previous.TotalBalance().Compare(block.TotalBalance()) == types.BalanceCompSmaller {
				return BalanceMismatch, nil
			}
		}
		return Progress, nil
	case types.Change, types.Online:
		// check previous
		if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			//check balance
			if block.Balance.Compare(previous.Balance) != types.BalanceCompEqual {
				return BalanceMismatch, nil
			}
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
				return BalanceMismatch, nil
			}
		}
		return Progress, nil
	default:
		return Other, invalidBlockType(block.GetType().String())
	}
}

type cacheBlockBalanceCheck struct {
}

func (cacheBlockBalanceCheck) balance(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Send:
		// check previous
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			if !(previous.Balance.Compare(block.Balance) == types.BalanceCompBigger) {
				return BalanceMismatch, nil
			}
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
				return BalanceMismatch, nil
			}
		}
		return Progress, nil
	case types.Change, types.Online:
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			//check balance
			if block.Balance.Compare(previous.Balance) != types.BalanceCompEqual {
				return BalanceMismatch, nil
			}
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
				return BalanceMismatch, nil
			}
		}
		return Progress, nil
	case types.ContractSend:
		// check previous
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			//check totalBalance
			if previous.TotalBalance().Compare(block.TotalBalance()) == types.BalanceCompSmaller {
				return BalanceMismatch, nil
			}
		}
		return Progress, nil
	default:
		return Other, invalidBlockType(block.GetType().String())
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
			return GapPrevious, nil
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
						return BalanceMismatch, nil
					}
					//check vote,network,storage,oracle
					if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
						previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
						previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
						previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
						return BalanceMismatch, nil
					}
				} else {
					return Other, err
				}
			} else if err == ledger.ErrPendingNotFound {
				return UnReceivable, nil
			} else {
				return Other, err
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
				return BalanceMismatch, nil
			}
			//check vote,network,storage,oracle
			vote := block.GetVote()
			network := block.GetNetwork()
			storage := block.GetStorage()
			oracle := block.GetOracle()
			if !vote.IsZero() || !network.IsZero() ||
				!storage.IsZero() || !oracle.IsZero() {
				return BalanceMismatch, nil
			}
		} else if err == ledger.ErrPendingNotFound {
			return UnReceivable, nil
		} else {
			return Other, err
		}
		return Progress, nil
	case types.ContractReward:
		return checkContractPending(lv, block)
	default:
		return Other, invalidBlockType(block.GetType().String())
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
						return BalanceMismatch, nil
					}
					//check vote,network,storage,oracle
					if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
						previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
						previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
						previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
						return BalanceMismatch, nil
					}
				} else {
					return Other, err
				}
			} else if err == ledger.ErrPendingNotFound {
				return UnReceivable, nil
			} else {
				return Other, err
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
				return BalanceMismatch, nil
			}
			//check vote,network,storage,oracle
			vote := block.GetVote()
			network := block.GetNetwork()
			storage := block.GetStorage()
			oracle := block.GetOracle()
			if !vote.IsZero() || !network.IsZero() ||
				!storage.IsZero() || !oracle.IsZero() {
				return BalanceMismatch, nil
			}
		} else if err == ledger.ErrPendingNotFound {
			return UnReceivable, nil
		} else {
			return Other, err
		}
		return Progress, nil
	case types.ContractReward:
		return checkContractPending(lv, block)
	default:
		return Other, invalidBlockType(block.GetType().String())
	}
}

func checkContractPending(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	input, err := lv.l.GetStateBlockConfirmed(block.GetLink())
	if err != nil {
		return GapSource, nil
	}

	pendingKey := types.PendingKey{
		Address: block.GetAddress(),
		Hash:    block.GetLink(),
	}

	// check pending
	if c, ok, err := contract.GetChainContract(types.Address(input.Link), input.Data); ok && err == nil {
		d := c.GetDescribe()
		switch d.GetVersion() {
		case contract.SpecVer1:
			if d.WithPending() {
				if pending, err := lv.l.GetPending(&pendingKey); err == nil {
					if pendingKey, pendingInfo, err := c.DoPending(input); err == nil && pendingKey != nil {
						if pending.Type == pendingInfo.Type && pending.Amount.Equal(pendingInfo.Amount) && pending.Source == pendingInfo.Source {
							return Progress, nil
						} else {
							lv.logger.Errorf("pending from chain, %s, %s, %s, %s, pending from contract, %s, %s, %s, %s",
								pending.Type, pending.Source, pending.Amount, pendingInfo.Type, pendingInfo.Source, pendingInfo.Amount)
							return InvalidData, nil
						}
					}
				} else if err == ledger.ErrPendingNotFound {
					return UnReceivable, nil
				} else {
					return Other, err
				}
			}
		case contract.SpecVer2:
			if d.WithPending() {
				if pending, err := lv.l.GetPending(&pendingKey); err == nil {
					vmCtx := vmstore.NewVMContext(lv.l)
					if pendingKey, pendingInfo, err := c.ProcessSend(vmCtx, input); err == nil && pendingKey != nil {
						if pending.Type == pendingInfo.Type && pending.Amount.Equal(pendingInfo.Amount) && pending.Source == pendingInfo.Source {
							return Progress, nil
						} else {
							lv.logger.Errorf("pending from chain, %s, %s, %s, %s, pending from contract, %s, %s, %s, %s",
								pending.Type, pending.Source, pending.Amount, pendingInfo.Type, pendingInfo.Source, pendingInfo.Amount)
							return InvalidData, nil
						}
					}
				} else if err == ledger.ErrPendingNotFound {
					return UnReceivable, nil
				} else {
					return Other, err
				}
			}
		default:
			return Other, fmt.Errorf("unsupported chain contract version %d", d.GetVersion())
		}
	} else {
		return Other, fmt.Errorf("can not find chain contract %s", input.GetLink().String())
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
			return GapSource, nil
		}
		return Progress, nil
	case types.ContractReward:
		// implement in contract data check
		return Progress, nil
	default:
		return Other, invalidBlockType(block.GetType().String())
	}
}

type cacheBlockSourceCheck struct {
}

func (cacheBlockSourceCheck) source(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.Receive, types.Open:
		if b, err := lv.l.HasStateBlock(block.Link); !b && err == nil {
			return GapSource, nil
		}
		return Progress, nil
	case types.ContractReward:
		// implement in contract data check
		return Progress, nil
	default:
		return Other, invalidBlockType(block.GetType().String())
	}
}

// check block contract
type contractCheck interface {
	contract(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error)
}

type blockContractCheck struct {
}

func (blockContractCheck) contract(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.ContractSend:
		return checkContractSendBlock(lv, block)
	case types.ContractReward:
		// check receive link
		input, err := lv.l.GetStateBlockConfirmed(block.GetLink())
		if err != nil {
			return GapSource, nil
		}

		address := types.Address(input.GetLink())
		if c, ok, err := contract.GetChainContract(address, input.Data); ok && err == nil {
			clone := block.Clone()
			//TODO:verify extra hash and commit to db
			vmCtx := vmstore.NewVMContext(lv.l)
			if g, e := c.DoReceive(vmCtx, clone, input); e == nil {
				if len(g) > 0 {
					amount, err := lv.l.CalculateAmount(block)
					if err != nil {
						lv.logger.Error("calculate amount error:", err)
					}
					if bytes.EqualFold(g[0].Block.Data, block.Data) && g[0].Token == block.Token &&
						g[0].Amount.Compare(amount) == types.BalanceCompEqual && g[0].ToAddress == block.Address {
						return Progress, nil
					} else {
						lv.logger.Errorf("data from contract, %s, %s, %s, %s, data from block, %s, %s, %s, %s",
							g[0].Block.Data, g[0].Token, g[0].Amount, g[0].ToAddress, block.Data, block.Token, amount, block.Address)
						return InvalidData, nil
					}
				} else {
					return Other, fmt.Errorf("can not generate receive block")
				}
			} else {
				if address == types.MintageAddress && e == vmstore.ErrStorageNotFound {
					return GapTokenInfo, nil
				} else {
					lv.logger.Error("DoReceive error ", e)
					return Other, e
				}
			}
		} else {
			//call vm.Run();
			return Other, fmt.Errorf("can not find chain contract %s", address.String())
		}
	default:
		return Other, invalidBlockType(block.GetType().String())
	}
}

type cacheBlockContractCheck struct {
}

func (cacheBlockContractCheck) contract(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	switch block.GetType() {
	case types.ContractSend:
		return checkContractSendBlock(lv, block)
	case types.ContractReward:
		//check smart c exist
		input, err := lv.l.GetStateBlock(block.GetLink())
		if err != nil {
			return GapSource, nil
		}

		address := types.Address(input.GetLink())
		if c, ok, err := contract.GetChainContract(address, input.Data); ok && err == nil {
			clone := block.Clone()
			//TODO:verify extra hash and commit to db
			vmCtx := vmstore.NewVMContext(lv.l)
			if g, e := c.DoReceive(vmCtx, clone, input); e == nil {
				if len(g) > 0 {
					amount, err := lv.l.CalculateAmount(block)
					if err != nil {
						lv.logger.Error("calculate amount error:", err)
					}
					if bytes.EqualFold(g[0].Block.Data, block.Data) && g[0].Token == block.Token &&
						g[0].Amount.Compare(amount) == types.BalanceCompEqual && g[0].ToAddress == block.Address {
						return Progress, nil
					} else {
						lv.logger.Errorf("data from contract, %s, %s, %s, %s, data from block, %s, %s, %s, %s",
							g[0].Block.Data, g[0].Token, g[0].Amount, g[0].ToAddress, block.Data, block.Token, amount, block.Address)
						return InvalidData, nil
					}
				} else {
					return Other, fmt.Errorf("can not generate receive block")
				}
			} else {
				if address == types.MintageAddress && e == vmstore.ErrStorageNotFound {
					return GapTokenInfo, nil
				} else {
					lv.logger.Error("DoReceive error ", e)
					return Other, e
				}
			}
		} else {
			//call vm.Run();
			return Other, fmt.Errorf("can not find chain contract %s", address.String())
		}
	default:
		return Other, invalidBlockType(block.GetType().String())
	}
}

func checkContractSendBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//check smart c exist
	address := types.Address(block.GetLink())

	if !contract.IsChainContract(address) {
		if b, err := lv.l.HasSmartContractBlock(address.ToHash()); !b && err == nil {
			return GapSmartContract, nil
		}
	}

	//verify data
	if c, ok, err := contract.GetChainContract(address, block.Data); ok && err == nil {
		clone := block.Clone()
		vmCtx := vmstore.NewVMContext(lv.l)
		d := c.GetDescribe()
		switch d.GetVersion() {
		case contract.SpecVer1:
			if err := c.DoSend(vmCtx, clone); err == nil {
				if bytes.EqualFold(block.Data, clone.Data) {
					return Progress, nil
				} else {
					lv.logger.Errorf("data not equal: %s, %s", block.Data, clone.Data)
					return InvalidData, nil
				}
			} else {
				lv.logger.Errorf("v1 ProcessSend error, block: %s, err: ", block.GetHash(), err)
				return Other, err
			}
		case contract.SpecVer2:
			if types.IsRewardContractAddress(types.Address(block.GetLink())) {
				h, err := c.DoGapPov(vmCtx, clone)
				if err != nil {
					lv.logger.Errorf("do gapPov error: %s", err)
					return Other, err
				}
				if h != 0 {
					return GapPovHeight, nil
				}
			}

			if types.PubKeyDistributionAddress == types.Address(block.GetLink()) {
				info := new(abi.OracleInfo)
				err := abi.PublicKeyDistributionABI.UnpackMethod(info, abi.MethodNamePKDOracle, block.GetData())
				if err == nil {
					if has, _ := lv.l.HasStateBlockConfirmed(info.Hash); !has {
						return GapPublish, nil
					}
				}
			}

			if _, _, err := c.ProcessSend(vmCtx, clone); err == nil {
				if bytes.EqualFold(block.Data, clone.Data) {
					return Progress, nil
				} else {
					lv.logger.Errorf("data not equal: %v, %v", block.Data, clone.Data)
					return InvalidData, nil
				}
			} else {
				lv.logger.Errorf("v2 ProcessSend error, block: %s, err: ", block.GetHash(), err)
				return Other, err
			}
		default:
			return Other, fmt.Errorf("unsupported chain contract version %d", d.GetVersion())
		}
	} else {
		//call vm.Run();
		return Other, fmt.Errorf("can not find chain contract %s", address.String())
	}
}

func invalidBlockType(typ string) error {
	return fmt.Errorf("invalid process block type, %s", typ)
}

type blockCheck interface {
	Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error)
}

package dpos

import (
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/common/vmcontract/mintage"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type Node struct {
	cfgPath string
	config  *config.Config
	t       *testing.T
	ctx     *context.ChainContext
	ledger  *ledger.Ledger
	cons    *consensus.Consensus
	dps     *DPoS
}

// Init node
// dir is path of binary file of node, if set empty, will be default path
// bootNode can be nil
func InitNode(t *testing.T) (*Node, error) {
	cfg, err := config.DefaultConfig(filepath.Join(config.QlcTestDataDir(), "nodes", uuid.New().String()))
	if err != nil {
		return nil, err
	}
	setDefaultConfig(cfg)

	cfgPath, err := save(cfg)
	if err != nil {
		return nil, err
	}

	return &Node{
		cfgPath: cfgPath,
		config:  cfg,
		t:       t,
	}, nil
}

func setDefaultConfig(cfg *config.Config) {
	cfg.PoV.PovEnabled = true
	cfg.LogLevel = "debug"
}

func InitNodes(count int, t *testing.T) ([]*Node, error) {
	nodes := make([]*Node, 0)

	for i := 0; i < count; i++ {
		node, err := InitNode(t)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)

		node.RunNode(i)
		node.InitLedger(t)
		node.InitStatus()
	}

	return nodes, nil
}

func StopNodes(nodes []*Node) {
	for i, n := range nodes {
		n.t.Logf("node %d stopping ...  ", i)

		n.StopNodeAndRemoveDir()

		n.t.Logf("node %d stopped", i)
	}
}

func save(cfg *config.Config) (string, error) {
	err := util.CreateDirIfNotExist(cfg.DataDir)
	if err != nil {
		return "", err
	}

	s := util.ToIndentString(cfg)
	cfgPath := filepath.Join(cfg.DataDir, config.QlcConfigFile)
	if err := ioutil.WriteFile(cfgPath, []byte(s), 0600); err != nil {
		return "", err
	}
	return cfgPath, nil
}

func calcWork(hash types.Hash) types.Work {
	var work types.Work
	worker, _ := types.NewWorker(work, hash)
	return worker.NewWork()
}

func (n *Node) CommitConfig() error {
	_, err := save(n.config)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) startServices() {
	log.Setup(n.config)
	n.startLedgerService()
	n.startConsensusService()
}

func (n *Node) stopServices() {
	n.stopConsensusService()
	if err := n.ledger.Close(); err != nil {
		n.t.Fatalf("node close ledger error %s", err)
	}
	log.Teardown()
}

func (n *Node) startLedgerService() {
	l := n.ledger

	genesisInfos := config.GenesisInfos()
	if len(genesisInfos) == 0 {
		n.t.Fatal("no genesis info")
	} else if config.ChainToken() == types.ZeroHash || config.GasToken() == types.ZeroHash {
		n.t.Fatal("no chain token info or gas token info")
	} else {
		if c, _ := l.CountStateBlocks(); c != 0 {
			chainHash := config.GenesisBlockHash()
			gasHash := config.GasBlockHash()
			b1, _ := l.HasStateBlockConfirmed(chainHash)
			b2, _ := l.HasStateBlockConfirmed(gasHash)
			if !b1 || !b2 {
				n.t.Fatal("chain token info or gas token info mismatch")
			}
		}
	}

	ctx := vmstore.NewVMContext(l, &contractaddress.MintageAddress)
	for _, v := range genesisInfos {
		mb := v.Mintage
		gb := v.Genesis
		err := ctx.SetStorage(contractaddress.MintageAddress[:], gb.Token[:], gb.Data)
		if err != nil {
			n.t.Fatal(err)
		}

		verifier := process.NewLedgerVerifier(l)
		if b, err := l.HasStateBlock(mb.GetHash()); !b && err == nil {
			if err := l.AddStateBlock(&mb); err != nil {
				n.t.Fatal(err)
			}
		} else {
			if err != nil {
				n.t.Fatal(err)
			}
		}

		if b, err := l.HasStateBlock(gb.GetHash()); !b && err == nil {
			if err := verifier.BlockProcess(&gb); err != nil {
				n.t.Fatal(err)
			}
		} else {
			if err != nil {
				n.t.Fatal(err)
			}
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		n.t.Fatal(err)
	}
}

func (n *Node) startConsensusService() {
	DPoS := NewDPoS(n.ctx.ConfigFile())
	DPoS.localRepAccount.Store(TestAccount.Address(), TestAccount)
	cons := consensus.NewConsensus(DPoS, n.ctx.ConfigFile())
	n.cons = cons
	n.dps = DPoS
	cons.Init()
	cons.Start()
}

func (n *Node) stopConsensusService() {
	n.cons.Stop()
}

// Run a node without account
func (n *Node) RunNode(i int) {
	n.t.Logf("node %d starting ...", i)

	n.ctx = context.NewChainContext(n.cfgPath)
	n.ctx.Init(func() error {
		return nil
	})

	n.ledger = ledger.NewLedger(n.ctx.ConfigFile())
	n.startServices()

	n.t.Logf("node %d started", i)
}

// Stop a node
func (n *Node) StopNodeAndRemoveDir() {
	n.stopServices()

	dir, err := filepath.Abs(filepath.Dir(n.cfgPath))
	if err != nil {
		n.t.Fatal(err)
	}

	os.RemoveAll(dir)
}

func (n *Node) InitStatus() {
	ticker := time.NewTimer(100 * time.Millisecond)
	n.ctx.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)

	for {
		if n.ctx.PoVState() == topic.SyncDone {
			return
		}

		select {
		case <-ticker.C:
			n.t.Fatal("pov sync err")
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func (n *Node) GenerateSendBlock(from *types.Account, to types.Address, amount types.Balance, tokenName string) *types.StateBlock {
	if tokenName == "" {
		tokenName = "QLC"
	}

	info, err := n.ledger.GetTokenByName(tokenName)
	if err != nil {
		n.t.Fatal(err)
	}

	b := new(types.StateBlock)
	b.Address = from.Address()
	b.Link = to.ToHash()
	b.Token = info.TokenId

	block, err := n.ledger.GenerateSendBlock(b, amount, from.PrivateKey())
	if err != nil {
		n.t.Fatal(err)
	}

	return block
}

func (n *Node) GenerateReceiveBlock(sendBlock *types.StateBlock, account *types.Account) *types.StateBlock {
	block, err := n.ledger.GenerateReceiveBlock(sendBlock, account.PrivateKey())
	if err != nil {
		n.t.Fatal(err)
	}
	return block
}

func (n *Node) GenerateChangeBlock(account *types.Account, representative types.Address) *types.StateBlock {
	block, err := n.ledger.GenerateChangeBlock(account.Address(), representative, account.PrivateKey())
	if err != nil {
		n.t.Fatal(err)
	}
	return block
}

func (n *Node) GenerateContractSendBlock(from, to *types.Account, ca types.Address, method string, param interface{}) *types.StateBlock {
	switch ca {
	case contractaddress.MintageAddress:
		if method == mintage.MethodNameMintage {
			totalSupply := big.NewInt(1000)
			decimals := uint8(8)
			tokenName := "testToken"
			tokenSymbol := "testToken"
			NEP5tTxId := random.RandomHexString(32)
			tokenId := mintage.NewTokenHash(from.Address(), param.(types.Hash), tokenName)

			data, err := mintage.MintageABI.PackMethod(mintage.MethodNameMintage, tokenId, tokenName, tokenSymbol, totalSupply, decimals, to.Address(), NEP5tTxId)
			if err != nil {
				n.t.Fatal(err)
			}

			am, err := n.ledger.GetAccountMeta(from.Address())
			if err != nil {
				n.t.Fatal(err)
			}

			tm := am.Token(config.ChainToken())
			if tm == nil {
				n.t.Fatal()
			}

			minPledgeAmount := types.Balance{Int: contract.MinPledgeAmount}
			if tm.Balance.Compare(minPledgeAmount) == types.BalanceCompSmaller {
				n.t.Fatal()
			}

			send := &types.StateBlock{
				Type:           types.ContractSend,
				Token:          tm.Type,
				Address:        from.Address(),
				Balance:        tm.Balance.Sub(minPledgeAmount),
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.MintageAddress),
				Representative: tm.Representative,
				Data:           data,
				PoVHeight:      0,
				Timestamp:      common.TimeNow().Unix(),
			}
			send.Signature = from.Sign(send.GetHash())
			send.Work = calcWork(send.Root())

			return send
		} else if method == mintage.MethodNameMintageWithdraw {
			tm, _ := n.ledger.GetTokenMeta(from.Address(), config.ChainToken())
			if tm == nil {
				n.t.Fatal()
			}
			data, err := mintage.MintageABI.PackMethod(mintage.MethodNameMintageWithdraw, param.(types.Hash))
			if err != nil {
				n.t.Fatal(err)
			}

			send := &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: from.Address(),
				Balance: tm.Balance,
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.MintageAddress),
				Representative: tm.Representative,
				Data:           data,
				PoVHeight:      0,
				Timestamp:      common.TimeNow().Unix(),
			}
			send.Signature = from.Sign(send.GetHash())
			send.Work = calcWork(send.Root())

			return send
		} else {
			n.t.Fatal()
		}
	default:
		n.t.Fatal()
	}

	return nil
}

func (n *Node) GenerateContractReceiveBlock(to *types.Account, ca types.Address, method string, send *types.StateBlock) *types.StateBlock {
	switch ca {
	case contractaddress.MintageAddress:
		if method == mintage.MethodNameMintage {
			recv := &types.StateBlock{}
			mintage := &contract.Mintage{}
			vmContext := vmstore.NewVMContext(n.ledger, &contractaddress.MintageAddress)
			contract.SetMinMintageTime(0, 0, 0, 0, 0, 1)

			blocks, err := mintage.DoReceive(vmContext, recv, send)
			if err != nil {
				n.t.Fatal(err)
			}

			if len(blocks) > 0 {
				recv.Timestamp = common.TimeNow().Unix()
				h := vmstore.TrieHash(blocks[0].VMContext)
				recv.Extra = h
			}

			recv.Signature = to.Sign(recv.GetHash())
			recv.Work = calcWork(recv.Root())

			return recv
		} else if method == mintage.MethodNameMintageWithdraw {
			recv := &types.StateBlock{}
			withdraw := &contract.WithdrawMintage{}
			vmContext := vmstore.NewVMContext(n.ledger, &contractaddress.MintageAddress)
			blocks, err := withdraw.DoReceive(vmContext, recv, send)
			if err != nil {
				n.t.Fatal(err)
			}

			if len(blocks) > 0 {
				recv.Timestamp = common.TimeNow().Unix()
				h := vmstore.TrieHash(blocks[0].VMContext)
				recv.Extra = h
			}

			recv.Signature = to.Sign(recv.GetHash())
			recv.Work = calcWork(recv.Root())

			return recv
		} else {
			n.t.Fatal()
		}
	default:
		n.t.Fatal()
	}

	return nil
}

// process block to node
func (n *Node) ProcessBlock(block *types.StateBlock) {
	l := n.ledger
	eb := n.ctx.EventBus()

	verifier := process.NewLedgerVerifier(l)
	flag, err := verifier.BlockCacheCheck(block)
	if flag == process.Other {
		n.t.Fatal(err)
	}

	if flag == process.Progress {
		hash := block.GetHash()

		err := verifier.BlockCacheProcess(block)
		if err != nil {
			n.t.Fatalf("Block %s add to blockCache error[%s]", hash, err)
		}

		eb.Publish(topic.EventAddBlockCache, block)
		eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.PublishReq, Message: block})
		eb.Publish(topic.EventGenerateBlock, block)
	}
}

func (n *Node) ProcessBlockLocal(block *types.StateBlock) {
	verifier := process.NewLedgerVerifier(n.ledger)
	r, err := verifier.Process(block)
	if err != nil || r != process.Progress {
		n.t.Fatal("process block err", err, r)
	}
}

func (n *Node) GenerateSendBlockAndProcess(from *types.Account, to types.Address, amount types.Balance, tokenName string) *types.StateBlock {
	blk := n.GenerateSendBlock(from, to, amount, tokenName)
	n.ProcessBlock(blk)
	return blk
}

func (n *Node) GenerateReceiveBlockAndProcess(sendBlock *types.StateBlock, acc *types.Account) *types.StateBlock {
	blk := n.GenerateReceiveBlock(sendBlock, acc)
	n.ProcessBlock(blk)
	return blk
}

func (n *Node) GenerateChangeBlockAndProcess(account *types.Account, representative types.Address) *types.StateBlock {
	blk := n.GenerateChangeBlock(account, representative)
	n.ProcessBlock(blk)
	return blk
}

// process block to node , and wait for block consensus confirmed
// if can not confirmed in one minute, return timeout error
func (n *Node) ProcessBlockAndWaitConfirmed(block *types.StateBlock) {
	n.ProcessBlock(block)
	n.WaitBlockConfirmed(block.GetHash())
}

func (n *Node) TokenTransactionAndConfirmed(from, to *types.Account, amount types.Balance, tokenName string) (send, recv *types.StateBlock) {
	sendBlk := n.GenerateSendBlock(from, to.Address(), amount, tokenName)
	n.ProcessBlockAndWaitConfirmed(sendBlk)

	recvBlk := n.GenerateReceiveBlock(sendBlk, to)
	n.ProcessBlockAndWaitConfirmed(recvBlk)

	return sendBlk, recvBlk
}

// Wait for block consensus confirmed
// if can not confirmed in one minute, return timeout error
func (n *Node) WaitBlockConfirmed(hash types.Hash) {
	t := time.NewTimer(time.Second * 30)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			n.t.Fatal("confirm block timeout")
		default:
			if has, _ := n.ledger.HasStateBlockConfirmed(hash); has {
				return
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (n *Node) WaitBlocksConfirmed(hashes []types.Hash) {
	for _, hash := range hashes {
		n.WaitBlockConfirmed(hash)
	}
}

func (n *Node) CheckBlocksConfirmed(hashes []types.Hash) {
	for _, hash := range hashes {
		if has, _ := n.ledger.HasStateBlockConfirmed(hash); !has {
			n.t.Fatal()
		}
	}
}

func (n *Node) InitLedger(t *testing.T) {
	n.ProcessBlockLocal(&mock.TestSendBlock)
	n.ProcessBlockLocal(&mock.TestReceiveBlock)
	n.ProcessBlockLocal(&mock.TestSendGasBlock)
	n.ProcessBlockLocal(&mock.TestReceiveGasBlock)
	n.ProcessBlockLocal(&mock.TestChangeRepresentative)

	block, td := mock.GeneratePovBlock(nil, 0)
	block.Header.BasHdr.Height = 0
	if err := n.ledger.AddPovBlock(block, td); err != nil {
		t.Fatal(err)
	}
	if err := n.ledger.AddPovBestHash(block.GetHeight(), block.GetHash()); err != nil {
		t.Fatal(err)
	}
	if err := n.ledger.SetPovLatestHeight(block.GetHeight()); err != nil {
		t.Fatal(err)
	}
}

func (n *Node) VoteBlock(acc *types.Account, blk *types.StateBlock) {
	index := n.dps.getProcessorIndex(blk.Address)
	vi := &voteInfo{
		hash:    blk.GetHash(),
		account: acc.Address(),
	}
	n.dps.processors[index].acks <- vi
}

func (n *Node) TestWithTimeout(err string, to time.Duration, fn func() bool) {
	ti := time.NewTimer(to)
	defer ti.Stop()
	for {
		select {
		case <-ti.C:
			n.t.Fatal(err)
		default:
			if fn() {
				return
			}
			time.Sleep(time.Second)
		}
	}
}

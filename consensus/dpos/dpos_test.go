package dpos

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/consensus/pov"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/miner"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type Node struct {
	cfgPath   string
	config    *config.Config
	account   *types.Account
	t         *testing.T
	isBoot    bool
	ctx       *context.ChainContext
	ledger    *ledger.Ledger
	povEngine *pov.PoVEngine
	miner     *miner.Miner
	cons      *consensus.Consensus
	p2p       *p2p.QlcService
	dps       *DPoS
}

func InitBootNode(t *testing.T) (*Node, error) {
	cfg, err := config.DefaultConfig(filepath.Join(config.DefaultDataDir(), uuid.New().String()))
	if err != nil {
		return nil, err
	}
	port := generateRangeNum(10000, 10999)
	setDefaultConfig(cfg, port)

	cfgPath, err := save(cfg)
	return &Node{
		cfgPath: cfgPath,
		config:  cfg,
		t:       t,
		isBoot:  true,
	}, nil
}

// Init node
// dir is path of binary file of node, if set empty, will be default path
// bootNode can be nil
func InitNode(bootNode *Node, t *testing.T) (*Node, error) {
	cfg, err := config.DefaultConfig(filepath.Join(config.DefaultDataDir(), uuid.New().String()))
	if err != nil {
		return nil, err
	}
	port := generateRangeNum(10999, 39999)
	setDefaultConfig(cfg, port)

	if bootNode != nil {
		cfg.P2P.BootNodes = []string{fmt.Sprintf("%s/ipfs/%s", bootNode.config.P2P.Listen, bootNode.config.P2P.ID.PeerID)}
		if len(bootNode.config.P2P.BootNodes) == 0 {
			bootNode.config.P2P.BootNodes = []string{fmt.Sprintf("%s/ipfs/%s", cfg.P2P.Listen, cfg.P2P.ID.PeerID)}
			if _, err = save(bootNode.config); err != nil {
				return nil, err
			}
		}
	}

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

func setDefaultConfig(cfg *config.Config, port int) {
	cfg.P2P.Listen = fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", strconv.Itoa(port))
	cfg.P2P.BootNodes = []string{}
	cfg.RPC.HTTPEndpoint = fmt.Sprintf("tcp4://0.0.0.0:%s", strconv.Itoa(port+1))
	cfg.RPC.WSEndpoint = fmt.Sprintf("tcp4://0.0.0.0:%s", strconv.Itoa(port+2))
	cfg.RPC.IPCEnabled = false
	cfg.P2P.SyncInterval = 12000
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.PoV.PovEnabled = true
	cfg.LogLevel = "error"
}

func InitNodes(count int, t *testing.T) ([]*Node, error) {
	nodes := make([]*Node, 0)

	bootNode, err := InitBootNode(t)
	if err != nil {
		return nil, err
	}

	nodes = append(nodes, bootNode)
	for i := 0; i < count-1; i++ {
		node, err := InitNode(bootNode, t)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func generateRangeNum(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(max-min) + min
	return randNum
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
	n.startPoVService()
	n.startP2PService()
	n.startConsensusService()
}

func (n *Node) stopServices() {
	n.stopP2PService()
	n.stopConsensusService()
	n.stopPoVService()
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

	ctx := vmstore.NewVMContext(l)
	for _, v := range genesisInfos {
		mb := v.Mintage
		gb := v.Genesis
		err := ctx.SetStorage(types.MintageAddress[:], gb.Token[:], gb.Data)
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
	_ = ctx.SaveStorage()
}

func (n *Node) startP2PService() {
	ps, err := p2p.NewQlcService(n.ctx.ConfigFile())
	if err != nil {
		n.t.Fatal(err)
	}
	n.p2p = ps

	err = ps.Start()
	if err != nil {
		n.t.Fatal(err)
	}
}

func (n *Node) stopP2PService() {
	n.p2p.Stop()
}

func (n *Node) startConsensusService() {
	DPoS := NewDPoS(n.ctx.ConfigFile())
	cons := consensus.NewConsensus(DPoS, n.ctx.ConfigFile())
	n.cons = cons
	n.dps = DPoS
	cons.Init()
	cons.Start()
}

func (n *Node) stopConsensusService() {
	n.cons.Stop()
}

func (n *Node) startPoVService() {
	pov.CheckPeerStatusTime = time.Second
	pov.ForceSyncTimeInSec = time.Second

	povEngine, _ := pov.NewPovEngine(n.ctx.ConfigFile(), false)
	err := povEngine.Init()
	if err != nil {
		n.t.Fatal(err)
	}
	n.povEngine = povEngine

	er := povEngine.Start()
	if er != nil {
		n.t.Fatal(er)
	}

	m := miner.NewMiner(n.ctx.ConfigFile(), povEngine)
	n.miner = m
	m.Init()
	m.Start()
}

func (n *Node) stopPoVService() {
	er := n.povEngine.Stop()
	if er != nil {
		n.t.Fatal(er)
	}

	n.miner.Stop()
}

// Run a node without account
func (n *Node) RunNode() {
	n.t.Log("node starting ... ", n.config.RPC.HTTPEndpoint)

	n.ctx = context.NewChainContext(n.cfgPath)
	n.ctx.Init(func() error {
		return nil
	})

	// save accounts to context
	var accounts []*types.Account
	accounts = append(accounts, mock.TestAccount)

	if n.isBoot {
		n.ctx.SetAccounts(accounts)
	}

	n.ledger = ledger.NewLedger(n.ctx.ConfigFile())
	n.startServices()

	n.t.Log("node started ", n.config.RPC.HTTPEndpoint)
}

// Stop a node
func (n *Node) StopNodeAndRemoveDir() {
	n.t.Log("node stopping ...  ", n.config.RPC.HTTPEndpoint)

	n.stopServices()

	dir, err := filepath.Abs(filepath.Dir(n.cfgPath))
	if err != nil {
		n.t.Fatal(err)
	}

	if err := os.RemoveAll(dir); err != nil {
		n.t.Fatal(err)
	}

	n.t.Log("node stopped ", n.config.RPC.HTTPEndpoint)
}

func (n *Node) InitStatus() {
	ticker := time.NewTicker(300 * time.Second)
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

	vmContext := vmstore.NewVMContext(n.ledger)
	info, err := abi.GetTokenByName(vmContext, tokenName)
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
	case types.MintageAddress:
		if method == abi.MethodNameMintage {
			totalSupply := big.NewInt(1000)
			decimals := uint8(8)
			tokenName := "testToken"
			tokenSymbol := "testToken"
			NEP5tTxId := random.RandomHexString(32)
			tokenId := abi.NewTokenHash(from.Address(), param.(types.Hash), tokenName)

			data, err := abi.MintageABI.PackMethod(abi.MethodNameMintage, tokenId, tokenName, tokenSymbol, totalSupply, decimals, to.Address(), NEP5tTxId)
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
				Link:           types.Hash(types.MintageAddress),
				Representative: tm.Representative,
				Data:           data,
				PoVHeight:      0,
				Timestamp:      common.TimeNow().Unix(),
			}
			send.Signature = from.Sign(send.GetHash())
			send.Work = calcWork(send.Root())

			return send
		} else if method == abi.MethodNameMintageWithdraw {
			tm, _ := n.ledger.GetTokenMeta(from.Address(), config.ChainToken())
			if tm == nil {
				n.t.Fatal()
			}
			data, err := abi.MintageABI.PackMethod(abi.MethodNameMintageWithdraw, param.(types.Hash))
			if err != nil {
				n.t.Fatal(err)
			}

			send := &types.StateBlock{
				Type:           types.ContractSend,
				Token:          tm.Type,
				Address:        from.Address(),
				Balance:        tm.Balance,
				Vote:           types.ZeroBalance,
				Network:        types.ZeroBalance,
				Storage:        types.ZeroBalance,
				Oracle:         types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(types.MintageAddress),
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
	case types.MintageAddress:
		if method == abi.MethodNameMintage {
			recv := &types.StateBlock{}
			mintage := &contract.Mintage{}
			vmContext := vmstore.NewVMContext(n.ledger)

			blocks, err := mintage.DoReceive(vmContext, recv, send)
			if err != nil {
				n.t.Fatal(err)
			}

			if len(blocks) > 0 {
				recv.Timestamp = common.TimeNow().Unix()
				h := blocks[0].VMContext.Cache.Trie().Hash()
				recv.Extra = *h
			}

			recv.Signature = to.Sign(recv.GetHash())
			recv.Work = calcWork(recv.Root())

			return recv
		} else if method == abi.MethodNameMintageWithdraw {
			recv := &types.StateBlock{}
			withdraw := &contract.WithdrawMintage{}
			vmContext := vmstore.NewVMContext(n.ledger)
			blocks, err := withdraw.DoReceive(vmContext, recv, send)
			if err != nil {
				n.t.Fatal(err)
			}

			if len(blocks) > 0 {
				recv.Timestamp = common.TimeNow().Unix()
				h := blocks[0].VMContext.Cache.Trie().Hash()
				recv.Extra = *h
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
	if err != nil {
		n.t.Fatal(err)
	}

	n.t.Log("process result, ", flag)
	switch flag {
	case process.Progress:
		hash := block.GetHash()

		err := verifier.BlockCacheProcess(block)
		if err != nil {
			n.t.Fatalf("Block %s add to blockCache error[%s]", hash, err)
		}

		eb.Publish(topic.EventAddBlockCache, block)
		n.t.Log("broadcast block")

		//TODO: refine
		eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.PublishReq, Message: block})
		eb.Publish(topic.EventGenerateBlock, block)
		return
	case process.BadWork:
		n.t.Fatalf("bad work")
	case process.BadSignature:
		n.t.Fatalf("bad signature")
	case process.Old:
		n.t.Fatalf("old block")
	case process.Fork:
		n.t.Fatalf("fork")
	case process.GapSource:
		n.t.Fatalf("gap source block")
	case process.GapPrevious:
		n.t.Fatalf("gap previous block")
	case process.BalanceMismatch:
		n.t.Fatalf("balance mismatch")
	case process.UnReceivable:
		n.t.Fatalf("unReceivable")
	case process.GapSmartContract:
		n.t.Fatalf("gap SmartContract")
	case process.InvalidData:
		n.t.Fatalf("invalid data")
	case process.ReceiveRepeated:
		n.t.Fatalf("generate receive block repeatedly ")
	default:
		n.t.Fatalf("error processing block")
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
	t := time.NewTimer(time.Second * 180)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			n.t.Fatal("confirm block timeout")
		default:
			if has, _ := n.ledger.HasStateBlockConfirmed(hash); has {
				return
			} else {
				time.Sleep(1 * time.Second)
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

func (n *Node) InitLedger() {
	n.ProcessBlockLocal(&mock.TestSendBlock)
	n.ProcessBlockLocal(&mock.TestReceiveBlock)
	n.ProcessBlockLocal(&mock.TestSendGasBlock)
	n.ProcessBlockLocal(&mock.TestReceiveGasBlock)
	n.ProcessBlockLocal(&mock.TestChangeRepresentative)
}

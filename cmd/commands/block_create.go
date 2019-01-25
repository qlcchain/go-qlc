package commands

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/test/mock"
	"github.com/spf13/cobra"
	cmn "github.com/tendermint/tmlibs/common"
)

var (
	blockcount int
	seedbegin  int
	seedend    int
)

var bctCmd = &cobra.Command{
	Use:   "bct",
	Short: "block create",
	Run: func(cmd *cobra.Command, args []string) {
		err := createBlock()
		if err != nil {
			cmd.Println(err)
		} else {
			//cmd.Printf("block create")
		}
	},
}

func init() {

	bctCmd.Flags().IntVarP(&blockcount, "blockcount", "b", 100, "maxinum block count")
	bctCmd.Flags().IntVarP(&seedbegin, "seedbegin", "s", 0, "seed for send account begin index")
	bctCmd.Flags().IntVarP(&seedend, "seedend", "e", 10, "seed for send account end index")

	rootCmd.AddCommand(bctCmd)
}

func createBlock() error {

	if cfgPath == "" {
		cfgPath = config.DefaultDataDir()
	}
	cm := config.NewCfgManager(cfgPath)
	cfg, err := cm.Load()
	if err != nil {
		return err
	}

	err = initNode(types.ZeroAddress, "", cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	services, err = startNode()
	if err != nil {
		fmt.Println(err)
	}

	seeds, err := getSeeds()
	if err != nil {
		return err
	}

	token := mock.GetChainTokenType()
	c := make(chan types.Block)
	processSendBlock(c)
	processReceiveBlock(seeds)

	sendAccs := seeds[seedbegin-1 : seedend]
	receivesAccs := append(seeds[0:seedbegin-1], seeds[seedend:len(seeds)]...)
	err = BatchCreateSendBlock(sendAccs, receivesAccs, blockcount, token, c)
	if err != nil {
		return err
	}
	cmn.TrapSignal(func() {
		stopNode(services)
	})
	return nil
}

func BatchCreateSendBlock(froms, tos []*SeedData, count int, token types.Hash, c chan types.Block) error {
	fromCount := len(froms) - 1
	toCount := len(tos) - 1
	for i := 0; i < count; i++ {
		f, _ := random.Intn(fromCount)
		from := froms[f]
		t, _ := random.Intn(toCount)
		to := tos[t]
		b, _ := random.Intn(1000)
		balance := types.Balance{Int: big.NewInt(int64(b + 1))}
		block, err := ctx.Ledger.Ledger.GenerateSendBlock(from.address, token, to.address, balance, from.privateKey)
		if err != nil {
			return err
		}
		c <- block
	}
	return nil
}

func processSendBlock(c chan types.Block) {
	go func() {
		for {
			block := <-c
			fmt.Println("process send block")
			//err := ctx.Ledger.Ledger.BlockProcess(block)
			//fmt.Println(err)
			//continue
			flag, err := ctx.Ledger.Ledger.Process(block)
			if err != nil {
				fmt.Println(err)
				return
			}

			switch flag {
			case ledger.Progress:
				pushBlock := protos.PublishBlock{
					Blk: block,
				}
				bytes, err := protos.PublishBlockToProto(&pushBlock)
				if err != nil {
					fmt.Println(err)
					return
				} else {
					fmt.Println("broadcast block")
					ctx.DPosService.GetP2PService().Broadcast(p2p.PublishReq, bytes)
				}
			default:
				fmt.Println(flag)
			}

		}
	}()
}

func processReceiveBlock(seeds []*SeedData) {
	_ = ctx.NetService.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmedBlock, func(v interface{}) {
		if b, ok := v.(*types.StateBlock); ok {
			if b.Address.ToHash() != b.Link {
				receAddr := types.Address(b.Link)
				receSeed := findAccount(seeds, receAddr)
				if receSeed != nil {
					balance, _ := mock.RawToBalance(b.Balance, "QLC")
					fmt.Printf("receive block from [%s] to[%s] amount[%d]\n", b.Address.String(), receSeed.address, balance)
					err := receiveblock(b, receSeed.address, receSeed.privateKey)
					if err != nil {
						fmt.Printf("err[%s] when generate receive block.\n", err)
					}
				}
			}
		}
	})
}

type SeedData struct {
	address    types.Address
	privateKey ed25519.PrivateKey
	seed       types.Seed
}

func getSeeds() ([]*SeedData, error) {
	dir := filepath.Join(config.DefaultDataDir(), "seed.txt")
	fmt.Println(dir)
	_, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	seeds := make([]*SeedData, 0)
	//acs := make([]*types.Account, 0)
	if r, err := ioutil.ReadAll(f); err == nil {
		seedStrs := strings.Split(string(r), "\n")
		for _, s := range seedStrs {
			r, _ := regexp.Compile(`PrivateKey:\s*([0-9a-f]*)`)
			match := r.FindStringSubmatch(s)
			if len(match) == 0 {
				continue
			}
			prkStr := match[1]

			r, _ = regexp.Compile(`Seed:\s*(\S*)\s*,`)
			match = r.FindStringSubmatch(s)
			if len(match) == 0 {
				continue
			}
			seedStr := match[1]

			r, _ = regexp.Compile(`Address:\s*(\S*)\s*,`)
			match = r.FindStringSubmatch(s)
			if len(match) == 0 {
				continue
			}
			addrStr := match[1]

			addr, err := types.HexToAddress(addrStr)
			if err != nil {
				return nil, err
			}
			prk, err := hex.DecodeString(prkStr)
			if err != nil {
				return nil, err
			}
			sbyte, err := hex.DecodeString(seedStr)
			if err != nil {
				return nil, err
			}
			seed, err := types.BytesToSeed(sbyte)
			if err != nil {
				return nil, err
			}
			s := SeedData{addr, prk, *seed}
			seeds = append(seeds, &s)
		}
	}
	return seeds, nil
}

func findAccount(seeds []*SeedData, addr types.Address) *SeedData {
	for _, ac := range seeds {
		if ac.address == addr {
			return ac
		}
	}
	return nil
}

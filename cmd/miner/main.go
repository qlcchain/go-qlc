package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
	"log"
	"time"
)

var flagNodeUrl string
var flagPriKey string
var flagSeed string
var flagAlgo string

func main() {
	flag.StringVar(&flagNodeUrl, "nodeurl", "http://127.0.0.1:9735", "RPC URL of node")
	flag.StringVar(&flagPriKey, "privatekey", "", "private key of miner account")
	flag.StringVar(&flagSeed, "seed", "", "seed of miner account")
	flag.StringVar(&flagAlgo, "algo", "SHA256D", "algo name, such as SHA256D/X11/SCRYPT")
	flag.Parse()

	var minerAccount *types.Account

	if len(flagPriKey) > 0 {
		pkBytes, err := hex.DecodeString(flagPriKey)
		if err != nil {
			log.Printf("decode private key, err %s", err)
			return
		}
		if len(pkBytes) != ed25519.PrivateKeySize {
			log.Printf("invalid private key size %d", len(pkBytes))
			return
		}

		minerAccount = types.NewAccount(pkBytes)
		if minerAccount == nil {
			log.Println("failed to new account")
			return
		}
	}

	if len(flagSeed) > 0 {
		seedBytes, err := hex.DecodeString(flagSeed)
		if err != nil {
			log.Printf("decode seed, err %s", err)
			return
		}

		minerSeed, err := types.BytesToSeed(seedBytes)
		if err != nil {
			log.Printf("failed to new seed, err %s", err)
			return
		}

		minerAccount, err = minerSeed.Account(0)
		if minerAccount == nil {
			log.Println("failed to new account")
			return
		}
	}

	if minerAccount == nil {
		log.Printf("miner account not exist")
		return
	}

	nodeClient, err := rpc.Dial(flagNodeUrl)
	if err != nil {
		log.Println(err)
		return
	}
	defer nodeClient.Close()

	log.Printf("running miner, account:%s, algo:%s", minerAccount.Address(), flagAlgo)

	for {
		getWorkRsp := new(api.PovApiGetWork)
		err = nodeClient.Call(&getWorkRsp, "pov_getWork", minerAccount.Address(), flagAlgo)
		if err != nil {
			log.Println(err)
			time.Sleep(10 * time.Second)
			continue
		}
		log.Printf("getWork response: %s", util.ToString(getWorkRsp))

		submitWorkReq := doWork(nodeClient, minerAccount, getWorkRsp)
		if submitWorkReq == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("submitWork request: %s", util.ToString(submitWorkReq))
		err = nodeClient.Call(nil, "pov_submitWork", &submitWorkReq)
		if err != nil {
			log.Println(err)
			time.Sleep(10 * time.Second)
			continue
		}
	}
}

func doWork(nodeClient *rpc.Client, minerAccount *types.Account, getWorkRsp *api.PovApiGetWork) *api.PovApiSubmitWork {
	povHeader := new(types.PovHeader)
	povHeader.BasHdr.Version = getWorkRsp.Version
	povHeader.BasHdr.Previous = getWorkRsp.Previous
	povHeader.BasHdr.Timestamp = uint32(time.Now().Unix())
	povHeader.BasHdr.Bits = getWorkRsp.Bits

	cbTxExtBuf := new(bytes.Buffer)
	cbTxExtBuf.Write(util.LE_Uint64ToBytes(getWorkRsp.Height))
	cbTxExtBuf.WriteString("QLC CPU Miner")

	cbTxDataBuf := new(bytes.Buffer)
	cbTxDataBuf.Write(getWorkRsp.CoinBaseData)
	cbTxDataBuf.Write(cbTxExtBuf.Bytes())
	cbTxHash := types.Sha256D_HashData(cbTxDataBuf.Bytes())

	povHeader.BasHdr.MerkleRoot = merkle.CalcCoinbaseMerkleRoot(&cbTxHash, getWorkRsp.MerkleBranch)

	targetIntAlgo := types.CompactToBig(getWorkRsp.AlgoBits)

	lastCheckTm := time.Now()

	for nonce := uint32(0); nonce < common.PovMaxNonce; nonce++ {
		povHeader.BasHdr.Nonce = nonce

		powHash := povHeader.ComputePowHash()
		powInt := powHash.ToBigInt()
		if powInt.Cmp(targetIntAlgo) <= 0 {
			log.Printf("workHash %s found nonce %d", getWorkRsp.WorkHash, nonce)
			submitWorkReq := new(api.PovApiSubmitWork)
			submitWorkReq.WorkHash = getWorkRsp.WorkHash

			submitWorkReq.CoinbaseExtra = cbTxExtBuf.Bytes()
			submitWorkReq.CoinbaseHash = cbTxHash
			submitWorkReq.CoinbaseSig = minerAccount.Sign(cbTxHash)

			submitWorkReq.MerkleRoot = povHeader.BasHdr.MerkleRoot
			submitWorkReq.Timestamp = povHeader.BasHdr.Timestamp
			submitWorkReq.Nonce = povHeader.BasHdr.Nonce

			submitWorkReq.BlockHash = povHeader.ComputeHash()
			return submitWorkReq
		}

		if time.Now().After(lastCheckTm.Add(10 * time.Second)) {
			latestHeader := getLatestHeader(nodeClient)
			if latestHeader != nil && latestHeader.GetHash() != getWorkRsp.Previous {
				log.Printf("workHash %s abort search nonce because latest block change", getWorkRsp.WorkHash)
				return nil
			}
			lastCheckTm = time.Now()
		}
	}

	log.Printf("workHash %s exhaust nonce", getWorkRsp.WorkHash)
	return nil
}

func getLatestHeader(nodeClient *rpc.Client) *api.PovApiHeader {
	latestHeaderRsp := new(api.PovApiHeader)
	err := nodeClient.Call(latestHeaderRsp, "pov_getLatestHeader")
	if err != nil {
		log.Println(err)
		return nil
	}
	return latestHeaderRsp
}

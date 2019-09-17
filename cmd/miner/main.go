package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
	"log"
	"time"
)

var flagNodeUrl string
var flagMiner string
var flagAlgo string
var flagAuxPow bool

func main() {
	flag.StringVar(&flagNodeUrl, "nodeurl", "http://127.0.0.1:9735", "RPC URL of node")
	flag.StringVar(&flagMiner, "miner", "", "address of miner account")
	flag.StringVar(&flagAlgo, "algo", "SHA256D", "algo name, such as SHA256D/X11/SCRYPT")
	flag.BoolVar(&flagAuxPow, "auxpow", false, "enable AuxPOW")
	flag.Parse()

	minerAddr, err := types.HexToAddress(flagMiner)
	if err != nil {
		log.Printf("invalid miner address")
		return
	}

	nodeClient, err := rpc.Dial(flagNodeUrl)
	if err != nil {
		log.Println(err)
		return
	}
	defer nodeClient.Close()

	log.Printf("running miner, account:%s, algo:%s", minerAddr, flagAlgo)

	for {
		getWorkRsp := new(api.PovApiGetWork)
		err = nodeClient.Call(&getWorkRsp, "pov_getWork", minerAddr, flagAlgo)
		if err != nil {
			log.Println(err)
			time.Sleep(10 * time.Second)
			continue
		}
		log.Printf("getWork response: %s", util.ToString(getWorkRsp))

		var submitWorkReq *api.PovApiSubmitWork
		if flagAuxPow {
			submitWorkReq = doWorkByAuxPow(nodeClient, minerAddr, getWorkRsp)
		} else {
			submitWorkReq = doWorkBySelf(nodeClient, minerAddr, getWorkRsp)
		}
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

func doWorkBySelf(nodeClient *rpc.Client, minerAddr types.Address, getWorkRsp *api.PovApiGetWork) *api.PovApiSubmitWork {
	povHeader := new(types.PovHeader)
	povHeader.BasHdr.Version = getWorkRsp.Version
	povHeader.BasHdr.Previous = getWorkRsp.Previous
	povHeader.BasHdr.Timestamp = uint32(time.Now().Unix())
	povHeader.BasHdr.Bits = getWorkRsp.Bits

	cbTxExtBuf := new(bytes.Buffer)
	cbTxExtBuf.Write(util.LE_Uint64ToBytes(getWorkRsp.Height))
	cbTxExtBuf.WriteString("/QLC CPU Miner/")
	cbTxExtData := cbTxExtBuf.Bytes()

	// hash data = coinbase1 + extra data + coinbase2
	// extra data = minerinfo + extranonce1 + extranonce2
	cbTxDataBuf := new(bytes.Buffer)
	cbData1, _ := hex.DecodeString(getWorkRsp.CoinBaseData1)
	cbTxDataBuf.Write(cbData1)

	cbTxDataBuf.Write(util.LE_EncodeVarInt(uint64(len(cbTxExtData))))
	cbTxDataBuf.Write(cbTxExtData)

	cbData2, _ := hex.DecodeString(getWorkRsp.CoinBaseData2)
	cbTxDataBuf.Write(cbData2)

	cbTxHash := types.Sha256D_HashData(cbTxDataBuf.Bytes())

	povHeader.BasHdr.MerkleRoot = merkle.CalcCoinbaseMerkleRoot(&cbTxHash, getWorkRsp.MerkleBranch)

	targetIntAlgo := types.CompactToBig(getWorkRsp.Bits)

	lastCheckTm := time.Now()

	for nonce := uint32(0); nonce < common.PovMaxNonce; nonce++ {
		povHeader.BasHdr.Nonce = nonce

		powHash := povHeader.ComputePowHash()
		powInt := powHash.ToBigInt()
		if powInt.Cmp(targetIntAlgo) <= 0 {
			log.Printf("workHash %s found nonce %d", getWorkRsp.WorkHash, nonce)
			submitWorkReq := new(api.PovApiSubmitWork)
			submitWorkReq.WorkHash = getWorkRsp.WorkHash

			submitWorkReq.CoinbaseExtra = hex.EncodeToString(cbTxExtBuf.Bytes())
			submitWorkReq.CoinbaseHash = cbTxHash

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

func doWorkByAuxPow(nodeClient *rpc.Client, minerAddr types.Address, getWorkRsp *api.PovApiGetWork) *api.PovApiSubmitWork {
	povHeader := new(types.PovHeader)
	povHeader.BasHdr.Version = getWorkRsp.Version
	povHeader.BasHdr.Previous = getWorkRsp.Previous
	povHeader.BasHdr.Timestamp = uint32(time.Now().Unix())
	povHeader.BasHdr.Bits = getWorkRsp.Bits

	cbTxExtBuf := new(bytes.Buffer)
	cbTxExtBuf.Write(util.LE_Uint64ToBytes(getWorkRsp.Height))
	cbTxExtBuf.WriteString("/QLC CPU AuxPOW/")
	cbTxExtData := cbTxExtBuf.Bytes()

	// hash data = coinbase1 + extra data + coinbase2
	// extra data = minerinfo + extranonce1 + extranonce2
	cbTxDataBuf := new(bytes.Buffer)
	cbData1, _ := hex.DecodeString(getWorkRsp.CoinBaseData1)
	cbTxDataBuf.Write(cbData1)

	cbTxDataBuf.Write(util.LE_EncodeVarInt(uint64(len(cbTxExtData))))
	cbTxDataBuf.Write(cbTxExtData)

	cbData2, _ := hex.DecodeString(getWorkRsp.CoinBaseData2)
	cbTxDataBuf.Write(cbData2)

	cbTxHash := types.Sha256D_HashData(cbTxDataBuf.Bytes())

	povHeader.BasHdr.MerkleRoot = merkle.CalcCoinbaseMerkleRoot(&cbTxHash, getWorkRsp.MerkleBranch)

	targetIntAlgo := types.CompactToBig(getWorkRsp.Bits)

	auxBlockHash := povHeader.ComputeHash()
	auxPow := GenerateAuxPow(auxBlockHash)

	lastCheckTm := time.Now()

	for nonce := uint32(0); nonce < common.PovMaxNonce; nonce++ {
		auxPow.ParBlockHeader.Nonce = nonce

		powHash := auxPow.ComputePowHash(povHeader.GetAlgoType())
		powInt := powHash.ToBigInt()
		if powInt.Cmp(targetIntAlgo) <= 0 {
			log.Printf("workHash %s found nonce %d", getWorkRsp.WorkHash, nonce)
			submitWorkReq := new(api.PovApiSubmitWork)
			submitWorkReq.WorkHash = getWorkRsp.WorkHash

			submitWorkReq.CoinbaseExtra = hex.EncodeToString(cbTxExtBuf.Bytes())
			submitWorkReq.CoinbaseHash = cbTxHash

			submitWorkReq.MerkleRoot = povHeader.BasHdr.MerkleRoot
			submitWorkReq.Timestamp = povHeader.BasHdr.Timestamp
			submitWorkReq.Nonce = povHeader.BasHdr.Nonce

			submitWorkReq.BlockHash = povHeader.ComputeHash()

			auxPow.ParentHash = auxPow.ParBlockHeader.ComputeHash()
			submitWorkReq.AuxPow = auxPow
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

func getBtcCoinbase(msgBlockHash types.Hash) *types.PovBtcTx {
	var magic [4]byte           // 4 byte
	var auxBlockHash types.Hash // 32 byte
	var merkleSize int32        // 4 byte
	var merkleNonce int32       // 4 byte

	magic = [4]byte{0xfa, 0xbe, 'm', 'm'}
	auxBlockHash = msgBlockHash
	merkleSize = 1
	merkleNonce = 0

	scriptSig := make([]byte, 0, 44) // 44 byte
	scriptSigBuf := bytes.NewBuffer(scriptSig)
	binary.Write(scriptSigBuf, binary.LittleEndian, magic)
	binary.Write(scriptSigBuf, binary.LittleEndian, auxBlockHash)
	binary.Write(scriptSigBuf, binary.LittleEndian, merkleSize)
	binary.Write(scriptSigBuf, binary.LittleEndian, merkleNonce)

	coinBaseTxin := types.PovBtcTxIn{
		PreviousOutPoint: types.PovBtcOutPoint{
			Hash:  types.ZeroHash,
			Index: uint32(0),
		},
		SignatureScript: scriptSigBuf.Bytes(),
		Sequence:        uint32(0),
	}

	btcTxin := make([]*types.PovBtcTxIn, 0)
	btcTxin = append(btcTxin, &coinBaseTxin)
	btcTxout := make([]*types.PovBtcTxOut, 0)

	coinbase := types.NewPovBtcTx(btcTxin, btcTxout)

	return coinbase
}

func GenerateAuxPow(msgBlockHash types.Hash) *types.PovAuxHeader {
	auxMerkleBranch := make([]*types.Hash, 0)
	auxMerkleIndex := 0
	parCoinbaseTx := getBtcCoinbase(msgBlockHash)
	parCoinBaseMerkle := make([]*types.Hash, 0)
	parMerkleIndex := 0
	parBlockHeader := types.PovBtcHeader{
		Version:    0x7fffffff,
		Previous:   types.ZeroHash,
		MerkleRoot: parCoinbaseTx.ComputeHash(),
		Timestamp:  uint32(time.Now().Unix()),
		Bits:       0, // do not care about parent block diff
		Nonce:      0, // to be solved
	}
	auxPow := &types.PovAuxHeader{
		AuxMerkleBranch:   auxMerkleBranch,
		AuxMerkleIndex:    auxMerkleIndex,
		ParCoinBaseTx:     *parCoinbaseTx,
		ParCoinBaseMerkle: parCoinBaseMerkle,
		ParMerkleIndex:    parMerkleIndex,
		ParBlockHeader:    parBlockHeader,
	}

	return auxPow
}

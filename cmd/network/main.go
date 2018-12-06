package main

import (
	"fmt"
	"math"
	"time"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/sync"
)

var (
	test_block = `{
    "type": "state",
	"addresses": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
	"representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"balance": "60000000000000000",
	"link": "D5BA6C7BB3F4F6545E08B03D6DA1258840E0395080378A890601991A2A9E3163",
	"token": "125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
	"signature": "AD57AA8819FA6A7811A13FF0684A79AFDEFB05077BCAD4EC7365C32D2A88D78C8C7C54717B40C0888A0692D05BF3771DF6D16A1F24AE612172922BBD4D93370F",
	"work": "13389dd67ced8429"
	}
	`
	test_block1 = `{
    "type": "state",
	"addresses": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"previous": "9d26b3bd88e9365ec6cc78c96a012dce3d4e11c6866fc45376d5a3a77d4db1e4",
	"representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"balance": "50000000000000000",
	"link": "D5BA6C7BB3F4F6545E08B03D6DA1258840E0395080378A890601991A2A9E3163",
	"token": "125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
	"signature": "AD57AA8819FA6A7811A13FF0684A79AFDEFB05077BCAD4EC7365C32D2A88D78C8C7C54717B40C0888A0692D05BF3771DF6D16A1F24AE612172922BBD4D93370F",
	"work": "13389dd67ced8429"
	}
	`
	test_block2 = `{
    "type": "state",
	"addresses": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
	"representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"balance": "80000000000000000",
	"link": "E4087493C2572DD52AB4AB67AE91C72DFF25CA2EFECECEC8205FF1764D75DDBB",
	"token": "3A938337C8B9F6BA8A6C94F3C53C02815E574E2BC2DCEC3EA2B60E67154FFECA",
	"signature": "B982719A99F2322FE7DE17AFAD9D20B02FC2CDB5BFA105E316C00BF9E1FA2119BB513511777F38B4914CADFBF0F2C721216ED456C96EF09C4FDB6542828CD207",
	"work": "23489dd67ced8256"
	}
	`
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	ledger := ledger.NewLedger()
	defer ledger.Close()

	session := ledger.NewLedgerSession(false)
	defer session.Close()
	node, err := p2p.NewQlcService(cfg, ledger)
	if err != nil {
		fmt.Println(err)
		return
	}
	node.Start()
	servicesync := sync.NewSyncService()
	servicesync.SetQlcService(node)
	servicesync.SetLedger(ledger)
	servicesync.Start()
	blk, err := types.NewBlock(byte(types.State))
	blk1, err := types.NewBlock(byte(types.State))
	blk2, err := types.NewBlock(byte(types.State))
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err = json.Unmarshal([]byte(test_block), &blk); err != nil {
		fmt.Println(err)
		return
	}
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err = json.Unmarshal([]byte(test_block1), &blk1); err != nil {
		fmt.Println(err)
		return
	}
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err = json.Unmarshal([]byte(test_block2), &blk2); err != nil {
		fmt.Println(err)
		return
	}

	err = session.AddBlock(blk)
	if err != nil {
		fmt.Println(err)
		return
	}
	fr := &types.Frontier{
		HeaderBlock: blk.GetHash(),
		OpenBlock:   blk.GetHash(),
	}
	err = session.AddFrontier(fr)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = session.AddBlock(blk1)
	if err != nil {
		fmt.Println(err)
		return
	}
	frs := &types.Frontier{
		HeaderBlock: blk1.GetHash(),
		OpenBlock:   blk.GetHash(),
	}
	err = session.DeleteFrontier(fr.HeaderBlock)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = session.AddFrontier(frs)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = session.AddBlock(blk2)
	if err != nil {
		fmt.Println(err)
		return
	}
	frs2 := &types.Frontier{
		HeaderBlock: blk2.GetHash(),
		OpenBlock:   blk2.GetHash(),
	}
	err = session.AddFrontier(frs2)
	if err != nil {
		fmt.Println(err)
		return
	}
	address := types.Address{}
	Req := sync.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	data, err := sync.FrontierReqToProto(Req)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		peerID, err := node.Node().StreamManager().RandomPeer()
		if err != nil {
			continue
		}
		node.SendMessageToPeer(sync.FrontierRequest, data, peerID)
		time.Sleep(time.Duration(60) * time.Second)
	}
	select {}
}

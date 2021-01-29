package apis

import (
	"math/big"

	"github.com/qlcchain/go-qlc/common/types"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

func toStringPoint(s string) *string {
	if s != "" {
		return &s
	}
	return nil
}

func toUInt32PointByProto(c *pb.UInt32) *uint32 {
	if c == nil {
		return nil
	}
	r := c.GetValue()
	return &r
}

func toOffsetByProto(o *pb.Offset) (int, *int) {
	count := int(o.GetCount())
	offset := int(o.GetOffset())
	return count, &offset
}

func toOffsetByValue(c int32, o int32) (int, *int) {
	count := int(c)
	offset := int(o)
	return count, &offset
}

// StateBlock

func toStateBlock(blk *types.StateBlock) *pbtypes.StateBlock {
	return &pbtypes.StateBlock{
		Type:           toBlockTypeValue(blk.GetType()),
		Token:          toHashValue(blk.GetToken()),
		Address:        toAddressValue(blk.GetAddress()),
		Balance:        toBalanceValue(blk.GetBalance()),
		Vote:           toBalanceValue(blk.GetVote()),
		Network:        toBalanceValue(blk.GetNetwork()),
		Storage:        toBalanceValue(blk.GetStorage()),
		Oracle:         toBalanceValue(blk.GetOracle()),
		Previous:       toHashValue(blk.GetPrevious()),
		Link:           toHashValue(blk.GetLink()),
		Sender:         blk.GetSender(),
		Receiver:       blk.GetReceiver(),
		Message:        toHashValue(blk.GetMessage()),
		Data:           blk.GetData(),
		PoVHeight:      blk.PoVHeight,
		Timestamp:      blk.GetTimestamp(),
		Extra:          toHashValue(blk.GetExtra()),
		Representative: toAddressValue(blk.GetRepresentative()),
		PrivateFrom:    blk.PrivateFrom,
		PrivateFor:     blk.PrivateFor,
		PrivateGroupID: blk.PrivateGroupID,
		Work:           toWorkValue(blk.GetWork()),
		Signature:      toSignatureValue(blk.GetSignature()),
		//Flag:           blk.Flag,
		//PrivateRecvRsp: blk.PrivateRecvRsp,
		//PrivatePayload: blk.PrivatePayload,
	}
}

func toStateBlocks(blocks []*types.StateBlock) *pbtypes.StateBlocks {
	blk := make([]*pbtypes.StateBlock, 0)
	for _, b := range blocks {
		blk = append(blk, toStateBlock(b))
	}
	return &pbtypes.StateBlocks{StateBlocks: blk}
}

func toOriginStateBlock(blk *pbtypes.StateBlock) (*types.StateBlock, error) {
	token, err := toOriginHashByValue(blk.GetToken())
	if err != nil {
		return nil, err
	}
	addr, err := toOriginAddressByValue(blk.GetAddress())
	if err != nil {
		return nil, err
	}
	pre, err := toOriginHashByValue(blk.GetPrevious())
	if err != nil {
		return nil, err
	}
	link, err := toOriginHashByValue(blk.GetLink())
	if err != nil {
		return nil, err
	}
	message, err := toOriginHashByValue(blk.GetMessage())
	if err != nil {
		return nil, err
	}
	extra, err := toOriginHashByValue(blk.GetExtra())
	if err != nil {
		return nil, err
	}
	rep, err := toOriginAddressByValue(blk.GetRepresentative())
	if err != nil {
		return nil, err
	}
	sign, err := toOriginSignatureByValue(blk.GetSignature())
	if err != nil {
		return nil, err
	}
	return &types.StateBlock{
		Type:           toOriginBlockValue(blk.GetType()),
		Token:          token,
		Address:        addr,
		Balance:        types.Balance{Int: big.NewInt(blk.GetBalance())},
		Vote:           types.ToBalance(types.Balance{Int: big.NewInt(blk.GetVote())}),
		Network:        types.ToBalance(types.Balance{Int: big.NewInt(blk.GetNetwork())}),
		Storage:        types.ToBalance(types.Balance{Int: big.NewInt(blk.GetStorage())}),
		Oracle:         types.ToBalance(types.Balance{Int: big.NewInt(blk.GetOracle())}),
		Previous:       pre,
		Link:           link,
		Sender:         blk.GetSender(),
		Receiver:       blk.GetReceiver(),
		Message:        types.ToHash(message),
		Data:           blk.GetData(),
		PoVHeight:      blk.GetPoVHeight(),
		Timestamp:      blk.GetTimestamp(),
		Extra:          &extra,
		Representative: rep,
		PrivateFrom:    blk.GetPrivateFrom(),
		PrivateFor:     blk.GetPrivateFor(),
		PrivateGroupID: blk.GetPrivateGroupID(),
		Work:           toOriginWorkByValue(blk.GetWork()),
		Signature:      sign,
		//Flag:           blk.GetFlag(),
		//PrivateRecvRsp: blk.GetPrivateRecvRsp(),
		//PrivatePayload: blk.GetPrivatePayload(),
	}, nil
}

func toOriginStateBlocks(blks []*pbtypes.StateBlock) ([]*types.StateBlock, error) {
	blocks := make([]*types.StateBlock, 0)
	for _, b := range blks {
		bt, err := toOriginStateBlock(b)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, bt)
	}
	return blocks, nil
}

// BlockType

func toBlockTypeValue(b types.BlockType) string {
	return b.String()
}

func toOriginBlockValue(b string) types.BlockType {
	return types.BlockTypeFromStr(b)
}

// Address

func toAddress(addr types.Address) *pbtypes.Address {
	return &pbtypes.Address{
		Address: addr.String(),
	}
}

func toAddresses(addrs []types.Address) *pbtypes.Addresses {
	as := make([]string, 0)
	for _, addr := range addrs {
		as = append(as, addr.String())
	}
	return &pbtypes.Addresses{Addresses: as}
}

func toAddressValue(addr types.Address) string {
	return addr.String()
}

func toAddressValues(addrs []types.Address) []string {
	r := make([]string, 0)
	for _, a := range addrs {
		r = append(r, toAddressValue(a))
	}
	return r
}

func toOriginAddress(addr *pbtypes.Address) (types.Address, error) {
	address, err := toOriginAddressByValue(addr.GetAddress())
	if err != nil {
		return types.ZeroAddress, nil
	}
	return address, nil
}

func toOriginAddresses(addrs *pbtypes.Addresses) ([]types.Address, error) {
	as := make([]types.Address, 0)
	for _, addr := range addrs.GetAddresses() {
		a, err := toOriginAddressByValue(addr)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, nil
}

func toOriginAddressByValue(addr string) (types.Address, error) {
	return types.HexToAddress(addr)
}

func toOriginAddressesByValues(addrs []string) ([]types.Address, error) {
	rs := make([]types.Address, 0)
	for _, r := range addrs {
		rt, err := toOriginAddressByValue(r)
		if err != nil {
			return nil, err
		}
		rs = append(rs, rt)
	}
	return rs, nil
}

// Hash

func toHash(hash types.Hash) *pbtypes.Hash {
	return &pbtypes.Hash{Hash: hash.String()}
}

func toHashes(hashes []types.Hash) *pbtypes.Hashes {
	hs := make([]string, 0)
	for _, h := range hashes {
		hs = append(hs, h.String())
	}
	return &pbtypes.Hashes{Hashes: hs}
}

func toHashValue(hash types.Hash) string {
	return hash.String()
}

func toHashesValues(hashes []types.Hash) []string {
	r := make([]string, 0)
	for _, h := range hashes {
		r = append(r, toHashValue(h))
	}
	return r
}

func toHashesValuesByPoint(hashes []*types.Hash) []string {
	r := make([]string, 0)
	for _, h := range hashes {
		r = append(r, toHashValue(*h))
	}
	return r
}

func toOriginHash(hash *pbtypes.Hash) (types.Hash, error) {
	h, err := toOriginHashByValue(hash.GetHash())
	if err != nil {
		return types.ZeroHash, err
	}
	return h, nil
}

func toOriginHashes(hashes *pbtypes.Hashes) ([]types.Hash, error) {
	hs := make([]types.Hash, 0)
	for _, h := range hashes.GetHashes() {
		h, err := toOriginHashByValue(h)
		if err != nil {
			return nil, err
		}
		hs = append(hs, h)
	}
	return hs, nil
}

func toOriginHashByValue(hash string) (types.Hash, error) {
	return types.NewHash(hash)
}

func toOriginHashesByValues(hs []string) ([]types.Hash, error) {
	rs := make([]types.Hash, 0)
	for _, r := range hs {
		rt, err := toOriginHashByValue(r)
		if err != nil {
			return nil, err
		}
		rs = append(rs, rt)
	}
	return rs, nil
}

// balance

func toBalance(b types.Balance) *pbtypes.Balance {
	if b.Int == nil {
		return &pbtypes.Balance{Balance: types.ZeroBalance.Int64()}
	}
	return &pbtypes.Balance{Balance: b.Int64()}
}

func toBalanceValue(b types.Balance) int64 {
	if b.Int == nil {
		return types.ZeroBalance.Int64()
	}
	return b.Int64()
}

func toBalanceValueByBigInt(b *big.Int) int64 {
	if b == nil {
		return 0
	}
	return b.Int64()
}

func toBalanceValueByBigNum(b *types.BigNum) int64 {
	if b == nil {
		return 0
	}
	return b.Int64()
}

func toOriginBalanceByValue(b int64) types.Balance {
	return types.Balance{Int: big.NewInt(b)}
}

// signature

func toSignatureValue(b types.Signature) string {
	return b.String()
}

func toOriginSignatureByValue(s string) (types.Signature, error) {
	sign, err := types.NewSignature(s)
	if err != nil {
		return types.ZeroSignature, err
	}
	return sign, nil
}

// work

func toWorkValue(b types.Work) uint64 {
	return uint64(b)
}

func toOriginWorkByValue(v uint64) types.Work {
	return types.Work(v)
}

// bytes

func toBytes(b []byte) *pb.Bytes {
	return &pb.Bytes{Value: b}
}

func toOriginBytes(b *pb.Bytes) []byte {
	return b.GetValue()
}

func toBoolean(b bool) *pb.Boolean {
	return &pb.Boolean{Value: b}
}

func toString(b string) *pb.String {
	return &pb.String{Value: b}
}

func toStrings(bs []string) *pb.Strings {
	return &pb.Strings{Value: bs}
}

func toOriginString(b *pb.String) string {
	return b.GetValue()
}

func toInt64(b int64) *pb.Int64 {
	return &pb.Int64{Value: b}
}

func toUInt64(b uint64) *pb.UInt64 {
	return &pb.UInt64{Value: b}
}

func toOriginUInt64(v *pb.UInt64) uint64 {
	return v.GetValue()
}

func toOriginUInt32(v *pb.UInt32) uint32 {
	return v.GetValue()
}

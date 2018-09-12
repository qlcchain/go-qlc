package types

//var (
//	stateBlock = &StateBlock{
//		Account:        util.Hex32ToBytes("3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic"),
//		PreviousHash:   util.Hex32ToBytes("0000000000000000000000000000000000000000000000000000000000000000"),
//		Representative: util.Hex32ToBytes("3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic"),
//		Balance:        ParseBalanceInts(0, 60000000000000000),
//		Link:           util.Hex32ToBytes("D5BA6C7BB3F4F6545E08B03D6DA1258840E0395080378A890601991A2A9E3163"),
//		Token:          util.Hex32ToBytes("125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36"),
//		Work:           0xf3389dd67ced8429,
//		Signature:      util.Hex64ToBytes("AD57AA8819FA6A7811A13FF0684A79AFDEFB05077BCAD4EC7365C32D2A88D78C8C7C54717B40C0888A0692D05BF3771DF6D16A1F24AE612172922BBD4D93370F"),
//	}
//)
//
//func TestBlockStateMarshal(t *testing.T) {
//	bytes, err := stateBlock.MarshalBinary()
//	if err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println(hex.EncodeToString(stateBlock.Account[:]))
//	var blk StateBlock
//	if err = blk.UnmarshalBinary(bytes); err != nil {
//		t.Fatal(err)
//	}
//	if blk.Hash() != stateBlock.Hash() || blk.Signature != stateBlock.Signature || blk.Work != stateBlock.Work {
//		t.Fatalf("blocks not equal")
//	}
//}

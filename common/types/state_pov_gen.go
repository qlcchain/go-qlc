package types

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *PovAccountState) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "a":
			err = dc.ReadExtension(&z.Account)
			if err != nil {
				err = msgp.WrapError(err, "Account")
				return
			}
		case "b":
			err = dc.ReadExtension(&z.Balance)
			if err != nil {
				err = msgp.WrapError(err, "Balance")
				return
			}
		case "v":
			err = dc.ReadExtension(&z.Vote)
			if err != nil {
				err = msgp.WrapError(err, "Vote")
				return
			}
		case "n":
			err = dc.ReadExtension(&z.Network)
			if err != nil {
				err = msgp.WrapError(err, "Network")
				return
			}
		case "s":
			err = dc.ReadExtension(&z.Storage)
			if err != nil {
				err = msgp.WrapError(err, "Storage")
				return
			}
		case "o":
			err = dc.ReadExtension(&z.Oracle)
			if err != nil {
				err = msgp.WrapError(err, "Oracle")
				return
			}
		case "ts":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				err = msgp.WrapError(err, "TokenStates")
				return
			}
			if cap(z.TokenStates) >= int(zb0002) {
				z.TokenStates = (z.TokenStates)[:zb0002]
			} else {
				z.TokenStates = make([]*PovTokenState, zb0002)
			}
			for za0001 := range z.TokenStates {
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						err = msgp.WrapError(err, "TokenStates", za0001)
						return
					}
					z.TokenStates[za0001] = nil
				} else {
					if z.TokenStates[za0001] == nil {
						z.TokenStates[za0001] = new(PovTokenState)
					}
					err = z.TokenStates[za0001].DecodeMsg(dc)
					if err != nil {
						err = msgp.WrapError(err, "TokenStates", za0001)
						return
					}
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *PovAccountState) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 7
	// write "a"
	err = en.Append(0x87, 0xa1, 0x61)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Account)
	if err != nil {
		err = msgp.WrapError(err, "Account")
		return
	}
	// write "b"
	err = en.Append(0xa1, 0x62)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Balance)
	if err != nil {
		err = msgp.WrapError(err, "Balance")
		return
	}
	// write "v"
	err = en.Append(0xa1, 0x76)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Vote)
	if err != nil {
		err = msgp.WrapError(err, "Vote")
		return
	}
	// write "n"
	err = en.Append(0xa1, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Network)
	if err != nil {
		err = msgp.WrapError(err, "Network")
		return
	}
	// write "s"
	err = en.Append(0xa1, 0x73)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Storage)
	if err != nil {
		err = msgp.WrapError(err, "Storage")
		return
	}
	// write "o"
	err = en.Append(0xa1, 0x6f)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Oracle)
	if err != nil {
		err = msgp.WrapError(err, "Oracle")
		return
	}
	// write "ts"
	err = en.Append(0xa2, 0x74, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.TokenStates)))
	if err != nil {
		err = msgp.WrapError(err, "TokenStates")
		return
	}
	for za0001 := range z.TokenStates {
		if z.TokenStates[za0001] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = z.TokenStates[za0001].EncodeMsg(en)
			if err != nil {
				err = msgp.WrapError(err, "TokenStates", za0001)
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PovAccountState) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 7
	// string "a"
	o = append(o, 0x87, 0xa1, 0x61)
	o, err = msgp.AppendExtension(o, &z.Account)
	if err != nil {
		err = msgp.WrapError(err, "Account")
		return
	}
	// string "b"
	o = append(o, 0xa1, 0x62)
	o, err = msgp.AppendExtension(o, &z.Balance)
	if err != nil {
		err = msgp.WrapError(err, "Balance")
		return
	}
	// string "v"
	o = append(o, 0xa1, 0x76)
	o, err = msgp.AppendExtension(o, &z.Vote)
	if err != nil {
		err = msgp.WrapError(err, "Vote")
		return
	}
	// string "n"
	o = append(o, 0xa1, 0x6e)
	o, err = msgp.AppendExtension(o, &z.Network)
	if err != nil {
		err = msgp.WrapError(err, "Network")
		return
	}
	// string "s"
	o = append(o, 0xa1, 0x73)
	o, err = msgp.AppendExtension(o, &z.Storage)
	if err != nil {
		err = msgp.WrapError(err, "Storage")
		return
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o, err = msgp.AppendExtension(o, &z.Oracle)
	if err != nil {
		err = msgp.WrapError(err, "Oracle")
		return
	}
	// string "ts"
	o = append(o, 0xa2, 0x74, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.TokenStates)))
	for za0001 := range z.TokenStates {
		if z.TokenStates[za0001] == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = z.TokenStates[za0001].MarshalMsg(o)
			if err != nil {
				err = msgp.WrapError(err, "TokenStates", za0001)
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovAccountState) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "a":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Account)
			if err != nil {
				err = msgp.WrapError(err, "Account")
				return
			}
		case "b":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Balance)
			if err != nil {
				err = msgp.WrapError(err, "Balance")
				return
			}
		case "v":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Vote)
			if err != nil {
				err = msgp.WrapError(err, "Vote")
				return
			}
		case "n":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Network)
			if err != nil {
				err = msgp.WrapError(err, "Network")
				return
			}
		case "s":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Storage)
			if err != nil {
				err = msgp.WrapError(err, "Storage")
				return
			}
		case "o":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Oracle)
			if err != nil {
				err = msgp.WrapError(err, "Oracle")
				return
			}
		case "ts":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "TokenStates")
				return
			}
			if cap(z.TokenStates) >= int(zb0002) {
				z.TokenStates = (z.TokenStates)[:zb0002]
			} else {
				z.TokenStates = make([]*PovTokenState, zb0002)
			}
			for za0001 := range z.TokenStates {
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					z.TokenStates[za0001] = nil
				} else {
					if z.TokenStates[za0001] == nil {
						z.TokenStates[za0001] = new(PovTokenState)
					}
					bts, err = z.TokenStates[za0001].UnmarshalMsg(bts)
					if err != nil {
						err = msgp.WrapError(err, "TokenStates", za0001)
						return
					}
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *PovAccountState) Msgsize() (s int) {
	s = 1 + 2 + msgp.ExtensionPrefixSize + z.Account.Len() + 2 + msgp.ExtensionPrefixSize + z.Balance.Len() + 2 + msgp.ExtensionPrefixSize + z.Vote.Len() + 2 + msgp.ExtensionPrefixSize + z.Network.Len() + 2 + msgp.ExtensionPrefixSize + z.Storage.Len() + 2 + msgp.ExtensionPrefixSize + z.Oracle.Len() + 3 + msgp.ArrayHeaderSize
	for za0001 := range z.TokenStates {
		if z.TokenStates[za0001] == nil {
			s += msgp.NilSize
		} else {
			s += z.TokenStates[za0001].Msgsize()
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PovContractState) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "sr":
			err = dc.ReadExtension(&z.StateRoot)
			if err != nil {
				err = msgp.WrapError(err, "StateRoot")
				return
			}
		case "ch":
			err = dc.ReadExtension(&z.CodeHash)
			if err != nil {
				err = msgp.WrapError(err, "CodeHash")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z PovContractState) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "sr"
	err = en.Append(0x82, 0xa2, 0x73, 0x72)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.StateRoot)
	if err != nil {
		err = msgp.WrapError(err, "StateRoot")
		return
	}
	// write "ch"
	err = en.Append(0xa2, 0x63, 0x68)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.CodeHash)
	if err != nil {
		err = msgp.WrapError(err, "CodeHash")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z PovContractState) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "sr"
	o = append(o, 0x82, 0xa2, 0x73, 0x72)
	o, err = msgp.AppendExtension(o, &z.StateRoot)
	if err != nil {
		err = msgp.WrapError(err, "StateRoot")
		return
	}
	// string "ch"
	o = append(o, 0xa2, 0x63, 0x68)
	o, err = msgp.AppendExtension(o, &z.CodeHash)
	if err != nil {
		err = msgp.WrapError(err, "CodeHash")
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovContractState) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "sr":
			bts, err = msgp.ReadExtensionBytes(bts, &z.StateRoot)
			if err != nil {
				err = msgp.WrapError(err, "StateRoot")
				return
			}
		case "ch":
			bts, err = msgp.ReadExtensionBytes(bts, &z.CodeHash)
			if err != nil {
				err = msgp.WrapError(err, "CodeHash")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z PovContractState) Msgsize() (s int) {
	s = 1 + 3 + msgp.ExtensionPrefixSize + z.StateRoot.Len() + 3 + msgp.ExtensionPrefixSize + z.CodeHash.Len()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PovPublishState) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "vh":
			z.VerifiedHeight, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "VerifiedHeight")
				return
			}
		case "vs":
			z.VerifiedStatus, err = dc.ReadInt8()
			if err != nil {
				err = msgp.WrapError(err, "VerifiedStatus")
				return
			}
		case "bf":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "BonusFee")
					return
				}
				z.BonusFee = nil
			} else {
				if z.BonusFee == nil {
					z.BonusFee = new(BigNum)
				}
				err = dc.ReadExtension(z.BonusFee)
				if err != nil {
					err = msgp.WrapError(err, "BonusFee")
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *PovPublishState) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "vh"
	err = en.Append(0x83, 0xa2, 0x76, 0x68)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.VerifiedHeight)
	if err != nil {
		err = msgp.WrapError(err, "VerifiedHeight")
		return
	}
	// write "vs"
	err = en.Append(0xa2, 0x76, 0x73)
	if err != nil {
		return
	}
	err = en.WriteInt8(z.VerifiedStatus)
	if err != nil {
		err = msgp.WrapError(err, "VerifiedStatus")
		return
	}
	// write "bf"
	err = en.Append(0xa2, 0x62, 0x66)
	if err != nil {
		return
	}
	if z.BonusFee == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteExtension(z.BonusFee)
		if err != nil {
			err = msgp.WrapError(err, "BonusFee")
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PovPublishState) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "vh"
	o = append(o, 0x83, 0xa2, 0x76, 0x68)
	o = msgp.AppendUint64(o, z.VerifiedHeight)
	// string "vs"
	o = append(o, 0xa2, 0x76, 0x73)
	o = msgp.AppendInt8(o, z.VerifiedStatus)
	// string "bf"
	o = append(o, 0xa2, 0x62, 0x66)
	if z.BonusFee == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = msgp.AppendExtension(o, z.BonusFee)
		if err != nil {
			err = msgp.WrapError(err, "BonusFee")
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovPublishState) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "vh":
			z.VerifiedHeight, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "VerifiedHeight")
				return
			}
		case "vs":
			z.VerifiedStatus, bts, err = msgp.ReadInt8Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "VerifiedStatus")
				return
			}
		case "bf":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.BonusFee = nil
			} else {
				if z.BonusFee == nil {
					z.BonusFee = new(BigNum)
				}
				bts, err = msgp.ReadExtensionBytes(bts, z.BonusFee)
				if err != nil {
					err = msgp.WrapError(err, "BonusFee")
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *PovPublishState) Msgsize() (s int) {
	s = 1 + 3 + msgp.Uint64Size + 3 + msgp.Int8Size + 3
	if z.BonusFee == nil {
		s += msgp.NilSize
	} else {
		s += msgp.ExtensionPrefixSize + z.BonusFee.Len()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PovRepState) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "a":
			err = dc.ReadExtension(&z.Account)
			if err != nil {
				err = msgp.WrapError(err, "Account")
				return
			}
		case "b":
			err = dc.ReadExtension(&z.Balance)
			if err != nil {
				err = msgp.WrapError(err, "Balance")
				return
			}
		case "v":
			err = dc.ReadExtension(&z.Vote)
			if err != nil {
				err = msgp.WrapError(err, "Vote")
				return
			}
		case "n":
			err = dc.ReadExtension(&z.Network)
			if err != nil {
				err = msgp.WrapError(err, "Network")
				return
			}
		case "s":
			err = dc.ReadExtension(&z.Storage)
			if err != nil {
				err = msgp.WrapError(err, "Storage")
				return
			}
		case "o":
			err = dc.ReadExtension(&z.Oracle)
			if err != nil {
				err = msgp.WrapError(err, "Oracle")
				return
			}
		case "t":
			err = dc.ReadExtension(&z.Total)
			if err != nil {
				err = msgp.WrapError(err, "Total")
				return
			}
		case "st":
			z.Status, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Status")
				return
			}
		case "he":
			z.Height, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "Height")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *PovRepState) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 9
	// write "a"
	err = en.Append(0x89, 0xa1, 0x61)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Account)
	if err != nil {
		err = msgp.WrapError(err, "Account")
		return
	}
	// write "b"
	err = en.Append(0xa1, 0x62)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Balance)
	if err != nil {
		err = msgp.WrapError(err, "Balance")
		return
	}
	// write "v"
	err = en.Append(0xa1, 0x76)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Vote)
	if err != nil {
		err = msgp.WrapError(err, "Vote")
		return
	}
	// write "n"
	err = en.Append(0xa1, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Network)
	if err != nil {
		err = msgp.WrapError(err, "Network")
		return
	}
	// write "s"
	err = en.Append(0xa1, 0x73)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Storage)
	if err != nil {
		err = msgp.WrapError(err, "Storage")
		return
	}
	// write "o"
	err = en.Append(0xa1, 0x6f)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Oracle)
	if err != nil {
		err = msgp.WrapError(err, "Oracle")
		return
	}
	// write "t"
	err = en.Append(0xa1, 0x74)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Total)
	if err != nil {
		err = msgp.WrapError(err, "Total")
		return
	}
	// write "st"
	err = en.Append(0xa2, 0x73, 0x74)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Status)
	if err != nil {
		err = msgp.WrapError(err, "Status")
		return
	}
	// write "he"
	err = en.Append(0xa2, 0x68, 0x65)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.Height)
	if err != nil {
		err = msgp.WrapError(err, "Height")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PovRepState) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 9
	// string "a"
	o = append(o, 0x89, 0xa1, 0x61)
	o, err = msgp.AppendExtension(o, &z.Account)
	if err != nil {
		err = msgp.WrapError(err, "Account")
		return
	}
	// string "b"
	o = append(o, 0xa1, 0x62)
	o, err = msgp.AppendExtension(o, &z.Balance)
	if err != nil {
		err = msgp.WrapError(err, "Balance")
		return
	}
	// string "v"
	o = append(o, 0xa1, 0x76)
	o, err = msgp.AppendExtension(o, &z.Vote)
	if err != nil {
		err = msgp.WrapError(err, "Vote")
		return
	}
	// string "n"
	o = append(o, 0xa1, 0x6e)
	o, err = msgp.AppendExtension(o, &z.Network)
	if err != nil {
		err = msgp.WrapError(err, "Network")
		return
	}
	// string "s"
	o = append(o, 0xa1, 0x73)
	o, err = msgp.AppendExtension(o, &z.Storage)
	if err != nil {
		err = msgp.WrapError(err, "Storage")
		return
	}
	// string "o"
	o = append(o, 0xa1, 0x6f)
	o, err = msgp.AppendExtension(o, &z.Oracle)
	if err != nil {
		err = msgp.WrapError(err, "Oracle")
		return
	}
	// string "t"
	o = append(o, 0xa1, 0x74)
	o, err = msgp.AppendExtension(o, &z.Total)
	if err != nil {
		err = msgp.WrapError(err, "Total")
		return
	}
	// string "st"
	o = append(o, 0xa2, 0x73, 0x74)
	o = msgp.AppendUint32(o, z.Status)
	// string "he"
	o = append(o, 0xa2, 0x68, 0x65)
	o = msgp.AppendUint64(o, z.Height)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovRepState) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "a":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Account)
			if err != nil {
				err = msgp.WrapError(err, "Account")
				return
			}
		case "b":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Balance)
			if err != nil {
				err = msgp.WrapError(err, "Balance")
				return
			}
		case "v":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Vote)
			if err != nil {
				err = msgp.WrapError(err, "Vote")
				return
			}
		case "n":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Network)
			if err != nil {
				err = msgp.WrapError(err, "Network")
				return
			}
		case "s":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Storage)
			if err != nil {
				err = msgp.WrapError(err, "Storage")
				return
			}
		case "o":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Oracle)
			if err != nil {
				err = msgp.WrapError(err, "Oracle")
				return
			}
		case "t":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Total)
			if err != nil {
				err = msgp.WrapError(err, "Total")
				return
			}
		case "st":
			z.Status, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Status")
				return
			}
		case "he":
			z.Height, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Height")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *PovRepState) Msgsize() (s int) {
	s = 1 + 2 + msgp.ExtensionPrefixSize + z.Account.Len() + 2 + msgp.ExtensionPrefixSize + z.Balance.Len() + 2 + msgp.ExtensionPrefixSize + z.Vote.Len() + 2 + msgp.ExtensionPrefixSize + z.Network.Len() + 2 + msgp.ExtensionPrefixSize + z.Storage.Len() + 2 + msgp.ExtensionPrefixSize + z.Oracle.Len() + 2 + msgp.ExtensionPrefixSize + z.Total.Len() + 3 + msgp.Uint32Size + 3 + msgp.Uint64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PovTokenState) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "t":
			err = dc.ReadExtension(&z.Type)
			if err != nil {
				err = msgp.WrapError(err, "Type")
				return
			}
		case "h":
			err = dc.ReadExtension(&z.Hash)
			if err != nil {
				err = msgp.WrapError(err, "Hash")
				return
			}
		case "r":
			err = dc.ReadExtension(&z.Representative)
			if err != nil {
				err = msgp.WrapError(err, "Representative")
				return
			}
		case "b":
			err = dc.ReadExtension(&z.Balance)
			if err != nil {
				err = msgp.WrapError(err, "Balance")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *PovTokenState) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "t"
	err = en.Append(0x84, 0xa1, 0x74)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Type)
	if err != nil {
		err = msgp.WrapError(err, "Type")
		return
	}
	// write "h"
	err = en.Append(0xa1, 0x68)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Hash)
	if err != nil {
		err = msgp.WrapError(err, "Hash")
		return
	}
	// write "r"
	err = en.Append(0xa1, 0x72)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Representative)
	if err != nil {
		err = msgp.WrapError(err, "Representative")
		return
	}
	// write "b"
	err = en.Append(0xa1, 0x62)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Balance)
	if err != nil {
		err = msgp.WrapError(err, "Balance")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PovTokenState) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "t"
	o = append(o, 0x84, 0xa1, 0x74)
	o, err = msgp.AppendExtension(o, &z.Type)
	if err != nil {
		err = msgp.WrapError(err, "Type")
		return
	}
	// string "h"
	o = append(o, 0xa1, 0x68)
	o, err = msgp.AppendExtension(o, &z.Hash)
	if err != nil {
		err = msgp.WrapError(err, "Hash")
		return
	}
	// string "r"
	o = append(o, 0xa1, 0x72)
	o, err = msgp.AppendExtension(o, &z.Representative)
	if err != nil {
		err = msgp.WrapError(err, "Representative")
		return
	}
	// string "b"
	o = append(o, 0xa1, 0x62)
	o, err = msgp.AppendExtension(o, &z.Balance)
	if err != nil {
		err = msgp.WrapError(err, "Balance")
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovTokenState) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "t":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Type)
			if err != nil {
				err = msgp.WrapError(err, "Type")
				return
			}
		case "h":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Hash)
			if err != nil {
				err = msgp.WrapError(err, "Hash")
				return
			}
		case "r":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Representative)
			if err != nil {
				err = msgp.WrapError(err, "Representative")
				return
			}
		case "b":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Balance)
			if err != nil {
				err = msgp.WrapError(err, "Balance")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *PovTokenState) Msgsize() (s int) {
	s = 1 + 2 + msgp.ExtensionPrefixSize + z.Type.Len() + 2 + msgp.ExtensionPrefixSize + z.Hash.Len() + 2 + msgp.ExtensionPrefixSize + z.Representative.Len() + 2 + msgp.ExtensionPrefixSize + z.Balance.Len()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PovVerifierState) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "tv":
			z.TotalVerify, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "TotalVerify")
				return
			}
		case "tr":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "TotalReward")
					return
				}
				z.TotalReward = nil
			} else {
				if z.TotalReward == nil {
					z.TotalReward = new(BigNum)
				}
				err = dc.ReadExtension(z.TotalReward)
				if err != nil {
					err = msgp.WrapError(err, "TotalReward")
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *PovVerifierState) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "tv"
	err = en.Append(0x82, 0xa2, 0x74, 0x76)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.TotalVerify)
	if err != nil {
		err = msgp.WrapError(err, "TotalVerify")
		return
	}
	// write "tr"
	err = en.Append(0xa2, 0x74, 0x72)
	if err != nil {
		return
	}
	if z.TotalReward == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteExtension(z.TotalReward)
		if err != nil {
			err = msgp.WrapError(err, "TotalReward")
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PovVerifierState) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "tv"
	o = append(o, 0x82, 0xa2, 0x74, 0x76)
	o = msgp.AppendUint64(o, z.TotalVerify)
	// string "tr"
	o = append(o, 0xa2, 0x74, 0x72)
	if z.TotalReward == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = msgp.AppendExtension(o, z.TotalReward)
		if err != nil {
			err = msgp.WrapError(err, "TotalReward")
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovVerifierState) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "tv":
			z.TotalVerify, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "TotalVerify")
				return
			}
		case "tr":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.TotalReward = nil
			} else {
				if z.TotalReward == nil {
					z.TotalReward = new(BigNum)
				}
				bts, err = msgp.ReadExtensionBytes(bts, z.TotalReward)
				if err != nil {
					err = msgp.WrapError(err, "TotalReward")
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *PovVerifierState) Msgsize() (s int) {
	s = 1 + 3 + msgp.Uint64Size + 3
	if z.TotalReward == nil {
		s += msgp.NilSize
	} else {
		s += msgp.ExtensionPrefixSize + z.TotalReward.Len()
	}
	return
}

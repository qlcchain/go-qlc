package settlement

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *CDRParam) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "i":
			z.Index, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "Index")
				return
			}
		case "dt":
			z.SmsDt, err = dc.ReadInt64()
			if err != nil {
				err = msgp.WrapError(err, "SmsDt")
				return
			}
		case "ac":
			z.Account, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Account")
				return
			}
		case "tx":
			z.Sender, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Sender")
				return
			}
		case "c":
			z.Customer, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Customer")
				return
			}
		case "d":
			z.Destination, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Destination")
				return
			}
		case "s":
			{
				var zb0002 int
				zb0002, err = dc.ReadInt()
				if err != nil {
					err = msgp.WrapError(err, "SendingStatus")
					return
				}
				z.SendingStatus = SendingStatus(zb0002)
			}
		case "ds":
			{
				var zb0003 int
				zb0003, err = dc.ReadInt()
				if err != nil {
					err = msgp.WrapError(err, "DlrStatus")
					return
				}
				z.DlrStatus = DLRStatus(zb0003)
			}
		case "ps":
			z.PreStop, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "PreStop")
				return
			}
		case "ns":
			z.NextStop, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "NextStop")
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
func (z *CDRParam) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 10
	// write "i"
	err = en.Append(0x8a, 0xa1, 0x69)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.Index)
	if err != nil {
		err = msgp.WrapError(err, "Index")
		return
	}
	// write "dt"
	err = en.Append(0xa2, 0x64, 0x74)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.SmsDt)
	if err != nil {
		err = msgp.WrapError(err, "SmsDt")
		return
	}
	// write "ac"
	err = en.Append(0xa2, 0x61, 0x63)
	if err != nil {
		return
	}
	err = en.WriteString(z.Account)
	if err != nil {
		err = msgp.WrapError(err, "Account")
		return
	}
	// write "tx"
	err = en.Append(0xa2, 0x74, 0x78)
	if err != nil {
		return
	}
	err = en.WriteString(z.Sender)
	if err != nil {
		err = msgp.WrapError(err, "Sender")
		return
	}
	// write "c"
	err = en.Append(0xa1, 0x63)
	if err != nil {
		return
	}
	err = en.WriteString(z.Customer)
	if err != nil {
		err = msgp.WrapError(err, "Customer")
		return
	}
	// write "d"
	err = en.Append(0xa1, 0x64)
	if err != nil {
		return
	}
	err = en.WriteString(z.Destination)
	if err != nil {
		err = msgp.WrapError(err, "Destination")
		return
	}
	// write "s"
	err = en.Append(0xa1, 0x73)
	if err != nil {
		return
	}
	err = en.WriteInt(int(z.SendingStatus))
	if err != nil {
		err = msgp.WrapError(err, "SendingStatus")
		return
	}
	// write "ds"
	err = en.Append(0xa2, 0x64, 0x73)
	if err != nil {
		return
	}
	err = en.WriteInt(int(z.DlrStatus))
	if err != nil {
		err = msgp.WrapError(err, "DlrStatus")
		return
	}
	// write "ps"
	err = en.Append(0xa2, 0x70, 0x73)
	if err != nil {
		return
	}
	err = en.WriteString(z.PreStop)
	if err != nil {
		err = msgp.WrapError(err, "PreStop")
		return
	}
	// write "ns"
	err = en.Append(0xa2, 0x6e, 0x73)
	if err != nil {
		return
	}
	err = en.WriteString(z.NextStop)
	if err != nil {
		err = msgp.WrapError(err, "NextStop")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *CDRParam) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 10
	// string "i"
	o = append(o, 0x8a, 0xa1, 0x69)
	o = msgp.AppendUint64(o, z.Index)
	// string "dt"
	o = append(o, 0xa2, 0x64, 0x74)
	o = msgp.AppendInt64(o, z.SmsDt)
	// string "ac"
	o = append(o, 0xa2, 0x61, 0x63)
	o = msgp.AppendString(o, z.Account)
	// string "tx"
	o = append(o, 0xa2, 0x74, 0x78)
	o = msgp.AppendString(o, z.Sender)
	// string "c"
	o = append(o, 0xa1, 0x63)
	o = msgp.AppendString(o, z.Customer)
	// string "d"
	o = append(o, 0xa1, 0x64)
	o = msgp.AppendString(o, z.Destination)
	// string "s"
	o = append(o, 0xa1, 0x73)
	o = msgp.AppendInt(o, int(z.SendingStatus))
	// string "ds"
	o = append(o, 0xa2, 0x64, 0x73)
	o = msgp.AppendInt(o, int(z.DlrStatus))
	// string "ps"
	o = append(o, 0xa2, 0x70, 0x73)
	o = msgp.AppendString(o, z.PreStop)
	// string "ns"
	o = append(o, 0xa2, 0x6e, 0x73)
	o = msgp.AppendString(o, z.NextStop)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CDRParam) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "i":
			z.Index, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Index")
				return
			}
		case "dt":
			z.SmsDt, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "SmsDt")
				return
			}
		case "ac":
			z.Account, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Account")
				return
			}
		case "tx":
			z.Sender, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Sender")
				return
			}
		case "c":
			z.Customer, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Customer")
				return
			}
		case "d":
			z.Destination, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Destination")
				return
			}
		case "s":
			{
				var zb0002 int
				zb0002, bts, err = msgp.ReadIntBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "SendingStatus")
					return
				}
				z.SendingStatus = SendingStatus(zb0002)
			}
		case "ds":
			{
				var zb0003 int
				zb0003, bts, err = msgp.ReadIntBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "DlrStatus")
					return
				}
				z.DlrStatus = DLRStatus(zb0003)
			}
		case "ps":
			z.PreStop, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "PreStop")
				return
			}
		case "ns":
			z.NextStop, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "NextStop")
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
func (z *CDRParam) Msgsize() (s int) {
	s = 1 + 2 + msgp.Uint64Size + 3 + msgp.Int64Size + 3 + msgp.StringPrefixSize + len(z.Account) + 3 + msgp.StringPrefixSize + len(z.Sender) + 2 + msgp.StringPrefixSize + len(z.Customer) + 2 + msgp.StringPrefixSize + len(z.Destination) + 2 + msgp.IntSize + 3 + msgp.IntSize + 3 + msgp.StringPrefixSize + len(z.PreStop) + 3 + msgp.StringPrefixSize + len(z.NextStop)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *CDRParamList) DecodeMsg(dc *msgp.Reader) (err error) {
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
			err = z.ContractAddress.DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, "ContractAddress")
				return
			}
		case "p":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				err = msgp.WrapError(err, "Params")
				return
			}
			if cap(z.Params) >= int(zb0002) {
				z.Params = (z.Params)[:zb0002]
			} else {
				z.Params = make([]*CDRParam, zb0002)
			}
			for za0001 := range z.Params {
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						err = msgp.WrapError(err, "Params", za0001)
						return
					}
					z.Params[za0001] = nil
				} else {
					if z.Params[za0001] == nil {
						z.Params[za0001] = new(CDRParam)
					}
					err = z.Params[za0001].DecodeMsg(dc)
					if err != nil {
						err = msgp.WrapError(err, "Params", za0001)
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
func (z *CDRParamList) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "a"
	err = en.Append(0x82, 0xa1, 0x61)
	if err != nil {
		return
	}
	err = z.ContractAddress.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "ContractAddress")
		return
	}
	// write "p"
	err = en.Append(0xa1, 0x70)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Params)))
	if err != nil {
		err = msgp.WrapError(err, "Params")
		return
	}
	for za0001 := range z.Params {
		if z.Params[za0001] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = z.Params[za0001].EncodeMsg(en)
			if err != nil {
				err = msgp.WrapError(err, "Params", za0001)
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *CDRParamList) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "a"
	o = append(o, 0x82, 0xa1, 0x61)
	o, err = z.ContractAddress.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "ContractAddress")
		return
	}
	// string "p"
	o = append(o, 0xa1, 0x70)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Params)))
	for za0001 := range z.Params {
		if z.Params[za0001] == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = z.Params[za0001].MarshalMsg(o)
			if err != nil {
				err = msgp.WrapError(err, "Params", za0001)
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CDRParamList) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
			bts, err = z.ContractAddress.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "ContractAddress")
				return
			}
		case "p":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Params")
				return
			}
			if cap(z.Params) >= int(zb0002) {
				z.Params = (z.Params)[:zb0002]
			} else {
				z.Params = make([]*CDRParam, zb0002)
			}
			for za0001 := range z.Params {
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					z.Params[za0001] = nil
				} else {
					if z.Params[za0001] == nil {
						z.Params[za0001] = new(CDRParam)
					}
					bts, err = z.Params[za0001].UnmarshalMsg(bts)
					if err != nil {
						err = msgp.WrapError(err, "Params", za0001)
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
func (z *CDRParamList) Msgsize() (s int) {
	s = 1 + 2 + z.ContractAddress.Msgsize() + 2 + msgp.ArrayHeaderSize
	for za0001 := range z.Params {
		if z.Params[za0001] == nil {
			s += msgp.NilSize
		} else {
			s += z.Params[za0001].Msgsize()
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *DLRStatus) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 int
		zb0001, err = dc.ReadInt()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = DLRStatus(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z DLRStatus) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteInt(int(z))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z DLRStatus) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendInt(o, int(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DLRStatus) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 int
		zb0001, bts, err = msgp.ReadIntBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = DLRStatus(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z DLRStatus) Msgsize() (s int) {
	s = msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SendingStatus) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 int
		zb0001, err = dc.ReadInt()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = SendingStatus(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z SendingStatus) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteInt(int(z))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z SendingStatus) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendInt(o, int(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SendingStatus) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 int
		zb0001, bts, err = msgp.ReadIntBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = SendingStatus(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z SendingStatus) Msgsize() (s int) {
	s = msgp.IntSize
	return
}

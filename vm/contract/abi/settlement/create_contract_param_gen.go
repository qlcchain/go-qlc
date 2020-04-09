package settlement

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *ContractService) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "id":
			z.ServiceId, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "ServiceId")
				return
			}
		case "mcc":
			z.Mcc, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "Mcc")
				return
			}
		case "mnc":
			z.Mnc, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "Mnc")
				return
			}
		case "t":
			z.TotalAmount, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "TotalAmount")
				return
			}
		case "u":
			z.UnitPrice, err = dc.ReadFloat64()
			if err != nil {
				err = msgp.WrapError(err, "UnitPrice")
				return
			}
		case "c":
			z.Currency, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Currency")
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
func (z *ContractService) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "id"
	err = en.Append(0x86, 0xa2, 0x69, 0x64)
	if err != nil {
		return
	}
	err = en.WriteString(z.ServiceId)
	if err != nil {
		err = msgp.WrapError(err, "ServiceId")
		return
	}
	// write "mcc"
	err = en.Append(0xa3, 0x6d, 0x63, 0x63)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.Mcc)
	if err != nil {
		err = msgp.WrapError(err, "Mcc")
		return
	}
	// write "mnc"
	err = en.Append(0xa3, 0x6d, 0x6e, 0x63)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.Mnc)
	if err != nil {
		err = msgp.WrapError(err, "Mnc")
		return
	}
	// write "t"
	err = en.Append(0xa1, 0x74)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.TotalAmount)
	if err != nil {
		err = msgp.WrapError(err, "TotalAmount")
		return
	}
	// write "u"
	err = en.Append(0xa1, 0x75)
	if err != nil {
		return
	}
	err = en.WriteFloat64(z.UnitPrice)
	if err != nil {
		err = msgp.WrapError(err, "UnitPrice")
		return
	}
	// write "c"
	err = en.Append(0xa1, 0x63)
	if err != nil {
		return
	}
	err = en.WriteString(z.Currency)
	if err != nil {
		err = msgp.WrapError(err, "Currency")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ContractService) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "id"
	o = append(o, 0x86, 0xa2, 0x69, 0x64)
	o = msgp.AppendString(o, z.ServiceId)
	// string "mcc"
	o = append(o, 0xa3, 0x6d, 0x63, 0x63)
	o = msgp.AppendUint64(o, z.Mcc)
	// string "mnc"
	o = append(o, 0xa3, 0x6d, 0x6e, 0x63)
	o = msgp.AppendUint64(o, z.Mnc)
	// string "t"
	o = append(o, 0xa1, 0x74)
	o = msgp.AppendUint64(o, z.TotalAmount)
	// string "u"
	o = append(o, 0xa1, 0x75)
	o = msgp.AppendFloat64(o, z.UnitPrice)
	// string "c"
	o = append(o, 0xa1, 0x63)
	o = msgp.AppendString(o, z.Currency)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ContractService) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "id":
			z.ServiceId, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "ServiceId")
				return
			}
		case "mcc":
			z.Mcc, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Mcc")
				return
			}
		case "mnc":
			z.Mnc, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Mnc")
				return
			}
		case "t":
			z.TotalAmount, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "TotalAmount")
				return
			}
		case "u":
			z.UnitPrice, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "UnitPrice")
				return
			}
		case "c":
			z.Currency, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Currency")
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
func (z *ContractService) Msgsize() (s int) {
	s = 1 + 3 + msgp.StringPrefixSize + len(z.ServiceId) + 4 + msgp.Uint64Size + 4 + msgp.Uint64Size + 2 + msgp.Uint64Size + 2 + msgp.Float64Size + 2 + msgp.StringPrefixSize + len(z.Currency)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ContractStatus) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 int
		zb0001, err = dc.ReadInt()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = ContractStatus(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ContractStatus) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteInt(int(z))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ContractStatus) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendInt(o, int(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ContractStatus) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 int
		zb0001, bts, err = msgp.ReadIntBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = ContractStatus(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ContractStatus) Msgsize() (s int) {
	s = msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *CreateContractParam) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "pa":
			err = z.PartyA.DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, "PartyA")
				return
			}
		case "pb":
			err = z.PartyB.DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, "PartyB")
				return
			}
		case "pre":
			err = dc.ReadExtension(&z.Previous)
			if err != nil {
				err = msgp.WrapError(err, "Previous")
				return
			}
		case "s":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				err = msgp.WrapError(err, "Services")
				return
			}
			if cap(z.Services) >= int(zb0002) {
				z.Services = (z.Services)[:zb0002]
			} else {
				z.Services = make([]ContractService, zb0002)
			}
			for za0001 := range z.Services {
				err = z.Services[za0001].DecodeMsg(dc)
				if err != nil {
					err = msgp.WrapError(err, "Services", za0001)
					return
				}
			}
		case "t1":
			z.SignDate, err = dc.ReadInt64()
			if err != nil {
				err = msgp.WrapError(err, "SignDate")
				return
			}
		case "t3":
			z.StartDate, err = dc.ReadInt64()
			if err != nil {
				err = msgp.WrapError(err, "StartDate")
				return
			}
		case "t4":
			z.EndDate, err = dc.ReadInt64()
			if err != nil {
				err = msgp.WrapError(err, "EndDate")
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
func (z *CreateContractParam) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 7
	// write "pa"
	err = en.Append(0x87, 0xa2, 0x70, 0x61)
	if err != nil {
		return
	}
	err = z.PartyA.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "PartyA")
		return
	}
	// write "pb"
	err = en.Append(0xa2, 0x70, 0x62)
	if err != nil {
		return
	}
	err = z.PartyB.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "PartyB")
		return
	}
	// write "pre"
	err = en.Append(0xa3, 0x70, 0x72, 0x65)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Previous)
	if err != nil {
		err = msgp.WrapError(err, "Previous")
		return
	}
	// write "s"
	err = en.Append(0xa1, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Services)))
	if err != nil {
		err = msgp.WrapError(err, "Services")
		return
	}
	for za0001 := range z.Services {
		err = z.Services[za0001].EncodeMsg(en)
		if err != nil {
			err = msgp.WrapError(err, "Services", za0001)
			return
		}
	}
	// write "t1"
	err = en.Append(0xa2, 0x74, 0x31)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.SignDate)
	if err != nil {
		err = msgp.WrapError(err, "SignDate")
		return
	}
	// write "t3"
	err = en.Append(0xa2, 0x74, 0x33)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.StartDate)
	if err != nil {
		err = msgp.WrapError(err, "StartDate")
		return
	}
	// write "t4"
	err = en.Append(0xa2, 0x74, 0x34)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.EndDate)
	if err != nil {
		err = msgp.WrapError(err, "EndDate")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *CreateContractParam) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 7
	// string "pa"
	o = append(o, 0x87, 0xa2, 0x70, 0x61)
	o, err = z.PartyA.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "PartyA")
		return
	}
	// string "pb"
	o = append(o, 0xa2, 0x70, 0x62)
	o, err = z.PartyB.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "PartyB")
		return
	}
	// string "pre"
	o = append(o, 0xa3, 0x70, 0x72, 0x65)
	o, err = msgp.AppendExtension(o, &z.Previous)
	if err != nil {
		err = msgp.WrapError(err, "Previous")
		return
	}
	// string "s"
	o = append(o, 0xa1, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Services)))
	for za0001 := range z.Services {
		o, err = z.Services[za0001].MarshalMsg(o)
		if err != nil {
			err = msgp.WrapError(err, "Services", za0001)
			return
		}
	}
	// string "t1"
	o = append(o, 0xa2, 0x74, 0x31)
	o = msgp.AppendInt64(o, z.SignDate)
	// string "t3"
	o = append(o, 0xa2, 0x74, 0x33)
	o = msgp.AppendInt64(o, z.StartDate)
	// string "t4"
	o = append(o, 0xa2, 0x74, 0x34)
	o = msgp.AppendInt64(o, z.EndDate)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CreateContractParam) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "pa":
			bts, err = z.PartyA.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "PartyA")
				return
			}
		case "pb":
			bts, err = z.PartyB.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "PartyB")
				return
			}
		case "pre":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Previous)
			if err != nil {
				err = msgp.WrapError(err, "Previous")
				return
			}
		case "s":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Services")
				return
			}
			if cap(z.Services) >= int(zb0002) {
				z.Services = (z.Services)[:zb0002]
			} else {
				z.Services = make([]ContractService, zb0002)
			}
			for za0001 := range z.Services {
				bts, err = z.Services[za0001].UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "Services", za0001)
					return
				}
			}
		case "t1":
			z.SignDate, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "SignDate")
				return
			}
		case "t3":
			z.StartDate, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "StartDate")
				return
			}
		case "t4":
			z.EndDate, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "EndDate")
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
func (z *CreateContractParam) Msgsize() (s int) {
	s = 1 + 3 + z.PartyA.Msgsize() + 3 + z.PartyB.Msgsize() + 4 + msgp.ExtensionPrefixSize + z.Previous.Len() + 2 + msgp.ArrayHeaderSize
	for za0001 := range z.Services {
		s += z.Services[za0001].Msgsize()
	}
	s += 3 + msgp.Int64Size + 3 + msgp.Int64Size + 3 + msgp.Int64Size
	return
}

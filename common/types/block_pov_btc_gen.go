package types

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *PovBtcHeader) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "v":
			z.Version, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Version")
				return
			}
		case "p":
			err = dc.ReadExtension(&z.Previous)
			if err != nil {
				err = msgp.WrapError(err, "Previous")
				return
			}
		case "mr":
			err = dc.ReadExtension(&z.MerkleRoot)
			if err != nil {
				err = msgp.WrapError(err, "MerkleRoot")
				return
			}
		case "ts":
			z.Timestamp, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Timestamp")
				return
			}
		case "b":
			z.Bits, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Bits")
				return
			}
		case "n":
			z.Nonce, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Nonce")
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
func (z *PovBtcHeader) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "v"
	err = en.Append(0x86, 0xa1, 0x76)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Version)
	if err != nil {
		err = msgp.WrapError(err, "Version")
		return
	}
	// write "p"
	err = en.Append(0xa1, 0x70)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Previous)
	if err != nil {
		err = msgp.WrapError(err, "Previous")
		return
	}
	// write "mr"
	err = en.Append(0xa2, 0x6d, 0x72)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.MerkleRoot)
	if err != nil {
		err = msgp.WrapError(err, "MerkleRoot")
		return
	}
	// write "ts"
	err = en.Append(0xa2, 0x74, 0x73)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	// write "b"
	err = en.Append(0xa1, 0x62)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Bits)
	if err != nil {
		err = msgp.WrapError(err, "Bits")
		return
	}
	// write "n"
	err = en.Append(0xa1, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Nonce)
	if err != nil {
		err = msgp.WrapError(err, "Nonce")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PovBtcHeader) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "v"
	o = append(o, 0x86, 0xa1, 0x76)
	o = msgp.AppendUint32(o, z.Version)
	// string "p"
	o = append(o, 0xa1, 0x70)
	o, err = msgp.AppendExtension(o, &z.Previous)
	if err != nil {
		err = msgp.WrapError(err, "Previous")
		return
	}
	// string "mr"
	o = append(o, 0xa2, 0x6d, 0x72)
	o, err = msgp.AppendExtension(o, &z.MerkleRoot)
	if err != nil {
		err = msgp.WrapError(err, "MerkleRoot")
		return
	}
	// string "ts"
	o = append(o, 0xa2, 0x74, 0x73)
	o = msgp.AppendUint32(o, z.Timestamp)
	// string "b"
	o = append(o, 0xa1, 0x62)
	o = msgp.AppendUint32(o, z.Bits)
	// string "n"
	o = append(o, 0xa1, 0x6e)
	o = msgp.AppendUint32(o, z.Nonce)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovBtcHeader) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "v":
			z.Version, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Version")
				return
			}
		case "p":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Previous)
			if err != nil {
				err = msgp.WrapError(err, "Previous")
				return
			}
		case "mr":
			bts, err = msgp.ReadExtensionBytes(bts, &z.MerkleRoot)
			if err != nil {
				err = msgp.WrapError(err, "MerkleRoot")
				return
			}
		case "ts":
			z.Timestamp, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Timestamp")
				return
			}
		case "b":
			z.Bits, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Bits")
				return
			}
		case "n":
			z.Nonce, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Nonce")
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
func (z *PovBtcHeader) Msgsize() (s int) {
	s = 1 + 2 + msgp.Uint32Size + 2 + msgp.ExtensionPrefixSize + z.Previous.Len() + 3 + msgp.ExtensionPrefixSize + z.MerkleRoot.Len() + 3 + msgp.Uint32Size + 2 + msgp.Uint32Size + 2 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PovBtcOutPoint) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "h":
			err = dc.ReadExtension(&z.Hash)
			if err != nil {
				err = msgp.WrapError(err, "Hash")
				return
			}
		case "i":
			z.Index, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Index")
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
func (z PovBtcOutPoint) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "h"
	err = en.Append(0x82, 0xa1, 0x68)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Hash)
	if err != nil {
		err = msgp.WrapError(err, "Hash")
		return
	}
	// write "i"
	err = en.Append(0xa1, 0x69)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Index)
	if err != nil {
		err = msgp.WrapError(err, "Index")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z PovBtcOutPoint) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "h"
	o = append(o, 0x82, 0xa1, 0x68)
	o, err = msgp.AppendExtension(o, &z.Hash)
	if err != nil {
		err = msgp.WrapError(err, "Hash")
		return
	}
	// string "i"
	o = append(o, 0xa1, 0x69)
	o = msgp.AppendUint32(o, z.Index)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovBtcOutPoint) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "h":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Hash)
			if err != nil {
				err = msgp.WrapError(err, "Hash")
				return
			}
		case "i":
			z.Index, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Index")
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
func (z PovBtcOutPoint) Msgsize() (s int) {
	s = 1 + 2 + msgp.ExtensionPrefixSize + z.Hash.Len() + 2 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PovBtcTx) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "v":
			z.Version, err = dc.ReadInt32()
			if err != nil {
				err = msgp.WrapError(err, "Version")
				return
			}
		case "ti":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				err = msgp.WrapError(err, "TxIn")
				return
			}
			if cap(z.TxIn) >= int(zb0002) {
				z.TxIn = (z.TxIn)[:zb0002]
			} else {
				z.TxIn = make([]*PovBtcTxIn, zb0002)
			}
			for za0001 := range z.TxIn {
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						err = msgp.WrapError(err, "TxIn", za0001)
						return
					}
					z.TxIn[za0001] = nil
				} else {
					if z.TxIn[za0001] == nil {
						z.TxIn[za0001] = new(PovBtcTxIn)
					}
					err = z.TxIn[za0001].DecodeMsg(dc)
					if err != nil {
						err = msgp.WrapError(err, "TxIn", za0001)
						return
					}
				}
			}
		case "to":
			var zb0003 uint32
			zb0003, err = dc.ReadArrayHeader()
			if err != nil {
				err = msgp.WrapError(err, "TxOut")
				return
			}
			if cap(z.TxOut) >= int(zb0003) {
				z.TxOut = (z.TxOut)[:zb0003]
			} else {
				z.TxOut = make([]*PovBtcTxOut, zb0003)
			}
			for za0002 := range z.TxOut {
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						err = msgp.WrapError(err, "TxOut", za0002)
						return
					}
					z.TxOut[za0002] = nil
				} else {
					if z.TxOut[za0002] == nil {
						z.TxOut[za0002] = new(PovBtcTxOut)
					}
					var zb0004 uint32
					zb0004, err = dc.ReadMapHeader()
					if err != nil {
						err = msgp.WrapError(err, "TxOut", za0002)
						return
					}
					for zb0004 > 0 {
						zb0004--
						field, err = dc.ReadMapKeyPtr()
						if err != nil {
							err = msgp.WrapError(err, "TxOut", za0002)
							return
						}
						switch msgp.UnsafeString(field) {
						case "v":
							z.TxOut[za0002].Value, err = dc.ReadInt64()
							if err != nil {
								err = msgp.WrapError(err, "TxOut", za0002, "Value")
								return
							}
						case "pks":
							z.TxOut[za0002].PkScript, err = dc.ReadBytes(z.TxOut[za0002].PkScript)
							if err != nil {
								err = msgp.WrapError(err, "TxOut", za0002, "PkScript")
								return
							}
						default:
							err = dc.Skip()
							if err != nil {
								err = msgp.WrapError(err, "TxOut", za0002)
								return
							}
						}
					}
				}
			}
		case "lt":
			z.LockTime, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "LockTime")
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
func (z *PovBtcTx) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "v"
	err = en.Append(0x84, 0xa1, 0x76)
	if err != nil {
		return
	}
	err = en.WriteInt32(z.Version)
	if err != nil {
		err = msgp.WrapError(err, "Version")
		return
	}
	// write "ti"
	err = en.Append(0xa2, 0x74, 0x69)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.TxIn)))
	if err != nil {
		err = msgp.WrapError(err, "TxIn")
		return
	}
	for za0001 := range z.TxIn {
		if z.TxIn[za0001] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = z.TxIn[za0001].EncodeMsg(en)
			if err != nil {
				err = msgp.WrapError(err, "TxIn", za0001)
				return
			}
		}
	}
	// write "to"
	err = en.Append(0xa2, 0x74, 0x6f)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.TxOut)))
	if err != nil {
		err = msgp.WrapError(err, "TxOut")
		return
	}
	for za0002 := range z.TxOut {
		if z.TxOut[za0002] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			// map header, size 2
			// write "v"
			err = en.Append(0x82, 0xa1, 0x76)
			if err != nil {
				return
			}
			err = en.WriteInt64(z.TxOut[za0002].Value)
			if err != nil {
				err = msgp.WrapError(err, "TxOut", za0002, "Value")
				return
			}
			// write "pks"
			err = en.Append(0xa3, 0x70, 0x6b, 0x73)
			if err != nil {
				return
			}
			err = en.WriteBytes(z.TxOut[za0002].PkScript)
			if err != nil {
				err = msgp.WrapError(err, "TxOut", za0002, "PkScript")
				return
			}
		}
	}
	// write "lt"
	err = en.Append(0xa2, 0x6c, 0x74)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.LockTime)
	if err != nil {
		err = msgp.WrapError(err, "LockTime")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PovBtcTx) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "v"
	o = append(o, 0x84, 0xa1, 0x76)
	o = msgp.AppendInt32(o, z.Version)
	// string "ti"
	o = append(o, 0xa2, 0x74, 0x69)
	o = msgp.AppendArrayHeader(o, uint32(len(z.TxIn)))
	for za0001 := range z.TxIn {
		if z.TxIn[za0001] == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = z.TxIn[za0001].MarshalMsg(o)
			if err != nil {
				err = msgp.WrapError(err, "TxIn", za0001)
				return
			}
		}
	}
	// string "to"
	o = append(o, 0xa2, 0x74, 0x6f)
	o = msgp.AppendArrayHeader(o, uint32(len(z.TxOut)))
	for za0002 := range z.TxOut {
		if z.TxOut[za0002] == nil {
			o = msgp.AppendNil(o)
		} else {
			// map header, size 2
			// string "v"
			o = append(o, 0x82, 0xa1, 0x76)
			o = msgp.AppendInt64(o, z.TxOut[za0002].Value)
			// string "pks"
			o = append(o, 0xa3, 0x70, 0x6b, 0x73)
			o = msgp.AppendBytes(o, z.TxOut[za0002].PkScript)
		}
	}
	// string "lt"
	o = append(o, 0xa2, 0x6c, 0x74)
	o = msgp.AppendUint32(o, z.LockTime)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovBtcTx) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "v":
			z.Version, bts, err = msgp.ReadInt32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Version")
				return
			}
		case "ti":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "TxIn")
				return
			}
			if cap(z.TxIn) >= int(zb0002) {
				z.TxIn = (z.TxIn)[:zb0002]
			} else {
				z.TxIn = make([]*PovBtcTxIn, zb0002)
			}
			for za0001 := range z.TxIn {
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					z.TxIn[za0001] = nil
				} else {
					if z.TxIn[za0001] == nil {
						z.TxIn[za0001] = new(PovBtcTxIn)
					}
					bts, err = z.TxIn[za0001].UnmarshalMsg(bts)
					if err != nil {
						err = msgp.WrapError(err, "TxIn", za0001)
						return
					}
				}
			}
		case "to":
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "TxOut")
				return
			}
			if cap(z.TxOut) >= int(zb0003) {
				z.TxOut = (z.TxOut)[:zb0003]
			} else {
				z.TxOut = make([]*PovBtcTxOut, zb0003)
			}
			for za0002 := range z.TxOut {
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					z.TxOut[za0002] = nil
				} else {
					if z.TxOut[za0002] == nil {
						z.TxOut[za0002] = new(PovBtcTxOut)
					}
					var zb0004 uint32
					zb0004, bts, err = msgp.ReadMapHeaderBytes(bts)
					if err != nil {
						err = msgp.WrapError(err, "TxOut", za0002)
						return
					}
					for zb0004 > 0 {
						zb0004--
						field, bts, err = msgp.ReadMapKeyZC(bts)
						if err != nil {
							err = msgp.WrapError(err, "TxOut", za0002)
							return
						}
						switch msgp.UnsafeString(field) {
						case "v":
							z.TxOut[za0002].Value, bts, err = msgp.ReadInt64Bytes(bts)
							if err != nil {
								err = msgp.WrapError(err, "TxOut", za0002, "Value")
								return
							}
						case "pks":
							z.TxOut[za0002].PkScript, bts, err = msgp.ReadBytesBytes(bts, z.TxOut[za0002].PkScript)
							if err != nil {
								err = msgp.WrapError(err, "TxOut", za0002, "PkScript")
								return
							}
						default:
							bts, err = msgp.Skip(bts)
							if err != nil {
								err = msgp.WrapError(err, "TxOut", za0002)
								return
							}
						}
					}
				}
			}
		case "lt":
			z.LockTime, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "LockTime")
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
func (z *PovBtcTx) Msgsize() (s int) {
	s = 1 + 2 + msgp.Int32Size + 3 + msgp.ArrayHeaderSize
	for za0001 := range z.TxIn {
		if z.TxIn[za0001] == nil {
			s += msgp.NilSize
		} else {
			s += z.TxIn[za0001].Msgsize()
		}
	}
	s += 3 + msgp.ArrayHeaderSize
	for za0002 := range z.TxOut {
		if z.TxOut[za0002] == nil {
			s += msgp.NilSize
		} else {
			s += 1 + 2 + msgp.Int64Size + 4 + msgp.BytesPrefixSize + len(z.TxOut[za0002].PkScript)
		}
	}
	s += 3 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PovBtcTxIn) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "pop":
			var zb0002 uint32
			zb0002, err = dc.ReadMapHeader()
			if err != nil {
				err = msgp.WrapError(err, "PreviousOutPoint")
				return
			}
			for zb0002 > 0 {
				zb0002--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					err = msgp.WrapError(err, "PreviousOutPoint")
					return
				}
				switch msgp.UnsafeString(field) {
				case "h":
					err = dc.ReadExtension(&z.PreviousOutPoint.Hash)
					if err != nil {
						err = msgp.WrapError(err, "PreviousOutPoint", "Hash")
						return
					}
				case "i":
					z.PreviousOutPoint.Index, err = dc.ReadUint32()
					if err != nil {
						err = msgp.WrapError(err, "PreviousOutPoint", "Index")
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						err = msgp.WrapError(err, "PreviousOutPoint")
						return
					}
				}
			}
		case "ss":
			z.SignatureScript, err = dc.ReadBytes(z.SignatureScript)
			if err != nil {
				err = msgp.WrapError(err, "SignatureScript")
				return
			}
		case "s":
			z.Sequence, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Sequence")
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
func (z *PovBtcTxIn) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "pop"
	// map header, size 2
	// write "h"
	err = en.Append(0x83, 0xa3, 0x70, 0x6f, 0x70, 0x82, 0xa1, 0x68)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.PreviousOutPoint.Hash)
	if err != nil {
		err = msgp.WrapError(err, "PreviousOutPoint", "Hash")
		return
	}
	// write "i"
	err = en.Append(0xa1, 0x69)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.PreviousOutPoint.Index)
	if err != nil {
		err = msgp.WrapError(err, "PreviousOutPoint", "Index")
		return
	}
	// write "ss"
	err = en.Append(0xa2, 0x73, 0x73)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.SignatureScript)
	if err != nil {
		err = msgp.WrapError(err, "SignatureScript")
		return
	}
	// write "s"
	err = en.Append(0xa1, 0x73)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Sequence)
	if err != nil {
		err = msgp.WrapError(err, "Sequence")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PovBtcTxIn) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "pop"
	// map header, size 2
	// string "h"
	o = append(o, 0x83, 0xa3, 0x70, 0x6f, 0x70, 0x82, 0xa1, 0x68)
	o, err = msgp.AppendExtension(o, &z.PreviousOutPoint.Hash)
	if err != nil {
		err = msgp.WrapError(err, "PreviousOutPoint", "Hash")
		return
	}
	// string "i"
	o = append(o, 0xa1, 0x69)
	o = msgp.AppendUint32(o, z.PreviousOutPoint.Index)
	// string "ss"
	o = append(o, 0xa2, 0x73, 0x73)
	o = msgp.AppendBytes(o, z.SignatureScript)
	// string "s"
	o = append(o, 0xa1, 0x73)
	o = msgp.AppendUint32(o, z.Sequence)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovBtcTxIn) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "pop":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "PreviousOutPoint")
				return
			}
			for zb0002 > 0 {
				zb0002--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					err = msgp.WrapError(err, "PreviousOutPoint")
					return
				}
				switch msgp.UnsafeString(field) {
				case "h":
					bts, err = msgp.ReadExtensionBytes(bts, &z.PreviousOutPoint.Hash)
					if err != nil {
						err = msgp.WrapError(err, "PreviousOutPoint", "Hash")
						return
					}
				case "i":
					z.PreviousOutPoint.Index, bts, err = msgp.ReadUint32Bytes(bts)
					if err != nil {
						err = msgp.WrapError(err, "PreviousOutPoint", "Index")
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						err = msgp.WrapError(err, "PreviousOutPoint")
						return
					}
				}
			}
		case "ss":
			z.SignatureScript, bts, err = msgp.ReadBytesBytes(bts, z.SignatureScript)
			if err != nil {
				err = msgp.WrapError(err, "SignatureScript")
				return
			}
		case "s":
			z.Sequence, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Sequence")
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
func (z *PovBtcTxIn) Msgsize() (s int) {
	s = 1 + 4 + 1 + 2 + msgp.ExtensionPrefixSize + z.PreviousOutPoint.Hash.Len() + 2 + msgp.Uint32Size + 3 + msgp.BytesPrefixSize + len(z.SignatureScript) + 2 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PovBtcTxOut) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "v":
			z.Value, err = dc.ReadInt64()
			if err != nil {
				err = msgp.WrapError(err, "Value")
				return
			}
		case "pks":
			z.PkScript, err = dc.ReadBytes(z.PkScript)
			if err != nil {
				err = msgp.WrapError(err, "PkScript")
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
func (z *PovBtcTxOut) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "v"
	err = en.Append(0x82, 0xa1, 0x76)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.Value)
	if err != nil {
		err = msgp.WrapError(err, "Value")
		return
	}
	// write "pks"
	err = en.Append(0xa3, 0x70, 0x6b, 0x73)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.PkScript)
	if err != nil {
		err = msgp.WrapError(err, "PkScript")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PovBtcTxOut) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "v"
	o = append(o, 0x82, 0xa1, 0x76)
	o = msgp.AppendInt64(o, z.Value)
	// string "pks"
	o = append(o, 0xa3, 0x70, 0x6b, 0x73)
	o = msgp.AppendBytes(o, z.PkScript)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PovBtcTxOut) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "v":
			z.Value, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Value")
				return
			}
		case "pks":
			z.PkScript, bts, err = msgp.ReadBytesBytes(bts, z.PkScript)
			if err != nil {
				err = msgp.WrapError(err, "PkScript")
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
func (z *PovBtcTxOut) Msgsize() (s int) {
	s = 1 + 2 + msgp.Int64Size + 4 + msgp.BytesPrefixSize + len(z.PkScript)
	return
}
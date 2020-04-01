package types

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *ContractValue) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "p":
			err = dc.ReadExtension(&z.BlockHash)
			if err != nil {
				err = msgp.WrapError(err, "BlockHash")
				return
			}
		case "r":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "Root")
					return
				}
				z.Root = nil
			} else {
				if z.Root == nil {
					z.Root = new(Hash)
				}
				err = dc.ReadExtension(z.Root)
				if err != nil {
					err = msgp.WrapError(err, "Root")
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
func (z *ContractValue) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "p"
	err = en.Append(0x82, 0xa1, 0x70)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.BlockHash)
	if err != nil {
		err = msgp.WrapError(err, "BlockHash")
		return
	}
	// write "r"
	err = en.Append(0xa1, 0x72)
	if err != nil {
		return
	}
	if z.Root == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteExtension(z.Root)
		if err != nil {
			err = msgp.WrapError(err, "Root")
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ContractValue) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "p"
	o = append(o, 0x82, 0xa1, 0x70)
	o, err = msgp.AppendExtension(o, &z.BlockHash)
	if err != nil {
		err = msgp.WrapError(err, "BlockHash")
		return
	}
	// string "r"
	o = append(o, 0xa1, 0x72)
	if z.Root == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = msgp.AppendExtension(o, z.Root)
		if err != nil {
			err = msgp.WrapError(err, "Root")
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ContractValue) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "p":
			bts, err = msgp.ReadExtensionBytes(bts, &z.BlockHash)
			if err != nil {
				err = msgp.WrapError(err, "BlockHash")
				return
			}
		case "r":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Root = nil
			} else {
				if z.Root == nil {
					z.Root = new(Hash)
				}
				bts, err = msgp.ReadExtensionBytes(bts, z.Root)
				if err != nil {
					err = msgp.WrapError(err, "Root")
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
func (z *ContractValue) Msgsize() (s int) {
	s = 1 + 2 + msgp.ExtensionPrefixSize + z.BlockHash.Len() + 2
	if z.Root == nil {
		s += msgp.NilSize
	} else {
		s += msgp.ExtensionPrefixSize + z.Root.Len()
	}
	return
}
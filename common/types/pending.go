package types

//go:generate msgp
type PendingKey struct {
	Address Address `msg:"account,extension" json:"account"`
	Hash    Hash    `msg:"hash,extension" json:"hash"`
}

//go:generate msgp
type PendingInfo struct {
	Source Address `msg:"source,extension" json:"source"`
	Amount Balance `msg:"amount,extension" json:"amount"`
	Type   Hash    `msg:"type,extension" json:"type"`
}

func (pk *PendingKey) Serialize() ([]byte, error) {
	var b []byte
	b = append(b, pk.Address.Bytes()...)
	b = append(b, pk.Hash.Bytes()...)
	return b, nil
}

func (pk *PendingKey) Deserialize(text []byte) error {
	addr, err := BytesToAddress(text[:AddressSize])
	if err != nil {
		return err
	}
	hash, err := BytesToHash(text[AddressSize:])
	if err != nil {
		return err
	}
	pk.Address = addr
	pk.Hash = hash
	return nil
}

func (pi *PendingInfo) Serialize() ([]byte, error) {
	return pi.MarshalMsg(nil)
}

func (pi *PendingInfo) Deserialize(text []byte) error {
	_, err := pi.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PendingKind byte

const (
	PendingNotUsed PendingKind = iota
	PendingUsed
)

//
//type Pending struct {
//	*PendingKey
//	*PendingInfo
//}
//
//func (p *Pending) TableName() string {
//	return "PENDING"
//}
//
//func (p *Pending) TableSchema() string {
//	str := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
//		(   id integer PRIMARY KEY AUTOINCREMENT,
//			address %s,
//			hash %s,
//			source %s,
//			amount %s,
//			type %s
//		)`,
//		p.TableName(),
//		ConvertSchemaType(p.Address),
//		ConvertSchemaType(p.Hash),
//		ConvertSchemaType(p.Source),
//		ConvertSchemaType(p.Amount),
//		ConvertSchemaType(p.Type),
//	)
//	return str
//}
//
//func (p *Pending) SetRelation() (string, []interface{}) {
//	var str string
//	str = fmt.Sprintf("INSERT INTO %s (address, hash,source,amount,type) VALUES ($1, $2, $3, $4, $5)", p.TableName())
//	val := []interface{}{p.Address.String(), p.Hash.String(), p.Source.String(), p.Amount, p.Type.String()}
//	return str, val
//}
//
//func (p *Pending) RemoveRelation() string {
//	return fmt.Sprintf("DELETE FROM %s WHERE hash = '%s'", p.TableName(), p.Hash.String())
//}

package types

const (
	FrontierSize = AddressSize + HashSize
)

//go:generate msgp
type Frontier struct {
	Address Address `msg:"address,extension" json:"address"`
	Hash    Hash    `msg:"hash,extension" json:"hash"`
}

func (f *Frontier) IsZero() bool {
	return f.Hash.IsZero()
}

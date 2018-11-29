package types

const (
	FrontierSize = AddressSize + HashSize
)

//go:generate msgp
type Frontier struct {
	HeaderBlock Hash `msg:"address,extension" json:"headerblock"`
	OpenBlock   Hash `msg:"hash,extension" json:"openblock"`
}

func (f *Frontier) IsZero() bool {
	return f.HeaderBlock.IsZero()
}

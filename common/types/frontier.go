package types

const (
	FrontierSize = AddressSize + HashSize
)

//go:generate msgp
type Frontier struct {
	HeaderBlock Hash `msg:"headerblock,extension" json:"headerblock"`
	OpenBlock   Hash `msg:"openblock,extension" json:"openblock"`
}

func (f *Frontier) IsZero() bool {
	return f.HeaderBlock.IsZero()
}

type Frontiers []*Frontier

func (fs Frontiers) Len() int {
	return len(fs)
}
func (fs Frontiers) Less(i, j int) bool {
	return fs[i].OpenBlock.String() < fs[j].OpenBlock.String()
}
func (fs Frontiers) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}

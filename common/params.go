package common

var (
	//vote right divisor
	DposVoteDivisor  = int64(200)
)

type nodeType uint

const (
	nodeTypeNormal nodeType = iota
	nodeTypeConfidant
)

func IsConfidantNode() bool {
	if NodeType == nodeTypeConfidant {
		return true
	}
	return false
}

func IsNormalNode() bool {
	if NodeType == nodeTypeNormal {
		return true
	}
	return false
}

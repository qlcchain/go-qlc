package common

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

const (
	//vote right divisor
	DposVoteDivisor = int64(200)

	DPosOnlinePeriod        = uint64(10)
	DPosOnlineRate          = uint64(60)
	DPosHeartCountPerPeriod = DPosOnlinePeriod / 2
	DPosOnlineSectionLeft   = 3
	DPosOnlineSectionRight  = 8
)

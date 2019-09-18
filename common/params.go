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

	DPosOnlinePeriod        = uint64(20)
	DPosOnlineRate          = uint64(60)
	DPosHeartCountPerPeriod = DPosOnlinePeriod / 2
	DPosOnlineSectionLeft   = 5
	DPosOnlineSectionRight  = 15
)

package common

import "strings"

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

	DPosOnlinePeriod        = uint64(120)
	DPosOnlineRate          = uint64(60)
	DPosHeartCountPerPeriod = DPosOnlinePeriod / 2
	DPosOnlineSectionLeft   = 30
	DPosOnlineSectionRight  = 90
)

var testMode string

func SetTestMode(mode string) {
	testMode = mode
}

func GetTestMode() string {
	return testMode
}

func CheckTestMode(flag string) bool {
	return strings.Contains(testMode, flag)
}

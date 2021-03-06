// Code generated by "stringer -type=ServiceStatus"; DO NOT EDIT.

package common

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Origin-0]
	_ = x[Initialing-1]
	_ = x[Inited-2]
	_ = x[Starting-3]
	_ = x[Started-4]
	_ = x[Stopping-5]
	_ = x[Stopped-6]
}

const _ServiceStatus_name = "OriginInitialingInitedStartingStartedStoppingStopped"

var _ServiceStatus_index = [...]uint8{0, 6, 16, 22, 30, 37, 45, 52}

func (i ServiceStatus) String() string {
	if i < 0 || i >= ServiceStatus(len(_ServiceStatus_index)-1) {
		return "ServiceStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ServiceStatus_name[_ServiceStatus_index[i]:_ServiceStatus_index[i+1]]
}

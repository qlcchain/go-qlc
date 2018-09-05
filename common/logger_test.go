package common

import "testing"

func TestNewLogger(t *testing.T) {
	logger1 := NewLogger("test1")
	logger1.Debug("debug1")
}

package common

import "testing"

func TestIsNormalNode(t *testing.T) {
	NodeType = nodeTypeNormal
	if !IsNormalNode() {
		t.Fatal()
	}

	NodeType = nodeTypeConfidant
	if !IsConfidantNode() {
		t.Fatal()
	}
}

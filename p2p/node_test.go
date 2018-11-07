package p2p

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/config"
)

func TestQlcNode_StartServices(t *testing.T) {
	cfg := config.DefaultlConfig
	node, err := NewNode(cfg)
	if err != nil {
		fmt.Println(err)
	}
	node.startHost()
}

package proto

import (
	"fmt"
	"testing"
)

func TestNewAccountAPIClient(t *testing.T) {
	createRequest := CreateRequest{}
	fmt.Println(createRequest.String())
}

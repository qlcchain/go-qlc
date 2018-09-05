package config

import (
	"testing"
	"fmt"
)

func TestInitConfig(t *testing.T) {
	InitConfig()
}
func TestDefaultConfig(t *testing.T) {
	DefaultConfig()
}
func TestGetBoolVaule(t *testing.T) {
	InitConfig()
	a := GetBoolVaule("rpc.rpc_enable")
	fmt.Println(a)
}
func TestGetIntVaule(t *testing.T) {
	InitConfig()
	a := GetIntVaule("port")
	fmt.Println(a)
}
func TestGetStringVaule(t *testing.T) {
	InitConfig()
	a := GetStringVaule("bootstrap_peers")
	fmt.Println(a)
}
func TestGetFloatVaule(t *testing.T) {

}
func TestSetBoolVaule(t *testing.T) {
	InitConfig()
	a := GetBoolVaule("rpc.rpc_enable")
	fmt.Println(a)
	SetBoolVaule("rpc.rpc_enable", false)
	b := GetBoolVaule("rpc.rpc_enable")
	fmt.Println(b)
}
func TestSetIntVaule(t *testing.T) {
	InitConfig()
	a := GetIntVaule("port")
	fmt.Println(a)
	SetIntVaule("port", 39734)
	b := GetIntVaule("port")
	fmt.Println(b)
}
func TestSetStringVaule(t *testing.T) {
	InitConfig()
	a := GetStringVaule("bootstrap_peers")
	fmt.Println(a)
	SetStringVaule("bootstrap_peers", "47.91.166.18")
	b := GetStringVaule("bootstrap_peers")
	fmt.Println(b)
}
func TestSetFloatVaule(t *testing.T) {

}

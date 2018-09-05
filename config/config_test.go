package config

import (
	"testing"
	"fmt"
	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	InitConfig()
}
func TestDefaultConfig(t *testing.T) {
	DefaultConfig()
}
func TestGetBoolVaule(t *testing.T) {
	InitConfig()
	a := GetBoolValue("rpc.rpc_enable")
	fmt.Println(a)
}
func TestGetIntVaule(t *testing.T) {
	InitConfig()
	a := GetIntValue("port")
	fmt.Println(a)
}
func TestGetStringVaule(t *testing.T) {
	InitConfig()
	a := GetStringValue("bootstrap_peers")
	fmt.Println(a)
}
func TestGetFloatVaule(t *testing.T) {

}
func TestGetSliceValueValue(t *testing.T) {
	InitConfig()
	a := viper.GetStringSlice("bootstrap_peers")
	for _, x := range a {
		fmt.Printf("%v\n", x)
	}
}
func TestSetBoolVaule(t *testing.T) {
	InitConfig()
	a := GetBoolValue("rpc.rpc_enable")
	fmt.Println(a)
	SetBoolValue("rpc.rpc_enable", false)
	b := GetBoolValue("rpc.rpc_enable")
	fmt.Println(b)
}
func TestSetIntVaule(t *testing.T) {
	InitConfig()
	a := GetIntValue("port")
	fmt.Println(a)
	SetIntValue("port", 39734)
	b := GetIntValue("port")
	fmt.Println(b)
}
func TestSetStringVaule(t *testing.T) {
	InitConfig()
	a := GetStringValue("bootstrap_peers")
	fmt.Println(a)
	SetStringValue("bootstrap_peers", "47.91.166.18")
	b := GetStringValue("bootstrap_peers")
	fmt.Println(b)
}

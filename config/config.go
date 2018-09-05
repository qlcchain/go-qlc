package config

import (
	"os"
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
)

var configName = "qlc.json"
var configpath string
var defaultconfig = []byte(`{
	"version": "1",
	"port": 29734,
	"bootstrap_peers": [
	"47.90.89.43",
	"47.91.166.18"
	],
	"rpc": {
	"rpc_enable": true,
	"host": "127.0.0.1",
	"rpc_port": 29735
	}
}`)

func InitConfig() int {
	exit := DefaultConfig()
	if exit == false {
		return -1
	}
	viper.SetConfigName("qlc")
	viper.SetConfigType("json")
	viper.AddConfigPath(configpath)
	err := viper.ReadInConfig() // 搜索路径，并读取配置数据
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
		return -2
	}
	return 0
}
func DefaultConfig() bool {
	configpath = DefaultPath()
	if configpath == "" {
		return false
	}
	defaultfilepath := filepath.Join(configpath, configName)
	_, err := os.Stat(defaultfilepath)
	if err != nil {
		if os.IsExist(err) {
			return true
		} else {
			err := os.MkdirAll(configpath, 0711)
			if err != nil {
				return false
			}
			f, err := os.Create(defaultfilepath)
			if err != nil {
				return false
			}

			defer f.Close()

			n2, err := f.Write(defaultconfig)
			if err != nil {
				return false
			}
			fmt.Printf("wrote %d bytes\n", n2)
			return true
		}
	}
	return true
}
func GetBoolValue(key string) bool {
	return viper.GetBool(key)
}
func GetIntValue(key string) int {
	return viper.GetInt(key)
}
func GetStringValue(key string) string {
	return viper.GetString(key)
}
func GetFloatValue(key string) float64 {
	return viper.GetFloat64(key)
}
func GetSliceValue(key string) []string {
	return viper.GetStringSlice(key)
}
func SetBoolValue(key string, value bool) {
	viper.Set(key, value)
}
func SetIntValue(key string, value int) {
	viper.Set(key, value)
}
func SetStringValue(key string, value string) {
	viper.Set(key, value)
}
func SetFloatValue(key string, value float64) {
	viper.Set(key, value)
}

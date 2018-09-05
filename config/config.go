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
	"bootstrap_peers": "48.90.89.43",
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
func GetBoolVaule(key string) bool {
	return viper.GetBool(key)
}
func GetIntVaule(key string) int {
	return viper.GetInt(key)
}
func GetStringVaule(key string) string {
	return viper.GetString(key)
}
func GetFloatVaule(key string) float64 {
	return viper.GetFloat64(key)
}
func SetBoolVaule(key string,vaule bool) {
	viper.Set(key, vaule)
}
func SetIntVaule(key string,vaule int) {
	viper.Set(key, vaule)
}
func SetStringVaule( key string,vaule string) {
	viper.Set(key, vaule)
}
func SetFloatVaule(key string,vaule float64)  {
	viper.Set(key, vaule)
}
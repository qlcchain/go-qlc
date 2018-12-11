package config

import (
	"fmt"
	"os"
	"path"
)

func InitConfig() (*Config, error) {
	filename := path.Join(DefaultDataDir(), QlcConfigFile)
	var cfg *Config
	manager := NewCfgManager()
	_, err := os.Stat(filename)
	if err == nil {
		fmt.Println("config file already exit !!")
		if err = manager.Read(&cfg); err != nil {
			return nil, fmt.Errorf("config load error: %s", err)
		}
	} else {
		fmt.Println("config file not exit, use the default option")
		cfg, err = DefaultConfig()
		if err != nil {
			return nil, err
		}
		if err = manager.Write(&cfg); err != nil {
			return nil, fmt.Errorf("config write error: %s", err)
		}
	}
	return cfg, nil
}

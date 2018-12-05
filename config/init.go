package config

import (
	"os"
	"path"

	"github.com/qlcchain/go-qlc/common/util"
)

func InitConfig() (*Config, error) {
	filename := path.Join(util.QlcDir(), QlcConfigFile)
	var cfg *Config
	manager := NewCfgManager()
	_, err := os.Stat(filename)
	if err == nil {
		log.Info("config file already exit !!")
		if err = manager.Read(&cfg); err != nil {
			log.Errorf("config load error: %s", err)
			return nil, err
		}
	} else {
		log.Info("config file not exit, use the default option")
		cfg, err = DefaultConfig()
		if err != nil {
			return nil, err
		}
		if err = manager.Write(&cfg); err != nil {
			log.Errorf("config write error: %s", err)
			return nil, err
		}
	}
	return cfg, nil
}

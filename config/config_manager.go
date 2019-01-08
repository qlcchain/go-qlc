package config

import (
	"fmt"
	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

type ConfigManager struct {
	cfgFile string
}

func NewCfgManager(file string) *ConfigManager {
	cfg := &ConfigManager{
		cfgFile: file,
	}
	return cfg
}

//Load the config file and will create default if config file no exist
func (c *ConfigManager) Load(migrations ...CfgMigrate) (*Config, error) {
	_, err := os.Stat(c.cfgFile)
	if err != nil {
		fmt.Printf("%s not exist, create default\n", c.cfgFile)
		cfg, err := DefaultConfig()
		if err != nil {
			return nil, err
		}
		err = c.save(cfg)
		if err != nil {
			return nil, err
		}
	}
	bytes, err := ioutil.ReadFile(c.cfgFile)
	if err != nil {
		return nil, err
	}

	var cfg Config

	err = jsoniter.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, err
	}
	flag := false
	// update cfg file
	sort.Sort(CfgMigrations(migrations))
	for _, m := range migrations {
		version := cfg.Version
		if version == m.StartVersion() {
			err := m.Migration(&cfg)
			if err != nil {
				fmt.Println(err)
			} else {
				flag = true
			}
		}
	}
	if flag {
		_ = c.save(&cfg)
	}
	return &cfg, nil
}

func (c *ConfigManager) save(cfg *Config) error {
	dir := filepath.Dir(c.cfgFile)
	err := util.CreateDirIfNotExist(dir)
	if err != nil {
		return err
	}

	bytes, err := jsoniter.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.cfgFile, bytes, 0600)
}

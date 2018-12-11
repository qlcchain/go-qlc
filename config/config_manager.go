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
	cfgFile    string
	migrations []CfgMigrate
}

func NewCfgManager(file string) *ConfigManager {
	cfg := &ConfigManager{
		cfgFile:    file,
		migrations: make([]CfgMigrate, 0),
	}
	return cfg
}

func (c *ConfigManager) Load() (*Config, error) {
	_, err := os.Stat(c.cfgFile)
	if err != nil {
		fmt.Printf("%s not exist, create default", c.cfgFile)
		cfg, err := DefaultConfig()
		if err != nil {
			return nil, err
		}
		err = c.Save(cfg)
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
	sort.Sort(CfgMigrations(c.migrations))
	for _, m := range c.migrations {
		err := m.Migration(&cfg)
		if err != nil {
			fmt.Println(err)
		} else {
			flag = true
		}
	}
	if flag {
		_ = c.Save(&cfg)
	}
	return &cfg, nil
}

func (c *ConfigManager) Save(cfg *Config) error {
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

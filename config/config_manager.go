package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"

	"github.com/qlcchain/go-qlc/common/util"
)

type CfgManager struct {
	cfgFile string
	cfgPath string
}

func NewCfgManager(path string) *CfgManager {
	file := filepath.Join(path, QlcConfigFile)
	cfg := &CfgManager{
		cfgFile: file,
		cfgPath: path,
	}
	return cfg
}

//Load the config file and will create default if config file no exist
func (c *CfgManager) Load(migrations ...CfgMigrate) (*Config, error) {
	_, err := os.Stat(c.cfgFile)
	if err != nil {
		err := c.createAndSave()
		if err != nil {
			return nil, err
		}
	}
	bytes, err := ioutil.ReadFile(c.cfgFile)
	if err != nil {
		return nil, err
	}

	version, err := c.parseVersion(bytes)
	if err != nil {
		fmt.Println(err)
		version = configVersion
		err := c.createAndSave()
		if err != nil {
			return nil, err
		}
	}

	flag := false
	sort.Sort(CfgMigrations(migrations))
	for _, m := range migrations {
		var err error
		if version == m.StartVersion() {
			fmt.Printf("migration cfg from v%d to v%d\n", m.StartVersion(), m.EndVersion())
			bytes, version, err = m.Migration(bytes, version)
			if err != nil {
				fmt.Println(err)
			} else {
				flag = true
			}
		}
	}

	// unmarshal as latest config
	var cfg Config
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, err
	}

	if flag {
		_ = c.save(&cfg)
	}

	return &cfg, nil
}

func (c *CfgManager) createAndSave() error {
	cfg, err := DefaultConfig(c.cfgPath)
	if err != nil {
		return err
	}

	err = c.save(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (c *CfgManager) save(cfg interface{}) error {
	dir := filepath.Dir(c.cfgFile)
	err := util.CreateDirIfNotExist(dir)
	if err != nil {
		return err
	}

	//bytes, err := json.Marshal(cfg)
	//if err != nil {
	//	return err
	//}
	s := util.ToIndentString(cfg)
	return ioutil.WriteFile(c.cfgFile, []byte(s), 0600)
}

func (c *CfgManager) parseVersion(data []byte) (int, error) {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(data, &objMap)
	if err != nil {
		return 0, err
	}

	if v, ok := objMap["version"]; ok {
		var version int
		if err := json.Unmarshal([]byte(*v), &version); err == nil {
			return version, nil
		} else {
			return 0, err
		}
	} else {
		return 0, errors.New("can not find any version")
	}
}

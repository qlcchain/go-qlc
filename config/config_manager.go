package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/validator.v2"
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

func NewCfgManagerWithName(path string, name string) *CfgManager {
	file := filepath.Join(path, name)
	cfg := &CfgManager{
		cfgFile: file,
		cfgPath: path,
	}
	return cfg
}

func (cm *CfgManager) Verify() error {
	cfg, err := cm.Config()
	if err != nil {
		return err
	}
	return validator.Validate(cfg)
}

func (cm *CfgManager) UpdateParams(params []string) (*Config, error) {
	s := strings.Split(filepath.Base(cm.cfgFile), ".")
	if len(s) != 2 {
		return nil, errors.New("get config path error")
	}

	cfg, err := cm.Load()
	if err != nil {
		return nil, err
	}
	viper.SetConfigName(s[0])
	viper.AddConfigPath(cfg.DataDir)
	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)
	err = viper.ReadConfig(r)
	if err != nil {
		return nil, err
	}

	for _, param := range params {
		k := strings.Split(param, "=")
		if len(k) != 2 || len(k[0]) == 0 || len(k[1]) == 0 {
			continue
		}
		if oldValue := viper.Get(k[0]); oldValue != nil {
			viper.Set(k[0], k[1])
		}
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	err = cm.Verify()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cm *CfgManager) Config() (*Config, error) {
	if cm.cfg != nil {
		return cm.cfg, nil
	} else {
		return cm.Load()
	}
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
		//fmt.Println(err)
		//version = configVersion
		//err := c.createAndSave()
		//if err != nil {
		//	return nil, err
		//}
		return nil, fmt.Errorf("parse config Version error : %s", err)
	}

	flag := false
	sort.Slice(migrations, func(i, j int) bool {
		if migrations[i].StartVersion() < migrations[j].StartVersion() {
			return true
		}

		if migrations[i].StartVersion() > migrations[j].StartVersion() {
			return false
		}

		return migrations[i].EndVersion() < migrations[j].EndVersion()
	})
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

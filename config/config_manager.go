package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/spf13/viper"
	"gopkg.in/validator.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type CfgManager struct {
	cfgFile string
	v       *viper.Viper
	cfg     *Config
}

func NewCfgManager(path string) *CfgManager {
	return NewCfgManagerWithName(path, QlcConfigFile)
}

func NewCfgManagerWithName(path string, name string) *CfgManager {
	file := filepath.Join(path, name)
	cfg := &CfgManager{
		cfgFile: file,
		v:       viper.New(),
	}
	return cfg
}

func (cm *CfgManager) ConfigDir() string {
	return filepath.Dir(cm.cfgFile)
}

func (cm *CfgManager) verify(data interface{}) error {
	if data == nil {
		cfg, err := cm.Config()
		if err != nil {
			return err
		}
		return validator.Validate(cfg)
	}
	return validator.Validate(data)
}

func (cm *CfgManager) UpdateParams(params []string) (*Config, error) {
	for _, param := range params {
		k := strings.Split(param, "=")
		if len(k) != 2 || len(k[0]) == 0 || len(k[1]) == 0 {
			continue
		}
		if oldValue := cm.v.Get(k[0]); oldValue != nil {
			cm.v.Set(k[0], k[1])
		}
	}
	err := cm.v.Unmarshal(cm.cfg)
	if err != nil {
		return nil, err
	}

	err = cm.verify(cm.cfg)
	if err != nil {
		return nil, err
	}
	return cm.cfg, nil
}

func (cm *CfgManager) Config() (*Config, error) {
	if cm.cfg != nil {
		return cm.cfg, nil
	} else {
		return cm.Load()
	}
}

//Load the config file and will create default if config file no exist
func (cm *CfgManager) Load(migrations ...CfgMigrate) (*Config, error) {
	_, err := os.Stat(cm.cfgFile)
	if err != nil {
		err := cm.createAndSave()
		if err != nil {
			return nil, err
		}
	}
	content, err := ioutil.ReadFile(cm.cfgFile)
	if err != nil {
		return nil, err
	}

	version, err := cm.parseVersion(content)
	if err != nil {
		fmt.Printf("parse config Version error : %s\n", err)
		// backup and create new default config
		version = configVersion
		cm.backUp(content)
		err := cm.createAndSave()
		if err != nil {
			return nil, err
		}
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
			content, version, err = m.Migration(content, version)
			if err != nil {
				fmt.Println(err)
			} else {
				flag = true
			}
		}
	}

	// unmarshal as latest config
	var cfg Config
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}
	cm.cfg = &cfg
	err = cm.verify(nil)
	if err != nil {
		cm.cfg = nil
		return nil, err
	}

	if flag {
		cm.backUp(content)
		_ = cm.Save()
	}

	err = cm.viper()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cm *CfgManager) viper() error {
	// prepare viper
	s := strings.Split(filepath.Base(cm.cfgFile), ".")
	if len(s) != 2 {
		return errors.New("get config path error")
	}
	cm.v.SetConfigName(s[0])
	cm.v.AddConfigPath(cm.cfg.DataDir)
	cm.v.AddConfigPath(cm.ConfigDir())
	b, err := json.Marshal(cm.cfg)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	err = cm.v.ReadConfig(r)
	if err != nil {
		return err
	}
	return nil
}

func (cm *CfgManager) backUp(content []byte) {
	backup := filepath.Join(filepath.Dir(cm.cfgFile),
		fmt.Sprintf("qlc_back_%s.json", time.Now().Format("2006-01-02T15-04")))
	_ = ioutil.WriteFile(backup, content, 0600)
}

func (cm *CfgManager) createAndSave() error {
	cfg, err := DefaultConfig(filepath.Dir(cm.cfgFile))
	if err != nil {
		return err
	}

	cm.cfg = cfg
	err = cm.Save()
	if err != nil {
		return err
	}

	return nil
}

// Save write config to file
func (cm *CfgManager) Save(data ...interface{}) error {
	dir := filepath.Dir(cm.cfgFile)
	err := util.CreateDirIfNotExist(dir)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		cfg, err := cm.Config()
		if err != nil {
			return err
		}
		s := util.ToIndentString(cfg)
		return ioutil.WriteFile(cm.cfgFile, []byte(s), 0600)
	}

	s := util.ToIndentString(data[0])
	return ioutil.WriteFile(cm.cfgFile, []byte(s), 0600)
}

func (cm *CfgManager) parseVersion(data []byte) (int, error) {
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

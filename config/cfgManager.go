package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/qlcchain/go-qlc/common/util"
)

type ConfigManager struct {
	dir      string
	filename string
}

func NewCfgManager() *ConfigManager {
	rootpath := util.QlcDir()
	filename := path.Join(rootpath, QlcConfigFile)
	cfg := &ConfigManager{
		dir:      util.QlcDir(),
		filename: filename,
	}
	return cfg
}

func (m *ConfigManager) Write(i interface{}) error {
	if _, err := os.Stat(m.dir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(m.dir, 0700); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if _, err := os.Stat(m.filename); err != nil {
		if os.IsNotExist(err) {
			return m.Save(i)
		}
		return err
	}
	return nil
}

func (m *ConfigManager) Read(i interface{}) error {
	bytes, err := ioutil.ReadFile(m.filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, i); err != nil {
		return err
	}

	return nil
}

func (m *ConfigManager) Save(i interface{}) error {
	bytes, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(m.filename, bytes, 0600)
}

func (m *ConfigManager) Dir() string {
	return m.dir
}

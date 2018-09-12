package config

import (
	"os/user"
	"os"
	"path"
	"io/ioutil"
	"encoding/json"
)

type ConfigManager struct {
	dir      string
	filename string
}

var defaultpath = &ConfigManager{
	dir:      ".qlcchain",
	filename: "qlc.json",
}

func NewCfgManager(dir, filename string) (*ConfigManager, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	if dir == "" || filename == "" {
		dir = defaultpath.dir
		filename = defaultpath.filename
	}
	cfg := &ConfigManager{dir: path.Join(usr.HomeDir, dir)}
	cfg.filename = path.Join(cfg.dir, filename)
	return cfg, nil
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

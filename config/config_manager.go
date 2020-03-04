package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"gopkg.in/validator.v2"

	"github.com/qlcchain/go-qlc/common/util"
)

type CfgManager struct {
	ConfigFile string
	v          *viper.Viper
	cfg        *Config
	cfgB       *Config
	locker     sync.Mutex
	isDirty    *atomic.Bool
}

func NewCfgManager(path string) *CfgManager {
	return NewCfgManagerWithName(path, QlcConfigFile)
}

func NewCfgManagerWithFile(cfgFile string) *CfgManager {
	return NewCfgManagerWithName(filepath.Dir(cfgFile), filepath.Base(cfgFile))
}

func NewCfgManagerWithName(path string, name string) *CfgManager {
	file := filepath.Join(path, name)
	cm := &CfgManager{
		ConfigFile: file,
		locker:     sync.Mutex{},
		isDirty:    atomic.NewBool(false),
	}
	_, _ = cm.Load()
	return cm
}

func NewCfgManagerWithConfig(cfgFile string, cfg *Config) *CfgManager {
	cm := &CfgManager{
		ConfigFile: cfgFile,
		locker:     sync.Mutex{},
		isDirty:    atomic.NewBool(false),
		cfg:        cfg,
	}
	return cm
}

func (cm *CfgManager) ConfigDir() string {
	return filepath.Dir(cm.ConfigFile)
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

func (cm *CfgManager) PatchParams(params []string, cfg *Config) (*Config, error) {
	cm.locker.Lock()
	defer cm.locker.Unlock()

	v := cm.v
	if v == nil {
		v = viper.New()
	}
	s := strings.Split(filepath.Base(cm.ConfigFile), ".")
	if len(s) != 2 {
		return nil, errors.New("get config path error")
	}
	v.SetConfigName(s[0])
	v.AddConfigPath(cfg.DataDir)
	v.AddConfigPath(cm.ConfigDir())
	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)
	err = v.ReadConfig(r)
	if err != nil {
		return nil, err
	}
	for _, param := range params {
		k := strings.Split(param, "=")
		if len(k) != 2 || len(k[0]) == 0 || len(k[1]) == 0 {
			continue
		}
		// TODO: clear slice, maybe there is a issue of mapstructure
		if oldValue := v.Get(k[0]); oldValue != nil {
			v.Set(k[0], k[1])
		}
	}
	err = v.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	err = cm.verify(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cm *CfgManager) UpdateParams(params []string) (*Config, error) {
	if cm.cfgB == nil {
		cfg, err := cm.Config()
		if err != nil {
			return nil, err
		}
		cm.cfgB, err = cfg.Clone()
	}
	cm.isDirty.Store(true)
	return cm.PatchParams(params, cm.cfgB)
}

func (cm *CfgManager) Discard() {
	cm.locker.Lock()
	defer cm.locker.Unlock()

	cm.cfgB = nil
	cm.v = nil
	cm.isDirty.Store(false)
}

// Commit changed cfg to runtime
func (cm *CfgManager) Commit() error {
	cm.locker.Lock()
	defer cm.locker.Unlock()

	return cm.commitCfg()
}

func (cm *CfgManager) commitCfg() error {
	if cm.isDirty.Load() && cm.cfgB != nil {
		if cfg, err := cm.cfgB.Clone(); err == nil {
			cm.cfg = cfg
			// clear buff vars
			cm.cfgB = nil
			cm.v = nil
			cm.isDirty.Store(false)
		} else {
			return err
		}
	}

	return nil
}

// CommitAndSave commit changed cfg to runtime and save to config file
func (cm *CfgManager) CommitAndSave() error {
	cm.locker.Lock()
	defer cm.locker.Unlock()

	if err := cm.commitCfg(); err == nil {
		if err := cm.Save(cm.cfg); err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

// Config get current used config
func (cm *CfgManager) Config() (*Config, error) {
	if cm.cfg != nil {
		return cm.cfg, nil
	} else {
		return nil, fmt.Errorf("invalid cfg ,cfg path is [%s]", cm.ConfigDir())
	}
}

// ParseDataDir parse dataDir from config file
func (cm *CfgManager) ParseDataDir() (string, error) {
	_, err := os.Stat(cm.ConfigFile)
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadFile(cm.ConfigFile)
	if err != nil {
		return "", err
	}

	var objMap map[string]*json.RawMessage
	err = json.Unmarshal(content, &objMap)
	if err != nil {
		return "", err
	}

	if v, ok := objMap["dataDir"]; ok {
		var dataDir string
		if err := json.Unmarshal(*v, &dataDir); err == nil {
			return dataDir, nil
		} else {
			return "", err
		}
	} else {
		return "", errors.New("can not parse dataDir")
	}
}

// Load the config file and will create default if config file no exist
func (cm *CfgManager) Load(migrations ...CfgMigrate) (*Config, error) {
	_, err := os.Stat(cm.ConfigFile)
	if err != nil {
		err := cm.createAndSave()
		if err != nil {
			return nil, err
		}
	}
	content, err := ioutil.ReadFile(cm.ConfigFile)
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

	loadGenesisAccount(&cfg)

	return &cfg, nil
}

func loadGenesisAccount(cfg *Config) {
	if cfg.Genesis != nil && len(cfg.Genesis.GenesisBlocks) > 0 {
		if len(genesisInfos) > 0 {
			genesisInfos = genesisInfos[:0]
		}

		for _, v := range cfg.Genesis.GenesisBlocks {
			if v != nil {
				genesisInfos = append(genesisInfos, v)
				if v.ChainToken {
					genesisAddress = v.Genesis.Address
					chainToken = v.Genesis.Token
					genesisMintageBlock = v.Mintage
					genesisMintageHash = v.Mintage.GetHash()
					genesisBlock = v.Genesis
					genesisBlockHash = v.Genesis.GetHash()
				}
				if v.GasToken {
					gasAddress = v.Genesis.Address
					gasToken = v.Genesis.Token
					gasMintageBlock = v.Mintage
					gasMintageHash = v.Mintage.GetHash()
					gasBlock = v.Genesis
					gasBlockHash = v.Genesis.GetHash()
				}
			}
		}
	}
}

// DiffOther diff runtime cfg with other `cfg`
func (cm *CfgManager) DiffOther(cfg *Config) (string, error) {
	used, err := cm.Config()
	if err != nil {
		return "", err
	}
	diff := cmp.Diff(cfg, used)

	return diff, nil
}

// Diff the changed config
func (cm *CfgManager) Diff() (string, error) {
	cm.locker.Lock()
	defer cm.locker.Unlock()

	if cm.isDirty.Load() && cm.cfgB != nil {
		cfg, err := cm.Config()
		if err != nil {
			return "", err
		}
		diff := cmp.Diff(cfg, cm.cfgB)
		return diff, nil
	}

	return "", errors.New("cfg not changed")
}

func (cm *CfgManager) backUp(content []byte) {
	backup := filepath.Join(filepath.Dir(cm.ConfigFile),
		fmt.Sprintf("qlc_back_%s.json", time.Now().Format("2006-01-02T15-04")))
	_ = ioutil.WriteFile(backup, content, 0600)
}

func (cm *CfgManager) createAndSave() error {
	cfg, err := DefaultConfig(filepath.Dir(cm.ConfigFile))
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
	dir := filepath.Dir(cm.ConfigFile)
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
		return ioutil.WriteFile(cm.ConfigFile, []byte(s), 0600)
	}

	s := util.ToIndentString(data[0])
	return ioutil.WriteFile(cm.ConfigFile, []byte(s), 0600)
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

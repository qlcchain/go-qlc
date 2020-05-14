package config

type ConfigV7 struct {
	ConfigV6  `mapstructure:",squash"`
	Privacy   *Privacy   `json:"privacy"`
	WhiteList *WhiteList `json:"whiteList"`
}

type Privacy struct {
	Enable  bool   `json:"enable"`
	PtmNode string `json:"ptmNode"`
}

type WhiteList struct {
	Enable         bool             `json:"enable"`
	WhiteListInfos []*WhiteListInfo `json:"whiteListInfo"`
}

type WhiteListInfo struct {
	PeerId  string `json:"peerId"`
	Addr    string `json:"addr"` //(for example, "192.0.2.1:25", "[2001:db8::1]:80")
	Comment string `json:"comment"`
}

func DefaultConfigV7(dir string) (*ConfigV7, error) {
	var cfg ConfigV7
	cfg6, _ := DefaultConfigV6(dir)
	cfg.ConfigV6 = *cfg6
	cfg.Privacy = defaultPrivacy()
	cfg.WhiteList = defaultWhiteList()
	cfg.RPC.PublicModules = append(cfg.RPC.PublicModules, "privacy")
	return &cfg, nil
}

func defaultPrivacy() *Privacy {
	return &Privacy{}
}

func defaultWhiteList() *WhiteList {
	return &WhiteList{
		Enable:         false,
		WhiteListInfos: []*WhiteListInfo{},
	}
}

package config

import (
	"os/user"
	"os"
	"path/filepath"
)

func DefaultPath() string {
	home := HomePath()
	if home != "" {
		return filepath.Join(home, "QLCChain")
	}
	return ""
}

func DefaultTestPath() string {
	home := HomePath()
	if home != "" {
		return filepath.Join(home, "QLCChainTest")
	}
	return ""
}

func HomePath() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

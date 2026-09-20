package config

import (
	"os"
	"path/filepath"
)

type AppConfig struct {
	DBPath 		string
	LogPath 	string
	LogLevel 	string
}

func Load() (*AppConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Join(homeDir, ".passmgr")
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, err
	}

	return &AppConfig{
		DBPath:   filepath.Join(baseDir, "passwords.db"),
        LogPath:  filepath.Join(baseDir, "passmgr.log"),
        LogLevel: "info",
    }, nil
}

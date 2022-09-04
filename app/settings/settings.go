package settings

import (
	"os"

	"github.com/nanu-c/axolotl/app/config"

	yaml "gopkg.in/yaml.v2"
)

// Model for application settings

type Settings struct {
	SendByEnter     bool   `yaml:"sendByEnter"`
	EncryptDatabase bool   `yaml:"encryptDatabase"`
	CountryCode     string `yaml:"countrysDatabase"`
	Registered      bool   `yaml:"registered"`
	DebugLog        bool   `yaml:"debugLog"`
	DarkMode        bool   `yaml:"darkMode"`
}

var SettingsModel *Settings

// Load the Settings
func LoadSettings() (*Settings, error) {
	s := &Settings{}

	b, err := os.ReadFile(config.SettingsFile)
	if err != nil {
		return s, err
	}
	err = yaml.Unmarshal(b, s)
	if err != nil {
		return s, err
	}
	SettingsModel = s

	return s, nil
}

// Save the Settings
func SaveSettings(s *Settings) error {
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(config.SettingsFile, b, 0600)
}

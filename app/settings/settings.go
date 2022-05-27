package settings

import (
	"io/ioutil"

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

//Load the Settings
func LoadSettings() (*Settings, error) {
	s := &Settings{}
	settingsFile := config.GetSettingsFile()

	b, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		return s, err
	}
	err = yaml.Unmarshal(b, s)
	if err != nil {
		return s, err
	}

	return s, nil
}

//Save the Settings ** Was SaveSettings( *Settings )
func (s *Settings) Save() error {
	settingsFile := config.GetSettingsFile()
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(settingsFile, b, 0600)
}

package settings

import (
	"io/ioutil"

	"github.com/nanu-c/textsecure-qml/app/config"
	yaml "gopkg.in/yaml.v2"
)

// Model for application settings

type Settings struct {
	SendByEnter     bool   `yaml:"sendByEnter"`
	EncryptDatabase bool   `yaml:"encryptDatabase"`
	CountryCode     string `yaml:"countrysDatabase"`
}

var SettingsModel *Settings

//Load the Settings
func LoadSettings() (*Settings, error) {
	s := &Settings{}

	b, err := ioutil.ReadFile(config.SettingsFile)
	if err != nil {
		return s, err
	}

	err = yaml.Unmarshal(b, s)
	if err != nil {
		return s, err
	}
	return s, nil
}

//Save the Settings
func SaveSettings(s *Settings) error {
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(config.SettingsFile, b, 0600)
}

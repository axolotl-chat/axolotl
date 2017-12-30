package store

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Model for application settings

type Settings struct {
	SendByEnter bool `yaml:"sendByEnter"`
}

var SettingsModel *Settings

func LoadSettings() (*Settings, error) {
	s := &Settings{}

	b, err := ioutil.ReadFile(SettingsFile)
	if err != nil {
		return s, err
	}

	err = yaml.Unmarshal(b, s)
	if err != nil {
		return s, err
	}
	return s, nil
}
func SaveSettings(s *Settings) error {
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(SettingsFile, b, 0600)
}

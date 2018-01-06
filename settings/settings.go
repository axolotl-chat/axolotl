package settings

import (
	"io/ioutil"

	"github.com/nanu-c/textsecure-qml/store"
	yaml "gopkg.in/yaml.v2"
)

// Model for application settings

type Settings struct {
	SendByEnter bool `yaml:"sendByEnter"`
}

var SettingsModel *Settings

func LoadSettings() (*Settings, error) {
	s := &Settings{}

	b, err := ioutil.ReadFile(store.SettingsFile)
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
	return ioutil.WriteFile(store.SettingsFile, b, 0600)
}

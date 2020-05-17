package spotifyplayer

import (
	"encoding/json"
	"io/ioutil"
)

// GetConfig either loads an existing config in or creates a new one
func GetConfig() *ConfigT {
	cfg := new(ConfigT)
	_ = cfg.Load(DefaultConfigFile)
	return cfg
}

// DefaultConfigFile is the default file read
// and save configs to.
const DefaultConfigFile = "../config.json"

// ConfigT is the type for a Config
type ConfigT struct {
	PlayerType string

	// WebPlayer Only
	ClientID        string
	ClientSecret    string
	RedirectURI     string
	AccessToken     string
	UserAccessToken string
}

// Load loads the configuration file into c.
func (c *ConfigT) Load(file string) error {
	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fileContent, c)
	if err != nil {
		return err
	}

	return nil
}

// Save saves the configuration defined by c to file.
func (c *ConfigT) Save(file string) error {
	serialized, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, serialized, 0644)
	if err != nil {
		return err
	}

	return nil
}

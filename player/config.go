package player

import (
	"encoding/json"
	"io/ioutil"

	"github.com/charlesyu108/spotify-cli/utils"
)

// ConfigT is the type for a Config
type ConfigT struct {
	AppClientID     string // Required
	AppClientSecret string // Required
	RedirectPort    string // Required
	Tokens          struct {
		AppAccessToken      string
		UserAccessToken     string
		UserRefreshToken    string
		UserTokenExpiration int64
	}
}

// LoadConfig loads up the config
func LoadConfig(configFile string) *ConfigT {
	cfg := new(ConfigT)
	fileContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		// No file found or corrupt
		return cfg
	}
	err = json.Unmarshal(fileContent, cfg)
	utils.Check(err)
	return cfg
}

// SaveConfig saves the Config defined by c to file.
func SaveConfig(c *ConfigT, configFile string) {
	serialized, err := json.Marshal(c)
	utils.Check(err)

	err = ioutil.WriteFile(configFile, serialized, 0644)
	utils.Check(err)
}

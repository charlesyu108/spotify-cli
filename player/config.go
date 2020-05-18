package player

import (
	"github.com/charlesyu108/spotify-cli/utils"
)

// ConfigT is the type for a Config
type ConfigT struct {
	AppClientID     string // Required
	AppClientSecret string // Required
	RedirectPort    string // Required
}

// LoadConfig loads up the config
func LoadConfig(configFile string) *ConfigT {
	cfg := new(ConfigT)
	utils.LoadJSON(configFile, cfg)
	return cfg
}

// SaveConfig saves the Config defined by c to file.
func SaveConfig(c *ConfigT, configFile string) {
	utils.SaveJSON(configFile, c)
}

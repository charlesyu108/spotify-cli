package spotify

import (
	"fmt"

	"github.com/charlesyu108/spotify-cli/utils"
)

// ConfigT is the type for a Config
type ConfigT struct {
	AppClientID     string // Required
	AppClientSecret string // Required
	RedirectPort    string // Required
}

// LoadConfig loads up the config
// Returns a ConfigT and a boolean denoting if the config file was created.
func LoadConfig(configFile string) (*ConfigT, bool) {
	cfg := new(ConfigT)
	err := utils.LoadJSON(configFile, cfg)
	if err != nil {
		return cfg, true
	}
	return cfg, false
}

// Validate that the config is good.
func (c *ConfigT) Validate() error {
	if c.AppClientID == "" {
		return fmt.Errorf("AppClientID must not be empty")
	}
	if c.AppClientSecret == "" {
		return fmt.Errorf("AppClientSecret must not be empty")
	}
	if c.RedirectPort == "" {
		return fmt.Errorf("Redirect must not be empty")
	}
	return nil
}

// SaveConfig saves the Config defined by c to file.
func SaveConfig(c *ConfigT, configFile string) {
	utils.SaveJSON(configFile, c)
}

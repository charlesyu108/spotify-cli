package spotifyplayer

import (
	"encoding/json"
	"io/ioutil"

	"github.com/charlesyu108/spotify-cli/utils"
)

// DefaultConfigFile is the default file read
// and save configs to.
const DefaultConfigFile = "config.json"

type configT struct {
	PlayerType string

	// WebPlayer Only
	ClientID        string
	ClientSecret    string
	RedirectURI     string
	AccessToken     string
	UserAccessToken string
}

// Load loads the configuration file into c.
func (c *configT) Load(file string) {
	fileContent, err := ioutil.ReadFile(file)
	utils.Check(err)

	err = json.Unmarshal(fileContent, c)
	utils.Check(err)
}

// Save saves the configuration defined by c to file.
func (c *configT) Save(file string) {
	serialized, err := json.Marshal(c)
	utils.Check(err)
	err = ioutil.WriteFile(file, serialized, 0644)
	utils.Check(err)
}

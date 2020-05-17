package spotifyplayer

import (
	"os"
	"testing"

	"github.com/charlesyu108/spotify-cli/utils"
)

const testFile = "test_123.json"

func cleanup() {
	err := os.Remove(testFile)
	utils.Check(err)
}

func TestConfigSaveAndLoad(t *testing.T) {
	t.Cleanup(cleanup)

	c1 := ConfigT{
		RedirectURI: "abcd",
		PlayerType:  "123",
	}
	c1.Save(testFile)

	d1 := ConfigT{}
	d1.Load(testFile)
	if d1 != c1 {
		t.FailNow()
	}
}

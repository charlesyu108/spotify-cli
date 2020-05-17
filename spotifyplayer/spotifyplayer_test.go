package spotifyplayer

import (
	"testing"
)

func TestNewOSXPlayer(t *testing.T) {
	mockConfig := configT{}
	player := NewOSXPlayer(&mockConfig)

	if player.Config.PlayerType != "OSX" {
		t.FailNow()
	}
}

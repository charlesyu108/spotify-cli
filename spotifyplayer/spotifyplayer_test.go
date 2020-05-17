package spotifyplayer

import (
	"testing"
)

func TestNewOSXPlayer(t *testing.T) {
	mockConfig := ConfigT{}
	player := NewOSXPlayer(&mockConfig)

	if player.Config.PlayerType != "OSX" {
		t.FailNow()
	}
}

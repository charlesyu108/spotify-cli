package main

import (
	"github.com/charlesyu108/spotify-cli/player"
)

func main() {
	defaultConfig := "config.json"
	spotify := player.NewSpotify(defaultConfig)
	spotify.Authorize()
	spotify.Play()
}

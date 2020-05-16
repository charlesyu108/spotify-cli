package main

import (
	"fmt"
	"os"

	"github.com/charlesyu108/spotify-cli/playback"
)

func main() {
	argsWithoutProg := os.Args[1:]
	command := argsWithoutProg[0]

	player := new(playback.SpotifyPlayer)

	output := "Command not understood."
	switch command {

	case "pp":
		output = player.PlayPause()

	case "play":
		output = player.Play()

	case "pause":
		output = player.Pause()

	case "next", "n":
		output = player.NextTrack()

	case "prev", "pv":
		output = player.PrevTrack()

	case "info":
		output = player.TrackInfo()
	}

	fmt.Println(output)
}

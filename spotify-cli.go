package main

import (
	"fmt"
	"os"

	"github.com/charlesyu108/spotify-cli/spotify"
)

// ConfigFile defines which config JSON file to load
const ConfigFile = "config.json"

func main() {

	args := os.Args[1:]

	Spotify := spotify.New(ConfigFile)
	Spotify.Authorize()

	switch cmd := args[0]; cmd {
	case "play":
		Spotify.Play()
	case "pause":
		Spotify.Pause()
	case "next":
		Spotify.NextTrack()
	case "prev":
		Spotify.PreviousTrack()
	case "devices":
		fmt.Println(Spotify.GetDevices())
	case "info":
		fmt.Println(Spotify.CurrentState())
	}

}

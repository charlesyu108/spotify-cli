package main

import (
	"fmt"
	"os"
	"strings"

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
	case "play-on":
		search := strings.ToLower(args[1])
		for _, device := range Spotify.GetDevices() {
			name, t := strings.ToLower(device.Name), strings.ToLower(device.Type)
			if strings.Contains(name, search) || strings.Contains(t, search) {
				Spotify.PlayOnDevice(device)
				break
			}
		}
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
	case "find":
		fmt.Println(Spotify.SimpleSearch(args[1], "track"))
	case "play-track":
		track := Spotify.SimpleSearch(args[1], "track")
		if track != "" {
			Spotify.PlayURI(track)
		}

	}

}

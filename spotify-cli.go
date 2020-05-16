package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	argsWithoutProg := os.Args[1:]

	// TODO: No command case

	command := argsWithoutProg[0]
	player := new(SpotifyPlayer)
	output := "Command not understood."

	switch command {

	case "pp":
		output = player.PlayPause()

	case "play":
		if len(argsWithoutProg) == 1 {
			output = player.Play()
		} else {
			loadConfig()
			getAuthToken()
			searchArg := argsWithoutProg[1]
			uri := search(searchArg)
			log.Println(uri)
			output = player.PlayResource(uri)
		}

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

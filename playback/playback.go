// Package playback manages the Spotify OSX Desktop Application Playback
package playback

import (
	"fmt"
	"os/exec"
	"strings"
)

type spotifyPlayback interface {
	Play() string
	Pause() string
	PlayPause() string
	NextTrack() string
	PrevTrack() string
}

// State reports the player's current state.
func State() string {
	state, _ := executeJXACommand(playbackState)
	status := formatString("Now %s", state[:len(state)-1])
	return status
}

// TrackInfo reports information about the current track.
func TrackInfo() string {
	title, _ := executeJXACommand(currentTrackName)
	artist, _ := executeJXACommand(currentTrackArtist)
	album, _ := executeJXACommand(currentTrackAlbum)

	title = strings.TrimSpace(title)

	info := formatString("%s FROM %s BY %s", title, album, artist)
	return info
}

// Play toggles the player to play music.
func Play() string {
	msg, err := executeJXACommand(play)
	if err != nil {
		return formatString("Error in playing music: %s", msg)
	}
	return State()
}

// Pause toggles the player to pause music.
func Pause() string {
	msg, err := executeJXACommand(pause)
	if err != nil {
		return formatString("Error in pausing music: %s", msg)
	}
	return State()
}

// PlayPause toggles the player's Playback state.
func PlayPause() string {
	msg, err := executeJXACommand(playPause)
	if err != nil {
		return formatString("Error in toggling Play/Pause: %s", msg)
	}
	return formatString("%s: %s", State(), TrackInfo())
}

// func NextTrack() string {
// 	msg, err := executeJXACommand(playPause)
// 	if err != nil {
// 		return formatString("Error in toggling Play/Pause: %s", msg)
// 	}
// 	return State()
// }

// Define JXA commands
const play = "Application('Spotify').play()"
const pause = "Application('Spotify').pause()"
const playPause = "Application('Spotify').playpause()"
const nextTrack = "Application('Spotify').nextTrack()"
const previousTrack = "Application('Spotify').previousTrack()"
const playbackState = "Application('Spotify').playerState()"

const currentTrackName = "Application('Spotify').currentTrack().name()"
const currentTrackArtist = "Application('Spotify').currentTrack().artist()"
const currentTrackAlbum = "Application('Spotify').currentTrack().album()"

// executeJXACommand executes the provided JXA command string.
func executeJXACommand(jxacmd string) (string, error) {
	command := exec.Command("osascript", "-l", "JavaScript", "-e", jxacmd)
	output, err := command.CombinedOutput()
	return string(output), err
}

// Sprintf except only for strings and strings are trimmed before formatting.
func formatString(template string, strs ...string) string {
	trimmed := make([]interface{}, len(strs))
	for idx, s := range strs {
		trimmed[idx] = strings.TrimSpace(s)
	}
	return fmt.Sprintf(template, trimmed...)
}

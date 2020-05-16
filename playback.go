package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// SpotifyPlayer controls the Spotify OSX Desktop Application
type SpotifyPlayer struct{}

// State reports the player's current state.
func (player *SpotifyPlayer) State() string {
	state, _ := executeJXACommand(playbackState)
	status := formatString("Now %s", state)
	return status
}

// TrackInfo reports information about the current track.
func (player *SpotifyPlayer) TrackInfo() string {
	title, _ := executeJXACommand(currentTrackName)
	artist, _ := executeJXACommand(currentTrackArtist)
	album, _ := executeJXACommand(currentTrackAlbum)
	info := formatString("[Track]: %s \n[Album]: %s \n[Artist]: %s", title, album, artist)
	return info
}

// Play toggles the player to play music.
func (player *SpotifyPlayer) Play() string {
	msg, err := executeJXACommand(play)
	if err != nil {
		return formatString("Error in playing music: %s", msg)
	}
	return player.State()
}

// Pause toggles the player to pause music.
func (player *SpotifyPlayer) Pause() string {
	msg, err := executeJXACommand(pause)
	if err != nil {
		return formatString("Error in pausing music: %s", msg)
	}
	return player.State()
}

// PlayResource plays the URI provided.
func (player *SpotifyPlayer) PlayResource(uri string) string {
	msg, err := executeJXACommand(formatString(playResourceTemplate, uri))
	if err != nil {
		return formatString("Error in playing specified URI %s: %s", uri, msg)
	}
	return formatString("%s.\n\n%s", player.State(), player.TrackInfo())
}

// PlayPause toggles the player's playback state.
func (player *SpotifyPlayer) PlayPause() string {
	msg, err := executeJXACommand(playPause)
	if err != nil {
		return formatString("Error in toggling Play/Pause: %s", msg)
	}
	return formatString("%s.\n\n%s", player.State(), player.TrackInfo())
}

// NextTrack advances the player to a new track.
func (player *SpotifyPlayer) NextTrack() string {
	msg, err := executeJXACommand(nextTrack)
	if err != nil {
		return formatString("Error in skipping to next track: %s", msg)
	}
	return player.TrackInfo()
}

// PrevTrack advances the player to a new track.
func (player *SpotifyPlayer) PrevTrack() string {
	msg, err := executeJXACommand(previousTrack)
	if err != nil {
		return formatString("Error in skipping to previous track: %s", msg)
	}
	return player.TrackInfo()
}

// Define JXA commands
const play = "Application('Spotify').play()"
const pause = "Application('Spotify').pause()"
const playPause = "Application('Spotify').playpause()"
const playResourceTemplate = "Application('Spotify').playTrack('%s')"
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

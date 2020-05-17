package spotifyplayer

import (
	"os/exec"

	"github.com/charlesyu108/spotify-cli/utils"
)

// OSXSpotifyPlayer controls the Spotify OSX Desktop Application
type OSXSpotifyPlayer struct {
	Config *configT
}

// State reports the player's current state.
func (player *OSXSpotifyPlayer) State() string {
	state, err := executeJXACommand(playbackState)
	utils.Check(err)
	status := utils.FormatString("Now %s", state)
	return status
}

// TrackInfo reports information about the current track.
func (player *OSXSpotifyPlayer) TrackInfo() string {
	title, _ := executeJXACommand(currentTrackName)
	artist, _ := executeJXACommand(currentTrackArtist)
	album, _ := executeJXACommand(currentTrackAlbum)
	info := utils.FormatString("[Track]: %s \n[Album]: %s \n[Artist]: %s", title, album, artist)
	return info
}

// Play toggles the player to play music.
func (player *OSXSpotifyPlayer) Play() string {
	msg, err := executeJXACommand(play)
	if err != nil {
		return utils.FormatString("Error in playing music: %s", msg)
	}
	return player.State()
}

// Pause toggles the player to pause music.
func (player *OSXSpotifyPlayer) Pause() string {
	msg, err := executeJXACommand(pause)
	if err != nil {
		return utils.FormatString("Error in pausing music: %s", msg)
	}
	return player.State()
}

// PlayResource plays the URI provided.
func (player *OSXSpotifyPlayer) PlayResource(uri string) string {
	msg, err := executeJXACommand(utils.FormatString(playResourceTemplate, uri))
	if err != nil {
		return utils.FormatString("Error in playing specified URI %s: %s", uri, msg)
	}
	return utils.FormatString("%s.\n\n%s", player.State(), player.TrackInfo())
}

// PlayPause toggles the player's playback state.
func (player *OSXSpotifyPlayer) PlayPause() string {
	msg, err := executeJXACommand(playPause)
	if err != nil {
		return utils.FormatString("Error in toggling Play/Pause: %s", msg)
	}
	return utils.FormatString("%s.\n\n%s", player.State(), player.TrackInfo())
}

// NextTrack advances the player to a new track.
func (player *OSXSpotifyPlayer) NextTrack() string {
	msg, err := executeJXACommand(nextTrack)
	if err != nil {
		return utils.FormatString("Error in skipping to next track: %s", msg)
	}
	return player.TrackInfo()
}

// PrevTrack advances the player to a new track.
func (player *OSXSpotifyPlayer) PrevTrack() string {
	msg, err := executeJXACommand(previousTrack)
	if err != nil {
		return utils.FormatString("Error in skipping to previous track: %s", msg)
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

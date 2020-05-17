package spotifyplayer

// The SpotifyPlayer interface
type SpotifyPlayer interface {
	State() string
	TrackInfo() string
	Play() string
	Pause() string
	PlayResource(uri string) string
	NextTrack() string
	PrevTrack() string
}

// NewOSXPlayer creates a new OSXPlayer from the passed in config
func NewOSXPlayer(config *ConfigT) *OSXPlayer {
	player := new(OSXPlayer)
	player.Config = config
	player.Config.PlayerType = "OSX"
	return player
}

// NewWebPlayer creates a new WebPlayer from the passed in config
func NewWebPlayer(config *ConfigT) *WebPlayer {
	player := new(WebPlayer)
	player.Config = config
	player.Config.PlayerType = "Web"
	return player
}

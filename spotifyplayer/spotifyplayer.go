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

// NewOSXPlayer creates a new OSXSpotifyPlayer from the passed in config
func NewOSXPlayer(config *configT) *OSXSpotifyPlayer {
	player := new(OSXSpotifyPlayer)
	player.Config = config
	player.Config.PlayerType = "OSX"
	return player
}

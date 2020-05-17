package spotifyplayer

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/charlesyu108/spotify-cli/utils"
)

// WebPlayer is the Spotify Web API Implementation of SpotifyPlayer
type WebPlayer struct {
	isAuthorized bool
	Config       *ConfigT
}

// AuthorizeApp exchanges information with the SpotifyAPI to
// get valid AccessToken credentials
func (player *WebPlayer) AuthorizeApp() {
	client := new(http.Client)
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	const spotifyAuthURL = "https://accounts.spotify.com/api/token"
	req, _ := http.NewRequest("POST", spotifyAuthURL, strings.NewReader(data.Encode()))
	appIdentity := []byte(player.Config.ClientID + ":" + player.Config.ClientSecret)
	b64encode := base64.StdEncoding.EncodeToString(appIdentity)
	req.Header.Set("Authorization", "Basic "+b64encode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	utils.Check(err)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)

	if resp.StatusCode != 200 {
		log.Panicf(
			"While performing Authorization, received code: %d. Message %s",
			resp.StatusCode,
			payload,
		)
	}
	// Success - Set these values
	player.Config.AccessToken = payload["access_token"].(string)
	player.isAuthorized = true
}

// search retrieves back the first result for the query q.
func (player *WebPlayer) search(q string, categories []string) string {
	client := new(http.Client)

	// By default cats is all types searchable
	cats := "track,artist,album,playlist"
	if categories != nil {
		cats = strings.Join(categories, ",")
	}
	searchURL := utils.FormatString("https://api.spotify.com/v1/search?q=%s&type=%s", url.QueryEscape(q), cats)
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("Authorization", "Bearer "+player.Config.AccessToken)

	resp, err := client.Do(req)
	utils.CheckHTTPResponse(resp, err, "Search Failed")

	defer resp.Body.Close()

	// Define Search resonse payload Shape
	type resourceT struct {
		Items []map[string]interface{}
	}

	type payloadT struct {
		Tracks    resourceT
		Artists   resourceT
		Albums    resourceT
		Playlists resourceT
	}

	var payload payloadT
	json.NewDecoder(resp.Body).Decode(&payload)

	if items := payload.Tracks.Items; len(items) > 0 {
		return items[0]["uri"].(string)
	}
	if items := payload.Artists.Items; len(items) > 0 {
		return items[0]["uri"].(string)
	}
	if items := payload.Albums.Items; len(items) > 0 {
		return items[0]["uri"].(string)
	}
	if items := payload.Playlists.Items; len(items) > 0 {
		return items[0]["uri"].(string)
	}

	// Return empty if nothing was found
	return ""
}

// State TODO
func (player *WebPlayer) State() string {
	// TODO
	return ""
}

// TrackInfo TODO
func (player *WebPlayer) TrackInfo() string {
	// TODO
	return ""
}

// Play TODO
func (player *WebPlayer) Play() string {
	//TODO
	return ""
}

// Pause TODO
func (player *WebPlayer) Pause() string {
	// TODO
	return ""
}

// PlayResource TODO
func (player *WebPlayer) PlayResource(uri string) string {
	// TODO
	return ""
}

// NextTrack TODO
func (player *WebPlayer) NextTrack() string {
	// TODO
	return ""
}

// PrevTrack TODO
func (player *WebPlayer) PrevTrack() string {
	// TODO
	return ""
}

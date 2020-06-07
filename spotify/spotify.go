package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/charlesyu108/spotify-cli/utils"
)

// tokensT defines tokens and related info
// for interfacing with the Spotify API
type tokensT struct {
	AppAccessToken      string
	UserAccessToken     string
	UserRefreshToken    string
	UserTokenExpiration int64
	AppTokenExpiration  int64
}

// authT defines a struct that encapsulates all resources
// required to obtain Authorization credentials
type authT struct {
	codeChan chan string
	server   *http.Server
}

// Spotify represents an interface to the Spotify API
type Spotify struct {
	Config    *ConfigT
	tokens    *tokensT
	tokenFile string
	auth      *authT
}

// New produces creates and initializes a new Spotify
func New(cfg *ConfigT) Spotify {
	spotify := Spotify{Config: cfg, tokens: new(tokensT)}
	spotify.tokenFile = filepath.Join(utils.GetProgFilesDir(), ".tokens")
	spotify.auth = &authT{
		codeChan: make(chan string),
		server: &http.Server{
			Addr:    ":" + spotify.Config.RedirectPort,
			Handler: http.DefaultServeMux,
		},
	}
	// Register Auth Redirect Handler
	http.HandleFunc("/", spotify.handleAuthorizeUserRedirect)
	// Start auth server
	go spotify.auth.server.ListenAndServe()

	return spotify
}

// handleAuthorizeUserRedirect is the HTTP Handler that listens for activity on the
// local authorization server & extracts the obtained user access token for OAuth.
func (spotify *Spotify) handleAuthorizeUserRedirect(w http.ResponseWriter, req *http.Request) {
	userCode := req.FormValue("code")
	if userCode != "" {
		w.Write([]byte("Success!"))
	} else {
		w.Write([]byte("User Authorization failed"))
	}
	spotify.auth.codeChan <- userCode
}

// Authorize performs the required client and user authorization steps for
// the app to work properly.
//
// NOTE: If the tokens are properly saved, they will cache authorization credentials
// to make this process more seamless.
func (spotify *Spotify) Authorize() {
	spotify.loadSavedTokens()
	defer spotify.saveTokens()

	tokens := spotify.tokens
	appClient, access, refresh := tokens.AppAccessToken, tokens.UserAccessToken, tokens.UserRefreshToken
	unixTimeNow := time.Now().Unix()
	appTokExpired, uTokExpired := unixTimeNow > tokens.AppTokenExpiration, unixTimeNow > tokens.UserTokenExpiration

	// Always want to make sure our App Client is authorized
	if appTokExpired || appClient == "" {
		spotify.acquireTokens("", "client")
	}

	// Case: user has existing tokens
	if !uTokExpired && access != "" {
		return
	}

	// Case: Existing user but tokens expired, refresh
	if uTokExpired && refresh != "" {
		spotify.acquireTokens(refresh, "refresh")
		return
	}

	// Case: New user - getting new auth and refresh tokens
	authCode := spotify.authorizeUser()
	spotify.acquireTokens(authCode, "auth")
}

// loadSavedTokens loads the cached tokens file (if it exists) into memory
func (spotify *Spotify) loadSavedTokens() {
	utils.LoadJSON(spotify.tokenFile, spotify.tokens)
}

// saveTokens saves the current tokens to a cached tokens file
func (spotify *Spotify) saveTokens() {
	err := utils.SaveJSON(spotify.tokenFile, spotify.tokens)
	utils.Check(err)
}

// acquireTokens exchanges AppClient or User AuthCode/Refresh tokens for
// access tokens that can be used to make Spotify API calls.
func (spotify *Spotify) acquireTokens(code string, tokenType string) {
	URL := "https://accounts.spotify.com/api/token"
	appIdentity := []byte(spotify.Config.AppClientID + ":" + spotify.Config.AppClientSecret)
	headers := map[string]string{
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString(appIdentity),
		"Content-Type":  "application/x-www-form-urlencoded",
	}
	form := url.Values{}

	switch tokenType {
	case "client":
		form.Set("grant_type", "client_credentials")

	case "auth":
		form.Set("grant_type", "authorization_code")
		form.Set("code", code)
		form.Set("redirect_uri", "http://localhost:"+spotify.Config.RedirectPort)

	case "refresh":
		form.Set("grant_type", "refresh_token")
		form.Set("refresh_token", code)
		form.Set("redirect_uri", "http://localhost:"+spotify.Config.RedirectPort)

	default:
		log.Fatalf("Bad value provided for tokenType arg to acquireTokens")
	}

	resp, _ := utils.MakeHTTPRequest("POST", URL, headers, form.Encode())
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("Error encountered during Authorization. INFO: %s", body)
	}
	var payload map[string]string
	json.NewDecoder(resp.Body).Decode(&payload)

	expiration := time.Now().Unix() + int64(3600)
	switch tokenType {
	case "client":
		if clientTok, ok := payload["access_token"]; ok {
			spotify.tokens.AppAccessToken = clientTok
			spotify.tokens.AppTokenExpiration = expiration
		}
	// Same logic otherwise
	default:
		if userTok, ok := payload["access_token"]; ok {
			spotify.tokens.UserAccessToken = userTok
			spotify.tokens.UserTokenExpiration = expiration
		}
		if refreshTok, ok := payload["refresh_token"]; ok {
			spotify.tokens.UserRefreshToken = refreshTok
		}
	}
}

// authorizeUser prompts the user to authorize his or her account
// and waits until the authServer has received and extracted the authCode.
func (spotify *Spotify) authorizeUser() string {
	authURL := utils.FormatString(
		"https://accounts.spotify.com/authorize?client_id=%s&"+
			"response_type=code&redirect_uri=%s&"+
			"scope=user-read-playback-state,user-modify-playback-state,user-read-currently-playing",
		spotify.Config.AppClientID,
		"http://localhost:"+spotify.Config.RedirectPort,
	)
	fmt.Printf("\nPlease navigate to this URL to Authorize Spotify:\n\n%s\n", authURL)
	_ = utils.OpenInBrowser(authURL)
	// Block while waiting for authorization code to be received
	// by redirect handler
	userAuthCode := <-spotify.auth.codeChan
	return userAuthCode
}

// SpotifyURI defines a reference to a playable Spotify resource
type SpotifyURI string

// Play starts/resumes playing music on the active device, if one exists. If not it
// tries to play on the first device that it comes across from the Devices API.
// NOTE: Play will return a 403 Forbbiden if Spotify already playing.
func (spotify *Spotify) Play() {
	device := spotify.activeOrFirstDevice()
	URL := utils.FormatString(
		"https://api.spotify.com/v1/me/player/play?device_id=%s",
		device.ID,
	)
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("PUT", URL, headers, "")
	handlePlaybackAPIErrorScenarios("Play", resp)
}

// PlayOnDevice starts/resumes playing music on the target device provided.
func (spotify *Spotify) PlayOnDevice(device Device) {
	URL := "https://api.spotify.com/v1/me/player/"
	body := utils.FormatString(
		`{"device_ids":["%s"], "play":true}`,
		device.ID,
	)

	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
		"Content-Type":  "application/json",
	}
	resp, _ := utils.MakeHTTPRequest("PUT", URL, headers, body)
	handlePlaybackAPIErrorScenarios("PlayOnDevice", resp)
}

// PlayURI starts playing the specified URI on the active device, if one exists. If not it
// tries to play on the first device that it comes across from the Devices API.
func (spotify *Spotify) PlayURI(uri SpotifyURI) {
	device := spotify.activeOrFirstDevice()
	// By default use the URI as a context_uri.
	body := utils.FormatString(`{"context_uri":"%s"}`, string(uri))
	// If URI is a track, different kind of body
	if strings.HasPrefix(string(uri), "spotify:track") {
		body = utils.FormatString(`{"uris":["%s"]}`, string(uri))
	}
	URL := utils.FormatString(
		"https://api.spotify.com/v1/me/player/play?device_id=%s",
		device.ID,
	)
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("PUT", URL, headers, body)
	handlePlaybackAPIErrorScenarios("PlayURI", resp)
}

// Pause pauses playing music on any device.
// NOTE: Pause will return a 403 Forbbiden if Spotify not already playing.
func (spotify *Spotify) Pause() {
	URL := "https://api.spotify.com/v1/me/player/pause"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("PUT", URL, headers, "")
	handlePlaybackAPIErrorScenarios("Pause", resp)
}

// NextTrack skips to the next track.
func (spotify *Spotify) NextTrack() {
	URL := "https://api.spotify.com/v1/me/player/next"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("POST", URL, headers, "")
	handlePlaybackAPIErrorScenarios("NextTrack", resp)
}

// PreviousTrack skips to the last track.
func (spotify *Spotify) PreviousTrack() {
	URL := "https://api.spotify.com/v1/me/player/previous"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("POST", URL, headers, "")
	handlePlaybackAPIErrorScenarios("PreviousTrack", resp)
}

// Volume adjusts the playback volume to the desired percentage [0..100].
func (spotify *Spotify) Volume(percent int) {
	URL := fmt.Sprintf("https://api.spotify.com/v1/me/player/volume?volume_percent=%d", percent)
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("PUT", URL, headers, "")
	handlePlaybackAPIErrorScenarios("Volume", resp)
}

// Device describes a device
type Device struct {
	ID           string `json:"id"`
	IsActive     bool   `json:"is_active"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	IsRestricted bool   `json:"is_restricted"`
}

// GetDevices returns all devices players
func (spotify *Spotify) GetDevices() []Device {
	URL := "https://api.spotify.com/v1/me/player/devices"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("GET", URL, headers, "")
	var payload struct {
		Devices []Device `json:"devices"`
	}
	json.NewDecoder(resp.Body).Decode(&payload)
	return payload.Devices
}

// SimpleSearch returns the first URI that matches the query string for the given
// resource type.
// NOTE `Type` must be one of { 'track', 'album', 'artist', 'playlist', 'show', 'episode' }
func (spotify *Spotify) SimpleSearch(q string, Type string) SpotifyURI {
	base, query := "https://api.spotify.com/v1/search", url.Values{}
	query.Set("q", q)
	query.Set("type", Type)
	query.Set("limit", "1")
	URL := utils.FormatString("%s?%s", base, query.Encode())
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	var payload map[string]struct {
		Items []struct {
			Uri SpotifyURI `json:"uri"`
		} `json:"items"`
	}
	resp, _ := utils.MakeHTTPRequest("GET", URL, headers, "")
	json.NewDecoder(resp.Body).Decode(&payload)

	if data, ok := payload[Type+"s"]; ok && len(data.Items) > 0 {
		return data.Items[0].Uri
	}

	log.Fatalf("SimpleSearch failed to find any '%s' matching search string '%s.'", Type, q)
	return ""
}

// Track describes a track
type Track struct {
	Album struct {
		Name string `json:"name"`
	} `json:"album"`
	Name    string     `json:"name"`
	URI     SpotifyURI `json:"uri"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
}

// StateInfo describes the current state of the Spotify playback
type StateInfo struct {
	IsPlaying bool  `json:"is_playing"`
	Track     Track `json:"item"`
}

// CurrentState fetches the current state of the Spotify playback
func (spotify *Spotify) CurrentState() StateInfo {
	URL := "https://api.spotify.com/v1/me/player/currently-playing"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("GET", URL, headers, "")
	var payload StateInfo
	json.NewDecoder(resp.Body).Decode(&payload)
	return payload

}

// ToggleShuffle toggles playback shuffle state.
func (spotify *Spotify) ToggleShuffle(active bool) {
	toggleState := "false"
	if active {
		toggleState = "true"
	}

	URL := fmt.Sprintf("https://api.spotify.com/v1/me/player/shuffle?state=%s", toggleState)
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("PUT", URL, headers, "")
	handlePlaybackAPIErrorScenarios("Shuffle", resp)
}

// activeOrFirstDevice returns the active device. If no active, return the first.
func (spotify *Spotify) activeOrFirstDevice() Device {
	devices := spotify.GetDevices()
	chosen := devices[0]
	for i := range devices {
		if devices[i].IsActive {
			chosen = devices[i]
		}
	}
	return chosen
}

// handlePlaybackAPIErrorScenarios handles HTTP response code error scenarios for
// the Spotify Player API methods.
func handlePlaybackAPIErrorScenarios(operation string, r *http.Response) {
	switch r.StatusCode {
	case 204:
		// ok
	case 400:
		body, _ := ioutil.ReadAll(r.Body)
		log.Fatalf("%s operation encountered unexpected client error. INFO: %s", operation, body)
	case 403:
		log.Fatalf("%s operation encountered 403 Forbidden. Is this operation allowed right now?", operation)
	case 404:
		log.Fatalf("%s operation encountered 404 Not Found. Are there active devices?", operation)
	}
}

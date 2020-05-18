package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/charlesyu108/spotify-cli/utils"
)

const tokenFile = ".tokens"

var userAuthCodeChan = make(chan string)

func handleAuthorizeUserRedirect(w http.ResponseWriter, req *http.Request) {
	userCode := req.FormValue("code")
	if userCode != "" {
		w.Write([]byte("Success!"))
	} else {
		w.Write([]byte("User Authorization failed"))
	}
	userAuthCodeChan <- userCode
}

// New produces creates and initializes a new Spotify
func New(configFile string) Spotify {
	cfg := LoadConfig(configFile)
	spotify := Spotify{Config: cfg, tokens: new(tokensT)}

	// Creating a Server for handling user Auth requests
	http.HandleFunc("/", handleAuthorizeUserRedirect)
	spotify.authServer = &http.Server{
		Addr:    ":" + spotify.Config.RedirectPort,
		Handler: http.DefaultServeMux,
	}
	// Start auth server
	go spotify.authServer.ListenAndServe()

	return spotify
}

type tokensT struct {
	AppAccessToken      string
	UserAccessToken     string
	UserRefreshToken    string
	UserTokenExpiration int64
	AppTokenExpiration  int64
}

type Spotify struct {
	Config     *ConfigT
	authServer *http.Server
	tokens     *tokensT
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

func (spotify *Spotify) loadSavedTokens() {
	utils.LoadJSON(tokenFile, spotify.tokens)
}

func (spotify *Spotify) saveTokens() {
	err := utils.SaveJSON(tokenFile, spotify.tokens)
	utils.Check(err)
}

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

	resp, _ := utils.MakeHTTPRequest("POST", URL, headers, form)
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

func (spotify *Spotify) authorizeUser() string {
	authURL := utils.FormatString(
		"https://accounts.spotify.com/authorize?client_id=%s&"+
			"response_type=code&redirect_uri=%s&"+
			"scope=user-read-playback-state,user-modify-playback-state",
		spotify.Config.AppClientID,
		"http://localhost:"+spotify.Config.RedirectPort,
	)
	fmt.Printf("\nPlease navigate to this URL to Authorize Spotify:\n\n%s\n", authURL)
	_ = utils.OpenInBrowser(authURL)
	// Block while waiting for authorization code to be received
	// by redirect handler
	userAuthCode := <-userAuthCodeChan
	return userAuthCode
}

func (spotify *Spotify) Play() {
	URL := utils.FormatString(
		"https://api.spotify.com/v1/me/player/play?device_id=%s",
		spotify.activeOrFirstDevice().ID,
	)
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("PUT", URL, headers, nil)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)
	// TODO err handling
}

func (spotify *Spotify) Pause() {
	URL := "https://api.spotify.com/v1/me/player/pause"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("PUT", URL, headers, nil)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)
	// TODO err handling
}

func (spotify *Spotify) NextTrack() {
	URL := "https://api.spotify.com/v1/me/player/next"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("POST", URL, headers, nil)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)
	// TODO err handling
}

func (spotify *Spotify) PreviousTrack() {
	URL := "https://api.spotify.com/v1/me/player/previous"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("POST", URL, headers, nil)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)
	// TODO err handling
}

type Device struct {
	ID           string `json:"id"`
	IsActive     bool   `json:"is_active"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	IsRestricted bool   `json:"is_restricted"`
}

func (spotify *Spotify) GetDevices() []Device {
	URL := "https://api.spotify.com/v1/me/player/devices"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("GET", URL, headers, nil)
	var payload struct {
		Devices []Device `json:"devices"`
	}
	json.NewDecoder(resp.Body).Decode(&payload)
	return payload.Devices
}

type Track struct {
	Album struct {
		Name string `json:"name"`
	} `json:"album"`
	Name    string `json:"name"`
	URI     string `json:"uri"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
}

type StateInfo struct {
	IsPlaying bool  `json:"is_playing"`
	Track     Track `json:"item"`
}

func (spotify *Spotify) CurrentState() StateInfo {
	URL := "https://api.spotify.com/v1/me/player/currently-playing"
	headers := map[string]string{
		"Authorization": "Bearer " + spotify.tokens.UserAccessToken,
	}
	resp, _ := utils.MakeHTTPRequest("GET", URL, headers, nil)
	var payload StateInfo
	json.NewDecoder(resp.Body).Decode(&payload)
	return payload
}

// Return the active device. If no active, return the first.
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

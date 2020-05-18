package player

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
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

func NewSpotify(configFile string) Spotify {
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
func (player *Spotify) Authorize() {
	player.loadSavedTokens()
	defer player.saveTokens()

	tokens := player.tokens
	appClient, access, refresh := tokens.AppAccessToken, tokens.UserAccessToken, tokens.UserRefreshToken
	unixTimeNow := time.Now().Unix()
	appTokExpired, uTokExpired := unixTimeNow > tokens.AppTokenExpiration, unixTimeNow > tokens.UserTokenExpiration

	// Always want to make sure our App Client is authorized
	if appTokExpired || appClient == "" {
		player.acquireTokens("", "client")
	}

	// Case: user has existing tokens
	if !uTokExpired && access != "" {
		return
	}

	// Case: Existing user but tokens expired, refresh
	if uTokExpired && refresh != "" {
		player.acquireTokens(refresh, "refresh")
		return
	}

	// Case: New user - getting new auth and refresh tokens
	authCode := player.authorizeUser()
	player.acquireTokens(authCode, "auth")
}

func (player *Spotify) loadSavedTokens() {
	utils.LoadJSON(tokenFile, player.tokens)
}

func (player *Spotify) saveTokens() {
	err := utils.SaveJSON(tokenFile, player.tokens)
	utils.Check(err)
}

func (player *Spotify) acquireTokens(code string, tokenType string) {
	URL := "https://accounts.spotify.com/api/token"
	appIdentity := []byte(player.Config.AppClientID + ":" + player.Config.AppClientSecret)
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
		form.Set("redirect_uri", "http://localhost:"+player.Config.RedirectPort)

	case "refresh":
		form.Set("grant_type", "refresh_token")
		form.Set("refresh_token", code)
		form.Set("redirect_uri", "http://localhost:"+player.Config.RedirectPort)

	default:
		log.Fatalf("Bad value provided for tokenType arg to acquireTokens")
	}

	resp, _ := makeRequest("POST", URL, headers, form)
	var payload map[string]string
	json.NewDecoder(resp.Body).Decode(&payload)

	expiration := time.Now().Unix() + int64(3600)
	switch tokenType {
	case "client":
		if clientTok, ok := payload["access_token"]; ok {
			player.tokens.AppAccessToken = clientTok
			player.tokens.AppTokenExpiration = expiration
		}
	// Same logic otherwise
	default:
		if userTok, ok := payload["access_token"]; ok {
			player.tokens.UserAccessToken = userTok
			player.tokens.UserTokenExpiration = expiration
		}
		if refreshTok, ok := payload["refresh_token"]; ok {
			player.tokens.UserRefreshToken = refreshTok
		}
	}
}

func (player *Spotify) authorizeUser() string {
	authURL := utils.FormatString(
		"https://accounts.spotify.com/authorize?"+
			"client_id=%s&"+
			"response_type=code&"+
			"redirect_uri=%s&"+
			"scope=user-read-playback-state,user-modify-playback-state",
		player.Config.AppClientID,
		"http://localhost:"+player.Config.RedirectPort,
	)
	fmt.Printf("\nPlease navigate to this URL to Authorize Spotify:\n\n%s\n", authURL)
	_ = utils.OpenInBrowser(authURL)
	userAuthCode := <-userAuthCodeChan
	return userAuthCode
}

func (player *Spotify) Play() {
	URL := "https://api.spotify.com/v1/me/player/play"
	headers := map[string]string{
		"Authorization": "Bearer " + player.tokens.UserAccessToken,
	}
	resp, err := makeRequest("PUT", URL, headers, nil)
	utils.Check(err)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)
	fmt.Println(payload)
}

func (player *Spotify) Pause() {
	client := new(http.Client)
	const URL = "https://api.spotify.com/v1/me/player/pause"
	req, _ := http.NewRequest("PUT", URL, nil)
	fmt.Println("\n\n\n" + player.tokens.UserAccessToken)
	req.Header.Set("Authorization", "Bearer "+player.tokens.UserAccessToken)
	resp, err := client.Do(req)
	utils.Check(err)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)
	fmt.Println(payload)
}

func makeRequest(method string, URL string, headers map[string]string, formData url.Values) (*http.Response, error) {
	client := new(http.Client)
	if formData == nil {
		formData = url.Values{}
	}
	req, _ := http.NewRequest(method, URL, strings.NewReader(formData.Encode()))

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return client.Do(req)
}

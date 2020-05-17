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

type ISpotify interface {
}

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
	spotify := Spotify{Config: cfg}

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

type Spotify struct {
	Config     *ConfigT
	authServer *http.Server
}

func (player *Spotify) Authorize() {
	tokens := player.Config.Tokens
	if tokens.AppAccessToken == "" {
		player.authorizeClient()
	}

	access, refresh := tokens.UserAccessToken, tokens.UserRefreshToken
	expired := time.Now().Unix() < tokens.UserTokenExpiration

	if !expired && access != "" {
		return
	}

	if expired && refresh != "" {
		player.exchangeUserCodeForAccessTokens(refresh)
		return
	}

	authCode := player.authorizeUser()
	player.exchangeUserCodeForAccessTokens(authCode)
}

func (player *Spotify) authorizeClient() {
	client := new(http.Client)
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	const appAuthURL = "https://accounts.spotify.com/api/token"
	req, _ := http.NewRequest("POST", appAuthURL, strings.NewReader(data.Encode()))
	appIdentity := []byte(player.Config.AppClientID + ":" + player.Config.AppClientSecret)
	b64encode := base64.StdEncoding.EncodeToString(appIdentity)
	req.Header.Set("Authorization", "Basic "+b64encode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	utils.Check(err)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)

	if resp.StatusCode != 200 {
		log.Panicf(
			"While performing Authorization, received Status Code: %d. Message %s",
			resp.StatusCode,
			payload,
		)
	}
	player.Config.Tokens.AppAccessToken = payload["access_token"].(string)
}

func (player *Spotify) exchangeUserCodeForAccessTokens(code string) (string, string) {
	client := new(http.Client)
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "http://localhost:"+player.Config.RedirectPort)

	const appAuthURL = "https://accounts.spotify.com/api/token"
	req, _ := http.NewRequest("POST", appAuthURL, strings.NewReader(data.Encode()))
	appIdentity := []byte(player.Config.AppClientID + ":" + player.Config.AppClientSecret)
	b64encode := base64.StdEncoding.EncodeToString(appIdentity)
	req.Header.Set("Authorization", "Basic "+b64encode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	utils.CheckHTTPResponse(resp, err, "Failure in exchanging user auth code for tokens.")

	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)
	access_token, refresh_token := payload["access_token"].(string), payload["refresh_token"].(string)

	player.Config.Tokens.UserRefreshToken = refresh_token
	player.Config.Tokens.UserAccessToken = access_token

	return access_token, refresh_token
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
	client := new(http.Client)
	const URL = "https://api.spotify.com/v1/me/player/play"
	req, _ := http.NewRequest("PUT", URL, nil)
	fmt.Println("\n\n\n" + player.Config.Tokens.UserAccessToken)
	req.Header.Set("Authorization", "Bearer "+player.Config.Tokens.UserAccessToken)
	resp, err := client.Do(req)
	utils.Check(err)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)
	fmt.Println(payload)
}

func (player *Spotify) Pause() {
	client := new(http.Client)
	const URL = "https://api.spotify.com/v1/me/player/pause"
	req, _ := http.NewRequest("PUT", URL, nil)
	fmt.Println("\n\n\n" + player.Config.Tokens.UserAccessToken)
	req.Header.Set("Authorization", "Bearer "+player.Config.Tokens.UserAccessToken)
	resp, err := client.Do(req)
	utils.Check(err)
	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)
	fmt.Println(payload)
}

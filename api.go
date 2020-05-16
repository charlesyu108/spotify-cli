package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var config configT

type configT struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
}

const configFile = "config.json"

// loadConfig unpacks the JSON config file
func loadConfig() {
	fileContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		log.Fatal(err)
	}
}

func getAuthToken() {
	client := new(http.Client)
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	appIdentity := []byte(config.ClientID + ":" + config.ClientSecret)
	b64encode := base64.StdEncoding.EncodeToString(appIdentity)
	req.Header.Set("Authorization", "Basic "+b64encode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)

	var payload map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&payload)

	config.AccessToken = payload["access_token"].(string)
}

func search(q string) string {
	client := new(http.Client)
	req, _ := http.NewRequest("GET", formatString("https://api.spotify.com/v1/search?q=%s&type=track,artist", q), nil)
	req.Header.Set("Authorization", "Bearer "+config.AccessToken)

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	type resourceT struct {
		Items []map[string]interface{}
	}

	type payloadT struct {
		Tracks  resourceT
		Artists resourceT
		Albums  resourceT
	}

	var payload payloadT
	json.NewDecoder(resp.Body).Decode(&payload)

	if len(payload.Tracks.Items) > 0 {
		return payload.Tracks.Items[0]["uri"].(string)
	}
	if len(payload.Artists.Items) > 0 {
		return payload.Artists.Items[0]["uri"].(string)
	}
	if len(payload.Albums.Items) > 0 {
		return payload.Albums.Items[0]["uri"].(string)
	}
	return ""

}

func processRedirect(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req)
}

func authorizeUser() {
	//display url
	authUrl := formatString(`https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=http://localhost:4567&scope=user-read-playback-state,user-modify-playback-state`, config.ClientID)
	log.Println(authUrl)
	http.HandleFunc("/", processRedirect)
	http.ListenAndServe(":4567", nil)
}

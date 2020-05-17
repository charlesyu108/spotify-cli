package spotifyplayer

import (
	"fmt"
	"testing"
)

func TestAuthorizeReal(t *testing.T) {
	cfg := GetConfig()
	player := NewWebPlayer(cfg)
	player.AuthorizeApp()

	if !player.isAuthorized {
		t.Fatalf("Player is not Authorized.")
	}

	if player.Config.AccessToken == "" {
		t.Fatalf("Expected Access Token but found empty string.")
	}
}

func TestAuthorizeBadAppClientConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	cfg := ConfigT{}
	player := NewWebPlayer(&cfg)
	player.AuthorizeApp()
	t.Fatalf("No error was thrown")
}

func TestSearchOk(t *testing.T) {
	// The Less I Know The Better - Tame Impala
	expected := "spotify:track:6K4t31amVTZDgR3sKmwUJJ"
	cfg := GetConfig()
	player := NewWebPlayer(cfg)
	player.AuthorizeApp()

	result := player.search("the less i know the better", []string{"track"})
	if result != expected {
		t.Fatalf("Found %s", result)
	}
}

func TestSearchNoCatProvided(t *testing.T) {
	// The Less I Know The Better - Tame Impala
	expected := "spotify:track:6K4t31amVTZDgR3sKmwUJJ"
	cfg := GetConfig()
	player := NewWebPlayer(cfg)
	player.AuthorizeApp()

	result := player.search("the less i know the better", nil)
	if result != expected {
		t.Fatalf("Found %s", result)
	}
}

func TestSearchBadCatProvided(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	// Empty string when nothing is found
	cfg := GetConfig()
	player := NewWebPlayer(cfg)
	player.AuthorizeApp()

	result := player.search("the less i know the better", []string{"trackssss"})
	t.Fatal("Should have errored.", result)
}

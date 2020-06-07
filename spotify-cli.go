package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charlesyu108/spotify-cli/spotify"
	"github.com/charlesyu108/spotify-cli/utils"
	"github.com/urfave/cli/v2"
)

// ConfigFile defines which config JSON file to load
const ConfigFile = "config.json"

func main() {

	app := &cli.App{
		Name:                 "spotify-cli",
		Usage:                "Use Spotify from the Command Line.",
		EnableBashCompletion: true,
	}

	app.Commands = []*cli.Command{
		// Define Playback category commands.
		{
			Name:     "play",
			Category: "Playback",
			Usage:    "Play/Resume playback. Can also specify something to play and switch playback device.",
			Action:   handlePlay,
			Aliases:  []string{"pl"},
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "device", Aliases: []string{"d"}, Usage: "Play music on a device. Use any partial identifier i.e. 'mbp', '064a', 'smartphone', etc."},
				&cli.StringFlag{Name: "track", Aliases: []string{"t"}, Usage: "A track to play."},
				&cli.StringFlag{Name: "album", Aliases: []string{"m"}, Usage: "An album to play."},
				&cli.StringFlag{Name: "artist", Aliases: []string{"r"}, Usage: "An artist to play."},
				&cli.StringFlag{Name: "playlist", Aliases: []string{"l"}, Usage: "An playlist to play."},
			},
		},
		{
			Name:     "pause",
			Category: "Playback",
			Usage:    "Pause playback.",
			Aliases:  []string{"ps"},
			Action:   handlePause,
		},
		{
			Name:     "next",
			Category: "Playback",
			Usage:    "Skip to next track.",
			Aliases:  []string{"nx"},
			Action:   handleNextTrack,
		},
		{
			Name:     "prev",
			Category: "Playback",
			Usage:    "Skip to last track.",
			Aliases:  []string{"pv"},
			Action:   handlePrevTrack,
		},
		{
			Name:      "volume",
			Category:  "Playback",
			Usage:     "Adjust the volume.",
			Aliases:   []string{"v"},
			Action:    handleVolume,
			ArgsUsage: "<volume-percent>",
		},
		{
			Name:      "shuffle",
			Category:  "Playback",
			Usage:     "Toggle shuffle.",
			Aliases:   []string{"s"},
			Action:    handleShuffle,
			ArgsUsage: "{on|off}",
		},
		// Define Info category commands.
		{
			Name:     "devices",
			Category: "Info",
			Usage:    "Show playable devices.",
			Aliases:  []string{"d"},
			Action:   handleDevices,
		},
		{
			Name:     "info",
			Category: "Info",
			Usage:    "Show what's currently playing and playback state.",
			Aliases:  []string{"i"},
			Action:   handleInfo,
		},
		// Define Config category commands.
		{
			Name:     "config",
			Category: "Configuration",
			Usage:    "Configure spotify-cli settings.",
			Aliases:  []string{"c"},
			Action:   handleConfig,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "set-app-client-id", Usage: "Set 'AppClientID'"},
				&cli.StringFlag{Name: "set-app-client-secret", Usage: "Set 'AppClientSecret'"},
				&cli.StringFlag{Name: "set-redirect-port", Usage: "Set 'RedirectPort'"},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getConfig() *spotify.ConfigT {

	progFilesDir := utils.GetProgFilesDir()
	// Make sure ~/.spotify-cli exists, create if not
	_ = os.Mkdir(progFilesDir, 0700)
	configPath := filepath.Join(progFilesDir, ConfigFile)

	cfg, created := spotify.LoadConfig(configPath)
	if created {
		fmt.Printf("No config file was found, so one was created for you at `%s`.\n", configPath)
		fmt.Printf("Edit the config file with your Spotify Application credentials or use the command `config` to help you.\n")
		spotify.SaveConfig(cfg, configPath)
		os.Exit(0)
	}
	return cfg
}

func handleConfig(c *cli.Context) error {
	progFilesDir := utils.GetProgFilesDir()
	// Make sure ~/.spotify-cli exists, create if not
	_ = os.Mkdir(progFilesDir, 0700)
	configPath := filepath.Join(progFilesDir, ConfigFile)
	cfg, _ := spotify.LoadConfig(configPath)

	id, secret, port := c.String("set-app-client-id"), c.String("set-app-client-secret"), c.String("set-redirect-port")

	if id != "" {
		cfg.AppClientID = id
		fmt.Printf("Set AppClientID.\n")
	}

	if secret != "" {
		cfg.AppClientSecret = secret
		fmt.Printf("Set AppClientSecret.\n")
	}

	if port != "" {
		cfg.RedirectPort = port
		fmt.Printf("Set RedirectPort.\n")
	}

	spotify.SaveConfig(cfg, configPath)

	if cfg.Validate() != nil {
		fmt.Printf("Configs were saved but errors were found.\n")
		fmt.Printf("For help setting these configs view the README.md or visit http://github.com/charlesyu108/spotify-cli.\n")
		os.Exit(1)
	}

	return nil
}

func handlePlay(c *cli.Context) error {
	cfg := getConfig()
	Spotify := spotify.New(cfg)
	Spotify.Authorize()

	device := c.String("device")
	track := c.String("track")
	album := c.String("album")
	artist := c.String("artist")
	playlist := c.String("playlist")

	switch true {
	case device != "":
		search := strings.ToLower(device)
		for _, d := range Spotify.GetDevices() {

			id, name, t := strings.ToLower(d.ID), strings.ToLower(d.Name), strings.ToLower(d.Type)
			if strings.Contains(id, search) ||
				strings.Contains(name, search) ||
				strings.Contains(t, search) {

				Spotify.PlayOnDevice(d)
				return nil
			}
		}
		fmt.Printf("Could not find any devices '%s'.\n", device)
		os.Exit(1)

	case track != "":
		uri := Spotify.SimpleSearch(track, "track")
		Spotify.PlayURI(uri)

	case album != "":
		uri := Spotify.SimpleSearch(album, "album")
		Spotify.PlayURI(uri)

	case artist != "":
		uri := Spotify.SimpleSearch(artist, "artist")
		Spotify.PlayURI(uri)

	case playlist != "":
		uri := Spotify.SimpleSearch(playlist, "playlist")
		Spotify.PlayURI(uri)

	default:
		Spotify.Play()
	}

	defer deferredTrackInfo(&Spotify)

	return nil
}

func handlePause(c *cli.Context) error {
	cfg := getConfig()
	Spotify := spotify.New(cfg)
	Spotify.Authorize()
	Spotify.Pause()
	return nil
}

func handleNextTrack(c *cli.Context) error {
	cfg := getConfig()
	Spotify := spotify.New(cfg)
	Spotify.Authorize()
	Spotify.NextTrack()
	defer deferredTrackInfo(&Spotify)
	return nil
}

func handlePrevTrack(c *cli.Context) error {
	cfg := getConfig()
	Spotify := spotify.New(cfg)
	Spotify.Authorize()
	Spotify.PreviousTrack()
	defer deferredTrackInfo(&Spotify)
	return nil
}

func handleVolume(c *cli.Context) error {
	volArg := c.Args().Get(0)
	if volArg == "" {
		fmt.Printf("Positional argument `volume-percent` not provided.\n")
		os.Exit(1)
	}
	vol, _ := strconv.Atoi(volArg)
	cfg := getConfig()
	Spotify := spotify.New(cfg)
	Spotify.Authorize()
	Spotify.Volume(vol)
	return nil
}

func handleDevices(c *cli.Context) error {
	cfg := getConfig()
	Spotify := spotify.New(cfg)
	Spotify.Authorize()
	fmt.Printf("[DeviceID]\t\t\t\t\tDeviceType\tName\n")
	for _, d := range Spotify.GetDevices() {
		fmt.Printf("[%s]\t%s\t%s\n", d.ID, d.Type, d.Name)
	}
	return nil
}

func handleInfo(c *cli.Context) error {
	cfg := getConfig()
	Spotify := spotify.New(cfg)
	Spotify.Authorize()
	displayTrackInfo(&Spotify)
	return nil
}

func displayTrackInfo(spotify *spotify.Spotify) {
	state := spotify.CurrentState()
	isPlayingDesc := "Paused"
	if state.IsPlaying {
		isPlayingDesc = "Playing"
	}

	var trackInfo string
	if state.Track.Name != "" {
		artistNames := []string{}
		for _, art := range state.Track.Artists {
			artistNames = append(artistNames, art.Name)
		}
		artistsString := strings.Join(artistNames, ", ")
		trackInfo = fmt.Sprintf(":: %s - %s", state.Track.Name, artistsString)
	}

	fmt.Printf("=> %s %s\n", isPlayingDesc, trackInfo)
}

func deferredTrackInfo(spotify *spotify.Spotify) {
	time.Sleep(200 * time.Millisecond)
	displayTrackInfo(spotify)
}

func handleShuffle(c *cli.Context) error {
	cfg := getConfig()
	Spotify := spotify.New(cfg)
	Spotify.Authorize()

	switch shuffleArg := c.Args().Get(0); shuffleArg {
	case "on":
		Spotify.ToggleShuffle(true)
		fmt.Printf("Shuffle toggled on.\n")
	case "off":
		Spotify.ToggleShuffle(false)
		fmt.Printf("Shuffle toggled off.\n")
	default:
		fmt.Printf("Positional argument `toggle` must be one of {on | off}.\n")
		os.Exit(1)
	}
	return nil
}

package spotifyplayer

import (
	"strings"
	"testing"
)

func player() *OSXPlayer {
	cfg := new(ConfigT)
	return NewOSXPlayer(cfg)
}

func TestState(t *testing.T) {
	p := player()
	p.State()
}

func TestTrackInfo(t *testing.T) {
	p := player()
	p.TrackInfo()
}

func TestPlay(t *testing.T) {
	p := player()
	t.Cleanup(func() { p.Pause() })

	p.Play()
	if ok := strings.Contains(p.State(), "playing"); !ok {
		t.FailNow()
	}
}

func TestPlayResource(t *testing.T) {
	p := player()
	t.Cleanup(func() { p.Pause() })

	// URI is The Less I Know The Better - Tame Impala
	p.PlayResource("spotify:track:6K4t31amVTZDgR3sKmwUJJ")
	if ok := strings.Contains(p.State(), "playing"); !ok {
		t.FailNow()
	}
	if ok := strings.Contains(p.TrackInfo(), "Tame Impala"); !ok {
		t.Fatalf(
			"Assertion Error. Expected to find 'Tame Impala' but got %s",
			p.TrackInfo(),
		)
	}
}

func TestPlayPause(t *testing.T) {
	p := player()
	t.Cleanup(func() { p.Pause() })

	stateBefore := p.State()
	p.PlayPause()
	stateAfter := p.State()
	if stateBefore == stateAfter {
		t.FailNow()
	}
}

func TestNextTrack(t *testing.T) {
	p := player()
	t.Cleanup(func() { p.Pause() })

	p.NextTrack()
	if ok := strings.Contains(p.State(), "playing"); !ok {
		t.FailNow()
	}
}

func TestPrevTrack(t *testing.T) {
	p := player()
	t.Cleanup(func() { p.Pause() })

	p.PrevTrack()
	if ok := strings.Contains(p.State(), "playing"); !ok {
		t.FailNow()
	}
}

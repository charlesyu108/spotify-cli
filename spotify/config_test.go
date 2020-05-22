package spotify

import (
	"os"
	"testing"
)

var validateTest = []struct {
	name        string
	cfg         *ConfigT
	expectError bool
}{
	{"No fields", &ConfigT{}, true},
	{"AppClientID Only", &ConfigT{AppClientID: "test123"}, true},
	{"AppClientSecret Only", &ConfigT{AppClientSecret: "test123"}, true},
	{"RedirectPort Only", &ConfigT{RedirectPort: "test123"}, true},
	{"Valid", &ConfigT{AppClientID: "test123", AppClientSecret: "test123", RedirectPort: "test123"}, false},
}

func TestValidateConfig(t *testing.T) {
	for _, tt := range validateTest {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.Validate()
			errWasFound := err != nil
			if errWasFound != tt.expectError {
				t.Errorf("Error Returned? %v but Expected %v", errWasFound, tt.expectError)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {

	t.Run("When proper file exists", func(t *testing.T) {
		file := ".tmp"
		fptr, _ := os.Create(file)
		fptr.WriteString(`{"AppClientID": "test123", "AppClientSecret": "test123", "RedirectPort": "test123"}\n`)
		fptr.Close()

		cfg, created := LoadConfig(file)
		if created || cfg.Validate() != nil {
			t.FailNow()
		}
		t.Cleanup(func() {
			os.Remove(file)
		})
	})

	t.Run("When file does not exist", func(t *testing.T) {
		file := ".tmpasdf123"
		cfg, created := LoadConfig(file)
		// A file should be created and an empty config is loaded
		if !created || (*cfg != ConfigT{}) {
			t.FailNow()
		}
		t.Cleanup(func() {
			os.Remove(file)
		})
	})
}

func TestSaveConfig(t *testing.T) {

	t.Run("Saves config to a file that didn't exist", func(t *testing.T) {
		file := ".tmpasdf123"
		cfg := &ConfigT{AppClientID: "test123", AppClientSecret: "test123", RedirectPort: "test123"}
		SaveConfig(cfg, file)

		fptr, err := os.Open(file)
		fptr.Close()

		if err != nil {
			t.FailNow()
		}

		t.Cleanup(func() {
			os.Remove(file)
		})
	})

	t.Run("Saves config to a file that did exist", func(t *testing.T) {
		file := ".tmpasdf123"
		fptr, err := os.Create(file)

		cfg := &ConfigT{AppClientID: "test123", AppClientSecret: "test123", RedirectPort: "test123"}
		SaveConfig(cfg, file)

		fptr, err = os.Open(file)
		fptr.Close()

		if err != nil {
			t.FailNow()
		}

		t.Cleanup(func() {
			os.Remove(file)
		})
	})
}

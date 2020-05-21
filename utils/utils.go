package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// Check for errors
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// CheckHTTPResponse runs Check and validates the response is good.
func CheckHTTPResponse(resp *http.Response, e error, errMsg string) {
	Check(e)
	if resp.StatusCode >= 400 {
		log.Panicf("Error in HTTP request. errMsg: %s. StatusCode: %d.", errMsg, resp.StatusCode)
	}
}

// FormatString is like Sprintf except it's only for strings and strings are trimmed before formatting.
func FormatString(template string, strs ...string) string {
	trimmed := make([]interface{}, len(strs))
	for idx, s := range strs {
		trimmed[idx] = strings.TrimSpace(s)
	}
	return fmt.Sprintf(template, trimmed...)
}

// OpenInBrowser opens a url in the default browser
// Source: https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func OpenInBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}

// LoadJSON loads the file into the struct defined by v.
func LoadJSON(fileName string, v interface{}) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(v)
}

// SaveJSON saves the struct defined by v to the file
func SaveJSON(fileName string, v interface{}) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	file.Truncate(0)
	file.Seek(0, 0)
	return json.NewEncoder(file).Encode(v)
}

// MakeHTTPRequest wraps http.NewRequest and client.Do to perform a request
func MakeHTTPRequest(method string, URL string, headers map[string]string, body string) (*http.Response, error) {
	client := new(http.Client)
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, _ := http.NewRequest(method, URL, reader)

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return client.Do(req)
}

// GetHomeDir gets the Current User's home directory.
func GetHomeDir() string {
	usr, _ := user.Current()
	return usr.HomeDir
}

// GetProgFilesDir gets the spotify-cli Program Files directory
func GetProgFilesDir() string {
	return filepath.Join(GetHomeDir(), ".spotify-cli")
}

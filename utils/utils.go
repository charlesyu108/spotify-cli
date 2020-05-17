package utils

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
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

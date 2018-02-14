// Package lokalise provides functions to access the Lokalise web API.
// Each function has some required arguments and an options argument to modify the API behaviour.
//
// An API token is at minimum required. Information on how to generate one can be found at the
// web API documentation at https://lokalise.co/apidocs.
package lokalise

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL    = "https://api.lokalise.co/api/"
	assetURL   = "https://s3-eu-west-1.amazonaws.com/lokalise-assets/"
	timeout    = 10 * time.Second
	timeFormat = "2006-01-02 15:04:05"
)

func callAPI(req *http.Request) (*http.Response, error) {
	client := http.Client{
		Timeout: timeout,
	}
	return client.Do(req)
}

func api(path string) string {
	return baseURL + path
}

type response struct {
	Status  string `json:"status"`
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

const ()

// A Time represents an instant in time with second precision.
// It follows the Lokalise time format "2006-01-02 15:04:05".
// The type embeds time.Time and can be used as such.
type Time struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		t.Time = time.Time{}
		return nil
	}
	time, err := time.Parse(timeFormat, s)
	if err != nil {
		return err
	}
	t.Time = time
	return nil
}

func jsonArray(s []string) *string {
	if len(s) == 0 {
		return nil
	}
	values := fmt.Sprintf("['%s']", strings.Join(s, "','"))
	return &values
}

func boolString(b *bool) *string {
	if b == nil {
		return nil
	}
	v := "0"
	if *b {
		v = "1"
	}
	return &v
}

package lokalise

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Bundle represents file locations for a project export bundle. If a webhook URL was
// specified in the ExportOptions the FullFile field is not set.
type Bundle struct {
	File     string `json:"file"`
	FullFile string `json:"full_file"`
}

type exportResponse struct {
	Bundle   Bundle   `json:"bundle"`
	Response response `json:"response"`
}

// Export initiates an export of project with ID projectID in file type fileType and returns the
// file locations for the export bundle.
//
// Customize the import by setting any ExportOptions.
//
// If option WithWebhookURL() is set the FullFile field is not set on the bundle.
//
// In case of API request errors an error of type Error is returned.
func Export(apiToken, projectID, fileType string, opts ...ExportOption) (Bundle, error) {
	form := &url.Values{}
	form.Add("api_token", apiToken)
	form.Add("id", projectID)
	form.Add("type", fileType)
	for _, opt := range opts {
		err := opt(form)
		if err != nil {
			return Bundle{}, err
		}
	}

	req, err := http.NewRequest("POST", api("project/export"), strings.NewReader(form.Encode()))
	if err != nil {
		return Bundle{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := callAPI(req)
	if err != nil {
		return Bundle{}, err
	}
	if err := errorFromStatus(resp); err != nil {
		return Bundle{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return Bundle{}, err
	}
	var dat exportResponse
	if err := json.Unmarshal(body, &dat); err != nil {
		return Bundle{}, err
	}
	if err := errorFromResponse(dat.Response); err != nil {
		return Bundle{}, err
	}
	if len(form.Get("webhook_url")) != 0 {
		dat.Bundle.FullFile = assetURL + dat.Bundle.File
	}
	return dat.Bundle, nil
}

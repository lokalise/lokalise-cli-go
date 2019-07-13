package lokalise

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// ImportResult represents the outcome of a file upload.
type ImportResult struct {
	Skipped  int64 `json:"skipped"`
	Inserted int64 `json:"inserted"`
	Updated  int64 `json:"updated"`
}

type importResponse struct {
	Result   ImportResult `json:"result"`
	Response response     `json:"response"`
}

// Import uploads a file with translations in language langISO to a Lokalise project with ID projectID.
//
// Customize the import by setting any ImportOptions.
//
// In case of API request errors an error of type Error is returned.
func Import(apiToken, projectID, file, langISO string, opts ...ImportOption) (ImportResult, error) {
	request, err := newfileUploadRequest(apiToken, projectID, file, langISO, opts...)
	if err != nil {
		return ImportResult{}, err
	}
	resp, err := callAPI(request)
	if err != nil {
		return ImportResult{}, err
	}
	if err := errorFromStatus(resp); err != nil {
		return ImportResult{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var dat importResponse
	if err := json.Unmarshal(body, &dat); err != nil {
		return ImportResult{}, err
	}

	if err := errorFromResponse(dat.Response); err != nil {
		return ImportResult{}, err
	}
	return dat.Result, nil
}

func newfileUploadRequest(apiToken, projectID, path, langISO string, opts ...ImportOption) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))

	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("api_token", apiToken)
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("id", projectID)
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("lang_iso", langISO)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		for _, opt := range opts {
			err := opt(writer)
			if err != nil {
				return nil, err
			}
		}
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", api("project/import"), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	return req, nil
}

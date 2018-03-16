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

// ImportOptions represents available options for importing files to a project.
type ImportOptions struct {
	Replace       	    *bool
	ConvertPlaceholders *bool
	FillEmpty           *bool
	Distinguish         *bool
	Hidden              *bool
	UseTransMem         *bool
	IncludePath         *bool
	Tags                []string
	ReplaceBreaks       *bool
	IcuPlurals          *bool
}

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
// Customize the import by setting any options in opts.
//
// In case of API request errors an error of type Error is returned.
func Import(apiToken, projectID, file, langISO string, opts *ImportOptions) (ImportResult, error) {
	request, err := newfileUploadRequest(apiToken, projectID, file, langISO, opts)
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

func newfileUploadRequest(apiToken, projectID, path, langISO string, opts *ImportOptions) (*http.Request, error) {
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

	err = multipartAdd(writer, "api_token", &apiToken)
	if err != nil {
		return nil, err
	}
	multipartAdd(writer, "id", &projectID)
	if err != nil {
		return nil, err
	}
	multipartAdd(writer, "lang_iso", &langISO)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		multipartAdd(writer, "tags", jsonArray(opts.Tags))
		if err != nil {
			return nil, err
		}
		multipartAdd(writer, "fill_empty", boolString(opts.FillEmpty))
		if err != nil {
			return nil, err
		}
		multipartAdd(writer, "hidden", boolString(opts.Hidden))
		if err != nil {
			return nil, err
		}
		multipartAdd(writer, "distinguish", boolString(opts.Distinguish))
		if err != nil {
			return nil, err
		}
		multipartAdd(writer, "replace", boolString(opts.Replace))
		if err != nil {
			return nil, err
		}
		multipartAdd(writer, "convert_placeholders", boolString(opts.ConvertPlaceholders))
		if err != nil {
			return nil, err
		}
		multipartAdd(writer, "use_trans_mem", boolString(opts.UseTransMem))
		if err != nil {
			return nil, err
		}
		multipartAdd(writer, "replace_breaks", boolString(opts.ReplaceBreaks))
		if err != nil {
			return nil, err
		}
		multipartAdd(writer, "icu_plurals", boolString(opts.IcuPlurals))
		if err != nil {
			return nil, err
		}
		if opts.IncludePath != nil {
			multipartAdd(writer, "filename", &path)
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

func multipartAdd(writer *multipart.Writer, field string, value *string) error {
	if value == nil {
		return nil
	}
	return writer.WriteField(field, *value)
}

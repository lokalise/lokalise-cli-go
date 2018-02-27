package lokalise

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// ExportOptions represents available options for a project export request.
type ExportOptions struct {
	Languages            []string
	UseOriginal          *bool
	Filter               []string
	BundleStructure      *string
	WebhookURL           *string
	ExportAll            *bool
	ExportEmpty          *string
	IncludeComments      *bool
	IncludePIDs          []string
	Tags                 []string
	ExportSort           *string
	ReplaceBreaks        *bool
	YAMLIncludeRoot      *bool
	JSONUnescapedSlashes *bool
	NoLanguageFolders    *bool
	Triggers             []string
	PluralFormat         []string
	ICUNumeric	     *bool
	PlaceholderFormat    []string
}

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
// If WebhookURL is set in opts the FullFile field is not set on the bundle.
//
// In case of API request errors an error of type Error is returned.
func Export(apiToken, projectID, fileType string, opts *ExportOptions) (Bundle, error) {
	form := &url.Values{}
	formAdd(form, "api_token", &apiToken)
	formAdd(form, "id", &projectID)
	formAdd(form, "type", &fileType)
	formAdd(form, "langs", jsonArray(opts.Languages))
	formAdd(form, "use_original", boolString(opts.UseOriginal))
	formAdd(form, "filter", jsonArray(opts.Filter))
	formAdd(form, "bundle_structure", opts.BundleStructure)
	formAdd(form, "webhook_url", opts.WebhookURL)
	formAdd(form, "export_all", boolString(opts.ExportAll))
	formAdd(form, "export_empty", opts.ExportEmpty)
	formAdd(form, "include_comments", boolString(opts.IncludeComments))
	formAdd(form, "include_pids", jsonArray(opts.IncludePIDs))
	formAdd(form, "tags", jsonArray(opts.Tags))
	formAdd(form, "export_sort", opts.ExportSort)
	formAdd(form, "replace_breaks", boolString(opts.ReplaceBreaks))
	formAdd(form, "no_language_folders", boolString(opts.NoLanguageFolders))
	formAdd(form, "yaml_include_root", boolString(opts.YAMLIncludeRoot))
	formAdd(form, "json_unescaped_slashes", boolString(opts.JSONUnescapedSlashes))
	formAdd(form, "triggers", jsonArray(opts.Triggers))
	formAdd(form, "plural_format", jsonArray(opts.PluralFormat))
	formAdd(form, "icu_numeric", boolString(opts.ICUNumeric))
	formAdd(form, "placeholder_format", jsonArray(opts.PlaceholderFormat))

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
	if opts.WebhookURL == nil {
		dat.Bundle.FullFile = assetURL + dat.Bundle.File
	}
	return dat.Bundle, nil
}

func formAdd(v *url.Values, field string, value *string) {
	if value == nil {
		return
	}
	v.Add(field, *value)
}

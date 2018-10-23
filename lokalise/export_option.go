package lokalise

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// ExportOption is a function setting options for an export request.
type ExportOption func(*url.Values) error

// WithLanguages returns an ExportOption setting what languages to export. If
// omitted all language are exported.
func WithLanguages(languages ...string) ExportOption {
	return stringArrayField("langs", languages)
}

// WithOriginal returns an ExportOption setting whether to use orignal
// filenames and formates languages to export.
func WithOriginal(enabled bool) ExportOption {
	return boolField("use_original", enabled)
}

// WithFilter returns an ExportOption setting a filter on the export data
// range. Allows values are 'translated', 'nonfuzzy' and 'nonhidden'.
func WithFilter(values ...string) ExportOption {
	return stringArrayField("filter", values, allowedSliceStrings("translated", "nonfuzzy", "nonhidden", "reviewed", "proofread"))
}

// WithBundleStructure returns an ExportOption setting the bundle structure.
// Used when exporting all keys to a single file per language with WithOriginal(false).
//
// Available placeholders are %LANG_ISO%, %LANG_NAME%, %FORMAT% and %PROJECT_NAME%.
//
// Example:
//   locale/%LANG_ISO%.%FORMAT%
//
// Option is ignored if WithOriginal(true) is set.
func WithBundleStructure(structure string) ExportOption {
	return stringField("bundle_structure", structure)
}

// WithDirectoryPrefix returns an ExportOption setting the directory prefix of
// the bundle.
// Used when exporting keys to previously assigned filenames with
// WithOriginal(true).
//
// Available placeholders are %LANG_ISO%.
//
// Example:
//   %LANG_ISO%/
//
// Option is ignored if WithOriginal(false) is set.
func WithDirectoryPrefix(prefix string) ExportOption {
	return stringField("directory_prefix", prefix)
}

// WithWebhookURL returns an ExportOption setting a webhook to call when
// the export is completed.
//
// The webhook is an HTTP POST request with payload:
//
//  file=export/Sample_locale.zip
//
// When receiving the webhook, prepend 'https://s3-eu-west-1.amazonaws.com/lokalise-assets/'
// to the filename in order to download the bundle.
//
// Use http.Request.PostFormValue() to get the filename:
//
//  func httpHandler(w http.ResponseWriter, req *http.Request) {
//    filename := req.PostFormValue("file")
//    // prepend assets URL and download bundle
//    w.WriteHeader(http.StatusOK)
//  }
func WithWebhookURL(webhook string) ExportOption {
	return stringField("webhook_url", webhook)
}

// WithAll returns an ExportOption setting whether to include all platform
// keys.
func WithAll(enabled bool) ExportOption {
	return boolField("export_all", enabled)
}

// WithEmpty returns an ExportOption setting empty string export preferences.
// Allowed values are "empty", "base" and "skip".
func WithEmpty(value string) ExportOption {
	return stringField("export_empty", value, allowedStrings("empty", "base", "skip"))
}

// WithComments returns an ExportOption setting whether to include key
// comments AND description (if supported by format).
func WithComments(enabled bool) ExportOption {
	return boolField("include_comments", enabled)
}

// WithDescription returns an ExportOption setting whether to include key
// description ONLY (if supported by format).
func WithDescription(enabled bool) ExportOption {
	return boolField("include_description", enabled)
}

// WithPIDs returns an ExportOption setting other projects ID's to be included
// with this export.
func WithPIDs(pids ...string) ExportOption {
	return stringArrayField("include_pids", pids)
}

// WithIncludeTags returns an ExportOption setting tags to limit export range.
// Only keys with provided tags are included in export.
func WithIncludeTags(tags ...string) ExportOption {
	return stringArrayField("include_tags", tags)
}

// WithExcludeTags returns an ExportOption setting tags to exclude in export range.
// Keys with provided tags are excluded in export.
func WithExcludeTags(tags ...string) ExportOption {
	return stringArrayField("exclude_tags", tags)
}

// WithSortOrder returns an ExportOption setting the sort order of exported keys.
//
// Allowed values are "first_added", "last_added", "last_updated", "a_z", "z_a".
func WithSortOrder(order string) ExportOption {
	return stringField("export_sort", order, allowedStrings("first_added", "last_added", "last_updated", "a_z", "z_a"))
}

// WithJavaPropertiesSeparator returns an ExportOption setting of the separator for .properties files export.
//
// Allowed values are ":", "=".
func WithJavaPropertiesSeparator(order string) ExportOption {
	return stringField("java_properties_separator", order, allowedStrings(":", "="))
}

// WithJavaPropertiesEncoding returns an ExportOption setting of the encoding for .properties files export.
//
// Allowed values are "utf-8", "latin-1".
func WithJavaPropertiesEncoding(order string) ExportOption {
	return stringField("java_properties_encoding", order, allowedStrings("utf-8", "latin-1"))
}

// WithExportReplaceBreaks returns an ExportOption setting whether to replace '\n' with
// line breaks.
func WithExportReplaceBreaks(enabled bool) ExportOption {
	return boolField("replace_breaks", enabled)
}

// WithYAMLRoot returns an ExportOption setting whether to include language ISO
// code as root key.
//
// Only available for YAML exports set with the type argument to 'yaml'.
func WithYAMLRoot(enabled bool) ExportOption {
	return boolField("yaml_include_root", enabled)
}

// WithJSONUnescapedSlashes returns an ExportOption setting whether to leave
// forward slashes unescaped.
//
// Only available for JSON exports set with the type argument to 'json'.
func WithJSONUnescapedSlashes(enabled bool) ExportOption {
	return boolField("json_unescaped_slashes", enabled)
}

// WithNoLanguageFolders returns an ExportOption setting whether to not use a
// directory prefix.
//
// This is a legacy option. Use WithDirectoryPrefix("") instead.
func WithNoLanguageFolders(enabled bool) ExportOption {
	return boolField("no_language_folders", enabled)
}

// WithTriggers returns an ExportOption setting what integration exports to trigger.
//
// Ensure this feature is enabled in project settings before use.
//
// Allowed values are "amazon", "gcs" and "github".
func WithTriggers(triggers ...string) ExportOption {
	return stringArrayField("triggers", triggers, allowedSliceStrings("amazons3", "gcs", "github", "gitlab", "bitbucket"))
}

// WithRepos returns an ExportOption setting what repos to include when repo integrations are triggered.
//
func WithRepos(repos ...string) ExportOption {
	return stringArrayField("repos", repos)
}

// WithPluralFormat returns an ExportOption overriding the default plural
// format for the file type.
//
// Allowed values are "json_string", "icu", "array" and "generic".
func WithPluralFormat(format string) ExportOption {
	return stringField("plural_format", format, allowedStrings("json_string", "icu", "array", "generic", "symfony"))
}

// WithICUNumeric returns an ExportOption setting whether the plural forms
// "zero", "one" and "two" is replaced with "=0", "=1", "=2" respectively.
//
// Only available for when setting WithPluralFormat("icu")
func WithICUNumeric(enabled bool) ExportOption {
	return boolField("icu_numeric", enabled)
}

// WithPercentEscape returns an ExportOption setting whether all universal percent
// placeholders "[%]" will be always exported as "%%".
//
// Only works for printf placeholder format.
func WithPercentEscape(enabled bool) ExportOption {
	return boolField("escape_percent", enabled)
}

// WithIndentation returns an ExportOption overriding the default indentation
//
// Allowed values are "1sp", "2sp", "3sp", "4sp", "5sp", "6sp", "7sp", "8sp", "tab",
func WithIndentation(format string) ExportOption {
	return stringField("indentation", format, allowedStrings("1sp", "2sp", "3sp", "4sp", "5sp", "6sp", "7sp", "8sp", "tab"))
}

// WithPlaceholderFormat returns an ExportOption overriding the default
// placeholder format for the file type.
//
// Allowed values are "printf", "ios", "icu" and "net".
func WithPlaceholderFormat(format string) ExportOption {
	return stringField("placeholder_format", format, allowedStrings("printf", "ios", "icu", "net", "symfony"))
}

func boolField(field string, value bool) ExportOption {
	return func(v *url.Values) error {
		v.Add(field, boolString(value))
		return nil
	}
}

type validator func(string) error

func stringField(field, value string, validators ...validator) ExportOption {
	return func(v *url.Values) error {
		for _, validator := range validators {
			err := validator(value)
			if err != nil {
				return err
			}
		}
		v.Add(field, strings.TrimSpace(value))
		return nil
	}
}

func allowedStrings(allowedValues ...string) validator {
	return func(value string) error {
		for _, v := range allowedValues {
			if value == v {
				return nil
			}
		}
		return restrictedErr(allowedValues, value)
	}
}

type sliceValidator func([]string) error

func stringArrayField(field string, values []string, validators ...sliceValidator) ExportOption {
	return func(v *url.Values) error {
		for _, validator := range validators {
			err := validator(values)
			if err != nil {
				return err
			}
		}
		for i := range values {
			values[i] = strings.TrimSpace(values[i])
		}
		jsonValues, err := json.Marshal(values)
		if err != nil {
			return err
		}
		v.Add(field, string(jsonValues))
		return nil
	}
}

func allowedSliceStrings(allowedValues ...string) sliceValidator {
	return func(values []string) error {
		for _, value := range values {
			var allowed bool
			for _, v := range allowedValues {
				if value == v {
					allowed = true
					break
				}
			}
			if !allowed {
				return restrictedErr(allowedValues, value)
			}
		}
		return nil
	}
}

func restrictedErr(allowed []string, value string) error {
	return fmt.Errorf("lokalise: allowed values '%s': value '%s' not allowed", strings.Join(allowed, "', '"), value)
}

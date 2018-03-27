package lokalise

import (
	"encoding/json"
	"mime/multipart"
)

// ImportOption is a function setting options for an import request.
type ImportOption func(*multipart.Writer) error

// WithReplace returns an ImportOption setting whether to replace existing
// translations of the imported keys.
func WithReplace(enabled bool) ImportOption {
	return func(w *multipart.Writer) error {
		return w.WriteField("replace", boolString(enabled))
	}
}

// WithConvertPlaceholders returns an ImportOption setting whether to convert
// placeholders to Lokalise universal ones. Enabled by default.
// See https://docs.lokalise.co/developer-docs/universal-placeholders
func WithConvertPlaceholders(enabled bool) ImportOption {
	return func(w *multipart.Writer) error {
		return w.WriteField("convert_placeholders", boolString(enabled))
	}
}

// WithICUPlurals returns an ImportOption setting whether to automatically
// detect and parse ICU formatted plurals.
func WithICUPlurals(enabled bool) ImportOption {
	return func(w *multipart.Writer) error {
		return w.WriteField("icu_plurals", boolString(enabled))
	}
}

// WithFillEmpty returns an ImportOption setting whether to fill empty values
// with keys.
func WithFillEmpty(enabled bool) ImportOption {
	return func(w *multipart.Writer) error {
		return w.WriteField("fill_empty", boolString(enabled))
	}
}

// WithDistinguish returns an ImportOption setting whether to distinguish
// similar keys in different files.
func WithDistinguish(enabled bool) ImportOption {
	return func(w *multipart.Writer) error {
		return w.WriteField("distinguish", boolString(enabled))
	}
}

// WithTranslationMemory returns an ImportOption setting whether to use
// automatically fill 100% translation memory matches.
func WithTranslationMemory(enabled bool) ImportOption {
	return func(w *multipart.Writer) error {
		return w.WriteField("use_trans_mem", boolString(enabled))
	}
}

// WithHidden returns an ImportOption setting whether to hide newly added keys
// from contributors.
func WithHidden(enabled bool) ImportOption {
	return func(w *multipart.Writer) error {
		return w.WriteField("hidden", boolString(enabled))
	}
}

// WithTags returns an ImportOption setting a list of tags for newly added
// keys.
func WithTags(tags ...string) ImportOption {
	return func(w *multipart.Writer) error {
		jsonTags, err := json.Marshal(tags)
		if err != nil {
			return err
		}
		return w.WriteField("tags", string(jsonTags))
	}
}

// WithFilename returns an ImportOption setting an override of the filename.
func WithFilename(filename string) ImportOption {
	return func(w *multipart.Writer) error {
		return w.WriteField("filename", filename)
	}
}

// WithImportReplaceBreaks returns an ImportOption setting whether to replace '\n' with
// line breaks.
func WithImportReplaceBreaks(enabled bool) ImportOption {
	return func(w *multipart.Writer) error {
		return w.WriteField("replace_breaks", boolString(enabled))
	}
}

package lokalise

import (
	"fmt"
	"net/http"
)

// Code is an API request status code.
type Code string

const (
	// OK indicates a successful request.
	OK Code = "200"
	// MissingAPIToken indicates a missing API token in the request.
	MissingAPIToken Code = "401"
	// InvalidAPIToken indicates an invalid API token in the request.
	InvalidAPIToken Code = "4011"
	// NoData indicates no data in the body of an HTTP POST request.
	NoData Code = "4012"
	// AccessDenied indicates missing permissions to access the requested resource.
	AccessDenied Code = "403"
	// InvalidCall indicates an invalid API request.
	InvalidCall Code = "404"
	// Custom indicates a custom error. Refer to field Message of Error for details.
	Custom Code = "4040"
	// NotJSON indicates a non-JSON payload in a request where JSON was expected.
	NotJSON Code = "4042"
	// WrongLanguageCode indicates an invalid language code.
	WrongLanguageCode Code = "4043"
	// LanguageNotAvailable indicates that the requested language is not available for the project.
	LanguageNotAvailable Code = "4044"
	// LanguageNotSpecified indicates that the language is missing in the request.
	LanguageNotSpecified Code = "4045"
	// InvalidFile indicates that an unsupported file format is used in the request.
	InvalidFile Code = "4046"
	// InvalidExportType indicates an invalid export type is used in the request.
	InvalidExportType Code = "4047"
	// RateLimit indicates to many requests in a short period of time.
	RateLimit Code = "4048"
	// MissingRequestParameter indicates a missing required parameter.
	MissingRequestParameter Code = "4049"
	// LanguageExist indicates the specified language already exist for project.
	LanguageExist Code = "4050"
)

// Error represents an API request error. When the API is not able to complete a request
// an error code and possibly a message is returned indicating why the request failed.
type Error struct {
	Code    Code
	Message string
}

// Error implements the error interface.
func (err *Error) Error() string {
	return fmt.Sprintf("lokalise: %s %s", err.Code, err.Message)
}

// Assert that Error implements the error interface.
var _ error = &Error{}

func errorFromResponse(resp response) error {
	if resp.Status != "error" {
		return nil
	}
	return &Error{
		Code:    resp.Code,
		Message: resp.Message,
	}
}

func errorFromStatus(resp *http.Response) error {
	if resp == nil {
		return nil
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return &Error{
		Code:    Custom,
		Message: fmt.Sprintf("api request did not respond with status 200. Got %s", resp.Status),
	}
}

package lokalise

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Project is the data model for a Lokalise project.
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"desc"`
	Created     Time   `json:"created"`
	Owner       string `json:"owner"`
}

type listResponse struct {
	Projects []Project `json:"projects"`
	Response response  `json:"response"`
}

// List returns a slice of projects available for the apiToken.
//
// In case of API request errors an error of type Error is returned.
func List(apiToken string) ([]Project, error) {
	request, err := http.NewRequest(http.MethodGet, api("project/list?api_token="+apiToken), nil)
	if err != nil {
		return nil, err
	}
	resp, err := callAPI(request)
	if err != nil {
		return nil, err
	}
	if err := errorFromStatus(resp); err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var dat listResponse
	if err := json.Unmarshal(body, &dat); err != nil {
		return nil, err
	}
	if err := errorFromResponse(dat.Response); err != nil {
		return nil, err
	}
	return dat.Projects, nil
}

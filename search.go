package splunk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// SearchResponse is the response from a search
type SearchResponse struct {
	Preview bool         `json:"preview,omitempty"`
	Last    bool         `json:"lastrow,omitempty"`
	Offset  int          `json:"offset,omitempty"`
	Result  SearchResult `json:"result"`
}

// SearchResults is a collection of SearchResult
type SearchResults []SearchResult

// SearchResult is a map->interface of search results
type SearchResult map[string]interface{}

// Search will search a query (must start with "search" term), and then
// output the data in the specified format, JSON is default
func (client *Client) Search(search, output string) (SearchResults, []error) {
	var errors []error
	var results SearchResults
	// Create the data search object
	data := url.Values{}
	data.Set("search", search)

	// Set default output type ot CSV
	if output == "" {
		output = "json"
	}
	data.Set("output_mode", output)

	/* Authenticate */
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		}, // ignore expired SSL certificates
	}
	httpClient := &http.Client{Transport: transCfg}
	req, err := http.NewRequest("POST", client.buildURL(SearchEndpoint), bytes.NewBufferString(data.Encode()))
	if err != nil {
		errors = append(errors, err)
		return results, errors
	}

	// Add in authorization
	req.Header.Set("Authorization", fmt.Sprintf("Splunk %s", client.Auth.SessionKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		errors = append(errors, err)
		return results, errors
	}
	defer resp.Body.Close()

	switch output {
	case "json":
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errors = append(errors, err)
			return results, errors
		}
		lines := strings.Split(string(responseBody), "\n")
		for _, line := range lines {
			if line == "\n" || len(line) <= 1 {
				continue
			}
			// Decode all the JSON
			var result SearchResponse
			err := json.Unmarshal([]byte(line), &result)
			if err != nil {
				errors = append(errors, fmt.Errorf("Result Parsing error: %s", err))
				continue
			}
			results = append(results, result.Result)
		}
	}

	return results, errors
}

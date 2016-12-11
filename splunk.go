package splunk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// AuthEndpoint is the endpoint to authenticate to
var AuthEndpoint = "/servicesNS/%username%/search/auth/login"

// SearchEndpoint is the endpoint to search against
var SearchEndpoint = "/servicesNS/%username%/search/search/jobs/export"

type Client struct {
	Hostname string
	Context  string
	Auth     Auth
}

type Auth struct {
	XMLName    xml.Name `xml:"response"`
	Username   string
	SessionKey string `xml:"sessionKey"`
}

type SplunkSearchResult struct {
	Preview bool         `json:"preview,omitempty"`
	Last    bool         `json:"lastrow,omitempty"`
	Offset  int          `json:"offset,omitempty"`
	Result  SearchResult `json:"result"`
}

type SearchResults []SearchResult
type SearchResult map[string]interface{}

// buildURL generates a url for communications
func (client *Client) buildURL(url string) string {
	url = strings.Replace(url, "%username%", client.Auth.Username, -1)
	return fmt.Sprintf(client.Hostname + url)
}

// NewClient will create a new Splunk API client object
func NewClient(host, user, pass string) (Client, error) {
	// Create our client object
	client := Client{Hostname: host}
	client.Auth.Username = user

	// Attempt to login
	if err := client.Authenticate(user, pass); err != nil {
		return client, fmt.Errorf("Error authenticating: %s", err.Error())
	}

	return client, nil
}

func (client *Client) Authenticate(username, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Add("password", password)

	/* Authenticate */
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		}, // ignore expired SSL certificates
	}
	httpClient := &http.Client{Transport: transCfg}
	req, err := http.NewRequest("POST", client.buildURL(AuthEndpoint), bytes.NewBufferString(data.Encode()))
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if its
	if resp.Status != "200 OK" {
		return fmt.Errorf("Login failed")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Unmarshal the response
	var auth Auth
	err = xml.Unmarshal(body, &auth)
	if err != nil {
		return fmt.Errorf("> Error unmarshaling authentication string: %s", body)
	}

	// Get the sessionkey
	client.Auth.SessionKey = auth.SessionKey
	return nil
}

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
			var result SplunkSearchResult
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

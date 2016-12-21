package splunk

import (
	"bytes"
	"crypto/tls"
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

// Version set for releases
var Version = "v1.0.4"

// Client is used to persist context
type Client struct {
	Hostname string
	Context  string
	Auth     Auth
}

// Auth controls basic xml struct
type Auth struct {
	XMLName    xml.Name `xml:"response"`
	Username   string
	SessionKey string `xml:"sessionKey"`
}

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

// Authenticate will authenticate the client to the Splunk server
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
	if err != nil {
		return err
	}

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

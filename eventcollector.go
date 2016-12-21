package splunk

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// EventCollectorClient struct
type EventCollectorClient struct {
	Request *http.Request
}

// NewEvent will send a new event to the Splunk server
func NewEvent(host, token string) *EventCollectorClient {
	e := &EventCollectorClient{}

	// Start generating the new request
	req, err := http.NewRequest("POST", host+"/services/collector/event", nil)
	if err != nil {
		fmt.Printf(err.Error())
		return e
	}
	req.Header.Set("User-Agent", "Go-Splunk/"+Version)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Splunk "+token)

	// Set the request within Client
	e.Request = req
	return e
}

// Data will set the data within the request
func (e *EventCollectorClient) Data(data interface{}) error {
	formattedEvent := map[string]interface{}{"event": data}
	b, err := json.Marshal(formattedEvent)
	if err != nil {
		return err
	}

	// Set the body to your jsonize'd map[string]interface{}
	e.Request.Body = ioutil.NopCloser(strings.NewReader(string(b)))
	return nil
}

// Send will transport the request to the remote splunk server
func (e *EventCollectorClient) Send() error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Make request
	resp, err := client.Do(e.Request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If it's not 200OK, let the caller know
	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid response: %d", resp.StatusCode)
	}

	return nil
}

package main

import (
	"fmt"

	splunk "github.com/mikemackintosh/go-splunk"
)

var splunkHost = ""
var splunkHECToken = ""

func main() {
	e := splunk.NewEvent(splunkHost, splunkHECToken)
	e.Data(map[string]interface{}{"Test": "THIS IS AWESOME"})
	if err := e.Send(); err != nil {
		fmt.Printf("Error sending payload: %s", err)
	}
}

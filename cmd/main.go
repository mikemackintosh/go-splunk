package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	splunk "github.com/mikemackintosh/go-splunk"
)

func main() {
	if ok, err := configureEnvironment(".splunk.conf"); !ok {
		fmt.Printf("> ERROR - Config: %s", err)
		os.Exit(1)
	}

	client, err := splunk.NewClient(os.Getenv("SPLUNK_HOSTNAME"), os.Getenv("SPLUNK_USERNAME"), os.Getenv("SPLUNK_PASSWORD"))
	if err != nil {
		fmt.Printf("> ERROR - Client: %s", err)
		os.Exit(1)
	}

	res, errors := client.Search("SEARCH STRING HERE", "json")
	if len(errors) > 0 {
		for i, err := range errors {
			fmt.Printf("\033[38;5;196m>\033[0m ERROR %d - Client: %s", i, err)
		}
	}

	for _, r := range res {
		for k, v := range r {
			fmt.Printf("%s => %v\n", k, v)
		}
		break
	}

}

// configureEnvironment takes a configuration file and loads
// each entry into the process environment for inclusion
func configureEnvironment(configFile string) (bool, error) {
	if _, err := os.Stat(configFile); err == nil {
		file, err := os.Open(configFile)
		if err != nil {
			return true, fmt.Errorf("No configuration file found. Skipping.")
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// trim the line from all leading whitespace first
			line := strings.TrimLeft(scanner.Text(), " \t")
			// line is not empty, and not starting with '#'
			if len(line) > 0 && !strings.HasPrefix(line, "#") {
				data := strings.SplitN(line, "=", 2)

				// If it parsed correctly, grab the key, value
				if len(data) == 2 {
					os.Setenv(data[0], data[1])
				}
			}
		}

		// If there was an error running scan, then return
		if err := scanner.Err(); err != nil {
			return false, fmt.Errorf("Unable to read configuration file.")
		}
	}

	// Looks good, return
	return true, nil
}

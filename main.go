package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// HTTP endpoint configuration: name, url, method, headers, body
type Endpoint struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
}

// statistics for each HTTP endpoint
type Stats struct {
	totalRequests int
	upRequests int
}

func main() {
	// 1. Accept an input argument to a file path
	if len(os.Args) != 2 {
		log.Fatal("Please provide a file path")
	}
	// 2. Parse YAML file to extract HTTP endpoint configuration
	endpoints, err := parseFile(os.Args[1])
	if err != nil {
		log.Fatalf("Error parsing file: %v", err)
	}
	// 3. Initialize + populate a map to store statistics for each endpoint
	stats := make(map[string]*Stats)
	for _, endpoint := range endpoints {
		domain, err := getDomain(endpoint.URL)
		if err != nil {
			log.Fatalf("Error parsing domain: %v", err)
		}
		if _, exists := stats[domain]; !exists {
			stats[domain] = &Stats{}
		}
	}
	// 4. Run checks and log stats
	runCheck(endpoints, stats)
	printAvailability(stats)
	// 5. Initialize ticker to repeat every 15 seconds
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	// 6. Create channel to receive interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	for {
		select {
		case <-ticker.C:
			runCheck(endpoints, stats)
			printAvailability(stats)
		case <-sig:
			// fmt.Println("Received interrupt signal, exiting...")
			return
		}
	}
}

// YAML parsing
func parseFile(path string) ([]Endpoint, error) {
	// 1. Read input config file
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	var endpoints []Endpoint
	// 2. parse YAML into endpoints slice
	if err := yaml.Unmarshal(data, &endpoints); err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
	}
	// 3. fill in method - empty default to GET
	for i := range endpoints {
		if endpoints[i].Method == "" {
			endpoints[i].Method = http.MethodGet
		}
	}
	// print out for verification
	// for _, endpoint := range endpoints {
	// 	fmt.Printf("Name: %s, URL: %s, Method: %s, Headers: %v, Body: %s\n",
	// 		endpoint.Name, endpoint.URL, endpoint.Method, endpoint.Headers, endpoint.Body)
	// }
	return endpoints, nil
}

// Health check
func runCheck(endpoints []Endpoint, stats map[string]*Stats) {
	for _, endpoint := range endpoints {
		startTime := time.Now() // for calculating response latency
		// 1. Create HTTP request
		req, err := http.NewRequest(endpoint.Method, endpoint.URL, strings.NewReader(endpoint.Body))
		if err != nil {
			// since this is valid url from previou check -> assume DOWN
			updateStats(stats, endpoint.URL, false)
			continue;
		}
		// 2. Add headers to request
		for k, v := range endpoint.Headers {
			req.Header.Add(k, v)
		}
		// 3. Send request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// no response -> assume DOWN
			updateStats(stats, endpoint.URL, false)
			continue;
		}
		latency := time.Since(startTime)
		// 4. UP only when any 200–299 response code && latency < 500 ms
		defer resp.Body.Close()
		checkStatus := resp.StatusCode >= 200 && resp.StatusCode < 300
		checkLatency := latency < 500 * time.Millisecond
		if checkStatus && checkLatency {
			updateStats(stats, endpoint.URL, true)
		} else {
			updateStats(stats, endpoint.URL, false)
		}
	}
}

// Log availability percentages to the console
func printAvailability(stats map[string]*Stats) {
	// Extract keys and sort them
    keys := make([]string, 0, len(stats))
    for key := range stats {
        keys = append(keys, key)
    }
    sort.Strings(keys)

    // enforce ordering as Go map iteration is random
    for _, domain := range keys {
        stat := stats[domain]
        // round to nearest whole percentage
        availability := int(math.Round(float64(stat.upRequests) / float64(stat.totalRequests) * 100))
        fmt.Printf("%s has %d%% availability percentage\n", domain, availability)
    }
}

/***********************************************
 *  HELPERS
 **********************************************/
// extract domain from url
func getDomain(target string) (string, error) {
	parsedURL, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}

// update stats
func updateStats(stats map[string]*Stats, url string, up bool) {
	domain, _ := getDomain(url)
	stat, exists := stats[domain]
	if !exists { // should NEVER happen
		// stat = &Stats{}
		// stats[domain] = stat
		return
	}
	stat.totalRequests++
	if up {
		stat.upRequests++
	}
}




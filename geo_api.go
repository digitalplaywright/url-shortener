package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//GeoAPI interacts with telize to do reverse geoip lookups
type GeoAPI struct {
}

// fetchUrl This function fetch the content of a URL will return it as an
// array of bytes if retrieved successfully.
func (s *GeoAPI) fetchURL(url string) ([]byte, error) {
	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// At this point we're done - simply return the bytes
	return body, nil
}

// lookup will attempt to get the IP record for
// a given IP. If no errors occur, it will return a pair
// of the record and nil. If it was not successful, it will
// return a pair of nil and the error.
func (s *GeoAPI) lookup(ip string) (*IPRecord, error) {
	// Fetch the JSON content for that given IP
	content, err := s.fetchURL(
		fmt.Sprintf("http://www.telize.com/geoip/%s", ip))
	if err != nil {
		fmt.Println("Failed to fetch GeoIP information")
		// An error occurred while fetching the JSON
		return nil, err
	}
	// Fill the record with the data from the JSON
	var record IPRecord
	err = json.Unmarshal(content, &record)
	if err != nil {
		// An error occurred while converting our JSON to an object
		return nil, err
	}
	return &record, err
}

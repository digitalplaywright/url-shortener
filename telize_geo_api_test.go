package main

import "testing"

func TestLookup(t *testing.T) {
	api := TelizeGeoAPI{}

	_, err := api.lookup("74.125.239.40")

	if err != nil {
		t.Error("ip lookup failed")
	}
}

func TestFailedLookup(t *testing.T) {
	api := TelizeGeoAPI{}

	_, err := api.lookup("invalid")

	if err == nil {
		t.Error("ip lookup should have failed, but didn't")
	}
}

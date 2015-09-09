package main

import "testing"

var (
	db = RedisDB{geoAPI: &MockGeoAPI{}}
)

func openDB() {
	if db.db == nil {
		db.initDB()
	}
}

func TestShortenURL(t *testing.T) {

	openDB()
	longURL := "http://www.wired.com"
	shortURL, err := db.shortenURL(longURL)

	if err != nil {
		t.Error(shortURL)
	}
}

func TestGetLongURL(t *testing.T) {
	openDB()

	longURL := "http://www.wired.com"
	shortURL, _ := db.shortenURL(longURL)

	_, err := db.getLongURL(shortURL, "74.125.239.40")

	if err != nil {
		t.Error(err.Error())

	}

}

func TestInvalidGetLongURL(t *testing.T) {
	openDB()

	_, err := db.getLongURL("INVALID", "74.125.239.40")

	if err == nil {
		t.Error("invalid URL should return nothing")

	}
}

func TestGetStatistics(t *testing.T) {
	openDB()

	longURL := "http://www.wired.com"
	shortURL, _ := db.shortenURL(longURL)

	db.getLongURL(shortURL, "74.125.239.40")

	_, err := db.getStatistics(shortURL)

	if err != nil {
		t.Error(err.Error())

	}

}

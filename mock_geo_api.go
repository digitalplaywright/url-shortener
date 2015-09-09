package main

//MockGeoAPI mocks geo API lookups
type MockGeoAPI struct {
}

// lookup mocks getting the IP record for an ip address
func (s *MockGeoAPI) lookup(ip string) (*IPRecord, error) {
	return &IPRecord{IP: ip, CountryName: "USA", CountryCode: "US",
		RegionCode: "CA", City: "San Francisco"}, nil

}

package main

//IPRecord the geo ip information about ip address after a successful lookup
type IPRecord struct {
	// These two fields use the json: tag to specify which field they map to
	CountryName string `json:"country_name"`
	CountryCode string `json:"country_code"`

	RegionCode string `json:"region_code"`

	// These fields are mapped directly by name (note the different case)
	City string
	IP   string

	// As these fields can be nullable, we use a pointer
	// to a string rather than a string
	Lat *string
	Lng *string
}

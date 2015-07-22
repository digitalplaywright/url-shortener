package main

import "time"

//IPElement contains a ip-address with a record id and is
//used to communicate background jobs to process geo ip information
//from request handlers to ProcessIP
type IPElement struct {
	IP   string
	id   string
	Time time.Time
}

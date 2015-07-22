package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

//ProcessIP processes demographic info
type ProcessIP struct {

	//API to do geo ip lookups
	api interface {
		lookup(string) (*IPRecord, error)
	}

	pool           *redis.Pool
	ipAddressQueue chan IPElement
}

//isLocalHost determines if we are currently running in development
func (s *ProcessIP) isLocalHost(ipAddress string) bool {
	return strings.Contains(ipAddress, "::1") || strings.Contains(ipAddress, "127.0.0.1")
}

func (s *ProcessIP) start() {
	c := s.pool.Get()
	defer c.Close()

	for {
		ipRecord := <-s.ipAddressQueue

		//FIXME: A hack for testing purposes so that we can have a
		//valid IP Address during development on localhost.
		if s.isLocalHost(ipRecord.IP) == true {
			ipRecord.IP = "74.125.239.40"
		}

		record, err := s.api.lookup(ipRecord.IP)
		if err != nil {
			fmt.Println("lookup error - failed to do geoip lookup for ip address " + ipRecord.IP)

		} else {

			//increase the metrics in the database
			c.Do("INCR", "demoClick:"+ipRecord.id)
			c.Do("HINCRBY", "demoCountry:"+ipRecord.id, record.CountryCode, 1)
			c.Do("HINCRBY", "demoRegion:"+ipRecord.id, record.RegionCode, 1)

		}
		time.Sleep(time.Second * 1)
	}
}

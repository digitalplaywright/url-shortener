package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

//RedisDB is an abstraction of all redis operations
//
//The resulting redis schema will be, where `shortURL` is the shortened url
//for the current entry:
//    - url:shortURL -> { "shortURL" -> shortURL,
//                        "longURL"  -> "url to redirect to"   }
//    - demoClick:shortURL   -> number of clicks for url in total
//    - demoCountry:shortURL -> hash of number of clicks from country
//    - demoRegion:shortURL  -> hash of number of clicks from region
type RedisDB struct {
	db             *redis.Pool
	ipAddressQueue chan IPElement

	//geoAPI to do geo ip lookups
	geoAPI interface {
		lookup(string) (*IPRecord, error)
	}
}

//initDB initializes the redis pool
func (r *RedisDB) initDB() {

	//DB setup
	flag.Parse()

	r.db = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			return nil, err
		}

		return c, err
	}, *maxConnections)

	//jobs in the background for optimization purposes
	//
	//creates a queue to fetch demographic information from
	//the IP address in the backogrund
	r.ipAddressQueue = make(chan IPElement, 2)

	go r.backgroundJob()
}

//backgroundJob processes demographic info in the b
func (r *RedisDB) backgroundJob() {
	c := r.db.Get()
	defer c.Close()

	for {
		ipRecord := <-r.ipAddressQueue

		record, err := r.geoAPI.lookup(ipRecord.IP)
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

//shortenURL takes a site url and returns a shortened version
func (r *RedisDB) shortenURL(url string) (string, error) {
	c := r.db.Get()
	defer c.Close()

	//Get unique id
	c.Do("INCR", "next_url_id")

	key, urlIDErr := c.Do("GET", "next_url_id")
	if urlIDErr != nil {
		log.Fatal("Failed to fetch next id from redis")
	}

	urlKey, urlErr := strconv.ParseInt(string(key.([]byte)), 10, 0)

	if urlErr != nil {
		return "", urlErr
	}

	//store shortened url and render success page
	storeKey := strconv.FormatInt(urlKey, 10)

	c.Do("HSET", "url:"+storeKey, "shortUrl", storeKey)
	c.Do("HSET", "url:"+storeKey, "longUrl", url)
	c.Do("SET", "demoClick:"+storeKey, 0)

	return storeKey, nil
}

//getLongURL takes a shortened url and returns the external site URL
func (r *RedisDB) getLongURL(shortURL string, remoteAddr string) (string, error) {
	c := r.db.Get()
	defer c.Close()

	longURL, err := redis.String(c.Do("HGET", "url:"+shortURL, "longUrl"))

	if err == nil {
		r.ipAddressQueue <- IPElement{remoteAddr, shortURL, time.Now()}
	}

	return longURL, err

}

//getStatistics takes a shortened url and returns the external site URL
func (r *RedisDB) getStatistics(key string) (*map[string]interface{}, error) {

	c := r.db.Get()
	defer c.Close()

	longURL, lErr := redis.String(c.Do("HGET", "url:"+key, "longUrl"))
	if lErr != nil {
		return nil, lErr
	}

	demoClick, cErr := redis.String(c.Do("GET", "demoClick:"+key))
	if cErr != nil {
		return nil, cErr
	}

	demoCountry, coErr := redis.StringMap(c.Do("HGETALL", "demoCountry:"+key))
	if coErr != nil {
		return nil, coErr
	}

	demoRegion, drErr := redis.StringMap(c.Do("HGETALL", "demoRegion:"+key))
	if drErr != nil {
		return nil, drErr
	}

	return &map[string]interface{}{
		"longURL":     longURL,
		"shortURL":    "http://localhost:9999/" + key,
		"demoClick":   demoClick,
		"demoRegion":  demoRegion,
		"demoCountry": demoCountry,
	}, nil

}

package main

import (
	"flag"
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
}

//initDB initializes the redis pool
func (r *RedisDB) initDB() {

	redisAddress := flag.String("redis-address", ":6379", "Address to the Redis server")
	maxConnections := flag.Int("max-connections", 10, "Max connections to Redis")

	//DB setup
	flag.Parse()

	r.db = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			return nil, err
		}

		return c, err
	}, *maxConnections)

	r.initIPElementChannel()

}

//initIPElementChannel creates a queue to fetch demographic information from
//the IP address in the backogrund
func (r *RedisDB) initIPElementChannel() {

	r.ipAddressQueue = make(chan IPElement, 2)

	q := &ProcessIP{
		api:            &GeoAPI{},
		pool:           r.db,
		ipAddressQueue: r.ipAddressQueue,
	}

	go q.start()

}

//shortenURL takes a site url and returns a shortened version
func (r *RedisDB) shortenURL(url string) string {
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
		log.Fatal("Failed to parse integer from redis key")
	}

	//store shortened url and render success page
	storeKey := strconv.FormatInt(urlKey, 10)

	c.Do("HSET", "url:"+storeKey, "shortUrl", storeKey)
	c.Do("HSET", "url:"+storeKey, "longUrl", url)
	c.Do("SET", "demoClick:"+storeKey, 0)

	return storeKey
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

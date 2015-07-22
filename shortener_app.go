package main

import (
	"flag"
	"log"
	"net/http"

	"html/template"
	"path/filepath"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

//ShortenerApp Represents the app itself
type ShortenerApp struct {
	templates      map[string]*template.Template
	redisPool      *redis.Pool
	ipAddressQueue chan IPElement
}

//initTemplates parses all templates using the current layout
func (s *ShortenerApp) initTemplates() {
	s.templates = map[string]*template.Template{}

	tmplGlobs, err := filepath.Glob("templates/includes/*.html")
	if err != nil {
		log.Fatal(err)
	}

	// Generate our templates map from our layouts/ and includes/ directories
	for _, tmpl := range tmplGlobs {
		s.templates[filepath.Base(tmpl)] = template.Must(template.ParseFiles(tmpl, "templates/layouts/layout.html"))
	}
}

//initDB initializes the redis pool
func (s *ShortenerApp) initDB() {

	redisAddress := flag.String("redis-address", ":6379", "Address to the Redis server")
	maxConnections := flag.Int("max-connections", 10, "Max connections to Redis")

	//DB setup
	flag.Parse()

	s.redisPool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			return nil, err
		}

		return c, err
	}, *maxConnections)

}

//create a queue to fetch demographic information from
//the IP address in the backogrund
func (s *ShortenerApp) initIPElementChannel() {

	s.ipAddressQueue = make(chan IPElement, 2)

	q := &ProcessIP{
		api:            &GeoAPI{},
		pool:           s.redisPool,
		ipAddressQueue: s.ipAddressQueue,
	}

	go q.start()

}

func (s *ShortenerApp) start() {
	s.initDB()
	s.initIPElementChannel()
	s.initTemplates()

	defer s.redisPool.Close()

	server := &Server{
		db:             s.redisPool,
		ipAddressQueue: s.ipAddressQueue,
		templates:      s.templates,
	}

	router := mux.NewRouter()
	router.HandleFunc("/", server.RootHandler).Methods("GET")
	router.HandleFunc("/set", server.SetHandler).Methods("GET")
	router.HandleFunc("/set", server.SetPostHandler).Methods("POST")
	router.HandleFunc("/statistics/{id:[0-9]+}", server.StatisticsHandler).Methods("GET")
	router.HandleFunc("/{id:[0-9]+}", server.FetchURL).Methods("GET")

	log.Fatal(http.ListenAndServe(":9999", router))

}

//NewShortenerApp creates a new shortener app instance
func NewShortenerApp() *ShortenerApp {
	app := ShortenerApp{}

	return &app

}

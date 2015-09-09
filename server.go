package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

//Server holds the variables that are used in the
//request handlers.
type Server struct {

	//A general abstaction of a database
	//makes it easier to test in isolation
	db interface {
		shortenURL(string) (string, error)
		getLongURL(string, string) (string, error)
		getStatistics(string) (*map[string]interface{}, error)
	}

	//Parsed vesion of the templates
	templates map[string]*template.Template
}

//NewServer creates a new server instance
func NewServer() *Server {
	db := RedisDB{geoAPI: &TelizeGeoAPI{}}
	db.initDB()

	server := &Server{
		db: &db,
	}

	server.initTemplates()

	return server
}

//initTemplates parses all templates using the current layout
func (s *Server) initTemplates() {
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

//start initializes and starts serving the app
func (s *Server) start() {

	//defer s.redisPool.Close()

	router := mux.NewRouter()
	router.HandleFunc("/", s.GetRoot).Methods("GET")
	router.HandleFunc("/set", s.NewItem).Methods("GET")
	router.HandleFunc("/set", s.PostItem).Methods("POST")
	router.HandleFunc("/{id:[0-9]+}", s.GetItem).Methods("GET")
	router.HandleFunc("/statistics/{id:[0-9]+}", s.GetItemStatistics).Methods("GET")

	log.Fatal(http.ListenAndServe(":9999", router))

}

// renderTemplate is a wrapper around template.ExecuteTemplate.
// It writes into a bytes.Buffer before writing to the http.ResponseWriter to catch
// any errors resulting from populating the template.
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	// Ensure the template exists in the map.
	tmpl, ok := s.templates[name]
	if !ok {
		return fmt.Errorf("The template %s does not exist.", name)
	}

	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		return err
	}

	// Set the header and write the buffer to the http.ResponseWriter
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return nil

}

//RenderErrorPage renders an error page.
func (s *Server) RenderErrorPage(w http.ResponseWriter, message string) {
	w.WriteHeader(400)

	data := map[string]interface{}{}
	s.renderTemplate(w, "error.html", data)

}

//GetRoot serves the root route
func (s *Server) GetRoot(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	s.renderTemplate(w, "root.html", data)

}

//NewItem serves the form to submit a URL for shortening
func (s *Server) NewItem(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	s.renderTemplate(w, "set.html", data)
}

//PostItem shortens a URL
func (s *Server) PostItem(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	url := r.FormValue("url")

	storeKey, err := s.db.shortenURL(url)

	if err != nil {
		s.RenderErrorPage(w, err.Error())

	} else {
		redirectURL := "/statistics/" + storeKey
		http.Redirect(w, r, redirectURL, 303)

	}

}

//GetItem redirects a shortened URL to the location and collects
//information about the requestor
func (s *Server) GetItem(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["id"]

	url, err := s.db.getLongURL(key, r.RemoteAddr)

	if err != nil {
		s.RenderErrorPage(w, err.Error())

	} else {
		http.Redirect(w, r, url, 303)

	}

}

//GetItemStatistics represents the statistics about a URL
func (s *Server) GetItemStatistics(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["id"]

	data, err := s.db.getStatistics(key)

	if err != nil {
		s.RenderErrorPage(w, err.Error())

	} else {
		s.renderTemplate(w, "statistics.html", *data)

	}

}

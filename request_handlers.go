package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

//Server holds the variables that are used in the
//request handlers.
type Server struct {
	db             *redis.Pool
	ipAddressQueue chan IPElement
	templates      map[string]*template.Template
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

	c := s.db.Get()
	defer c.Close()

	//Get unique id
	c.Do("INCR", "next_url_id")
	key, urlIDErr := c.Do("GET", "next_url_id")

	if urlIDErr != nil {
		//I believe this can only happen if the redis db is down
		s.RenderErrorPage(w, "Could not shorten your url. Please try again.")

	} else {

		urlKey, urlErr := strconv.ParseInt(string(key.([]byte)), 10, 0)

		if urlErr != nil {
			//TODO: if client-side html5 validation does not work and the
			//form is submitted with an invalid URL we need to handle that
			//properly
			message := "Could not shorten your url. Please try again."
			s.RenderErrorPage(w, message)

		} else {
			//store shortened url and render success page
			storeKey := strconv.FormatInt(urlKey, 10)

			c.Do("HSET", "url:"+storeKey, "shortUrl", storeKey)
			c.Do("HSET", "url:"+storeKey, "longUrl", url)
			c.Do("SET", "demoClick:"+storeKey, 0)

			redirectURL := "/statistics/" + storeKey
			http.Redirect(w, r, redirectURL, 303)

		}
	}

}

//GetItem redirects a shortened URL to the location and collects
//information about the requestor
func (s *Server) GetItem(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["id"]

	c := s.db.Get()
	defer c.Close()

	url, err := redis.String(c.Do("HGET", "url:"+key, "longUrl"))

	if err != nil {
		message := fmt.Sprintf("Could not GET %s", key)
		s.RenderErrorPage(w, message)
	} else {

		s.ipAddressQueue <- IPElement{r.RemoteAddr, key, time.Now()}

		http.Redirect(w, r, url, 303)

	}
}

//GetItemStatistics represents the statistics about a URL
func (s *Server) GetItemStatistics(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["id"]

	c := s.db.Get()
	defer c.Close()

	shortURL, err := redis.String(c.Do("HGET", "url:"+key, "shortUrl"))
	shortURL = "http://localhost:3000/" + shortURL

	if err != nil {
		message := fmt.Sprintf("Could not GET %s", key)
		s.RenderErrorPage(w, message)

	} else {

		longURL, lErr := redis.String(c.Do("HGET", "url:"+key, "longUrl"))
		if lErr != nil {
			fmt.Fprint(w, "db error - failed to locate longUrl:"+key)
		}

		demoClick, cErr := redis.String(c.Do("GET", "demoClick:"+key))
		if cErr != nil {
			fmt.Fprint(w, "db error - failed to locate demoClick:"+key)
		}

		demoCountry, coErr := redis.StringMap(c.Do("HGETALL", "demoCountry:"+key))
		if coErr != nil {
			fmt.Fprint(w, "db error - failed to locate demoCountry:"+key)
		}

		demoRegion, drErr := redis.StringMap(c.Do("HGETALL", "demoRegion:"+key))
		if drErr != nil {
			fmt.Fprint(w, "db error - failed to locate demoRegion:"+key)
		}

		if err != nil {
			message := "Problems accessing url " + key
			s.RenderErrorPage(w, message)

		} else {

			data := map[string]interface{}{
				"longURL":     longURL,
				"shortURL":    shortURL,
				"demoClick":   demoClick,
				"demoRegion":  demoRegion,
				"demoCountry": demoCountry,
			}

			s.renderTemplate(w, "statistics.html", data)

		}

	}
}

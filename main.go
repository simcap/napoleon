package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

var elasticHost = "elastic"
var elasticPort = 9200

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/search", search)
	http.ListenAndServe(":8080", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("home.html")
	if err != nil {
		log.Printf("Error parsing template %v", err)
	}
	t.Execute(w, nil)
}

type searchResults struct {
	Hits struct {
		Hits []hits
	}
}

type hits struct {
	Source struct {
		Path string
		Text string
	} `json:"_source"`
}

func search(w http.ResponseWriter, r *http.Request) {
	results := new(searchResults)
	query := r.URL.Query().Get("q")

	if len(query) < 4 {
		log.Printf("search term '%s' too short\n", query)
		json.NewEncoder(w).Encode(results)
		return
	}

	elasticurl, err := url.Parse(fmt.Sprintf("http://%s:%d/codecivil/article/_search", elasticHost, elasticPort))
	parameters := url.Values{}
	parameters.Add("size", "100")
	parameters.Add("q", "Text:"+query)
	elasticurl.RawQuery = parameters.Encode()

	log.Printf("Searching for %s\n", elasticurl)

	resp, err := http.Get(elasticurl.String())
	if err != nil {
		log.Printf("error searching for %s\n%s", elasticurl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(results)
	json.NewEncoder(w).Encode(results)
}

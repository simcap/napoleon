package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
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

	url := fmt.Sprintf("http://%s:%d/codecivil/article/_search?size=100&q=Text:%s", elasticHost, elasticPort, query)

	log.Printf("Searching for %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error searching for %s\n%s", url, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(results)
	json.NewEncoder(w).Encode(results)
}

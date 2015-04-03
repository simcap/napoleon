package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"text/template"
)

var homeTemplate = template.Must(template.ParseFiles("home.html"))
var searchURL = "http://elastic:9200/codecivil/article/_search"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		homeTemplate.Execute(w, nil)
	})
	http.HandleFunc("/search", search)
	http.ListenAndServe(":8080", nil)
}

type searchResults struct {
	Hits struct {
		Hits []hits
	}
}

type hits struct {
	Article struct {
		Section string
		Text    string
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

	elasticurl := buildSearchUrl(query)
	log.Printf("Searching for %s\n", elasticurl)

	resp, err := http.Get(elasticurl)
	if err != nil {
		log.Printf("error searching for %s\n%s", elasticurl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(results)
	json.NewEncoder(w).Encode(results)
}

func buildSearchUrl(query string) string {
	u, _ := url.Parse(searchURL)
	parameters := url.Values{}
	parameters.Add("size", "100")
	parameters.Add("q", "Text:"+query)
	u.RawQuery = parameters.Encode()
	return u.String()
}

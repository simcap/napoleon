package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var articlePaths []string
var articleRegex = regexp.MustCompile(`\AArticle.+md\z`)
var client = new(http.Client)

func main() {
	host := flag.String("h", "elastic", "host address of Elastic search to index content")
	port := flag.Int("p", 9200, "port of Elastic search to index content")
	rootdir := flag.String("r", "", "root directory name of the code civil content")

	flag.Parse()

	if *host == "" || *rootdir == "" {
		flag.Usage()
		os.Exit(1)
	}

	errwalk := filepath.Walk(*rootdir, collectArticles)
	if errwalk != nil {
		log.Fatalf("Error while walking the files at %s. Error: %s", *rootdir, errwalk)
	}

	type article struct {
		Section string
		Text    string
	}

	start := time.Now()
	for _, path := range articlePaths {
		log.Printf("Processing article at %s", path)

		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		articleSection, _ := filepath.Rel(*rootdir, path)
		articleID := fmt.Sprintf("%x", sha1.Sum([]byte(articleSection)))

		json, err := json.Marshal(article{Section: articleSection, Text: string(content)})
		if err != nil {
			log.Fatal(err)
		}

		req, _ := http.NewRequest("PUT", fmt.Sprintf("http://%s:%d/codecivil/article/%s", *host, *port, articleID), bytes.NewBuffer(json))
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode < 200 && resp.StatusCode > 299 {
			log.Fatalf("Bad response %#v", resp)
		}

	}

	log.Printf("Finished. Indexed %d articles\n. Elapsed %.1f seconds", len(articlePaths), time.Since(start).Seconds())
}

func collectArticles(path string, file os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !file.IsDir() && articleRegex.MatchString(file.Name()) {
		articlePaths = append(articlePaths, path)
	}

	return nil
}

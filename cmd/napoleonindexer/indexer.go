package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var articlePaths []string

func visit(path string, file os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	isArticle, _ := regexp.MatchString(`\AArticle.+md\z`, file.Name())
	if !file.IsDir() && isArticle {
		articlePaths = append(articlePaths, path)
	}

	return nil
}

func main() {
	host := flag.String("h", "elastic", "host address of Elastic search to index content")
	port := flag.Int("p", 9200, "port of Elastic search to index content")
	rootdir := flag.String("r", "", "root directory name of the code civil content")

	flag.Parse()

	if *host == "" || *rootdir == "" {
		flag.Usage()
		os.Exit(1)
	}

	errwalk := filepath.Walk(*rootdir, visit)
	if errwalk != nil {
		log.Fatalf("Error while walking the files at %s. Error: %s", *rootdir, errwalk)
	}

	type indexed struct {
		Path string
		Text string
	}

	client := new(http.Client)

	for _, path := range articlePaths {
		log.Printf("Processing article at %s", path)

		file, err := os.Open(path)
		if err != nil {
			log.Fatalf("Cannot open file at %s\n%s", path, err)
		}
		writer := bytes.NewBufferString("")
		_, errcopy := io.Copy(writer, file)
		if errcopy != nil {
			log.Fatalf("Cannot copy content of file %s\n%s", file.Name(), errcopy)
		}
		file.Close()

		relPath, _ := filepath.Rel(*rootdir, path)
		articleID := fmt.Sprintf("%x", sha1.Sum([]byte(relPath)))
		articleContent := writer.String()

		json, err := json.Marshal(indexed{Path: relPath, Text: articleContent})
		if err != nil {
			log.Fatalf("Cannot marshal into json %s", err)
		}

		req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:%d/codecivil/article/%s", *host, *port, articleID), bytes.NewBuffer(json))
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error in PUT %s", err)
		}

		if resp.StatusCode < 200 && resp.StatusCode > 299 {
			log.Fatalf("Bad response %#v", resp)
		}

	}

	log.Printf("Finished. Indexed %d articles\n", len(articlePaths))
}

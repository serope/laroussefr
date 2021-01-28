// Package scrapeutil provides some convenience functions to be used alongside
// package github.com/yhat/scrape.
package scrapeutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	
	"golang.org/x/net/html"
)

// HTMLRoot takes an HTML page, as either a URL or a disk filepath, and returns
// the root node of its parse tree with all newline text nodes removed for
// easier parsing.
func HTMLRoot(in string) (*html.Node, error) {
	if in == "" {
		return nil, fmt.Errorf("HTMLRoot(%s)\n%s", in, "Empty in")
	}
	data, err := getHTMLData(in)
	if err != nil {
		return nil, fmt.Errorf("HTMLRoot(%s)\n%s", in, err.Error())
	}
	doc, err := dataToDoc(data)
	if err != nil {
		return nil, fmt.Errorf("HTMLRoot(%s)\n%s", in, err.Error())
	}
	return doc, nil
}

// dataToDoc takes a web page's contents as a byte slice and returns the root
// node of its parse tree with all newline text nodes removed for easier
// parsing.
func dataToDoc(data []byte) (*html.Node, error) {
	data = cleanPageData(data)
	reader := bytes.NewReader(data)
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("dataToDoc()\n%s", err.Error())
	}
	return doc, nil
}

// getHTMLData takes an HTML page, as either a URL or a disk filepath, and
// returns the page's contents as a byte slice.
func getHTMLData(in string) ([]byte, error) {
	var readingFunc func(string)([]byte,error)
	if FileExists(in) {
		readingFunc = ioutil.ReadFile
	} else {
		readingFunc = getHTMLDataFromURL
	}
	data, err := readingFunc(in)
	if err != nil {
		return nil, fmt.Errorf("getHTMLData(%s)\nEither the file wasn't found, or: %s", in, err.Error())
	}
	return data, nil
}

// getHTMLDataFromURL takes a URL and returns the page's contents as a byte
// slice.
func getHTMLDataFromURL(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getHTMLDataFromURL(%s)\nhttp.Get\n%s", url, err.Error())
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("getHTMLDataFromURL(%s)\nioutil.ReadAll\n%s", url, err.Error())
	}
	return data, err
}

// cleanPageData takes a web page's contents as a byte slice and removes all
// newlines and tabs.
func cleanPageData(page []byte) []byte {
	removeThese := []string{"\n", "\t", "\r"}
	for _, r := range removeThese {
		page = bytes.ReplaceAll(page, []byte(r), []byte(""))
	}
	return page
}

// FileExists returns true if the specified file exists.
func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

// Package laroussefr provides packages for web scraping Larousse
// (https://www.larousse.fr).
// 
// laroussefr.go contains common functions shared by packages definition and
// traduction.
package laroussefr

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"github.com/yhat/scrape"
)

// ErrWordNotFound is returned by functions that search for words on Larousse
// and end up encountering a "word not found" page.
var ErrWordNotFound error

// LfrError implements the Error interface.
// 
// This is for internal use. Exported functions always return normal errors.
type LfrError struct {
	function string
	arg      string
	message  string 
}

func (lfre LfrError) Error() string {
	return fmt.Sprintf("%s(%s)\n%s", lfre.function, lfre.arg, lfre.message)
}

// NewError takes a function name, an example of an argument passed to it, and
// a short message describing an error that occurred, returning a new LfrError.
// 
// This is for internal use. Exported functions always return normal errors.
func NewError(function, arg, message string) LfrError {
	return LfrError{function, arg, message}
}

// GetPageID takes the root node of a page and returns its ID.
func GetPageID(doc *html.Node) (int, error) {
	n, ok := scrape.Find(doc, isPageIDnode)
	if !ok {
		return -1, NewError("GetPageID", "", "Failed to find ID node")
	}
	link := scrape.Attr(n, "href")
	i := strings.LastIndexByte(link, '/')
	if i == -1 {
		return -1, NewError("GetPageID", "", "Failed to extract ID from link " + link)
	}
	pageID, err := strconv.Atoi(link[i+1:])
	if err != nil {
		return -1, NewError("GetPageID", "", "strconv.Atoi says " + err.Error())
	}
	return pageID, nil
}

// GetPageIDsFromURLs takes a slice of URLs and calls GetPageIDFromURL on each.
func GetPageIDsFromURLs(urls []string) ([]int, error) {
	out := make([]int, len(urls))
	for i, s := range urls {
		pageID, err := GetPageIDFromURL(s)
		if err != nil {
			return nil, NewError("GetPageIDsFromURLs", "", err.Error())
		}
		out[i] = pageID
	}
	return out, nil
}

// GetPageIDFromURL takes a Larousse URL in the form of
// "larousse.fr/dictionnaires/abc/xyz/12345" and returns the page ID at
// the end.
func GetPageIDFromURL(url string) (int, error) {
	i := strings.LastIndexByte(url, '/')
	if i == -1 {
		return -1, NewError("GetPageIDsFromURL", "", "strings.LastIndexByte returned -1")
	}
	pageID, err := strconv.Atoi(url[i+1:])
	if err != nil {
		return -1, NewError("GetPageIDsFromURL", "", err.Error())
	}
	return pageID, nil
}

// GetSimilarWords takes the root node of a page and returns the list of URLs
// found in the word carousel near the bottom.
func GetSimilarWords(doc *html.Node) ([]string, error) {
	nodes := scrape.FindAll(doc, scrape.ByClass("item-word"))
	if len(nodes) <= 1 {
		return nil, nil
	}
	var out []string
	for _, n := range nodes[1:] {
		m := n.FirstChild
		href := scrape.Attr(m, "href")
		if href == "" {
			continue
		}
		str, err := url.PathUnescape(href)
		if err != nil {
			return nil, NewError("GetSimilarWords", "", err.Error())
		}
		out = append(out, "https://larousse.fr" + str)
	}
	return out, nil
}

// GetSearchSuggestions takes a "word not found" page and returns a list of
// search suggestions, if any are provided.
func GetSearchSuggestions(doc *html.Node) []string {
	var out []string
	if IsWordNotFoundPage(doc) && hasSuggestions(doc) {
		n, _ := scrape.Find(doc, scrape.ByClass("corrector"))
		liNodes := scrape.FindAll(n, scrape.ByTag(atom.Li))
		for _, li := range liNodes {
			a, _ := scrape.Find(li, scrape.ByTag(atom.A))
			str := scrape.Attr(a, "href")
			out = append(out, "https://larousse.fr" + str)
		}
	}
	return out
}

// IsWordNotFoundPage returns true if doc is the root of a "word not found"
// page.
func IsWordNotFoundPage(doc *html.Node) bool {
	_, ok := scrape.Find(doc, scrape.ByClass("corrector"))
	return ok
}

// IsURL verifies if str is a valid URL to a Larousse dictionary page. If it is,
// true and "" are returned. Otherwise, false and a message describing the
// problem are returned.
func IsURL(str string) (bool, string) {
	_, err := url.PathUnescape(str)
	if err != nil {
		return false, err.Error()
	}
	
	url, err := url.Parse(str)
	if err != nil {
		return false, err.Error()
	} else if !urlHasAllowedScheme(url) {
		return false, "Scheme must be http or https"
	} else if !strings.Contains(url.Hostname(), "larousse.fr") {
		return false, "Hostname must contain larousse.fr"
	}
	
	i := strings.Index(str, "larousse.fr") // reject if has "//" after protocol
	substr := str[i+11:]
	if strings.Contains(substr, "//") {
		return false, "Found \"//\""
	} else if !strings.Contains(str, "larousse.fr/dictionnaires/") {
		return false, "URL must contain \"larousse.fr/dictionnaires/\""
	}
	
	return true, ""
}

// urlHasAllowedScheme returns true if in has an "http" or "https" scheme.
func urlHasAllowedScheme(in *url.URL) bool {
	allowed := [2]string{"http", "https"}
	for _, a := range allowed {
		if in.Scheme == a {
			return true
		}
	}
	return false
}

// GetAudioURL takes an <audio> node containing a link to a TTS audio file
// and extracts the URL from it.
// 
// All URLs to larousse.fr/dictionnaires-prononciation/x/tts/... always redirect
// to voix.larousse.fr.
func GetAudioURL(n *html.Node) string {
	src := scrape.Attr(n, "src")
	if src == "" {
		return ""
	}
	
	str := src[29:] // after "/dictionnaires-prononciation/"
	i := strings.IndexByte(str, '/')
	j := strings.LastIndexByte(str, '/')
	
	lang := str[:i]
	filename := str[j+1:]
	url := fmt.Sprintf("https://voix.larousse.fr/%s/%s.mp3", lang, filename)
	return url
}

// hasSuggestions returns true if this "word not found" page has search
// suggestions.
// 
// An example of a page that returns true:
// https://www.larousse.fr/dictionnaires/francais-anglais/verytbvfsjd
// 
// An example of a page that returns false:
// https://www.larousse.fr/dictionnaires/francais-anglais/vert%202
func hasSuggestions(doc *html.Node) bool {
	_, ok := scrape.Find(doc, scrape.ByClass("err"))
	return !ok
}

// isPageIDnode returns true if n is a <link> node containing the page URL, from
// which the page ID can be exracted.
func isPageIDnode(n *html.Node) bool {
	return n.DataAtom == atom.Link && scrape.Attr(n, "rel") == "canonical"
}

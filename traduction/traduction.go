// Package traduction provides functions for web scraping Larousse's French
// and English bilingual dictionaries.
// 
// 
// Page Layout Example
// 
// Consider the French->English page for "court"
// (https://www.laroussefr.fr/dictionnaires/francais-anglais/court/19844).
//
// A page consists of a page ID (shown in its URL; 19844 in this case), a list
// of words, and a list of similar words in a scrolling carousel at the bottom.
// This page has 7 words, starting with "court (f courte)" and ending with "tout
// court". Its similar word carousel contains "courtage", "courtaud", etc.
// 
// A word consists of a header and a list of subheaders.
// 
// A header contains a word's details, such as text and pronunciation. The first
// word's header is "court (f courte) [kur, kurt] \ adjectif".
// 
// A subheader is black, boldfaced, all-caps text which contains a numbered list
// of items. The first word has 3 subheaders: "[DANS L'ESPACE]", "[DANS LE
// TEMPS]", and "[FAIBLE, INSUFFISANT]".
// 
// Within a subheader is a numbered list of items. The first subheader,
// "[DANS L'ESPACE]", has 3 items.
// 
// An item has two parts, meanings and phrases.
//
// A meaning is red text in the original language, often (but not always)
// followed by boldfaced blue text in the target language. The first meaning on
// this page is "[en longueur - cheveux, ongles] \ short". Red text is written
// in a combination of square brackets, parentheses, and all caps; see the
// Meaning type for their distinctions.
// 
// A phrase is an example phrase, which consists of black text in the original
// language followed by blue text in the target language. It may also have red
// text like a meaning does. The 2nd phrase on this page is "la jupe est trop
// courte de trois centimètres \ the skirt is three centimetres too short".
// 
// A phrase sometimes has a list of subphrases in an alphabet-bullet list. The
// first phrase on this page, "court sur pattes", has two subphrases.
// 
// A phrase may also be an "expression", which means it's displayed in a blue
// box with the word "EXPR" in the corner. The first expression on this page is
// "cycle court \ course of studies leading to qualifications exclusive of
// university entrance".
package traduction

import (
	"fmt"
	"strings"
	
	"scraper/laroussefr"
	"scraper/laroussefr/scrapeutil"
	"scraper/laroussefr/traduction/parse"
	
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

// ErrWordNotFound is returned by New or NewFromFileOrURL if the requested word
// isn't found.
var ErrWordNotFound error = laroussefr.ErrWordNotFound

// Type Language is an enum type.
// 
// Values: En, Fr
type Language int

func (lang Language) String() string {
	switch lang {
		case En: return "anglais"
		case Fr: return "francais"
	}
	return ""
}

// Available values for Language.
const (
	En = iota
	Fr
)


// Type Result represents a page from Larousse's French and English bilingual
// dictionaries. 
// 
// PageID is a unique identifier, which can be seen in the URL.
// 
// Words is a slice of words defined on the page.
// 
// SeeAlso is a slice of URLs of similar words found in the word carousel near
// the bottom of the page. If a Result ends up being a "word not found" page,
// then SeeAlso will contain search suggestions, if any are provided.
type Result struct {
	PageID  int
	Words   []Word
	SeeAlso []string
}

// equals compares r and q. If they're equal, an empty string and true are
// returned. Otherwise, a message describing the inequality and false are
// returned.
// 
// When comparing SeeAlso strings, only the page IDs in the URLs are compared,
// due to the way the copyright symbol '®' is displayed in some URLs, e.g. for
// the Airbag link in "aire"
// (https://www.larousse.fr/dictionnaires/francais-anglais/aire/1944):
// 
// http.Get -> https://larousse.fr/dictionnaires/francais-anglais/Airbag<sup>®</sup>/82998
// wget     -> https://larousse.fr/dictionnaires/francais-anglais/AirbagAirbag/82998
func (r Result) equals(q Result) (string, bool) {
	comparisonFuncs := []func(Result)(string,bool) {
		r.equalPageIDs,
		r.equalLens,
		r.equalWords,
		r.equalSeeAlsoIDs,
	}
	
	for _, comp := range comparisonFuncs {
		message, ok := comp(q)
		if !ok {
			return message, false
		}
	}
	
	return "", true
}

// equalPageIDs returns true if r and q have identical page IDs.
func (r Result) equalPageIDs(q Result) (string, bool) {
	if r.PageID != q.PageID {
		return fmt.Sprintf("PageID\nr: %d\nq: %d", r.PageID, q.PageID), false
	}
	return "", true
}

// equalLens returns true if r and q have the same length for every slice field.
func (r Result) equalLens(q Result) (string, bool) {
	if len(r.Words) != len(q.Words){
		return fmt.Sprintf("len(Words)\nr: %d\nq: %d", len(r.Words), len(q.Words)), false
	}
	if len(r.SeeAlso) != len(q.SeeAlso) {
		return fmt.Sprintf("len(SeeAlso)\nr: %d\nq: %d", len(r.SeeAlso), len(q.SeeAlso)), false
	}
	return "", true
}

// equalWords returns true if r and q have identical Words slices.
func (r Result) equalWords(q Result) (string, bool) {
	for i := range r.Words {
		word1 := r.Words[i]
		word2 := q.Words[i]
		message, ok := word1.equals(word2)
		if !ok {
			return fmt.Sprintf("Words[%d]: %s", i, message), false
		}
	}
	return "", true
}

// equalSeeAlsoIDs returns true if the page IDs at the end of each URL in both
// r's and q's SeeAlso slices are equivalent.
func (r Result) equalSeeAlsoIDs(q Result) (string, bool) {
	rSeeAlsoIDs, err := laroussefr.GetPageIDsFromURLs(r.SeeAlso)
	if err != nil {
		return err.Error(), false
	}
	qSeeAlsoIDs, err := laroussefr.GetPageIDsFromURLs(q.SeeAlso)
	if err != nil {
		return err.Error(), false
	}
	if len(rSeeAlsoIDs) != len(qSeeAlsoIDs) {
		return fmt.Sprintf("len(xSeeAlsoIDs)\nr: %d\nq: %d", len(rSeeAlsoIDs), len(qSeeAlsoIDs)), false
	}
	
	for i := range r.SeeAlso {
		id1 := rSeeAlsoIDs[i]
		id2 := qSeeAlsoIDs[i]
		if id1 != id2 {
			url1 := r.SeeAlso[i]
			url2 := q.SeeAlso[i]
			return fmt.Sprintf("SeeAlso[%d] different page ID at end of URL\nr: %s\nq: %s", i, url1, url2), false
		}
	}
	return "", true
}

// Type Word represents a word, which consists of a code, a header, and
// subheaders.
// 
// Code is an integer assigned to some words, but is not a unique identifier.
// The first word on a page will always have a code which is equivalent to the
// page's ID, but subsequent words may have the same or different codes.
// Larousse tends to be inconsistent in this regard.
type Word struct {
	Code       int
	Header     Header
	Subheaders []Subheader
}

// equals compares w and u. If they're equal, an empty string and true are
// returned. Otherwise, a message describing the inequality and false are
// returned.
func (w Word) equals(u Word) (string, bool) {
	if w.Code != u.Code {
		return fmt.Sprintf("Code\nw: %d\nu: %d", w.Code, u.Code), false
	}
	if len(w.Subheaders) != len(u.Subheaders) {
		return fmt.Sprintf("len(Subheaders)\nw: %d\nu: %d", len(w.Subheaders), len(u.Subheaders)), false
	}
	message, ok := w.Header.equals(u.Header)
	if !ok {
		return fmt.Sprintf("Header: %s", message), false
	}
	message, ok = w.equalSubheaders(u)
	if !ok {
		return message, false
	}
	return "", true
}

// equalSubheaders returns true if w and u have identical Subheaders.
func (w Word) equalSubheaders(u Word) (string, bool) {
	for i := range w.Subheaders {
		sub1 := w.Subheaders[i]
		sub2 := u.Subheaders[i]
		message, ok := sub1.equals(sub2)
		if !ok {
			return fmt.Sprintf("Subheaders[%d]: %s", i, message), false
		}
	}
	return "", true
}

// Type Header represents the header block of a word where its information is
// displayed.
// 
// Text is the word's string.
//
// TextAlt is the word's alternate string, if any, shown in parentheses. For
// French (and other Romance languages supported by Larousse), this is typically
// the feminine form of a masculine word or vice-versa.
// 
// Phonetic is the IPA pronunciation text shown in small square brackets.
//
// Audio is the URL of the audio clip, if available.
// 
// Type is the word's grammatical type.
type Header struct {
	Text     string
	TextAlt  string
	Phonetic string
	Audio    string
	Type     string
}

// equals compares h and i. If they're equal, an empty string and true are
// returned. Otherwise, a message describing the inequality and false are
// returned.
func (h Header) equals(i Header) (string, bool) {
	switch {
		case h.Text != i.Text:
			return fmt.Sprintf("Text\nh: \"%s\"\ni: \"%s\"", h.Text, i.Text), false
		case h.TextAlt != i.TextAlt:
			return fmt.Sprintf("TextAlt\nh: \"%s\"\ni: \"%s\"", h.TextAlt, i.TextAlt), false
		case h.Phonetic != i.Phonetic:
			return fmt.Sprintf("Phonetic\nh: \"%s\"\ni: \"%s\"", h.Phonetic, i.Phonetic), false
		case h.Audio != i.Audio:
			return fmt.Sprintf("Audio\nh: \"%s\"\ni: \"%s\"", h.Audio, i.Audio), false
		case h.Type != i.Type:
			return fmt.Sprintf("Type\nh: \"%s\"\ni: \"%s\"", h.Type, i.Type), false
	}
	return "", true
}

// Type Subheader represents a subheader. Most words in the French-English
// dictionary have a single Subheader with an empty Title.
type Subheader struct {
	Title string
	Items []Item
}

// equals compares s and t. If they're equal, an empty string and true are
// returned. Otherwise, a message describing the inequality and false are
// returned.
func (s Subheader) equals(t Subheader) (string, bool) {
	if s.Title != t.Title {
		return fmt.Sprintf("Title\ns: %s\nt: %s", s.Title, t.Title), false
	}
	if len(s.Items) != len(t.Items) {
		return fmt.Sprintf("len(Items)\ns: %d\nt: %d", len(s.Items), len(t.Items)), false
	}
	message, ok := s.equalItems(t)
	if !ok {
		return message, false
	}
	return "", true
}

// equalItems returns true if s and t have identical Items slices.
func (s Subheader) equalItems(t Subheader) (string, bool) {
	for i := range s.Items {
		item1 := s.Items[i]
		item2 := t.Items[i]
		message, ok := item1.equals(item2)
		if !ok {
			return fmt.Sprintf("Items[%d]: %s", i, message), false
		}
	}
	return "", true
}

// Type Item represents an item within a subheader.
type Item struct {
	Meanings []Meaning
	Phrases  []Phrase
}

// equals compares i and t. If they're equal, an empty string and true are
// returned. Otherwise, a message describing the inequality and false are
// returned.
func (i Item) equals(t Item) (string, bool) {
	message, ok := i.equalLens(t)
	if !ok {
		return message, false
	}
	message, ok = i.equalMeanings(t)
	if !ok {
		return message, false
	}
	message, ok = i.equalPhrases(t)
	if !ok {
		return message, false
	}
	return "", true
}

// equalLens returns true if the slice fields of i and t have equivalent
// lengths.
func (i Item) equalLens(t Item) (string, bool) {
	if len(i.Meanings) != len(t.Meanings) {
		return fmt.Sprintf("len(Meanings)\ni: %d\nt: %d", len(i.Meanings), len(t.Meanings)), false
	}
	if len(i.Phrases) != len(t.Phrases) {
		return fmt.Sprintf("len(Phrases)\ni: %d\nt: %d", len(i.Phrases), len(t.Phrases)), false
	}
	return "", true
}

// equalMeanings returns true if i and t have identical Meanings slices.
func (i Item) equalMeanings(t Item) (string, bool) {
	for j := range i.Meanings {
		meaning1 := i.Meanings[j]
		meaning2 := t.Meanings[j]
		message, ok := meaning1.equals(meaning2)
		if !ok {
			return fmt.Sprintf("\n%v\n%v\nMeanings[%d]: %s", meaning1, meaning2, j, message), false
		}
	}
	return "", true
}

// equalPhrases returns true if i and t have equivalent Phrases slices.
func (i Item) equalPhrases(t Item) (string, bool) {
	for j := range i.Phrases {
		phrase1 := i.Phrases[j]
		phrase2 := t.Phrases[j]
		message, ok := phrase1.equals(phrase2)
		if !ok {
			return fmt.Sprintf("Phrases[%d]: %s", j, message), false
		}
	}
	return "", true
}

// Type Meaning represents a translation of a word.
// 
// Text is the meaning's text in the target language.
// 
// RedBrac is the meanings's context, displayed in red square brackets.
// 
// RedCaps is the meanings's "domain" context, displayed in red all caps. This
// is usually a context that's more specific than RedBrac.
// 
// RedMeta is the meaning's "meta" context, displayed in red parentheses. This
// is usually used to indicate whether a term is formal or informal, or if it's
// from a region-specific dialect.
type Meaning struct {
	Text    string // Traduction
	RedBrac string // Indicateur
	RedCaps string // IndicateurDomaine
	RedMeta string // Metalangue
}

// equals compares m and n. If they're equal, an empty string and true are
// returned. Otherwise, a message describing the inequality and false are
// returned.
func (m Meaning) equals(n Meaning) (string, bool) {
	switch {
		case m.Text != n.Text:
			return fmt.Sprintf("Text\nm: \"%s\"\nn: \"%s\"", m.Text, n.Text), false
		case m.RedBrac != n.RedBrac:
			return fmt.Sprintf("RedBrac\nm: \"%s\"\nn: \"%s\"", m.RedBrac, n.RedBrac), false
		case m.RedCaps != n.RedCaps:
			return fmt.Sprintf("RedCaps\nm: \"%s\"\nn: \"%s\"", m.RedCaps, n.RedCaps), false
		case m.RedMeta != n.RedMeta:
			return fmt.Sprintf("RedMeta\nm: \"%s\"\nn: \"%s\"", m.RedMeta, n.RedMeta),  false
	}
	return "", true
}

// isEmpty returns true if m consists entirely of empty strings.
func (m Meaning) isEmpty() bool {
	return m.Text=="" && m.RedBrac=="" && m.RedCaps=="" && m.RedMeta==""
}

// update takes a node containing a Meaning property and applies it to m.
func (m *Meaning) update(n *html.Node) {
	class := scrape.Attr(n, "class")
	switch class {
		case "Renvois":           m.Text = scrape.Text(n) // for "coup de fil" on fr->en coup
		case "Glose2":            m.Text = scrape.Text(n) // for en->fr "blue" POLITICS
		case "Traduction":        m.updateFromTraductionNode(n)
		case "Indicateur":        m.RedBrac = scrape.Text(n)
		case "IndicateurDomaine": m.RedCaps = strings.ToUpper(scrape.Text(n))
		case "Metalangue":        m.RedMeta = scrape.Text(n)
	}
}

// updateFromTraductionNode takes a "Traduction" node and applies it to m.
func (m *Meaning) updateFromTraductionNode(n *html.Node) {
		if m.Text != "" {
			m.Text += " "
		}
		m.Text += parse.Traduction(n)
}

// Type Phrase represents an example phrase.
// 
// Text1 and Text2 are the phrase's text in the original and target languages,
// respectively.
// 
// Audio1 and Audio2 are the URLs of the TTS audio clips corresponding to
// Text1 and Text2.
// 
// RedBrac is the phrase's context, displayed in red square brackets.
// 
// RedCaps is the phrase's "domain" context, displayed in red all caps. This is
// usually a context that's more specific than RedBrac.
// 
// RedMeta is the term's "meta" context, displayed in red parentheses. This is
// usually used to indicate whether a term is formal or informal, or if it's
// from a region-specific dialect.
// 
// IsBlue is true if the phrase is an expression. An expression is merely a
// phrase shown in a blue box with "EXPR" in the corner. If an expression has
// subphrases, their IsBlue values are true as well.
// 
// Subphrases is a slice of subphrases, which appear in an alphabet-bullet list.
// Each subphrase's Subphrases slice is nil.
type Phrase struct {
	Text1      string   // Locution2
	Text2      string   // Traduction2, Metalangue2
	Audio1     string   // lienson3
	Audio2     string   // lienson2
	RedBrac    string   // Indicateur
	RedCaps    string   // IndicateurDomaine
	RedMeta    string   // Metalangue
	IsBlue     bool     // true if inside BlocExpression
	Subphrases []Phrase // DivisionExpression
}

// equals compares p and q. If they're equal, an empty string and true are
// returned. Otherwise, a message describing the inequality and false are
// returned.
func (p Phrase) equals(q Phrase) (string, bool) {
	message, ok := p.equalStringFields(q)
	if !ok {
		return message, false
	}
	if len(p.Subphrases) != len(q.Subphrases) {
		return fmt.Sprintf("len(Subphrases)\np: %d\nq: %d", len(p.Subphrases), len(q.Subphrases)), false
	}
	message, ok = p.equalSubphrases(q)
	if !ok {
		return message, false
	}
	return "", true
}

// equalStringFields returns true if the string fields of p and q are identical.
func (p Phrase) equalStringFields(q Phrase) (string, bool) {
	switch {
		case p.Text1 != q.Text1:     return fmt.Sprintf("Text1\np: \"%s\"\nq: \"%s\"", p.Text1, q.Text1), false
		case p.Text2 != q.Text2:     return fmt.Sprintf("Text2\np: \"%s\"\nq: \"%s\"", p.Text2, q.Text2), false
		case p.Audio1 != q.Audio1:   return fmt.Sprintf("Audio1\np: \"%s\"\nq: \"%s\"", p.Audio1, q.Audio1), false
		case p.Audio2 != q.Audio2:   return fmt.Sprintf("Audio2\np: \"%s\"\nq: \"%s\"", p.Audio2, q.Audio2), false
		case p.RedBrac != q.RedBrac: return fmt.Sprintf("Text1\np: \"%s\"\nq: \"%s\"", p.RedBrac, q.RedBrac), false
		case p.RedCaps != q.RedCaps: return fmt.Sprintf("RedCaps\np: \"%s\"\nq: \"%s\"", p.RedCaps, q.RedCaps), false
		case p.RedMeta != q.RedMeta: return fmt.Sprintf("RedMeta\np: \"%s\"\nq: \"%s\"", p.RedMeta, q.RedMeta), false
		case p.IsBlue != q.IsBlue:   return fmt.Sprintf("IsBlue\np: %v\nq: %v", p.IsBlue, q.IsBlue), false
	}
	return "", true
}

// equalSubphrases returns true if p's and q's Subphrases slices are identical.
func (p Phrase) equalSubphrases(q Phrase) (string, bool) {
	for i := range p.Subphrases {
		sub1 := p.Subphrases[i]
		sub2 := q.Subphrases[i]
		message, ok := sub1.equals(sub2)
		if !ok {
			return fmt.Sprintf("Subphrases[%d]: %s", i, message), false
		}
	}
	return "", true
}

// update takes a node containing a Phrase property and applies it to p.
func (p *Phrase) update(n *html.Node) {
	class := scrape.Attr(n, "class")
	switch class {
		case "Locution2":
			p.Text1   = scrape.Text(n)
			audio1, ok := handleLocution2InnerLienson3(n)
			if ok {
				p.Audio1 = audio1
			}
		case "Glose2":            fallthrough
		case "Traduction2":       fallthrough
		case "Renvois":           fallthrough
		case "Metalangue2":
			if p.Text2 != "" {
				p.Text2 += " "
			}
			p.Text2 += parse.Traduction(n)
		case "lienson3":          p.Audio1  = parse.Lienson(n)
		case "lienson2":          p.Audio2  = parse.Lienson(n)
		case "Indicateur":        p.RedBrac = scrape.Text(n)
		case "IndicateurDomaine": p.RedCaps = strings.ToUpper(scrape.Text(n))
		case "Metalangue":        p.RedMeta = scrape.Text(n)
	}
}

// setBlue sets the IsBlue value for p and all of its Subphrases.
func (p *Phrase) setBlue(blue bool) {
	p.IsBlue = blue
	for i := range p.Subphrases {
		p.Subphrases[i].IsBlue = blue
	}
}

// handleLocution2nInnerLienson3 is used to find "lienson3" nodes buried within
// "Locution2" nodes. If a "lienson3" node exists, the appropriate Audio1 string
// and true are returned. This is necessary for some Phrase edge cases, such as:
// 
// 1. "vert de rage"
//    (https://www.larousse.fr/dictionnaires/francais-anglais/vert/80698)
// 2. "il y avait un brouillard à couper au couteau"
//    (https://www.larousse.fr/dictionnaires/francais-anglais/couper/19720)
// 3. "il a essayé de me faire le coup de la panne"
//    (https://www.larousse.fr/dictionnaires/francais-anglais/coup/19682)
func handleLocution2InnerLienson3(locution2Node *html.Node) (string, bool) {
	lienson3, ok := scrape.Find(locution2Node, scrape.ByClass("lienson3"))
	if ok {
		return parse.Lienson(lienson3), true
	}
	return "", false
}




// New takes a word, its language, and a target language and searches for its
// translation on Larousse.
// 
// If the word doesn't exist, an error ErrWordNotFound is returned. If Larousse
// provides search suggestions for this nonexistent word, they will be put into
// the returned Result's SeeAlso slice.
func New(word string, from, to Language) (Result, error) {
	err := checkNewArgs(word, from, to)
	if err != nil {
		return Result{}, laroussefr.NewError("New", word, err.Error())
	}
	if strings.ContainsRune(word, ' ') {
		word = strings.ReplaceAll(word, " ", "-")
	}
	url := fmt.Sprintf("https://www.larousse.fr/dictionnaires/%s-%s/%s", from, to, word)
	return NewFromFileOrURL(url)
}

// checkNewArgs checks the arguments passed to New, returning a non-nil error if
// they're invalid.
func checkNewArgs(word string, from, to Language) error {
	switch {
		case word == "":          return laroussefr.NewError("checkNewArgs", word, "Empty string")
		case from.String() == "": return laroussefr.NewError("checkNewArgs", word, "Unknown 'from' language")
		case to.String() == "":   return laroussefr.NewError("checkNewArgs", word, "Unknown 'to' language")
		case from == to:          return laroussefr.NewError("checkNewArgs", word, "Same 'from' and 'to' language: " + from.String())
	}
	return nil
}

// NewFromFileOrURL scrapes an English-French or French-English page given as
// either an HTML filepath or a URL.
// 
// If the result is a "word not found" page, an error ErrWordNotFound is
// returned. If the page provides search suggestions, they will be put into the
// returned Result's SeeAlso slice.
func NewFromFileOrURL(in string) (Result, error) {
	if !scrapeutil.FileExists(in) {
		ok, message := isURL(in)
		if !ok {
			return Result{}, laroussefr.NewError("NewFromFileOrURL", in, "Bad URL: " + message)
		}
	}
	
	doc, err := scrapeutil.HTMLRoot(in)
	if err != nil {
		return Result{}, laroussefr.NewError("NewFromFileOrURL", in, "Download step: " + err.Error())
	}
	
	if laroussefr.IsWordNotFoundPage(doc) {
		ErrWordNotFound = laroussefr.NewError("NewFromFileOrURL", in, "ErrWordNotFound")
		seeAlso := laroussefr.GetSearchSuggestions(doc)
		result := Result{-1, nil, seeAlso}
		return result, ErrWordNotFound
	}
	
	result, err := newResultFromRoot(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("NewFromFileOrURL", in, "Scrape step: " + err.Error())
	}
	return result, err
}

// isURL verifies if str is a valid URL to a French-English or English-French
// translation page on Larousse. If it is, then true and "" are returned.
// Otherwise, false and a message describing the problem are returned.
func isURL(str string) (bool, string) {
	ok, message := laroussefr.IsURL(str)
	if !ok {
		return false, message
	}
	
	sl := [2]string{
		"larousse.fr/dictionnaires/francais-anglais/",
		"larousse.fr/dictionnaires/anglais-francais/",
	}
	for _, s := range sl {
		if strings.Contains(str, s) && !strings.HasSuffix(str, s) {
			return true, ""
		}
	}
	return false, fmt.Sprintf("Must contain \"%s\" or \"%s\"", sl[0], sl[1])
}

// newResultFromRoot returns a new Result from an HTML root.
func newResultFromRoot(doc *html.Node) (Result, error) {
	pageID, err := laroussefr.GetPageID(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	words, err := scrapeWords(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	seeAlso, err := laroussefr.GetSimilarWords(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	result := Result{pageID, words, seeAlso}
	return result, nil
}

// scrapeWords takes a page root and scrapes all of its bigWords and smallWords
// into a Word slice.
func scrapeWords(doc *html.Node) ([]Word, error) {
	bigWords, err := scrapeBigWords(doc)
	if err != nil {
		return nil, laroussefr.NewError("scrapeWords", "", "bigWords step: " + err.Error())
	}
	
	smallWords, err := scrapeSmallWords(doc)
	if err != nil {
		return nil, laroussefr.NewError("scrapeWords", "", "smallWords step: " + err.Error())
	}
	
	var words []Word
	for _, bw := range bigWords {
		words = append(words, Word(bw))
	}
	for _, sw := range smallWords {
		words = append(words, sw.toWord())
	}
	return words, nil
}

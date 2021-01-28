// Package definition provides functions for scraping the Larousse's French
// dictionary.
//
// 
// Page Layout Example
// 
// Consider the page for "vert"
// (https://www.laroussefr.fr/dictionnaires/francais/vert).
// 
// Header section
// 
// This section shows:
//   1. the word, both male and female forms if applicable ("vert, verte")
//   2. its type ("adjectif")
//   3. an optional footnote, usually historical trivia ("latin viridis")
//   4. an audio link
// 
// DÉFINITIONS
// 
// A list of definitions.
// A definition may have a sample phrase following it, separated by a French
// semicolon (" : ").
// It may also have a context, written in red font above it. Note that one
// context may have multiple definitions in a numbered list.
//
// EXPRESSIONS
// 
// A list of common expressions; each expression is in blue font, followed by
// an explanation in normal font, separated by a comma.
// Like definitions, an expression can have a context with multiple items.
//
// SYNONYMES ET CONTRAIRES
// 
// A list of synonyms and/or antonyms -for a definition-. The definition text,
// as shown in this section, is usually a substring of its corresponding
// DÉFINITIONS text.
// The synonyms and antonyms may optionally be hyperlinked to their own pages.
// Very rarely, a word will have some synonyms and/or antonyms, but no
// definition (e.g. aguiche). I ignore these ones.
// 
// HOMONYMES
// 
// A list of homonyms and/or variants.)
// 
// DIFFICULTÉS
// 
// Describes irregularities and common mistakes.
package definition

import (
	"fmt"
	"strings"
	
	"scraper/laroussefr"
	"scraper/laroussefr/scrapeutil"
	"scraper/laroussefr/definition/match"
	"scraper/laroussefr/definition/parse"
	
	"golang.org/x/net/html"
	"github.com/yhat/scrape"
)

// ErrWordNotFound is returned by New or NewFromFileOrURL if the requested word
// isn't found.
var ErrWordNotFound error = laroussefr.ErrWordNotFound

// Type Result represents a page from Larousse's French dictionary.
type Result struct {
	PageID      int
	Header      Header
	Definitions []Definition
	Expressions []Expression
	Relations   []Relation // synonymes et contraires
	Homonymes   []Homonyme
	Difficultes []Difficulte
	Citations   []Citation
	SeeAlso     []string
}

// equals compares r and q. If they're equal, an empty string and true are
// returned. Otherwise, a message describing the inequality and false are
// returned.
// 
// When comparing SeeAlso strings, only the page IDs in the URLs are compared.
func (r Result) equals(q Result) (string, bool) {
	comparisonFuncs := []func(Result)(string,bool) {
		r.equalPageIDs,
		r.equalHeaders,
		r.equalLens,
		r.equalDefinitions,
		r.equalExpressions,
		r.equalRelations,
		r.equalHomonymes,
		r.equalDifficultes,
		r.equalCitations,
		r.equalSeeAlsoIDs,
	}
	
	for _, comp := range comparisonFuncs {
		message, ok := comp(q)
		if !ok {
			return message, false
		}
	}
	
	/*
	msg, ok := r.equalLens(q)
	if !ok {
		return msg, false
	}
	
	msg, ok = r.equalDefinitions(q)
	if !ok {
		return msg, false
	}
	
	msg, ok = r.equalExpressions(q)
	if !ok {
		return msg, false
	}
	
	msg, ok = r.equalRelations(q)
	if !ok {
		return msg, false
	}
	
	msg, ok = r.equalHomonymes(q)
	if !ok {
		return msg, false
	}
	
	msg, ok = r.equalDifficultes(q)
	if !ok {
		return msg, false
	}
	
	msg, ok = r.equalCitations(q)
	if !ok {
		return msg, false
	}
	*/
	
	return "", true
}

// equalPageIDs returns true if p and q have the same page ID.
func (r Result) equalPageIDs(q Result) (string, bool) {
	if r.PageID != q.PageID {
		return fmt.Sprintf("PageID\nr: %d\nq: %d", r.PageID, q.PageID), false
	}
	return "", true
}

// equalHeaders returns true if p and q have identical Headers.
func (r Result) equalHeaders(q Result) (string, bool) {
	message, ok := r.Header.equals(q.Header)
	if !ok {
		return fmt.Sprintf("Header: %s", message), false
	}
	return "", true
}

// equalLens returns true if p and q have the same length for every slice field.
func (r Result) equalLens(q Result) (string, bool) {
	switch {
	case len(r.Definitions) != len(q.Definitions): return fmt.Sprintf("len(Definitions)\nr: %d\nq: %d", len(r.Definitions), len(q.Definitions)), false
	case len(r.Expressions) != len(q.Expressions): return fmt.Sprintf("len(Expressions)\nr: %d\nq: %d", len(r.Expressions), len(q.Expressions)), false
	case len(r.Relations) != len(q.Relations):     return fmt.Sprintf("len(Relations)\nr: %d\nq: %d", len(r.Relations), len(q.Relations)), false
	case len(r.Homonymes) != len(q.Homonymes):     return fmt.Sprintf("len(Homonymes)\nr: %d\nq: %d", len(r.Homonymes), len(q.Homonymes)), false
	case len(r.Difficultes) != len(q.Difficultes): return fmt.Sprintf("len(Difficultes)\nr: %d\nq: %d", len(r.Difficultes), len(q.Difficultes)), false
	case len(r.Citations) != len(q.Citations):     return fmt.Sprintf("len(Citations)\nr: %d\nq: %d", len(r.Citations), len(q.Citations)), false
	case len(r.SeeAlso) != len(q.SeeAlso):         return fmt.Sprintf("len(SeeAlso)\nr: %d\nq: %d", len(r.SeeAlso), len(q.SeeAlso)), false
	}
	return "", true
}

// equalDefinitions returns true p and q have identical Definitions slices.
func (r Result) equalDefinitions(q Result) (string, bool) {
	for i := range r.Definitions {
		def1 := r.Definitions[i]
		def2 := q.Definitions[i]
		message, ok := def1.equals(def2)
		if !ok {
			return fmt.Sprintf("Definitions[%d]: %s", i, message), false
		}
	}
	return "", true
}

// equalExpressions returns true if p and q have identical Expressions slices.
func (r Result) equalExpressions(q Result) (string, bool) {
	for i := range r.Expressions {
		exp1 := r.Expressions[i]
		exp2 := q.Expressions[i]
		message, ok := exp1.equals(exp2)
		if !ok {
			return fmt.Sprintf("Expressions[%d]: %s", i, message), false
		}
	}
	return "", true
}

// equalRelations returns true if p and q have identical Relations slices.
func (r Result) equalRelations(q Result) (string, bool) {
	for i := range r.Relations {
		rel1 := r.Relations[i]
		rel2 := q.Relations[i]
		message, ok := rel1.equals(rel2)
		if !ok {
			return fmt.Sprintf("Relations[%d]: %s", i, message), false
		}
	}
	return "", true
}

// equalHomonymes returns true if p and q have identical Homonymes slices.
func (r Result) equalHomonymes(q Result) (string, bool) {
	for i := range r.Homonymes {
		hom1 := r.Homonymes[i]
		hom2 := q.Homonymes[i]
		message, ok := hom1.equals(hom2)
		if !ok {
			return fmt.Sprintf("Homonymes[%d]: %s", i, message), false
		}
	}
	return "", true
}

// equalDifficultes returns true if p and q have identical Difficultes slices.
func (r Result) equalDifficultes(q Result) (string, bool) {
	for i := range r.Difficultes {
		dif1 := r.Difficultes[i]
		dif2 := q.Difficultes[i]
		message, ok := dif1.equals(dif2)
		if !ok {
			return fmt.Sprintf("Difficultes[%d]: %s", i, message), false
		}
	}
	return "", true
}

// equalCitations returns true if p and q have identical Citations slices.
func (r Result) equalCitations(q Result) (string, bool) {
	for i := range r.Citations {
		cit1 := r.Citations[i]
		cit2 := q.Citations[i]
		message, ok := cit1.equals(cit2)
		if !ok {
			return fmt.Sprintf("Citations[%d]: %s", i, message), false
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

// Type Header represents the header area of a page.
type Header struct {
	Texte  string
	Audio  string
	Type   string
}

// equals returns true if h and i are identical.
func (h Header) equals(i Header) (string, bool) {
	switch {
	case h.Texte != i.Texte: return fmt.Sprintf("Texte: h:%s\ni:%s", h.Texte, i.Texte), false
	case h.Audio != i.Audio: return fmt.Sprintf("Audio: h:%s\ni:%s", h.Audio, i.Audio), false
	case h.Type != i.Type:   return fmt.Sprintf("Type: h:%s\ni:%s", h.Type, i.Type), false
	}
	return "", true
}

// Type Relation represents an item from a page's SYNONYMES ET CONTRAIRES
// section.
// 
// Texte is often, but not always, equivalent to the Texte of an item from
// DÉFINITIONS or EXPRESSIONS.
type Relation struct {
	Texte      string
	Synonymes  []string
	Contraires []string
}

// equals returns true if r and q are identical.
func (r Relation) equals(q Relation) (string, bool) {
	if r.Texte != q.Texte {
		return fmt.Sprintf("Texte: r:%s\nq:%s", r.Texte, q.Texte), false
	}
	
	if len(r.Synonymes) != len(q.Synonymes) {
		return fmt.Sprintf("len(Synonymes)\nr: %d\nq: %d", len(r.Synonymes), len(q.Synonymes)), false
	}
	
	if len(r.Contraires) != len(q.Contraires) {
		return fmt.Sprintf("len(Contraires)\nr: %d\nq: %d", len(r.Contraires), len(q.Contraires)), false
	}
	
	message, ok := r.equalSynonymes(q)
	if !ok {
		return message, false
	}
	
	message, ok = r.equalContraires(q)
	if !ok {
		return message, false
	}
	
	return "", true
}

// equalSynonymes returns true if r and q have identical Synonymes slices.
func (r Relation) equalSynonymes(q Relation) (string, bool) {
	for i := range r.Synonymes {
		syn1 := r.Synonymes[i]
		syn2 := q.Synonymes[i]
		if syn1 != syn2 {
			return fmt.Sprintf("Synonymes[%d] \n r:%s \n q:%s", i, syn1, syn2), false
		}
	}
	return "", true
}

// equalContraires returns true if r and q have identical Contraires slices.
func (r Relation) equalContraires(q Relation) (string, bool) {
	for i := range r.Contraires {
		con1 := r.Contraires[i]
		con2 := q.Contraires[i]
		if con1 != con2 {
			return fmt.Sprintf("Contraires[%d] \n r:%s \n q:%s", i, con1, con2), false
		}
	}
	return "", true
}

func (r Relation) hasSynonymes() bool {
	return len(r.Synonymes) > 0
}

func (r Relation) hasContraires() bool {
	return len(r.Contraires) > 0
}

// Type Definition represents an item from a page's DÉFINITIONS section.
// 
// Texte is the definition text, typically with the meaning in black font and
// one or more example phrases in blue font, separated by a French semicolon
// (" : ").
// 
// RedBig is the definition's context written in large, red, boldfaced text
// above the definition text.
// 
// RedSmall is more specific context written in red text preceeding the
// definition text.
type Definition struct {
	Texte    string
	RedBig   string
	RedSmall string
}

// equals returns true if d and e are identical.
func (d Definition) equals(e Definition) (string, bool) {
	switch {
	case d.Texte != e.Texte:       return fmt.Sprintf("Texte: d:%s\ne:%s", d.Texte, e.Texte), false
	case d.RedBig != e.RedBig:     return fmt.Sprintf("RedBig: d:%s\ne:%s", d.RedBig, e.RedBig), false
	case d.RedSmall != e.RedSmall: return fmt.Sprintf("RedSmall: d:%s\ne:%s", d.RedSmall, e.RedSmall), false
	}
	return "", true
}

// Type Expression represents an item from a page's EXPRESSIONS section.
// 
// Textes is the expression text.
// 
// RedBig is the definition's context written in large, red, boldfaced text
// above the definition text.
// 
// RedSmall is more specific context written in red text preceeding the
// definition text.
type Expression struct {
	Texte    string
	RedBig   string
	RedSmall string
}

// equals returns true if e and f are identical.
func (e Expression) equals(f Expression) (string, bool) {
	switch {
	case e.Texte != f.Texte:       return fmt.Sprintf("Texte: e:%s\nf:%s", e.Texte, f.Texte), false
	case e.RedBig != f.RedBig:     return fmt.Sprintf("RedBig: e:%s\nf:%s", e.RedBig, f.RedBig), false
	case e.RedSmall != f.RedSmall: return fmt.Sprintf("RedSmall: e:%s\nf:%s", e.RedSmall, f.RedSmall), false
	}
	return "", true
}

// Type Homonyme represents an item from a page's HOMONYMES section.
type Homonyme struct {
	Texte string
	Type  string
}

// equals returns true if h and i are identical.
func (h Homonyme) equals(i Homonyme) (string, bool) {
	switch {
	case h.Texte != i.Texte: return fmt.Sprintf("Texte: h:%s\ni:%s", h.Texte, i.Texte), false
	case h.Type != i.Type:   return fmt.Sprintf("Type: h:%s\ni:%s", h.Type, i.Type), false
	}
	return "", true
}

// Type Difficulte represents an item from a page's DIFFICULTÉS section.
type Difficulte struct {
	Type  string
	Texte string
}

// equals returns true if d and e are identical.
func (d Difficulte) equals(e Difficulte) (string, bool) {
	switch {
	case d.Type != e.Type:         return fmt.Sprintf("Type: d:%s\ne:%s", d.Type, e.Type), false
	case d.Texte != e.Texte:       return fmt.Sprintf("Texte: d:%s\ne:%s", d.Texte, e.Texte), false
	}
	return "", true
}

// Type Citation represents an item from a page's CITATIONS section.
type Citation struct {
	ID         int
	Auteur     string
	InfoAuteur string
	Texte      string
	Info       string
}

// equals returns true if c and d are identical.
func (c Citation) equals(d Citation) (string, bool) {
	switch {
	case c.ID != d.ID:                 return fmt.Sprintf("ID: c:%d\nd:%d", c.ID, d.ID), false
	case c.Auteur != d.Auteur:         return fmt.Sprintf("Auteur: c:%s\nd:%s", c.Auteur, d.Auteur), false
	case c.InfoAuteur != d.InfoAuteur: return fmt.Sprintf("InfoAuteur: c:%s\nd:%s", c.InfoAuteur, d.InfoAuteur), false
	case c.Texte != d.Texte:           return fmt.Sprintf("Texte: c:%s\nd:%s", c.Texte, d.Texte), false
	case c.Info != d.Info:             return fmt.Sprintf("Info: c:%s\nd:%s", c.Info, d.Info), false
	}
	return "", true
}


// New takes a French word and searches for its definition on Larousse.
// 
// If the word doesn't exist, an error ErrWordNotFound is returned. If Larousse
// provides search suggestions for this nonexistent word, they will be put into
// the returned Result's SeeAlso slice.
func New(word string) (Result, error) {
	if word == "" {
		return Result{}, laroussefr.NewError("New", word, "Empty string")
	}
	if strings.ContainsRune(word, ' ') {
		word = strings.ReplaceAll(word, " ", "-")
	}
	url := "https://www.larousse.fr/dictionnaires/francais/" + word
	return NewFromFileOrURL(url)
}

// NewFromFileOrURL scrapes a French definition page given as either an HTML
// filepath or a URL.
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
		var res Result
		res.SeeAlso = laroussefr.GetSearchSuggestions(doc)
		return res, ErrWordNotFound
	}
	
	res, err := newResultFromRoot(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("NewFromFileOrURL", in, "Scrape step: " + err.Error())
	}
	return res, err
}

// isURL verifies if str is a valid URL to a French dictionary page on Larousse.
// If it is, then true and "" are returned. Otherwise, false and a message
// describing the problem are returned.
func isURL(str string) (bool, string) {
	ok, message := laroussefr.IsURL(str)
	if !ok {
		return false, message
	}
	
	substr := "larousse.fr/dictionnaires/francais/"
	if !strings.Contains(str, substr) {
		return false, fmt.Sprintf("Must contain \"%s\"", substr)
	}
	
	if strings.HasSuffix(str, substr) {
		return false, "Missing protocol (http:// or https://)"
	}
	return true, ""
}

// newPageFromRoot returns a new Result from an HTML root.
func newResultFromRoot(doc *html.Node) (Result, error) {
	pageID, err := laroussefr.GetPageID(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	
	head, err := findHeader(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	
	defs, err := findDefinitions(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	
	exprs, err := findExpressions(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	
	rels, err := findRelations(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	
	homs, err := findHomonymes(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	
	diffis, err := findDifficultes(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	
	cits, err := findCitations(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	
	seeAlso, err := laroussefr.GetSimilarWords(doc)
	if err != nil {
		return Result{}, laroussefr.NewError("newResultFromRoot", "", err.Error())
	}
	
	res := Result{pageID, head, defs, exprs, rels, homs, diffis, cits, seeAlso}
	return res, nil
}

// findHeader returns a word's Header.
func findHeader(doc *html.Node) (Header, error) {
	texte, err := findHeaderTexte(doc)
	if err != nil {
		return Header{}, laroussefr.NewError("findHeader", "", err.Error())
	}
	
	audio, err := findHeaderAudio(doc)
	if err != nil {
		return Header{}, laroussefr.NewError("findHeader", "", err.Error())
	}
	
	typ:= findHeaderType(doc)
	
	head := Header{texte, audio, typ}
	return head, nil
}

// findHeaderTexte returns a word's text (e.g. vert -> []string{"vert", "verte"} ).
func findHeaderTexte(doc *html.Node) (string, error) {
	nodes := scrape.FindAll(doc, match.HeaderTexteNode)
	if len(nodes) == 0 {
		return "", laroussefr.NewError("findHeaderTexte", "",  "failed to find HeaderTexte nodes")
	}
	
	var out string
	for i, n := range nodes {
		if i > 0 && !strings.HasSuffix(out, ",") {
			out += ", "
		}
		out += scrape.Text(n)
	}
	return out, nil
}

// findHeaderAudio returns a word's audio URL (e.g. vert -> 
// https://laroussefr.fr/dictionnaires-prononciation/francais/tts/64636fra2).
func findHeaderAudio(doc *html.Node) (string, error) {
	n, ok := scrape.Find(doc, match.HeaderAudioNode)
	if !ok {
		return "", laroussefr.NewError("findHeaderAudio", "", "failed to find audio node")
	}
	url := laroussefr.GetAudioURL(n)
	return url, nil
}

// findHeaderType returns a word's Type as a string.
// 
// Note: This field could be empty (see page for "auto" or "cotentin").
func findHeaderType(doc *html.Node) string {
	n, ok := scrape.Find(doc, match.HeaderTypeNode)
	if ok {
		return n.Data
	}
	return ""
}

// findDefinitions returns a word's DÉFINITIONS list.
func findDefinitions(doc *html.Node) ([]Definition, error) {
	var out []Definition
	defNodes := scrape.FindAll(doc, match.DefinitionNode)
	for _, n := range defNodes {
		arr, err := parse.DefinitionNode(n)
		if err != nil {
			return nil, laroussefr.NewError("findDefinitions", "", err.Error())
		}
		def := Definition{arr[0], arr[1], arr[2]}
		out = append(out, def)
	}
	return out, nil
}

// findDefinitionsFull returns a word's DÉFINITIONS list merged with the
// corresponding items in the SYNONYMES ET CONTRAIRES list.
/*
func findDefinitionsFull(doc *html.Node) ([]Definition, error) {
	defs, err := findDefinitions(doc)
	if err != nil {
		return nil, laroussefr.NewError("findDefinitionsFull", "", err.Error())
	} else if defs == nil {
		return nil, nil
	}
	
	rels, err := findRelations(doc)
	if err != nil {
		return nil, laroussefr.NewError("findDefinitionsFull", "", err.Error())
	}
	out := mergeDefinitionsAndRelations(defs, rels)
	return out, nil
}*/

// findExpressions returns a word's EXPRESSIONS list.
func findExpressions(doc *html.Node) ([]Expression, error) {
	var out []Expression
	nodes := scrape.FindAll(doc, match.ExpressionNode)
	for _, n := range nodes {
		textes, redBig, redSmall, err := parse.ExpressionNode(n)
		if err != nil {
			return nil, laroussefr.NewError("findExpressions", "", err.Error())
		}
		exp := Expression{textes, redBig, redSmall}
		out = append(out, exp)
	}
	return out, nil
}

// findRelations returns a word's SYNONYMES ET CONTRAIRES list.
func findRelations(doc *html.Node) ([]Relation, error) {
	var out []Relation
	nodes := scrape.FindAll(doc, match.RelationNode)
	
	for _, n := range nodes {
		texte, syns, conts, err := parse.RelationNode(n)
		if err != nil {
			return nil, laroussefr.NewError("findRelations", "", err.Error())
		}
		rel := Relation{texte, syns, conts}
		out = append(out, rel)
	}
	return out, nil
}

// findHomonymes returns a word's HOMONYMES list.
func findHomonymes(doc *html.Node) ([]Homonyme, error) {
	var out []Homonyme
	nodes := scrape.FindAll(doc, match.HomonymeNode)
	
	for _, n := range nodes {
		texte, typ, err := parse.HomonymeNode(n)
		if err != nil {
			return nil, laroussefr.NewError("findHomonymes", "", err.Error())
		}
		hom := Homonyme{texte, typ}
		out = append(out, hom)
	}
	return out, nil
}

// findDifficultes returns a word's DIFFICULTÉS list.
func findDifficultes(doc *html.Node) ([]Difficulte, error) {
	var out []Difficulte
	diffNodes := scrape.FindAll(doc, match.DifficulteNode)
	
	for _, n := range diffNodes {
		categorie, texte, err := parse.DifficulteNode(n)
		if err != nil {
			return nil, laroussefr.NewError("findDifficultes", "", err.Error())
		}
		diff := Difficulte{categorie, texte}
		out = append(out, diff)
	}
	return out, nil
}

// findCitations returns a word's CITATIONS list.
func findCitations(doc *html.Node) ([]Citation, error) {
	var out []Citation
	citationNodes := scrape.FindAll(doc, match.CitationNode)
	
	for _, n := range citationNodes {
		id, arr, err := parse.CitationNode(n)
		if err != nil {
			return nil, laroussefr.NewError("findCitations", "", err.Error())
		}
		cit := Citation{id, arr[0], arr[1], arr[2], arr[3]}
		out = append(out, cit)
	}
	return out, nil
}

// mergeDefinitionsAndRelations returns a new slice of Definitions, which is
// identical to defs but with rels's Synonymes and Contraires.
/*
func mergeDefinitionsAndRelations(defs []Definition, rels []Relation) []Definition {
	var out []Definition
	for _, d := range defs {
		for _, r := range rels {
			r.Texte = strings.TrimRight(r.Texte, " .")
			if strings.HasPrefix(d.Texte, r.Texte) {
				if r.hasSynonymes() {
					d.Synonymes = r.Synonymes
				}
				if r.hasContraires() {
					d.Contraires = r.Contraires
				}
				if r.hasSynonymes() || r.hasContraires() {
					break
				}
			}
		}
		out = append(out, d)
	}
	return out
}*/

// Package match provides matcher functions to be used with package
// github.com/yhat/scrape.
package match

import "golang.org/x/html"

// PhraseProp returns true if n is a node containing a Phrase property.
func PhraseProp(n *html.Node) bool {
	switch scrape.Attr(n, "class") {
		case "Locution2":          fallthrough
		case "Glose2":             fallthrough
		case "Traduction2":        fallthrough
		case "Renvois":            fallthrough
		case "Metalangue2":        fallthrough
		case "lienson3":           fallthrough
		case "lienson2":           fallthrough
		case "Indicateur":         fallthrough
		case "IndicateurDomaine":  fallthrough
		case "Metalangue":         fallthrough
		case "DivisionExpression": return true
	}
	return false
}

// MeaningProp returns true if n is a node containing a Meaning property.
func MeaningProp(n *html.Node) bool {
	switch scrape.Attr(n, "class") {
		case "Traduction":         fallthrough
		case "Glose2":             fallthrough // for en->fr "blue" POLITICS
		case "Indicateur":         fallthrough
		case "IndicateurDomaine":  fallthrough
		case "Metalangue":         return true
	}
	return false
}

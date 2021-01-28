// Package match contains matcher functions to be used with package
// github.com/yhat/scrape.
package match

import (
	"github.com/yhat/scrape"
	
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// class returns n's "class" attribute.
func class(n *html.Node) string {
	return scrape.Attr(n, "class")
}

// HeaderTexteNode returns true if n is a node containing a header's Texte.
func HeaderTexteNode(n *html.Node) bool {
	if n.Type != html.TextNode {
		return false
	}
	
	prev := n.PrevSibling
	if prev == nil {
		return false
	}
	
	return prev.DataAtom == atom.Audio
}

// HeaderAudioNode returns true if n is an <audio> node, which contains a
// header's Audio.
func HeaderAudioNode(n *html.Node) bool {
	return n.DataAtom == atom.Audio
}

// HeaderTypeNode returns true if n contains a header's Type.
func HeaderTypeNode(n *html.Node) bool {
	par := n.Parent
	if par == nil {
		return false
	}
	return par.DataAtom == atom.P && class(par) == "CatgramDefinition"
}

// DefinitionNode returns true if n is an item in the DÉFINITIONS sections.
func DefinitionNode(n *html.Node) bool {
	return n.DataAtom == atom.Li && class(n) == "DivisionDefinition" && n.FirstChild != nil
}

// ExpressionNode returns true if n is an item on the EXPRESSIONS list.
// This returns true for -any- expression item, not just the first one.
func ExpressionNode(n *html.Node) bool {
	return n.DataAtom == atom.Li && class(n) == "Locution"
}

// RelationNode returns true if n is an item on the SYNONYMES ET CONTRAIRES
// list.
// This returns true for -any- relation item, not just the first one.
//
// NOTE: Some words, such as 'aguiche', have a synonyme with no corresponding
// definition.
// 
// NOTE 2: At the end, check for both DivisionDefinition and <b> (see final item
// on "beau" page).
func RelationNode(n *html.Node) bool {
	if n.DataAtom != atom.Div {
		return false
	}
	if class(n) != "SensSynonymes" {
		return false
	}
	
	m := n.FirstChild
	if m == nil {
		return false
	}
	
	// NOTE 2
	if m.DataAtom == atom.P && class(m) == "DivisionDefinition" {
		return true
	}
	return m.DataAtom == atom.B
}

// SuggestionsNode returns true if n contains the "word not found - try these
// suggestions" text.
func SuggestionsNode(n *html.Node) bool {
	return n.DataAtom == atom.H1 && class(n) == "icon-question-sign" && scrape.Text(n) == "Suggestions proposées par le correcteur"
}

// NoSuggestionsNode returns true if n contains the "word not found - no
// suggestions found" text.
func NoSuggestionsNode(n *html.Node) bool {
	return n.DataAtom == atom.P && class(n) == "err" && scrape.Text(n) == "Nous n'avons aucune suggestion pour votre recherche"
}

// HomonymeNode returns true if n is an item on the HOMONYMES list.
func HomonymeNode(n *html.Node) bool {
	return n.DataAtom == atom.Li && class(n) == "Homonyme"
}

// DifficulteNode returns true if n is an item on the DIFFICULTÉS list.
func DifficulteNode(n *html.Node) bool {
	return n.DataAtom == atom.Li && class(n) == "Difficulte"
}

// DifficulteTypeNode returns true if n holds the Type field of a DIFFICULTÉ.
func DifficulteTypeNode(n *html.Node) bool {
	return n.DataAtom == atom.P && class(n) == "TypeDifficulte"
}

// DifficulteTexteNode returns true if n holds the Texte field of a DIFFICULTÉ.
func DifficulteTexteNode(n *html.Node) bool {
	return n.DataAtom == atom.P && class(n) == "DefinitionDifficulte"
}

// CitationNode returns true if n is an item on the CITATIONS list.
func CitationNode(n *html.Node) bool {
	return n.DataAtom == atom.Li && class(n) == "Citation"
}

// CitationAuteurNode returns true if n is an Auteur node within a CITATION
// node.
func CitationAuteurNode(n *html.Node) bool {
	return n.DataAtom == atom.Span && class(n) == "AuteurCitation"
}

// CitationInfoAutuerNode returns true if n is an InfoAuteur node within a
// CITATION node.
func CitationInfoAuteurNode(n *html.Node) bool {
	return n.DataAtom == atom.Span && class(n) == "InfoAuteurCitation"
}

// CitationTexteNode returns true if n is a Texte node within a CITATION node.
func CitationTexteNode(n *html.Node) bool {
	return n.DataAtom == atom.Span && class(n) == "TexteCitation"
}

// CitationInfoNode returns true if n is an Info node within a CITATION node.
func CitationInfoNode(n *html.Node) bool {
	return n.DataAtom == atom.Span && class(n) == "InfoCitation"
}

// RubriqueDefinitionNode returns true if n is a <p> element of class
// RubriqueDefinition.
func RubriqueDefinitionNode(n *html.Node) bool {
	return n.DataAtom == atom.P && class(n) == "RubriqueDefinition"
}

// IndicateurDefinitionNode returne true if n is a <span> node of class
// indicateurDefinition (note the lowercase 'i').
func IndicateurDefinitionNode(n *html.Node) bool {
	return n.DataAtom == atom.Span && class(n) == "indicateurDefinition"
}

// IndicateurLocutionnNode returne true if n is a <span> node of class
// IndicateurLocution.
func IndicateurLocutionNode(n *html.Node) bool {
	return n.DataAtom == atom.Span && class(n) == "IndicateurLocution"
}

// ExampleDefinitionNode return true if n is a <span> element of class
// ExempleDefinition.
func ExempleDefinitionNode(n *html.Node) bool {
	return n.DataAtom == atom.Span && class(n) == "ExempleDefinition"
}

// AdresseLocutionNode returns true if n is an <h2> element of class
// AdresseLocution, which holds a single Textes element of an Expression.
func AdresseLocutionNode(n *html.Node) bool {
	return n.DataAtom == atom.H2 && class(n) == "AdresseLocution"
}

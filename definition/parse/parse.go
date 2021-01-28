// parse.go contains functions for parsing nodes into desired data.
package parse

import (
	"strconv"
	"strings"
	
	"github.com/serope/laroussefr"
	"github.com/serope/laroussefr/definition/match"
	
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// DefinitionNode takes a DEFINITION node and returns the fields for a
// Definition object.
// 
// Note: Some pages have a single DÉFINITION node without any child nodes (see
// old page for "delà").
func DefinitionNode(n *html.Node) ([3]string, error) {
	m := n.FirstChild
	if m == nil {
		return [3]string{}, laroussefr.NewError("DefinitionNode", "", "nil FirstChild")
	}
	
	var texte, redBig, redSmall string
	for m != nil {
		switch {
			case match.RubriqueDefinitionNode(m):
				redBig = scrape.Text(m)
			case match.IndicateurDefinitionNode(m):
				redSmall = scrape.Text(m)
			default:
				if shouldGetSpace(texte) {
					texte += " "
				}
				texte += scrape.Text(m)
		}
		m = m.NextSibling
	}
	return [3]string{texte, redBig, redSmall}, nil
}

// shouldGetSpace returns true if str should be appended with a space (that is,
// if it's non-empty and doesn't end with a space).
func shouldGetSpace(str string) bool {
	if len(str) == 0 {
		return false
	}
	i := len(str)-1
	return str[i] != ' '
}

// ExpressionNode takes an EXPRESSION ("Locution") node and returns the string
// fields for an Expression object.
func ExpressionNode(n *html.Node) (string, string, string, error) {
	var textes []string
	var redBig, redSmall string
	
	// redBig
	rbn, ok := scrape.Find(n, match.RubriqueDefinitionNode)
	if ok {
		redBig = scrape.Text(rbn)
	}
	
	nodes := scrape.FindAll(n, match.AdresseLocutionNode)
	for _, n := range nodes {
		// redSmall
		indiloc, ok := scrape.Find(n, match.IndicateurLocutionNode)
		if ok {
			redSmall = scrape.Text(indiloc)
		}
		
		// texte
		texte := scrape.Text(n)
		if n.NextSibling != nil {
			texte += " "
			texte += scrape.Text(n.NextSibling)
		}
		
		textes = append(textes, texte)
	}
	
	texte := strings.Join(textes, " ")
	if strings.HasPrefix(texte, redSmall) {
		texte = strings.Replace(texte, redSmall, "", 1)
	}
	texte = expressionCleanupTexte(texte)
	return texte, redBig, redSmall, nil
}

// expressionCleanupTexte cleans up the texte parsed in ExpressionNode.
func expressionCleanupTexte(texte string) string {
	replace := map[string]string{
		"' " : "'",
		" ." : ".",
	}
	for k, v := range replace {
		if strings.Contains(texte, k) {
			texte = strings.ReplaceAll(texte, k, v)
		}
	}
	texte = strings.Trim(texte, " ")
	return texte
}

// isExpressionTexteNode returns true if n is part of the Texte portion of an
// EXPRESSIONS node.
func isExpressionTexteNode(n *html.Node) bool {
	if n.DataAtom == atom.Span && scrape.Attr(n, "class") == "TexteLocution" {
		return true
	}
	if match.AdresseLocutionNode(n) {
		return true
	}
	if n.Type == html.TextNode {
		m := n.Parent
		if m == nil {
			return false
		}
		return match.AdresseLocutionNode(m)
	}
	return false
}

// HomonymeNode takes a HOMONYMES node and returns the Texte and Type fields
// for a Homonyme object.
func HomonymeNode(n *html.Node) (string, string, error) {
	m, ok := scrape.Find(n, scrape.ByClass("Renvois"))
	if !ok {
		m, ok = scrape.Find(n, scrape.ByTag(atom.B))
		if !ok {
			return "", "", laroussefr.NewError("HomonymeNode", "", "can't find texte")
		}
	}
	texte := scrape.Text(m)
	
	m, ok = scrape.Find(n, scrape.ByClass("CatGramHomonyme"))
	var typ string // typ is optional (see "brique")
	if ok {
		typ = scrape.Text(m)
	}
	
	return texte, typ, nil
}

// RelationNode parses a single SYNONYMES ET CONTRAIRES node into the fields
// for a Relation object.
func RelationNode(n *html.Node) (string, []string, []string, error) {
	texte, err := parseRelationNodeTexte(n)
	if err != nil {
		return "", nil, nil, laroussefr.NewError("RelationNode", "", err.Error())
	}
	lists, err := parseRelationNodeLists(n)
	if err != nil {
		return "", nil, nil, laroussefr.NewError("RelationNode", "", err.Error())
	}
	return texte, lists[0], lists[1], nil
}

// parseRelationText retrieves the Texte from a relation node.
func parseRelationNodeTexte(n *html.Node) (string, error) {
	m := n.FirstChild
	if m == nil {
		return "", laroussefr.NewError("parseRelationNodeTexte", "", "nil FirstChild")
	}
	return scrape.Text(m), nil
}

// parseRelationNodeLists returns both the SYNONYMES list and CONTRAIRES list
// from a relation node, in that order.
func parseRelationNodeLists(n *html.Node) ([2][]string, error) {
	var out [2][]string
	
	m := n.FirstChild
	if m == nil {
		return out, laroussefr.NewError("parseRelationNodeLists", "", "nil FirstChild")
	}
	
	m = m.NextSibling
	if m == nil {
		return out, laroussefr.NewError("parseRelationNodeLists", "", "nil NextSibling")
	}
	
	var i int
	if strings.HasPrefix(scrape.Text(m), "Synonyme") {
		i = 0
	} else {
		i = 1
	}
	m = m.NextSibling
	out[i] = strings.Split(scrape.Text(m), " - ")
	if i == 1 || m.NextSibling == nil {
		return out, nil
	}
	
	m = m.NextSibling.NextSibling
	out[1] = strings.Split(scrape.Text(m), " - ")
	return out, nil
}

// DifficulteNode takes a DIFFICULTÉ node and returns the text fields for a
// Difficulte object.
func DifficulteNode(n *html.Node) (string, string, error) {
	// Type
	var typ string
	typeNode, ok := scrape.Find(n, match.DifficulteTypeNode)
	if !ok {
		return "", "", laroussefr.NewError("DifficulteNode", "", "Can't find Type")
	}
	typ = scrape.Text(typeNode)
	
	var texte string
	m := typeNode.NextSibling
	for m != nil {
		texte += scrape.Text(m)
		m = m.NextSibling
	}
	
	return typ, texte, nil
}


// CitationNode takes a CITATION node and returns the ID and string fields for
// a Citation object.
func CitationNode(n *html.Node) (int, [4]string, error) {
	id, err := getNodeID(n)
	if err != nil {
		return -1, [4]string{}, laroussefr.NewError("CitationNode", "", err.Error())
	}
	
	auteurNode, ok := scrape.Find(n, match.CitationAuteurNode)
	var auteur string // auteur optional; see "arbre" page
	if ok {
		auteur = scrape.Text(auteurNode)
	}
	
	infoAuteurNode, ok := scrape.Find(n, match.CitationInfoAuteurNode)
	var infoAuteur string // infoAuteur optional; see "arbre" page
	if ok {
		infoAuteur = scrape.Text(infoAuteurNode)
	}
	
	texteNode, ok := scrape.Find(n, match.CitationTexteNode)
	if !ok {
		return -1, [4]string{}, laroussefr.NewError("CitationNode", "", "can't find Texte node")
	}
	texte := scrape.Text(texteNode)
	
	infoNode, ok := scrape.Find(n, match.CitationInfoNode)
	var info string // info optional; see "voici"
	if ok {
		info = scrape.Text(infoNode)
	}
	
	return id, [4]string{auteur, infoAuteur, texte, info}, nil
}

// getNodeID takes a node with an "id" attribute and returns it as an integer.
func getNodeID(n *html.Node) (int, error) {
	idStr := scrape.Attr(n, "id")
	if idStr == "" {
		return -1, laroussefr.NewError("getNodeID", "", "empty idStr")
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return -1, laroussefr.NewError("getNodeID", "", err.Error())
	}
	return idInt, nil
}



// isTexteLocutionNode returns true if n is a <span> element of class
// TexteLocution, which holds the Description of an Expression.
func isTexteLocutionNode(n *html.Node) bool {
	return n.Data == "span" && scrape.Attr(n, "class") == "TexteLocution"
}


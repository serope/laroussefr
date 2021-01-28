// Package parse contains functions for parsing certain kinds of nodes.
package parse

import (
	"strings"
	
	"scraper/laroussefr"
	
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"github.com/yhat/scrape"
)

// Traduction takes a "Traduction" node and returns its inner text.
func Traduction(n *html.Node) string {
	var out string
	m := n.FirstChild
	for m != nil {
		text := scrape.Text(m)
		class := scrape.Attr(m, "class")
		if class == "Genre" || strings.HasSuffix(out, ",") {
			out += " "
		}
		
		if isOuBienNode(m) {
			out += " ou "
		} else if class != "lienconj2" && class != "Metalangue2" {
			if strings.HasPrefix(text, "(") {
				out += " "
			}
			out += text
		}
		m = m.NextSibling
	}
	return out
}

// isOuBienNode is true if n is a <span class="oubien"> node.
func isOuBienNode(n *html.Node) bool {
	return n.DataAtom == atom.Span && scrape.Attr(n, "class") == "oubien"
}

// isSpace returns true if n is a text node consisting of a single space.
func isSpace(n *html.Node) bool {
	return n.Type == html.TextNode && n.Data == " "
}

// ZoneEntree takes a "ZoneEntree" node and returns a [5]string array containing
// the values to be assigned to a Header object.
//
// [0] Texte
// [1] TexteAlt
// [2] Phonetique
// [3] Audio
// [4] Type
func ZoneEntree(n *html.Node) ([5]string, error) {
	texte, err := parseEntreeTexte(n)
	if err != nil {
		return [5]string{}, laroussefr.NewError("ZoneEntree", "", err.Error())
	}
	texteAlt := parseEntreTexteAlt(n)
	phonetique := parseEntreePhonetique(n)
	audio := parseEntreeAudio(n)
	typ := parseEntreeType(n)
	return [5]string{texte, texteAlt, phonetique, audio, typ}, nil
}

// parseEntreeTexte takes a "ZoneEntree" node and returns the value to be
// assigned to the Texte field.
func parseEntreeTexte(n *html.Node) (string, error) {
	adresseNode, ok := scrape.Find(n, scrape.ByClass("Adresse"))
	if !ok {
		return "", laroussefr.NewError("parseEntreeTexte", "", "Failed to find Adresse node")
	}
	return scrape.Text(adresseNode), nil
}

// parseEntreeTexteAlt takes a "ZoneEntree" node and returns the value to be
// assigned to the TexteAlt field.
func parseEntreTexteAlt(n *html.Node) string {
	formeFlechieAdresseNode, ok := scrape.Find(n, scrape.ByClass("FormeFlechieAdresse"))
	if !ok {
		return ""
	}
	str := scrape.Text(formeFlechieAdresseNode)
	if strings.HasPrefix(str, "( ") {
		str = "(" + str[2:]
	}
	return str
}

// parseEntreePhonetique takes a "ZoneEntree" node and returns the value to be
// assigned to the Phonetique field.
func parseEntreePhonetique(n *html.Node) string {
	phonetiqueNodes := scrape.FindAll(n, scrape.ByClass("Phonetique"))
	var out string
	for _, p := range phonetiqueNodes {
		out += scrape.Text(p)
	}
	return out
}

// parseEntreeAudio takes a "ZoneEntree" node and returns the value to be
// assigned to the Audio field.
func parseEntreeAudio(n *html.Node) string {
	lienson, ok := scrape.Find(n, scrape.ByClass("lienson"))
	if !ok {
		return ""
	}
	return Lienson(lienson)
}

// parseEntreeType takes a "ZoneEntree" node and returns the value to be
// assigned to the Type field.
func parseEntreeType(n *html.Node) string {
	m, ok := scrape.Find(n, scrape.ByClass("ZoneGram"))
	if !ok {
		m, ok = scrape.Find(n, scrape.ByClass("CategorieGrammaticale"))
		if !ok {
			return ""
		}
	}
	out := scrape.Text(m)
	out = strings.ReplaceAll(out, "Conjugaison", "")
	out = strings.ReplaceAll(out, "  ", " ")
	out = strings.Trim(out, " ")
	return out
}

// Lienson takes a "lienson", "lienson2" or "lienson3" span node and returns the
// URL to the audio clip.
// 
// Note that the URL in the "src" attribute redirects to a voix.laroussefr.fr
// address.
func Lienson(n *html.Node) string {
	m := n.NextSibling
	if m == nil {
		return ""
	} else if m.Type == html.TextNode {
		m = m.NextSibling
	} else if m.DataAtom != atom.Audio {
		return ""
	}
	return laroussefr.GetAudioURL(m)
}

// Adresse takes an "Adresse" node and returns the header values for a
// tableWord or a monoWord.
//
// [0] Text
// [1] TextAlt
// [2] Phonetic
// [3] Audio
// [4] Type
func Adresse(n *html.Node) [5]string {
	text := scrape.Text(n)
	audio := parseAdresseAudio(n)
	phonetic := parseAdressePhonetic(n)
	typ := parseAdresseType(n)
	return [5]string{text, "", phonetic, audio, typ}
}

// parseAdresseAudio takes an Adresse node and returns the string to be assigned
// to a thinWord's Audio field.
func parseAdresseAudio(n *html.Node) string {
	m := n
	for i:=0; i<3; i++ {
		m = m.PrevSibling
		if m == nil {
			break
		} else if scrape.Attr(m, "class") == "lienson" {
			return Lienson(m)
		}
	}
	return ""
}

// parseAdressePhonetic takes an Adresse node and returns the string to be
// assigned to a thinWord's Phonetic field.
func parseAdressePhonetic(n *html.Node) string {
	m := n
	for i:=0; i<3; i++ {
		m = m.NextSibling
		if m == nil {
			break
		} else if scrape.Attr(m, "class") == "Phonetique" {
			return scrape.Text(m)
		}
	}
	return ""
}

// parseAdresseType takes an Adresse node and returns the string to be
// assigned to a thinWord's Type field.
func parseAdresseType(n *html.Node) string {
	var out string
	m := n
	
	for i:=0; i<5; i++ {
		m = m.NextSibling
		if m == nil {
			break
		} else if m.DataAtom == atom.Span && scrape.Attr(m, "class") == "CategorieGrammaticale" {
			out = scrape.Text(m)
			if strings.HasSuffix(out, " Conjugaison") {
				i := strings.LastIndexByte(out, ' ')
				out = out[:i]
			}
			break
		}
	}
	
	return out
}

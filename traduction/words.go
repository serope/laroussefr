// words.go contains functions pertaining to the internal types bigWord and
// smallWord, both of which eventually become Words before being exported.
package traduction

import (
	"strconv"
	"strings"
	"unicode"
	
	"github.com/serope/laroussefr"
	"github.com/serope/laroussefr/traduction/parse"
	
	"github.com/yhat/scrape"
	"golang.org/x/net/html/atom"
	"golang.org/x/net/html"
)

// Type bigWord represents a word with subheaders.
// 
// The word header is in a "ZoneEntree" node and its definitions and phrases
// are both in a "ZoneTexte" node after it. The "ZoneTexte" consists of items
// stored in "itemBLSEM" nodes, which are further divided into "itemZONESEM"
// nodes.
// 
// An example of a bigWord is the first word on the fr->en "coup" page
// (https://www.larousse.fr/dictionnaires/francais-anglais/coup/19682).
type bigWord Word

// Type smallWord represents a word that doesn't have subheaders, but is
// otherwise structured similarly to a bigWord.
// 
// The word header is in a "ZoneEntree" node and its definitions and phrases
// are both in a "ZoneTexte" node after it. The "ZoneTexte" consists of items
// stored in "itemZONESEM" nodes.
// 
// An example of a smallWord is every word but the first on the same page
// linked above.
type smallWord struct {
	Code   int
	Header Header
	Items  []Item
}

func (sw smallWord) toWord() Word {
	sh := Subheader{"", sw.Items}
	return Word{sw.Code, sw.Header, []Subheader{sh}}
}



// scrapeSmallWords gets all the smallWords from this page.
func scrapeSmallWords(doc *html.Node) ([]smallWord, error) {
	zoneEntreeNodes, err := getSmallWordZoneEntreeNodes(doc)
	if err != nil {
		return nil, laroussefr.NewError("scrapeSmallWords", "", err.Error())
	}
	
	var out []smallWord
	for i, zoneEntreeNode := range zoneEntreeNodes {
		// Code
		code := getWordCode(i, doc, zoneEntreeNode)
		
		// Entree
		arr, err := parse.ZoneEntree(zoneEntreeNode)
		if err != nil {
			return nil, laroussefr.NewError("scrapeSmallWords", "", err.Error())
		}
		header := Header{arr[0], arr[1], arr[2], arr[3], arr[4]}
		
		// ZoneTexte
		zoneTexteNode := zoneEntreeNode.NextSibling
		itemNodes := scrape.FindAll(zoneTexteNode, scrape.ByClass("itemZONESEM"))
		if len(itemNodes) == 0 {
			itemNodes = []*html.Node{zoneTexteNode}
		}
		items, err := scrapeItems(itemNodes)
		if err != nil {
			return nil, laroussefr.NewError("scrapeSmallWords", "", err.Error())
		}
		
		sw := smallWord{code, header, items}
		out = append(out, sw)
	}
	
	return out, nil
}

// scrapeBigWords gets all the bigWords from this page, if any are present.
func scrapeBigWords(doc *html.Node) ([]bigWord, error) {
	if !hasBigWords(doc) {
		return nil, nil
	}
	
	zoneEntreeNodes, err := getBigWordZoneEntreeNodes(doc)
	if err != nil {
		return nil, laroussefr.NewError("scrapeBigWords", "", err.Error())
	}
	
	var out []bigWord
	for i, zoneEntreeNode := range zoneEntreeNodes {
		// Code
		code := getWordCode(i, doc, zoneEntreeNode)
		
		// Entree
		arr, err := parse.ZoneEntree(zoneEntreeNode)
		if err != nil {
			return nil, laroussefr.NewError("scrapeBigWords", "", err.Error())
		}
		header := Header{arr[0], arr[1], arr[2], arr[3], arr[4]}
		
		// ZoneTexte
		zoneTexteNode := zoneEntreeNode.NextSibling
		blackNodes := getBlackNodes(zoneTexteNode)
		blacks, err := scrapeBlackNodes(blackNodes)
		if err != nil {
			return nil, laroussefr.NewError("scrapeBigWords", "", err.Error())
		}
		
		bw := bigWord{code, header, blacks}
		out = append(out, bw)
	}
	
	return out, nil
}

// hasBigWords returns true of this page contains bigWords.
func hasBigWords(doc *html.Node) bool {
	itemBLSEMnodes := scrape.FindAll(doc, scrape.ByClass("itemBLSEM1"))
	if len(itemBLSEMnodes) > 0 {
		return true
	}
	return false
}

// getBlackNodes returns all "itemBLSEM1" and "itemBLSEM" nodes, which are
// used to create black (Subheader) objects.
func getBlackNodes(doc *html.Node) []*html.Node {
	a := scrape.FindAll(doc, scrape.ByClass("itemBLSEM1"))
	b := scrape.FindAll(doc, scrape.ByClass("itemBLSEM"))
	return append(a, b...)
}

// scrapeBlackNodes takes a slice of black nodes returns a Subheader slice.
func scrapeBlackNodes(blackNodes []*html.Node) ([]Subheader, error) {
	var out []Subheader
	for _, blackNode := range blackNodes {
		bl, err := scrapeBlackNode(blackNode)
		if err != nil {
			return nil, laroussefr.NewError("scrapeBlackNodes", "", err.Error())
		}
		out = append(out, bl)
	}
	return out, nil
}

// scrapeBlackNode takes a black node ("itemBLSEM1" or "itemBLSEM") and returns
// a Subheader.
func scrapeBlackNode(blackNode *html.Node) (Subheader, error) {
	title := scrapeBlackTitle(blackNode)
	itemNodes := scrape.FindAll(blackNode, scrape.ByClass("itemZONESEM"))
	if len(itemNodes) == 0 {
		itemNodes = []*html.Node{blackNode}
	}
	items, err := scrapeItems(itemNodes)
	if err != nil {
		return Subheader{}, laroussefr.NewError("scrapeBlackNode", blackNode.Data, err.Error())
	}
	return Subheader{title, items}, nil
}

// scrapeBlackTitle returns the Title from a black node.
// 
// Note: No longer returns error if no title found; en->fr "make" has subheaders
// without titles!
func scrapeBlackTitle(blackNode *html.Node) string {
	indicateur2, ok := scrape.Find(blackNode, scrape.ByClass("Indicateur2"))
	var title string
	if ok {
		title = scrape.Text(indicateur2)
	}
	return title
}

// getBigWordZoneEntreeNodes returns all ZoneEntree nodes associated with
// bigWords.
func getBigWordZoneEntreeNodes(doc *html.Node) ([]*html.Node, error) {
	zoneEntreeNodes := scrape.FindAll(doc, scrape.ByClass("ZoneEntree"))
	var out []*html.Node
	for _, zoneEntreeNode := range zoneEntreeNodes {
		zoneTexteNode := zoneEntreeNode.NextSibling
		if zoneTexteNode == nil {
			return nil, laroussefr.NewError("getBigWordZoneEntreeNodes", "", "nil sibling node after ZoneEntree")
		}
		if hasBigWords(zoneTexteNode) {
			out = append(out, zoneEntreeNode)
		}
	}
	
	return out, nil
}

// getSmallWordZoneEntreeNodes returns all ZoneEntree nodes associated with
// smallWords.
func getSmallWordZoneEntreeNodes(doc *html.Node) ([]*html.Node, error) {
	zoneEntreeNodes := scrape.FindAll(doc, scrape.ByClass("ZoneEntree"))
	var out []*html.Node
	for _, zoneEntreeNode := range zoneEntreeNodes {
		zoneTexteNode := zoneEntreeNode.NextSibling
		if zoneTexteNode == nil {
			return nil, laroussefr.NewError("getSmallWordZoneEntreeNodes", "", "nil sibling node after ZoneEntree")
		}
		if hasBigWords(zoneTexteNode) {
			continue
		}
		out = append(out, zoneEntreeNode)
	}
	
	return out, nil
}

// scrapeItems takes a slice of "itemZONESEM" nodes and returns an Item slice.
func scrapeItems(itemNodes []*html.Node) ([]Item, error) {
	var out []Item
	for _, itemNode := range itemNodes {
		item := scrapeItem(itemNode)
		out = append(out, item)
	}
	return out, nil
}

// scrapeItem takes an "itemZONESEM" node and returns an Item.
func scrapeItem(itemNode *html.Node) Item {
	meanings := scrapeMeanings(itemNode)
	phrases := scrapePhrases(itemNode)
	return Item{meanings, phrases}
}

// scrapePhrases takes an "itemZONESEM" node and returns a Phrase slice.
func scrapePhrases(n *html.Node) []Phrase {
	a := scrape.FindAll(n, scrape.ByClass("ZoneExpression1"))
	b := scrape.FindAll(n, scrape.ByClass("ZoneExpression"))
	if len(b) > 0 {
		a = append(a, b...)
	}
	exprNodes := a
	
	var out []Phrase
	for _, e := range exprNodes {
		phrase := getPhraseFromZoneExpression(e)
		out = append(out, phrase)
	}
	
	// blues
	blues := scrapeExpressions(n)
	out = append(out, blues...)
	return out
}

// scrapeExpressions takes an "itemZONESEM" node and returns a Phrase slice of
// expressions, if any exist.
func scrapeExpressions(n *html.Node) []Phrase {
	blocExpressionNode, ok := scrape.Find(n, scrape.ByClass("BlocExpression"))
	if !ok {
		return nil
	}
	
	firstPhrase := getPhraseFromZoneExpression(blocExpressionNode)
	firstPhrase.setBlue(true)
	out := []Phrase{firstPhrase}
	
	exprNodes := scrape.FindAll(n, scrape.ByClass("ZoneExpression2"))
	for _, e := range exprNodes {
		phrase := getPhraseFromZoneExpression(e)
		phrase.setBlue(true)
		out = append(out, phrase)
	}
	return out
}

// scrapeMeanings takes an item node ("itemZONESEM") and returns a list of
// Meanings in this node.
func scrapeMeanings(itemNode *html.Node) []Meaning {
	// 1st genre/meaning strings
	n := itemNode.FirstChild
	if n.Type == html.TextNode && isWhitespace(n.Data) {
		n = n.NextSibling
	}
	
	var m Meaning
	for stillOnFirstMeaningStrings(n) {
		m.update(n)
		n = n.NextSibling
	}
	
	// 1st done
	out := []Meaning{m}
	
	// other genres/meanings
	semantiqueNodes := scrape.FindAll(itemNode, scrape.ByClass("division-semantique"))
	for _, s := range semantiqueNodes {
		if s == itemNode {
			continue
		}
		// meaning := scrapeMeanings(s)[0]
		meanings := scrapeMeanings(s)
		if len(meanings) > 0 {
			out = append(out, meanings[0])
		}
	}
	
	// end
	if len(out) == 1 && out[0].isEmpty() {
		out = nil
	}
	return out
}

// getWordCode returns the code associated with the ith "ZoneEntree" node on
// this page, starting at i=0.
func getWordCode(i int, doc *html.Node, zoneEntreeNode *html.Node) int {
	var getCodeFunc func(*html.Node)(int,error)
	var n *html.Node
	
	if i == 0 {
		getCodeFunc = laroussefr.GetPageID
		n = doc
	} else {
		getCodeFunc = getWordCodeFromZoneEntreeNode
		n = zoneEntreeNode
	}
	
	code, err := getCodeFunc(n)
	if err != nil {
		// if something goes wrong, use the previous word's code
		code = getWordCode(i-1, doc, zoneEntreeNode)
	}
	return code
}

// getWordCodeFromZoneEntreeNode takes a "ZoneEntree" node and returns the
// current word's code.
func getWordCodeFromZoneEntreeNode(n *html.Node) (int, error) {
	var str string
	
	m := n.PrevSibling
	if m == nil {
		m = n.Parent
		if m == nil {
			return -1, laroussefr.NewError("getWordCodeFromZoneEntreeNode", n.Data, "nil prev sibling and parent")
		}
		str = scrape.Attr(m, "link")
		if str == "" {
			return -1, laroussefr.NewError("getWordCodeFromZoneEntreeNode", n.Data, "failed to find <span> with \"link\" attr")
		}
		str = str[1:]
	} else {
		str = scrape.Attr(m, "id")
		if str == "" {
			return -1, laroussefr.NewError("getWordCodeFromZoneEntreeNode", n.Data, "Code wasn't found in " + m.Data + " node")
		}
	}
	
	out, err := strconv.Atoi(str)
	if err != nil {
		return -1, laroussefr.NewError("getWordCodeFromZoneEntreeNode", n.Data, "strconv.Atoi says " + err.Error())
	}
	
	return out, nil
}

// getPhraseFromZoneExpression takes a "ZoneExpression" or "ZoneExpression1"
// node and returns a Phrase.
func getPhraseFromZoneExpression(zoneExpressionNode *html.Node) Phrase {
	var p Phrase
	n := zoneExpressionNode.FirstChild
	for n != nil {
		p.update(n)
		if scrape.Attr(n, "class") == "DivisionExpression" {
			liNodes := scrape.FindAll(n, scrape.ByTag(atom.Li))
			for _, li := range liNodes {
				subphrase := getPhraseFromZoneExpression(li)
				p.Subphrases = append(p.Subphrases, subphrase)
			}
		}
		n = n.NextSibling
	}
	return p
}

// stillOnFirstMeaningStrings returns true if n is a node containing data
// relavent to the current Meaning, which also happens to be the first Meaning
// in this item.
func stillOnFirstMeaningStrings(n *html.Node) bool {
	if n == nil {
		return false
	} else if n.DataAtom == atom.Audio  || isWhitespace(n.Data) {
		return true
	} else if n.Type == html.TextNode && strings.ContainsRune(n.Data, '→') { // for "coup de fil" on fr->en coup
		return true
	}
		
	classes := []string{"Indicateur", "lienson2", "Traduction", "IndicateurDomaine", "Metalangue", "Indicateur2", "Renvois", "Glose2"}
	for _, c := range classes {
		if c == scrape.Attr(n, "class") {
			return true
		}
	}
	return false
}

// isWhiteSpace returns true if str consists entirely of whitespace runes.
func isWhitespace(str string) bool {
	for _, r := range []rune(str) {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

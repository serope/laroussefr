// larousse_test.go tests some common functions defined in laroussefr.go.
// 
// Most functions are tested in passing by the tests in packages definition and
// traduction.
package laroussefr

import (
	"fmt"
	"testing"
)

// TestIsURL tests IsURL on good and bad values.
func TestIsURL(t *testing.T) {
	cases := map[string]bool {
		"":false,
		" ":false,
		"asdfasdfsadfdasfaafsd":false,
		"https://fr.wikipedia.org":false,
		"ftp://larousse.fr/dictionnaires/francais/vert":false,
		"https://larousse.jp/dictionnaires/francais/vert":false,
		"http2://larousse.fr/dictionnaires/francais/rouge":false,
		
		"https://larousse.fr/dictionnaires/francais/bonjour":true,
		"http://www.larousse.fr/dictionnaires/francais/rose":true,
		"http://www.larousse.fr/dictionnaires/francais-anglais/ciel":true,
	}
	
	for k, v := range cases {
		fmt.Print(k, "\t")
		ok, message := IsURL(k)
		if ok != v {
			fmt.Println("FAIL - ", message);
		} else {
			fmt.Println("OK")
		}
	}
}

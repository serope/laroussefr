# laroussefr

Package laroussefr provides functions for web scraping [Larousse](https://www.larousse.fr).

### Example: Definition

```go
package main

import (
        "github.com/serope/laroussefr/definition"
        "fmt"
)

func main() {
        result, err := definition.New("déneiger")
        if err != nil {
                panic(err)
        }
        fmt.Println(result.Definitions[0])
        // "Débarrasser une surface de la neige qui la recouvre : Déneiger une route."
}
```

### Example: Translation

```go
package main

import (
        "github.com/serope/laroussefr/traduction"
        "fmt"
)

func main() {
        result, err := traduction.New("mountain", traduction.En, traduction.Fr)
        if err != nil {
                panic(err)
        }
        fmt.Println(result.Words)
        // print all Words defined on this page
}
```


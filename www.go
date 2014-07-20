// www.go

package www

import (
	"fmt"
	"github.com/davsk/davsknet/mandlebrot"
	"net/http"
)

var testIt int = mandlebrot.TestInt

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world! from davsk.net ", string(testIt))
}

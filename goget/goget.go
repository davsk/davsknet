// goget.go

package goget

import (
	"fmt"
	"net/http"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	resp := `
	<html>
  <head>
	<meta name="go-import" content="davsk.net/%s git https://github.com/davsk/%s.git">
	</head>
  </html>
`
	// Split the path to return a meta tag with only the first portion. So if
	// a request is received for /dae/user package (which is part of the `dae`
	// package), then we return a meta tag pointing to /dae.
	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Improper path for goget: "+r.URL.Path)
	} else {
		fmt.Fprintf(w, resp, segments[1], segments[1])
	}
}

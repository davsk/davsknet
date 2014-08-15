// page.go

package page

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

type Nav struct {
	Link        string
	Description string
}

type Data struct {
	Title       string
	Description string
	Author      string
	Content     string
	Navigation  []Nav
}

var PageTmpl = template.Must(template.ParseFiles("page.html"))

func (d *Data) Execute(w http.ResponseWriter, r *http.Request) {
	b := new(bytes.Buffer)

	if err := PageTmpl.Execute(b, d); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "tmpl.Execute failed: %v", err)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
	b.WriteTo(w)
}

type TextPlain []byte

func (t TextPlain) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write(t)
}

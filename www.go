// www.go

package www

import (
	"fmt"
	_ "mandelbrot"
	"net/http"
	"page"
	_ "wiki"
)

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/robots.txt", robotsHandler)
	http.HandleFunc("/sitemap.xml", sitemapHandler)
}

const WWW = "www.davsk.net"

func handler(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Host == WWW {
		var HomePage = page.Data{"Davsk.Net", "David Skinner Family Site", "David Skinner", "Hello World!", nil}
		HomePage.Execute(rw, req)
	} else {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, "<h1>www.davsk.net</h1><div>%s</div>", http.StatusText(http.StatusNotFound))
	}
}

func robotsHandler(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Host == WWW {
		http.ServeFile(rw, req, "page/robots.txt")
	} else {
		http.ServeFile(rw, req, "page/norobots.txt")
	}
}

func sitemapHandler(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Host == WWW {
		http.ServeFile(rw, req, "page/sitemap.xml")
	} else {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, "<h1>www.davsk.net</h1><div>%s</div>", http.StatusText(http.StatusNotFound))
	}
}

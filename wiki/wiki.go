// wiki.go

package wiki

import (
	"appengine"
	appengineuser "appengine/user"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type Article struct {
	Name        string
	Subdomain   string
	Description string
	Author      string
	Address     string
	Created     time.Time
	Modified    time.Time
	PreteenFlag bool
	IndexFlag   bool
	FollowFlag  bool
	Title       string
	Heading     string
	Subheading  string
	Nav         []byte
	Content     []byte
	Extra       []byte
}

var ArticleTmpl = template.Must(template.ParseFiles("wiki/article.html"))

func init() {
	http.HandleFunc("/view/", viewHandler)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	//var subdomain string

	if r.URL.Host == "www.davsk.net" {
		//subdomain = "www"
	} else {
		u := "http://www.davsk.net" + r.URL.Path
		http.Redirect(w, r, u, http.StatusMovedPermanently)
		return
	}

	name := r.URL.Path[len("/view/"):]

	//Creates an App Engine Context - required to access App Engine services.
	c := appengine.NewContext(r)

	//Acquire the current user
	user := appengineuser.Current(c)
	/*if user == nil {
	      logUrl, _ := appengineuser.LoginURL(c, "/")
	      logLabel := "LogIn"
	  } else {
	      logUrl, _ := appengineuser.LogoutURL(c, "/")
	      logLabel = "LogOut"
	  }
	*/
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	if user == nil {
		url, _ := appengineuser.LoginURL(c, r.URL.Path)
		fmt.Fprintf(w, `<h1>%s</h1><a href="%s">Sign in</a>`, name, url)
	} else {
		url, _ := appengineuser.LogoutURL(c, r.URL.Path)
		fmt.Fprintf(w, `<h1>%s</h1>Welcome, %s! (<a href="%s">sign out</a>)`, name, user, url)
	}
	//article, _ := loadPage(name)
	//fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", article.Title, article.Content)
}

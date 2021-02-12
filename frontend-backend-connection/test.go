package main

// cd test-app
// npm install
// npm run build
// cd ..
// go run test.go

import (
	"net/http"
	"fmt"
	"strings"
)

type indexHtmlDefaultFs struct {
	http.FileSystem
}

func (fs indexHtmlDefaultFs) Open(name string) (http.File, error) {
	file, err := fs.FileSystem.Open(name)
	if err != nil {
		if name == "index.html" {
			return nil, err
		} else {
			return fs.Open("index.html")
		}
	}
	return file, err
}

// parse url of the form: "/ignore/a=b/x=sdaff/jskdfsdlkj=sdijs/"
// tolerates redundant "/"s
// returns nil if fails
func parseUrl(ignore string, urlPath string) map[string]string {
	urlMap := make(map[string]string)
	for _, s := range strings.Split(urlPath,"/") {
		if s == "" || s == ignore {
			continue
		}
		spl := strings.Split(s, "=")
		if len(spl) != 2 {
			return nil
		}
		urlMap[spl[0]] = spl[1]
	}
	return urlMap
}

// ASSUMPTION: database fields don't have "/"s or "="s in them
// ANOTHER ASSUMPTION: sending db query over http doesnt compromise privacy
func dbHandler(w http.ResponseWriter, r *http.Request) {
	urlMap := parseUrl("db", r.URL.Path)
	if urlMap == nil {
		return
	}
	// do some fancy database stuff
	fmt.Println("eventually will query database for following key value pairs:", urlMap)
	fmt.Fprintf(w,"[]")
}

// dont know exactly how this is gonna work
func calHandler(w http.ResponseWriter, r *http.Request) {
	urlMap := parseUrl("cal", r.URL.Path)
	if urlMap == nil {
		return
	}
	// do some fancy calendar stuff
	fmt.Println("eventually will look at calendar for following key value pairs:", urlMap)
	fmt.Fprintf(w,"[]")
}

func main() {
	http.HandleFunc("/db/", dbHandler)
	http.HandleFunc("/cal/", calHandler)
	fs := indexHtmlDefaultFs{http.Dir("test-app/build")}
	http.Handle("/", http.FileServer(fs))
	http.ListenAndServe(":8080", nil)
}

package main

// git clone https://github.com/changuo0/foodmap-frontend
// cd foodmap-frontend/mayor-app
// npm install
// npm run build
// mv build ../../website
// cd ../..
// gcloud app deploy

import (
	"net/http"
	"fmt"
	"strings"
	"os"
	"log"
	"bytes"

	"encoding/json"
	"io/ioutil"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
        // The file token.json stores the user's access and refresh tokens, and is
        // created automatically when the authorization flow completes for the first
        // time.
        tokFile := "token.json"
        tok, err := tokenFromFile(tokFile)
        if err != nil {
                tok = getTokenFromWeb(config)
                saveToken(tokFile, tok)
        }
        return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
        authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
        fmt.Printf("Go to the following link in your browser then type the "+
                "authorization code: \n%v\n", authURL)

        var authCode string
        if _, err := fmt.Scan(&authCode); err != nil {
                log.Fatalf("Unable to read authorization code: %v", err)
        }

        tok, err := config.Exchange(context.TODO(), authCode)
        if err != nil {
                log.Fatalf("Unable to retrieve token from web: %v", err)
        }
        return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
        f, err := os.Open(file)
        if err != nil {
                return nil, err
        }
        defer f.Close()
        tok := &oauth2.Token{}
        err = json.NewDecoder(f).Decode(tok)
        return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
        fmt.Printf("Saving credential file to: %s\n", path)
        f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
        if err != nil {
                log.Fatalf("Unable to cache oauth token: %v", err)
        }
        defer f.Close()
        json.NewEncoder(f).Encode(token)
}



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

var sheetSrv *sheets.Service

func loadSpreadsheet() *sheets.Service {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return srv
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Panicf("%s environment variable not set.", k)
	}
	return v
}

// ASSUMPTION: database fields don't have "/"s or "="s in them
func sheetHandler(w http.ResponseWriter, r *http.Request) {
	srv := sheetSrv

	urlMap := parseUrl("sheet", r.URL.Path)
	if urlMap == nil {
		return
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	spreadsheetId := "1lLlkGGKVapJ1Ot0QDoXKM-24vJWmOnA_BS5Rro17BA0"
	readRange := "Sheet1!A1:B"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	w.Header().Set("Content-Type", "text/plain")
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "[")
	first := true
	for _, row := range resp.Values {
		if !first {
			fmt.Fprintf(buf,",")
		}
		fmt.Fprintf(buf, "{\"firstThing\":\"%s\", \"secondThing\":\"%s\"}\n", row[0],row[1])
		first = false
	}
	fmt.Fprintf(buf, "]")
	w.Write(buf.Bytes())
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
	sheetSrv = loadSpreadsheet()

	http.HandleFunc("/sheet/", sheetHandler)
	http.HandleFunc("/cal/", calHandler)
	fs := indexHtmlDefaultFs{http.Dir("website")}
	http.Handle("/", http.FileServer(fs))
	http.ListenAndServe(":8080", nil)
}

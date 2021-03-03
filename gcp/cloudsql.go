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
	"os"
	"log"
	"bytes"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
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

var db *sql.DB

// DB gets a connection to the database.
// This can panic for malformed database connection strings, invalid credentials, or non-existance database instance.
func DB() *sql.DB {
	var (
		connectionName = mustGetenv("CLOUDSQL_CONNECTION_NAME")
		user           = mustGetenv("CLOUDSQL_USER")
		dbName         = os.Getenv("CLOUDSQL_DATABASE_NAME") // NOTE: dbName may be empty
		password       = os.Getenv("CLOUDSQL_PASSWORD")      // NOTE: password may be empty
		socket         = os.Getenv("CLOUDSQL_SOCKET_PREFIX")
	)
	// /cloudsql is used on App Engine.
	if socket == "" {
		socket = "/cloudsql"
	}
	// MySQL Connection, comment out to use PostgreSQL.
	// connection string format: USER:PASSWORD@unix(/cloudsql/PROJECT_ID:REGION_ID:INSTANCE_ID)/[DB_NAME]
	dbURI := fmt.Sprintf("%s:%s@unix(%s/%s)/%s", user, password, socket, connectionName, dbName)
	conn, err := sql.Open("mysql", dbURI)
	if err != nil {
		panic(fmt.Sprintf("DB: %v", err))
	}
	return conn
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Panicf("%s environment variable not set.", k)
	}
	return v
}

// ASSUMPTION: database fields don't have "/"s or "="s in them
// ANOTHER ASSUMPTION: sending db query over http doesnt compromise privacy
func dbHandler(w http.ResponseWriter, r *http.Request) {
	urlMap := parseUrl("db", r.URL.Path)
	if urlMap == nil {
		return
	}

	query := "SELECT * FROM organizations"
	if len(urlMap) > 0 {
		query += " WHERE "
		for k,v := range urlMap {
			query += k + " = " + v + " AND "
		}
		query = query[:len(query) - len(" AND ")]
	}
	query += ";"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Could not query db: %v", err)
		http.Error(w, "error at location 1: " + err.Error(), 500)
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/plain")
	buf := bytes.NewBufferString("")
	for rows.Next() {
		var name string
		var contact string
		var notes string
		var location string
		var zip int
		var foodtype string
		var id int
		if err := rows.Scan(&name,&contact,&notes,&location,&zip,&foodtype,&id); err != nil {
			log.Printf("Could not scan result: %v", err)
			http.Error(w, "error at location 2: " + err.Error(), 500)
			return
		}
		fmt.Fprintf(buf, "%s, %s, %s, %s, %d, %s, %d\n", name, contact, notes, location, zip, foodtype, id)
	}
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
	db = DB()
	db.Exec("USE foodmap;")

	http.HandleFunc("/db/", dbHandler)
	http.HandleFunc("/cal/", calHandler)
	fs := indexHtmlDefaultFs{http.Dir("foodmap-frontend/build")}
	http.Handle("/", http.FileServer(fs))
	http.ListenAndServe(":8080", nil)
}

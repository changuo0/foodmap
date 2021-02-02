// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START gae_cloudsql]

// Sample cloudsql demonstrates connection to a Cloud SQL instance from App Engine standard.
package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	// MySQL library, comment out to use PostgreSQL.
	_ "github.com/go-sql-driver/mysql"
	// PostgreSQL Library, uncomment to use.
	// _ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	db = DB()

	http.HandleFunc("/", indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

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

	// PostgreSQL Connection, uncomment to use.
	// connection string format: user=USER password=PASSWORD host=/cloudsql/PROJECT_ID:REGION_ID:INSTANCE_ID/[ dbname=DB_NAME]
	// dbURI := fmt.Sprintf("user=%s password=%s host=/cloudsql/%s dbname=%s", user, password, connectionName, dbName)
	// conn, err := sql.Open("postgres", dbURI)

	if err != nil {
		panic(fmt.Sprintf("DB: %v", err))
	}

	return conn
}

// indexHandler responds to requests with our list of available databases.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	db.Exec("USE foodmap;")
	rows, err := db.Query("SELECT * FROM organizations;")
	if err != nil {
		log.Printf("Could not query db: %v", err)
		http.Error(w, "error at location 1: " + err.Error(), 500)
		return
	}
	defer rows.Close()

	buf := bytes.NewBufferString("data:\n")
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
		fmt.Fprintf(buf, "- %s, %s, %s, %s, %d, %s, %d\n", name, contact, notes, location, zip, foodtype, id)
	}
	w.Write(buf.Bytes())
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Panicf("%s environment variable not set.", k)
	}
	return v
}

// [END gae_cloudsql]

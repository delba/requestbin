package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"text/template"

	_ "github.com/lib/pq"

	"github.com/delba/requestbin/model"
	"github.com/jinzhu/gorm"
)

var db gorm.DB

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	var requests []model.Request
	db.Find(&requests)

	layout := path.Join("templates", "layouts", "application.html")
	index := path.Join("templates", "bins", "index.html")
	t, err := template.ParseFiles(layout, index)

	handle(err)

	err = t.Execute(w, requests)
	handle(err)
}

func create(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	handle(err)
	defer r.Body.Close()

	request := model.Request{Body: body}
	db.Create(&request)
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		index(w, r)
	case "POST":
		create(w, r)
	default:
		index(w, r)
	}
}

func init() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "dbname=requestbin sslmode=disable"
	}

	db = func() gorm.DB {
		db, err := gorm.Open("postgres", url)
		handle(err)
		return db
	}()

	db.CreateTable(&model.Request{})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}

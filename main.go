package main

import (
	"io/ioutil"
	"net/http"
	"path"
	"text/template"

	"github.com/delba/requestbin/model"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
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

	t, err := template.ParseFiles(path.Join("templates", "index.html"))
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
	db = func() gorm.DB {
		db, err := gorm.Open("postgres", "dbname=requestbin sslmode=disable")
		handle(err)
		return db
	}()
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

package main

import (
	"io/ioutil"
	"net/http"
	"path"
	"text/template"

	"github.com/delba/requestbin/db"
	"github.com/delba/requestbin/model"
	_ "github.com/lib/pq"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	requests, err := db.GetAllRequests()
	handle(err)

	templatePath := path.Join("templates", "index.html")
	t, err := template.ParseFiles(templatePath)
	handle(err)

	err = t.Execute(w, requests)
	handle(err)
}

func create(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	handle(err)
	defer r.Body.Close()

	err = db.CreateRequest(model.Request{Body: body})
	handle(err)
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

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

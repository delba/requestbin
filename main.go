package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"

	"github.com/delba/requestbin/model"
	"github.com/jinzhu/gorm"
)

var db gorm.DB
var store = sessions.NewCookieStore([]byte("something-very-secret"))

func handle(err error) {
	if err != nil {
		panic(err)
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

	db.DropTable(&model.Bin{})
	db.DropTable(&model.Request{})
	db.CreateTable(&model.Bin{})
	db.CreateTable(&model.Request{})

	db.LogMode(true)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := mux.NewRouter()

	r.HandleFunc("/", BinsIndex).Methods("GET")
	r.HandleFunc("/favicon.ico", ServeFileHandler).Methods("GET")
	r.HandleFunc("/bins", BinsCreate).Methods("POST")
	r.HandleFunc("/{token}", BinsShow).Methods("GET")
	r.HandleFunc("/{token}", RequestsCreate).Methods("POST")

	http.Handle("/", r)
	http.ListenAndServe(":"+port, nil)
}

func ServeFileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serve static file")
}

func BinsIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Println("bins#index")

	tokens := getTokens(r)

	layout := path.Join("templates", "layouts", "application.html")
	index := path.Join("templates", "bins", "index.html")
	t, err := template.ParseFiles(layout, index)
	handle(err)

	err = t.Execute(w, tokens)
	handle(err)
}

func BinsCreate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("bins#create")

	var bin model.Bin
	db.Create(&bin)

	addToken(bin.Token, w, r)

	http.Redirect(w, r, "/"+bin.Token, 302)
}

func BinsShow(w http.ResponseWriter, r *http.Request) {
	fmt.Println("bins#show")

	token := mux.Vars(r)["token"]
	fmt.Println(token)
	bin := findBin(r)

	var requests []model.Request
	db.Model(&bin).Related(&requests)

	layout := path.Join("templates", "layouts", "application.html")
	show := path.Join("templates", "bins", "show.html")
	t, err := template.ParseFiles(layout, show)
	handle(err)

	type Data struct {
		Bin      model.Bin
		Requests []model.Request
	}

	data := Data{Bin: bin, Requests: requests}

	err = t.Execute(w, data)
	handle(err)
}

func RequestsCreate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("requests#create")

	bin := findBin(r)
	fmt.Println(bin)

	body, err := ioutil.ReadAll(r.Body)
	handle(err)
	defer r.Body.Close()

	fmt.Println("Bin id is", bin.ID)
	request := model.Request{Body: body, BinID: bin.ID}
	db.Create(&request)

	http.Redirect(w, r, "/"+bin.Token, 302)
}

func findBin(r *http.Request) model.Bin {
	token := mux.Vars(r)["token"]

	var bin model.Bin
	db.Where(&model.Bin{Token: token}).First(&bin)

	return bin
}

func getTokens(r *http.Request) []string {
	sessions, _ := store.Get(r, "session-name")

	tokens := sessions.Values["tokens"]

	if tokens == nil {
		tokens = []string{}
	}

	return tokens.([]string)
}

func addToken(token string, w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	tokens := append(getTokens(r), token)

	session.Values["tokens"] = tokens

	session.Save(r, w)
}

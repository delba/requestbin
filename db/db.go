package db

import (
	"database/sql"

	"github.com/delba/requestbin/model"
	_ "github.com/lib/pq"
)

var db *sql.DB

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	db = func() *sql.DB {
		db, err := sql.Open("postgres", "dbname=requestbin sslmode=disable")
		handle(err)
		return db
	}()
}

func GetAllRequests() ([]model.Request, error) {
	var requests []model.Request
	var err error

	rows, err := db.Query(`SELECT * FROM requests`)
	if err != nil {
		return requests, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var body []byte

		err = rows.Scan(&id, &body)
		if err != nil {
			return requests, err
		}

		request := model.Request{
			ID:   id,
			Body: body,
		}

		requests = append(requests, request)
	}

	return requests, err
}

func CreateRequest(r model.Request) error {
	var err error

	rows, err := db.Query(`INSERT INTO requests (body) VALUES ($1)`, r.Body)
	defer rows.Close()

	return err
}

func DeleteAll() error {
	_, err := db.Query(`TRUNCATE FROM requests`)

	return err
}

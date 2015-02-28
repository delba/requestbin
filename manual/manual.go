package manual

import (
	"io"
	"net/http"
	"os"
)

const url = "http://localhost:8080"

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var dataUrl string

	if len(os.Args) > 1 {
		dataUrl = os.Args[1]
	} else {
		dataUrl = "https://github.com/flori/json/raw/master/data/example.json"
	}

	r, err := getReader(dataUrl)
	handle(err)

	_, err = http.DefaultClient.Post(url, "application/json", r)
	handle(err)
}

func getReader(url string) (io.Reader, error) {
	res, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Request struct {
	ID   int
	Body []byte
}

func (r *Request) FormattedBody() string {
	var buf bytes.Buffer

	if err := json.Indent(&buf, r.Body, "", "  "); err != nil {
		fmt.Println(err)
	}

	return buf.String()
}

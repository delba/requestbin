package model

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

type Bin struct {
	ID       int
	Token    string `sql:"unique"`
	Requests []Request
}

func (b *Bin) BeforeCreate() (err error) {
	// TODO loop to make sure that the token is unique

	buf := make([]byte, 4)

	if _, err = io.ReadFull(rand.Reader, buf); err != nil {
		err = err
	}

	b.Token = hex.EncodeToString(buf)

	return err
}

func (b *Bin) GenerateToken() string {
	buf := make([]byte, 4)

	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		fmt.Println(err)
	}

	return hex.EncodeToString(buf)
}

type Request struct {
	ID    int
	Body  []byte
	BinID int
}

func (r *Request) FormattedBody() string {
	var buf bytes.Buffer

	if err := json.Indent(&buf, r.Body, "", "  "); err != nil {
		fmt.Println(err)
	}

	return buf.String()
}

package storage

import (
	"bytes"

	"golang.org/x/net/html"
)

func htmlToText(r []byte) string {
	b := bytes.Buffer{}
	z := html.NewTokenizer(bytes.NewReader(r))

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		if tt == html.TextToken {
			b.Write(z.Text())
			b.WriteString(" ")
		}
	}

	return b.String()
}

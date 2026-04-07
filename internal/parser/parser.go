package parser

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/Pix3l-Guy21/search-engine/internal/pipeline"
)

var wordRegex = regexp.MustCompile(`[a-z0-9]+`)

// parser.go
func Parser(done <-chan struct{}, paths <-chan string) <-chan pipeline.Document {
	docs := make(chan pipeline.Document, 10)
	go func() {
		defer close(docs)
		for path := range paths {
			fs, err := os.Stat(path)
			if err != nil || fs.Size() > 50*1024*1024 {
				continue
			}

			// Just send the metadata, NOT the extracted text/tokens yet
			docs <- pipeline.Document{
				Path:  path,
				Title: fs.Name(),
				Size:  fs.Size(),
			}
		}
	}()
	return docs
}

func ExtractTextFromPdf(path string) (text string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pdf library panicked on file %s: %v", path, r)
		}
	}()
	_, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	fonts := make(map[string]*pdf.Font)
	numPages := r.NumPage()
	for i := 1; i <= numPages; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		plainText, err := p.GetPlainText(fonts)
		if err != nil {
			continue
		}
		b.WriteString(plainText)
	}
	return b.String(), nil
}

func Tokenize(text string) []string {
	text = strings.ToLower(text)
	words := wordRegex.FindAllString(text, -1)

	encountered := make(map[string]bool)
	uniqueTokens := make([]string, 0, len(words)/2)
	for _, w := range words {
		if len(w) > 2 && !encountered[w] {
			encountered[w] = true
			uniqueTokens = append(uniqueTokens, w)
		}
	}
	return uniqueTokens
}

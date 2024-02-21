package domain

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"time"

	"github.com/go-shiori/go-readability"
	"github.com/monzo/slog"
)

type Article struct {
	ID                string
	Title             string
	Description       string
	CompressedContent []byte
	Content           readability.Article
	ImageURL          string
	Link              string
	Author            string
	SourceID          int64
	Source			  Source
	Timestamp         time.Time
	TS                string
	LayoutID		  int64
	Layout            Layout

	decompressed []byte
}

type Element struct {
	Type  string
	Value string
}

func (a *Article) Size() int {
	n := len(a.Content.TextContent)
	return n
}

func (a *Article) Trim(size int) {
	if len(a.Content.TextContent) > size {
		a.Content.TextContent = a.Content.TextContent[:size] + "..."
	}
}

func (a *Article) RawHTML() template.HTML {
	if len(a.decompressed) != 0 {
		return template.HTML(a.decompressed)
	}
	slog.Debug(context.Background(), "Decompressing %s", a.ID)
	r, _ := gzip.NewReader(bytes.NewReader(a.CompressedContent))
	buf, _ := ioutil.ReadAll(r)
	a.decompressed = buf
	return template.HTML(buf)
}

func (article *Article) SetHTMLContent(body string) []byte {
    buf := new(bytes.Buffer)
    w := gzip.NewWriter(buf)
    w.Write([]byte(body))
    w.Flush()
    w.Close()
    article.CompressedContent = buf.Bytes()
    return article.CompressedContent
}


type JSONArticle struct {
	Title         string
	Byline        string
	Content       string
	TextContent   string
	Length        int
	Excerpt       string
	SiteName      string
	Image         string
	Favicon       string
	Language      string
	PublishedTime *time.Time
	ModifiedTime  *time.Time
}

func ConvertToJSONArticle(article readability.Article) JSONArticle {
	return JSONArticle{
		Title:         article.Title,
		Byline:        article.Byline,
		Content:       article.Content,
		TextContent:   article.TextContent,
		Length:        article.Length,
		Excerpt:       article.Excerpt,
		SiteName:      article.SiteName,
		Image:         article.Image,
		Favicon:       article.Favicon,
		Language:      article.Language,
		PublishedTime: article.PublishedTime,
		ModifiedTime:  article.ModifiedTime,
	}
}


func CompressContent(article readability.Article) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)

	encoder := json.NewEncoder(w)
	err := encoder.Encode(ConvertToJSONArticle(article))
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ConvertToReadabilityArticle(jsonArticle JSONArticle) readability.Article {
	return readability.Article{
		Title:         jsonArticle.Title,
		Byline:        jsonArticle.Byline,
		Content:       jsonArticle.Content,
		TextContent:   jsonArticle.TextContent,
		Length:        jsonArticle.Length,
		Excerpt:       jsonArticle.Excerpt,
		SiteName:      jsonArticle.SiteName,
		Image:         jsonArticle.Image,
		Favicon:       jsonArticle.Favicon,
		Language:      jsonArticle.Language,
		PublishedTime: jsonArticle.PublishedTime,
		ModifiedTime:  jsonArticle.ModifiedTime,
	}
}

func DecompressContent(data []byte) (readability.Article, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return readability.Article{}, err
	}
	defer r.Close()

	decoder := json.NewDecoder(r)
	var jsonArticle JSONArticle
	err = decoder.Decode(&jsonArticle)
	if err != nil {
		return readability.Article{}, err
	}

	// Convert JSONArticle back to readability.Article
	article := ConvertToReadabilityArticle(jsonArticle)

	return article, nil
}
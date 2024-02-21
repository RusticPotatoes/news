package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/RusticPotatoes/news/dao"
	"github.com/monzo/slog"
)

func handleRefreshArticle(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		id  = r.URL.Query().Get("id")
	)

	a, err := dao.GetArticle(ctx, id)
	if err != nil {
		slog.Error(ctx, "Error decoding JSON: %s", err)
		return
	}

	slogParams := map[string]string{
		"url": a.Link,
	}
	res, err := c.Get(fmt.Sprintf("https://readability-server.russellsaw.io/?url=%s", url.QueryEscape(a.Link)))
	if err != nil {
		slog.Error(ctx, "Error fetching article: %s", err, slogParams)
		return
	}
	var article = struct {
		Body     string `json:"body"`
		BodyText string `json:"body_text"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&article)
	if err != nil {
		slog.Error(ctx, "Error fetching article: %s", err, slogParams)
		return
	}

	// a.Content = toElements(article.BodyText, "\n")
	a.SetHTMLContent(article.Body)

	// sa, err := swan.FromHTML(a.Link, []byte(article.Body))
	// if err != nil {
	// 	slog.Error(ctx, "Error parsing article: %s", err, slogParams)
	// 	return
	// }
	// if sa.Img != nil {
	// 	a.ImageURL = sa.Img.Src
	// }
	err = dao.SetArticle(ctx, a)
	if err != nil {
		slog.Error(ctx, "Error storing article: %s", err, slogParams)
		return
	}
	slog.Info(ctx, "Stored article: %s - %s", a.ID, a.Title)
	http.Redirect(w, r, fmt.Sprintf("/article?id=%s", a.ID), 307)
}

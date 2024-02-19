package handler

import (
	"net/http"
	"strings"

	"github.com/RusticPotatoes/news/dao"
	"github.com/RusticPotatoes/news/domain"
	"github.com/monzo/slog"
)

func handleAddSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseForm()

	u := domain.UserFromContext(ctx)
	if u == nil {
		http.Error(w, "not logged in", 400)
		return
	}
	name := r.Form.Get("name")
	homepage := r.Form.Get("homepage")
	feedURL := r.Form.Get("feed_url")
	categories := r.Form.Get("categories")

	src := domain.Source{
		// ID:         idgen.New("src"),
		OwnerID:    u.ID,
		Name:       name,
		URL:        homepage,
		FeedURL:    feedURL,
		Categories: strings.Split(categories, ","),
	}

	err := dao.SetSource(ctx, &src)
	if err != nil {
		slog.Error(ctx, "Error storing source: %s", err)
		http.Error(w, "error storing source", 500)
	}
	http.Redirect(w, r, "/settings", 307)
}

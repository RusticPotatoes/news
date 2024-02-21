package handler

import (
	"html/template"
	"net/http"

	"github.com/RusticPotatoes/news/dao"
	"github.com/RusticPotatoes/news/domain"
	"github.com/monzo/slog"
)

type settingsPage struct {
	Sources []domain.Source
	base
}

type base struct {
	User       *domain.User
	Error      string
	Categories []string
	ID         string
	Name       string
	Title      string
	Meta       Meta
}

type Meta struct {
	Title       string
	Description string
	Image       string
	URL         string
}

func handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	t := template.New("frame.html")
	t, err := t.ParseFiles("tmpl/frame.html", "tmpl/meta.html", "tmpl/settings.html")
	if err != nil {
		slog.Error(ctx, "Error parsing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	u := domain.UserFromContext(ctx)
	if u == nil {
		http.Error(w, "Not logged in", 400)
		return
	}
	sources, err := dao.GetSources(ctx, u.Name)
	if err != nil {
		http.Error(w, "Couldn't get sources", 500)
		return
	}

	s := settingsPage{
		Sources: sources,
		base: base{
			ID:   "Settings",
			User: u,
		},
	}

	err = t.Execute(w, &s)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

package handler

import (
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/RusticPotatoes/news/dao"
	"github.com/RusticPotatoes/news/domain"
	"github.com/monzo/slog"
)

func handleArticle(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    t := template.New("frame.html").Funcs(template.FuncMap{
        "safeHTML": safeHTML,
    })
    t, err := t.ParseFiles("tmpl/frame.html", "tmpl/meta.html", "tmpl/article.html")
    if err != nil {
        slog.Error(ctx, "Error parsing template: %s", err)
        http.Error(w, err.Error(), 500)
        return
    }

    article, err := dao.GetArticle(ctx, r.URL.Query().Get("id"))
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Uncompress the content
    // uncompressedContend, err := domain.DecompressContent(article.Content)
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// 	return
	// }

    // Set the uncompressed content
    // article.Content = uncompressedContend
	// article.SetHTMLContent(uncompressedContend)

	// u := domain.UserFromContext(ctx)
	var sources []domain.Source
	// if u != nil {
	sources, err = dao.GetSources(ctx, "admin")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// } else {
	// 	sources = domain.GetSources()
	// }
	for _, src := range sources {
		if src.FeedURL == article.Source.FeedURL {
			article.Source = src
		}
	}

	a := articlePage{
		Article: article,
		base: base{
			User: domain.UserFromContext(ctx),
			Meta: Meta{
				Title:       article.Title + " - " + article.Source.Name,
				Description: preview([]domain.Element{{Type: "text", Value: article.Content.TextContent}}),
				Image:       article.ImageURL,
				URL:         r.URL.String(),
			},
		},
	}

	// Uncompress the content and add it to the Content field
	uncompressedContent, err:=  domain.DecompressContent(article.CompressedContent)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	a.Article.Content = uncompressedContent

	byFeedURL := make(map[string]domain.Source)
	smap := make(map[string]struct{})
	for _, s := range sources {
		byFeedURL[s.FeedURL] = s
		for _, cat := range s.Categories {
			smap[cat] = struct{}{}
		}
	}
	for cat := range smap {
		a.Categories = append(a.Categories, cat)
	}
	sort.Strings(a.Categories)

	err = t.Execute(w, a)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

type articlePage struct {
	Article *domain.Article
	base
}

func preview(es []domain.Element) string {
	var out string
	for _, e := range es {
		if e.Type != "text" {
			continue
		}
		if strings.TrimSpace(e.Value) == "" {
			continue
		}
		out += e.Value
	}
	if len(out) <= 400 {
		return out
	}
	return out[:400]
}

func safeHTML(text string) template.HTML {
    return template.HTML(text)
}
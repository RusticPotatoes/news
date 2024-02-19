package handler

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/RusticPotatoes/news/dao"
	"github.com/RusticPotatoes/news/domain"
)

type result struct {
    Article domain.Article `json:"article"`
    // HitText string        `json:"hitText"`
}

type searchPage struct {
	Results []result
	Query   string
	url     string
}

func handleSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
    ctx := r.Context()
    query := r.URL.Query().Get("q")

	searchResults, err := dao.SearchInCache(ctx, query)
	if err != nil {
		return nil, err
	}
	
	p := searchPage{
		Query: query,
		url:   r.URL.String(),
	}
	for _, article := range searchResults {
		p.Results = append(p.Results, result{
			Article: article,
		})
	}

	sort.Slice(p.Results, func(i, j int) bool {
		return p.Results[i].Article.Timestamp.After(p.Results[j].Article.Timestamp)
	})
	return p, nil
}

func (p searchPage) Meta() Meta {
	return Meta{
		Title: fmt.Sprintf("Search for %s on The Webpage", p.Query),
		URL:   p.url,
		Image: "/static/images/preview.png",
	}
}


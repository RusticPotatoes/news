package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/RusticPotatoes/news/domain"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// import "github.com/elastic/go-elasticsearch/v8"

var (
	client       *elasticsearch.Client
	mu           sync.RWMutex
	articleCache = make(map[string]*domain.Article)
)

func Init(ctx context.Context) error {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}

	client = es

	return nil
}

func Client() *elasticsearch.Client {
	return client
}

func GetArticle(ctx context.Context, id string) (*domain.Article, error) {
	mu.RLock()
	article, ok := articleCache[id]
	if ok {
		mu.RUnlock()
		return &article, nil
	}
	mu.RUnlock()

	req := esapi.GetRequest{
		Index:      "articles",
		DocumentID: id,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var a domain.Article
	if err := json.NewDecoder(res.Body).Decode(&a); err != nil {
		return nil, err
	}

	mu.Lock()
	articleCache[a.ID] = a
	mu.Unlock()

	return &a, nil
}

func SetArticle(ctx context.Context, a *domain.Article) error {
	req := esapi.IndexRequest{
		Index:      "articles",
		DocumentID: a.ID,
		Body:       strings.NewReader(a.Body),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New(res.String())
	}

	return nil
}

func GetSource(ctx context.Context, id string) (*domain.Source, error) {
	req := esapi.GetRequest{
		Index:      "sources",
		DocumentID: id,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var s domain.Source
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

func SetSource(ctx context.Context, s *domain.Source) error {
	req := esapi.IndexRequest{
		Index:      "sources",
		DocumentID: s.ID,
		Body:       strings.NewReader(s.Body),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New(res.String())
	}

	return nil
}

func DeleteSource(ctx context.Context, id string) error {
	req := esapi.DeleteRequest{
		Index:      "sources",
		DocumentID: id,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New(res.String())
	}

	return nil
}

func GetSources(ctx context.Context) ([]*domain.Source, error) {
	req := esapi.SearchRequest{
		Index: []string{"sources"},
		Body:  strings.NewReader(`{"query": {"match_all": {}}}`),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
	sources := make([]*domain.Source, len(hits))
	for i, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		sources[i] = &domain.Source{
			ID:   source["id"].(string),
			Body: source["body"].(string),
		}
	}

	return sources, nil
}

func GetAllSources(ctx context.Context) ([]*domain.Source, error) {
	req := esapi.SearchRequest{
		Index: []string{"sources"},
		Body:  strings.NewReader(`{"query": {"match_all": {}}}`),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
	sources := make([]*domain.Source, len(hits))
	for i, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		sources[i] = &domain.Source{
			ID:   source["id"].(string),
			Body: source["body"].(string),
		}
	}

	return sources, nil
}

func GetArticlesForOwner(ctx context.Context, owner string) ([]*domain.Article, error) {
	req := esapi.SearchRequest{
		Index: []string{"articles"},
		Body:  strings.NewReader(fmt.Sprintf(`{"query": {"match": {"owner": "%s"}}}`, owner)),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
	articles := make([]*domain.Article, len(hits))
	for i, hit := range hits {
		article := hit.(map[string]interface{})["_source"].(map[string]interface{})
		articles[i] = &domain.Article{
			ID:    article["id"].(string),
			Title: article["title"].(string),
			// Add other fields as necessary.
		}
	}

	return articles, nil
}

func GetEditionForTime(ctx context.Context, t time.Time) (*domain.Edition, error) {
	req := esapi.SearchRequest{
		Index: []string{"editions"},
		Body:  strings.NewReader(fmt.Sprintf(`{"query": {"range": {"time": {"gte": "%s"}}}}`, t.Format(time.RFC3339))),
		Sort:  []string{"time:asc"},
		Size:  1,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
	if len(hits) == 0 {
		return nil, nil
	}

	edition := hits[0].(map[string]interface{})["_source"].(map[string]interface{})
	return &domain.Edition{
		ID: edition["id"].(string),
		// Add other fields as necessary.
	}, nil
}

func SetEdition(ctx context.Context, e *domain.Edition) error {
	req := esapi.IndexRequest{
		Index:      "editions",
		DocumentID: e.ID,
		Body:       strings.NewReader(e.Body),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New(res.String())
	}

	return nil
}

func GetEdition(ctx context.Context, id string) (*domain.Edition, error) {
	req := esapi.GetRequest{
		Index:      "editions",
		DocumentID: id,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var e domain.Edition
	if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
		return nil, err
	}

	return &e, nil
}

func GetArticleByURL(ctx context.Context, url string) (*domain.Article, error) {
	req := esapi.SearchRequest{
		Index: []string{"articles"},
		Body:  strings.NewReader(fmt.Sprintf(`{"query": {"match": {"url": "%s"}}}`, url)),
		Size:  1,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
	if len(hits) == 0 {
		return nil, nil
	}

	article := hits[0].(map[string]interface{})["_source"].(map[string]interface{})
	return &domain.Article{
		ID:    article["id"].(string),
		Title: article["title"].(string),
		// Add other fields as necessary.
	}, nil
}

func GetArticlesByTime(ctx context.Context, t time.Time) ([]*domain.Article, error) {
	req := esapi.SearchRequest{
		Index: []string{"articles"},
		Body:  strings.NewReader(fmt.Sprintf(`{"query": {"range": {"time": {"gte": "%s"}}}}`, t.Format(time.RFC3339))),
		Sort:  []string{"time:asc"},
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
	articles := make([]*domain.Article, len(hits))
	for i, hit := range hits {
		article := hit.(map[string]interface{})["_source"].(map[string]interface{})
		articles[i] = &domain.Article{
			ID:    article["id"].(string),
			Title: article["title"].(string),
			// Add other fields as necessary.
		}
	}

	return articles, nil
}

func GetUser(ctx context.Context, id string) (*domain.User, error) {
	req := esapi.GetRequest{
		Index:      "users",
		DocumentID: id,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var u domain.User
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return nil, err
	}

	return &u, nil
}

func SetUser(ctx context.Context, u *domain.User) error {
	userJSON, err := json.Marshal(u)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      "users",
		DocumentID: u.ID,
		Body:       strings.NewReader(string(userJSON)),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New(res.String())
	}

	return nil
}

func GetUserByName(ctx context.Context, name string) (*domain.User, error) {
	req := esapi.SearchRequest{
		Index: []string{"users"},
		Body:  strings.NewReader(fmt.Sprintf(`{"query": {"match": {"name": "%s"}}}`, name)),
		Size:  1,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
	if len(hits) == 0 {
		return nil, nil
	}

	user := hits[0].(map[string]interface{})["_source"].(map[string]interface{})
	return &domain.User{
		ID:   user["id"].(string),
		Name: user["name"].(string),
		// Add other fields as necessary.
	}, nil
}


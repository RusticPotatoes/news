package dao

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"database/sql"

	"github.com/RusticPotatoes/news/domain"
)

var (
	db           *sql.DB
	mu           sync.RWMutex
	articleCache = make(map[string]domain.Article)
)

func Init(ctx context.Context) error {
	var err error
	db, err = sql.Open("sqlite3", "./editions.db")
	return err
}

// type storedEdition struct {
// 	ID         string
// 	Name       string
// 	Date       string
// 	StartTime  time.Time
// 	EndTime    time.Time
// 	Created    time.Time
// 	Sources    []domain.Source
// 	Articles   []string
// 	Categories []string
// 	Metadata   map[string]string
// }

type storedEdition struct {
	ID         string
	Name       string
	Date       string
	StartTime  time.Time
	EndTime    time.Time
	Created    time.Time
	Sources    string
	Articles   string
	Categories string
	Metadata   string
}

func Client() *sql.DB {
    return db
}

func GetEditionForTime(ctx context.Context, t time.Time, allowRecent bool) (*domain.Edition, error) {
	rows, err := db.Query("SELECT * FROM editions WHERE EndTime > ? ORDER BY EndTime DESC", t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := []*domain.Edition{}
	var maxEdition storedEdition

	for rows.Next() {
		var s storedEdition
		err = rows.Scan(&s.ID, &s.Name, &s.Date, &s.StartTime, &s.EndTime, &s.Created, /* ... other fields ... */)
		if err != nil {
			return nil, err
		}

		if s.EndTime.After(maxEdition.EndTime) {
			maxEdition = s
		}

		if s.EndTime.After(t) {
			e, err := editionFromStored(ctx, s)
			if err != nil {
				return nil, err
			}
			candidates = append(candidates, e)
		}
	}

	if len(candidates) == 0 {
		if maxEdition.ID != "" && allowRecent {
			return editionFromStored(ctx, maxEdition)
		}
	}

	selected := &domain.Edition{}
	for _, e := range candidates {
		if e.Created.After(selected.Created) {
			selected = e
		}
	}
	if selected.ID == "" {
		return nil, nil
	}
	return selected, nil
}

func SetEdition(ctx context.Context, e *domain.Edition) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, a := range e.Articles {
		err := SetArticle(ctx, &a)
		if err != nil {
			return errors.Wrap(err, "storing article: "+a.ID)
		}
	}

	stored, err := editionToStored(e)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO editions (ID, Name, Date, StartTime, EndTime, Created, /* ... other fields ... */) VALUES (?, ?, ?, ?, ?, ?, /* ... other fields ... */)",
		stored.ID, stored.Name, stored.Date, stored.StartTime, stored.EndTime, stored.Created, /* ... other fields ... */)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}


func editionToStored(e *domain.Edition) (storedEdition, error) {
	sources, err := json.Marshal(e.Sources)
	if err != nil {
		return storedEdition{}, err
	}

	articles, err := json.Marshal(e.Articles)
	if err != nil {
		return storedEdition{}, err
	}

	categories, err := json.Marshal(e.Categories)
	if err != nil {
		return storedEdition{}, err
	}

	metadata, err := json.Marshal(e.Metadata)
	if err != nil {
		return storedEdition{}, err
	}

	s := storedEdition{
		ID:         e.ID,
		Name:       e.Name,
		Date:       e.Date,
		Sources:    string(sources),
		StartTime:  e.StartTime,
		EndTime:    e.EndTime,
		Created:    e.Created,
		Categories: string(categories),
		Metadata:   string(metadata),
		Articles:   string(articles),
	}

	return s, nil
}

func editionFromStored(ctx context.Context, s storedEdition) (*domain.Edition, error) {
	var sources []domain.Source
	err := json.Unmarshal([]byte(s.Sources), &sources)
	if err != nil {
		return nil, err
	}

	var articles []string
	err = json.Unmarshal([]byte(s.Articles), &articles)
	if err != nil {
		return nil, err
	}

	var categories []string
	err = json.Unmarshal([]byte(s.Categories), &categories)
	if err != nil {
		return nil, err
	}

	var metadata map[string]string
	err = json.Unmarshal([]byte(s.Metadata), &metadata)
	if err != nil {
		return nil, err
	}

	e := domain.Edition{
		ID:         s.ID,
		Name:       s.Name,
		Date:       s.Date,
		Sources:    sources,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
		Created:    s.Created,
		Categories: categories,
		Metadata:   metadata,
	}

	e.Articles = make([]domain.Article, len(articles))
	g := errgroup.Group{}
	for i, id := range articles {
		i, id := i, id
		g.Go(func() error {
			a, err := GetArticle(ctx, id)
			if err != nil {
				return err
			}
			e.Articles[i] = *a
			return nil
		})
	}
	return &e, g.Wait()
}

func GetEdition(ctx context.Context, id string) (*domain.Edition, error) {
	row := db.QueryRow("SELECT * FROM editions WHERE ID = ?", id)

	var s storedEdition
	var sources, articles, categories, metadata string
	err := row.Scan(&s.ID, &s.Name, &s.Date, &s.StartTime, &s.EndTime, &s.Created, &sources, &articles, &categories, &metadata)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching edition found
			return nil, nil
		}
		return nil, err
	}

	var srcs []domain.Source
	err = json.Unmarshal([]byte(sources), &srcs)
	if err != nil {
		return nil, err
	}

	var arts []string
	err = json.Unmarshal([]byte(articles), &arts)
	if err != nil {
		return nil, err
	}

	var cats []string
	err = json.Unmarshal([]byte(categories), &cats)
	if err != nil {
		return nil, err
	}

	var meta map[string]string
	err = json.Unmarshal([]byte(metadata), &meta)
	if err != nil {
		return nil, err
	}

	e := domain.Edition{
		ID:         s.ID,
		Name:       s.Name,
		Date:       s.Date,
		Sources:    srcs,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
		Created:    s.Created,
		Categories: cats,
		Metadata:   meta,
	}

	e.Articles = make([]domain.Article, len(arts))
	g := errgroup.Group{}
	for i, id := range arts {
		i, id := i, id
		g.Go(func() error {
			a, err := GetArticle(ctx, id)
			if err != nil {
				return err
			}
			e.Articles[i] = *a
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func SetArticle(ctx context.Context, a *domain.Article) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Convert the article to JSON
	articleJson, err := json.Marshal(a)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO articles (ID, Data) VALUES (?, ?) ON CONFLICT(ID) DO UPDATE SET Data = ?",
		a.ID, articleJson, articleJson)
	if err != nil {
		tx.Rollback()
		return err
	}

	mu.Lock()
	delete(articleCache, a.ID)
	mu.Unlock()

	return tx.Commit()
}
func GetArticle(ctx context.Context, id string) (*domain.Article, error) {
	mu.RLock()
	a, ok := articleCache[id]
	if ok {
		mu.RUnlock()
		return &a, nil
	}
	mu.RUnlock()

	row := db.QueryRow("SELECT Data FROM articles WHERE ID = ?", id)

	var articleJson string
	err := row.Scan(&articleJson)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching article found
			return nil, nil
		}
		return nil, err
	}

	a = domain.Article{}
	err = json.Unmarshal([]byte(articleJson), &a)
	if err != nil {
		return nil, err
	}

	mu.Lock()
	articleCache[a.ID] = a
	mu.Unlock()

	return &a, nil
}
func GetArticleByURL(ctx context.Context, url string) (*domain.Article, error) {
	row := db.QueryRow("SELECT Data FROM articles WHERE Link = ?", url)

	var articleJson string
	err := row.Scan(&articleJson)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching article found
			return nil, nil
		}
		return nil, err
	}

	a := domain.Article{}
	err = json.Unmarshal([]byte(articleJson), &a)
	if err != nil {
		return nil, err
	}

	mu.Lock()
	articleCache[a.ID] = a
	mu.Unlock()

	return &a, nil
}
func GetArticlesByTime(ctx context.Context, start, end time.Time) ([]domain.Article, error) {
	rows, err := db.Query("SELECT Data FROM articles WHERE Timestamp > ? AND Timestamp < ? ORDER BY Timestamp", start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Article{}
	for rows.Next() {
		var articleJson string
		err = rows.Scan(&articleJson)
		if err != nil {
			return nil, err
		}

		a := domain.Article{}
		err = json.Unmarshal([]byte(articleJson), &a)
		if err != nil {
			return nil, err
		}

		mu.Lock()
		articleCache[a.ID] = a
		mu.Unlock()
		out = append(out, a)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func GetUser(ctx context.Context, id string) (*domain.User, error) {
	row := db.QueryRow("SELECT Data FROM users WHERE ID = ?", id)

	var userJson string
	err := row.Scan(&userJson)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching user found
			return nil, nil
		}
		return nil, err
	}

	u := domain.User{}
	err = json.Unmarshal([]byte(userJson), &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
func SetUser(ctx context.Context, u *domain.User) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Convert the user to JSON
	userJson, err := json.Marshal(u)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO users (ID, Data) VALUES (?, ?) ON CONFLICT(ID) DO UPDATE SET Data = ?",
		u.ID, userJson, userJson)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func GetUserByName(ctx context.Context, name string) (*domain.User, error) {
	row := db.QueryRow("SELECT Data FROM users WHERE Name = ?", name)

	var userJson string
	err := row.Scan(&userJson)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching user found
			return nil, nil
		}
		return nil, err
	}

	u := domain.User{}
	err = json.Unmarshal([]byte(userJson), &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func GetSource(ctx context.Context, id string) (*domain.Source, error) {
	row := db.QueryRow("SELECT Data FROM sources WHERE ID = ?", id)

	var sourceJson string
	err := row.Scan(&sourceJson)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching source found
			return nil, nil
		}
		return nil, err
	}

	s := domain.Source{}
	err = json.Unmarshal([]byte(sourceJson), &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func SetSource(ctx context.Context, s *domain.Source) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Convert the source to JSON
	sourceJson, err := json.Marshal(s)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO sources (ID, Data) VALUES (?, ?) ON CONFLICT(ID) DO UPDATE SET Data = ?",
		s.ID, sourceJson, sourceJson)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func DeleteSource(ctx context.Context, id string) error {
	_, err := db.Exec("DELETE FROM sources WHERE ID = ?", id)
	return err
}

func GetSources(ctx context.Context, ownerID string) ([]domain.Source, error) {
	rows, err := db.Query("SELECT Data FROM sources WHERE OwnerID = ?", ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []domain.Source{}
	for rows.Next() {
		var sourceJson string
		err = rows.Scan(&sourceJson)
		if err != nil {
			return nil, err
		}

		s := domain.Source{}
		err = json.Unmarshal([]byte(sourceJson), &s)
		if err != nil {
			return nil, err
		}

		sources = append(sources, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}

func GetAllSources(ctx context.Context) ([]domain.Source, error) {
	rows, err := db.Query("SELECT Data FROM sources")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []domain.Source{}
	for rows.Next() {
		var sourceJson string
		err = rows.Scan(&sourceJson)
		if err != nil {
			return nil, err
		}

		s := domain.Source{}
		err = json.Unmarshal([]byte(sourceJson), &s)
		if err != nil {
			return nil, err
		}

		sources = append(sources, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}

func GetArticlesForOwner(ctx context.Context, ownerID string, start, end time.Time) ([]domain.Article, []domain.Source, error) {
	var (
		sources []domain.Source
		err     error
	)
	if ownerID != "" {
		sources, err = GetSources(ctx, ownerID)
		if err != nil {
			return nil, nil, err
		}
	} else {
		sources, err = GetAllSources(ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	out := []domain.Article{}
	for _, s := range sources {
		rows, err := db.Query("SELECT Data FROM articles WHERE Source.FeedURL = ? AND Timestamp > ? AND Timestamp < ? ORDER BY Timestamp", s.FeedURL, start, end)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var articleJson string
			err = rows.Scan(&articleJson)
			if err != nil {
				return nil, nil, err
			}

			a := domain.Article{}
			err = json.Unmarshal([]byte(articleJson), &a)
			if err != nil {
				return nil, nil, err
			}

			out = append(out, a)
		}

		if err = rows.Err(); err != nil {
			return nil, nil, err
		}
	}

	return out, sources, nil
}
type feedCache struct {
	ttl time.Duration
}

func (c *feedCache) Get(url string) ([]domain.Article, bool) {
	rows, err := db.Query("SELECT Data FROM feed_cache WHERE URL = ?", url)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, false
	}

	var articlesJson string
	err = rows.Scan(&articlesJson)
	if err != nil {
		return nil, false
	}

	as := []domain.Article{}
	err = json.Unmarshal([]byte(articlesJson), &as)
	if err != nil {
		return nil, false
	}

	return as, true
}

func (c *feedCache) Set(url string, as []domain.Article) {
	articlesJson, err := json.Marshal(as)
	if err != nil {
		return
	}

	_, err = db.Exec("INSERT INTO feed_cache (URL, Data, Expiry) VALUES (?, ?, ?) ON CONFLICT(URL) DO UPDATE SET Data = ?, Expiry = ?",
		url, articlesJson, time.Now().Add(c.ttl), articlesJson, time.Now().Add(c.ttl))
	if err != nil {
		return
	}
}

func (c *feedCache) Delete(url string) {
	_, err := db.Exec("DELETE FROM feed_cache WHERE URL = ?", url)
	if err != nil {
		return
	}
}

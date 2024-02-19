package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"database/sql"

	"github.com/RusticPotatoes/news/domain"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	// _ "github.com/ncruces/go-sqlite3"
)

var (
	db           *sql.DB
	mu           sync.RWMutex
	articleCache = make(map[string]domain.Article)
)

func Init(ctx context.Context) error {
	var err error
	db, err = sql.Open("sqlite3", "./data/news.db")
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return fmt.Errorf("error in %s:%d: %v", file, line, err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating password hash: %v", err)
		return err
	}

	// Insert the default user
	query := `
		INSERT INTO users (name, password_hash, is_admin) 
		VALUES (?, ?, ?) 
		ON CONFLICT(name) DO NOTHING
	`
	params := []interface{}{"admin", hashedPassword, true}

	// interpolatedQuery := interpolateParams(query, params)
	// log.Printf("Executing query: %s", interpolatedQuery)

	_, err = db.Exec(query, params...)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		return err
	}
	// log.Println("Admin user inserted successfully")
	err = InitSources(ctx)
	if err != nil {
		return err
	}

	sources, err := GetAllSources(ctx)
	if err != nil {
		return err
	}

	cache := &feedCache{
		ttl: 1 * time.Hour,  // Cache feeds for 1 hour
	}

	// Populate the cache with feeds from all sources
	for _, source := range sources {
		// Get a feed from the cache
		articles, ok := cache.Get(source.FeedURL)
		if !ok {
			// If the feed is not in the cache, fetch it from the source and add it to the cache
			articles, err = GetArticlesForSource(ctx, source.FeedURL)
			if err != nil {
				return err
			}
			cache.Set(source.FeedURL, articles)
		}

		// Use the articles
		for _, article := range articles {
			fmt.Println(article.Title)
		}
	}

	return nil
}

// func interpolateParams(query string, params []interface{}) string {
// 	paramStrs := make([]interface{}, len(params))
// 	for i, param := range params {
// 		switch v := param.(type) {
// 		case string:
// 			paramStrs[i] = fmt.Sprintf("'%s'", v)
// 		case []byte:
// 			paramStrs[i] = fmt.Sprintf("'%x'", v)
// 		default:
// 			paramStrs[i] = fmt.Sprintf("%v", v)
// 		}
// 	}
// 	return fmt.Sprintf(query, paramStrs...)
// }

func Close() {
    if db != nil {
        db.Close()
    }
}

func HashPassword(password string) ([]byte, error) {
    return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func CheckPassword(password string, hashedPassword []byte) error {
    return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}

func GetArticlesForSource(ctx context.Context, link string) ([]domain.Article, error) {
    rows, err := db.Query("SELECT id, title, description, link, image_url, source_id, timestamp FROM articles WHERE link = ? ORDER BY timestamp", link)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

	articles := []domain.Article{}
	for rows.Next() {
		var a domain.Article
		err = rows.Scan(&a.ID, &a.Title, &a.Description, &a.Link, &a.ImageURL, &a.Source, &a.Timestamp)
		if err != nil {
			return nil, err
		}

		articles = append(articles, a)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func InitSources(ctx context.Context) error {
	sources, err := GetAllSources(ctx)
	if err != nil {
		return err
	}

	// If there are no sources in the database, insert the default sources
	if len(sources) == 0 {
		defaultSources := domain.GetSources()

		for _, source := range defaultSources {
			err := SetSource(ctx, &source)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Edition struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	Date       string    `db:"date"`
	StartTime  time.Time `db:"start_time"`
	EndTime    time.Time `db:"end_time"`
	Created    time.Time `db:"created"`
	Sources    []Source  `db:"sources"`
	Articles   []Article `db:"articles"`
	Categories []string  `db:"categories"`
	Metadata   map[string]string `db:"metadata"`
}

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

type Analytics struct {
	UserID              string    `db:"user_id"`
	InsertionTimestamp  time.Time `db:"insertion_timestamp"`
	Payload             string    `db:"payload"`
}

type Article struct {
	ID                string    `db:"id"`
	Title             string    `db:"title"`
	Description       string    `db:"description"`
	CompressedContent []byte    `db:"compressed_content"`
	ImageURL          string    `db:"image_url"`
	Link              string    `db:"link"`
	Author            string    `db:"author"`
	SourceID          int64     `db:"source_id"`
	Timestamp         time.Time `db:"timestamp"`
	Ts                string    `db:"ts"`
	Layout            string    `db:"layout"`
}

type Source struct {
	ID          string   `db:"id"`
	OwnerID     string   `db:"owner_id"`
	Name        string   `db:"name"`
	URL         string   `db:"url"`
	FeedURL     string   `db:"feed_url"`
	Categories  []string `db:"categories"`
	DisableFetch bool  `db:"disable_fetch"`
	LastFetchTime time.Time `db:"last_fetch_time"`
}

type storedSource struct {
	ID          	string
	OwnerID     	string
	Name        	string
	URL         	string
	FeedURL     	string
	Categories  	string
	DisableFetch 	bool
	LastFetchTime 	time.Time
	LayoutID    	string
}

type User struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	Created      time.Time `db:"created"`
	PasswordHash []byte    `db:"password_hash"`
	IsAdmin      bool      `db:"is_admin"`
}

func Client() *sql.DB {
    return db
}

func GetEditionForTime(ctx context.Context, t time.Time, allowRecent bool) (*domain.Edition, error) {
	rows, err := db.Query("SELECT * FROM edition WHERE EndTime > ? ORDER BY EndTime DESC", t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := []*domain.Edition{}
	var maxEdition storedEdition

	for rows.Next() {
		var s storedEdition
		err = rows.Scan(&s.ID, &s.Name, &s.Date, &s.StartTime, &s.EndTime, &s.Created, &s.Sources, &s.Articles, &s.Categories, &s.Metadata)
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

	_, err = tx.Exec(`
		INSERT INTO edition (name, date, start_time, end_time, created, sources, articles, categories, metadata) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(name, date) DO UPDATE SET 
		name = excluded.name, 
		date = excluded.date, 
		start_time = excluded.start_time, 
		end_time = excluded.end_time, 
		created = excluded.created, 
		sources = excluded.sources, 
		articles = excluded.articles, 
		categories = excluded.categories, 
		metadata = excluded.metadata
	`, 
		stored.Name, 
		stored.Date, 
		stored.StartTime, 
		stored.EndTime, 
		stored.Created, 
		stored.Sources, 
		stored.Articles, 
		stored.Categories, 
		stored.Metadata,
	)
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
		StartTime:  e.StartTime,
		EndTime:    e.EndTime,
		Created:    e.Created,
		Sources:    string(sources),
		Articles:   string(articles),
		Categories: string(categories),
		Metadata:   string(metadata),
	}

	return s, nil
}

func editionFromStored(ctx context.Context, s storedEdition) (*domain.Edition, error) {
	var sources []domain.Source
	err := json.Unmarshal([]byte(s.Sources), &sources)
	if err != nil {
		return nil, err
	}

	var articles []domain.Article
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
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
		Created:    s.Created,
		Sources:    sources,
		Articles:   articles,
		Categories: categories,
		Metadata:   metadata,
	}

	return &e, nil
}

func GetEdition(ctx context.Context, id string) (*domain.Edition, error) {
	row := db.QueryRow("SELECT * FROM edition WHERE ID = ?", id)

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

	_, err = tx.Exec(`
		INSERT INTO articles (title, description, compressed_content, image_url, link, author, source_id, timestamp, ts, layout_id) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) 
		ON CONFLICT(link) DO UPDATE SET 
			title = excluded.title, 
			description = excluded.description, 
			compressed_content = excluded.compressed_content, 
			image_url = excluded.image_url, 
			link = excluded.link, 
			author = excluded.author, 
			source_id = excluded.source_id, 
			timestamp = excluded.timestamp, 
			ts = excluded.ts, 
			layout_id = excluded.layout_id
	`, 
		a.Title, 
		a.Description, 
		a.CompressedContent, 
		a.ImageURL, 
		a.Link, 
		a.Author, 
		a.SourceID, 
		a.Timestamp, 
		a.Timestamp, 
		a.LayoutID, // Assuming a.LayoutID holds the layout ID
	)
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

	row := db.QueryRow("SELECT id, title, description, compressed_content, image_url, link, author, source_id, timestamp, ts, layout_id FROM articles WHERE ID = ?", id)

	a = domain.Article{}
	err := row.Scan(&a.ID, &a.Title, &a.Description, &a.CompressedContent, &a.ImageURL, &a.Link, &a.Author, &a.Source, &a.Timestamp, &a.TS, &a.Layout)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching article found
			return nil, nil
		}
		return nil, err
	}

	mu.Lock()
	articleCache[a.ID] = a
	mu.Unlock()

	return &a, nil
}
func GetArticleByURL(ctx context.Context, url string) (*domain.Article, error) {
	row := db.QueryRow("SELECT id, title, description, compressed_content, image_url, link, author, source_id, timestamp, ts, layout_id FROM articles WHERE Link = ?", url)

	a := domain.Article{}
	err := row.Scan(&a.ID, &a.Title, &a.Description, &a.CompressedContent, &a.ImageURL, &a.Link, &a.Author, &a.Source, &a.Timestamp, &a.Timestamp, &a.Layout)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching article found
			return nil, nil
		}
		return nil, err
	}

	mu.Lock()
	articleCache[a.ID] = a
	mu.Unlock()

	return &a, nil
}
func GetArticlesByTime(ctx context.Context, start, end time.Time) ([]domain.Article, error) {
	rows, err := db.Query("SELECT id, title, description, compressed_content, image_url, link, author, source_id, timestamp, ts, layout_id FROM articles WHERE Timestamp > ? AND Timestamp < ? ORDER BY Timestamp", start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Article{}
	for rows.Next() {
		a := domain.Article{}
		err = rows.Scan(&a.ID, &a.Title, &a.Description, &a.CompressedContent, &a.ImageURL, &a.Link, &a.Author, &a.Source, &a.Timestamp, &a.Timestamp, &a.Layout)
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
	row := db.QueryRow("SELECT id, name, password_hash FROM users WHERE ID = ?", id)

	u := domain.User{}
	err := row.Scan(&u.ID, &u.Name, &u.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching user found
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}

func GetUserByName(ctx context.Context, name string) (*domain.User, error) {
	row := db.QueryRow("SELECT id, name, password_hash FROM users WHERE Name = ?", name)

	u := domain.User{}
	err := row.Scan(&u.ID, &u.Name, &u.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching user found
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}

func SetUser(ctx context.Context, u *domain.User) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO users (name, password_hash, is_admin) 
		VALUES (?, ?, ?) 
		ON CONFLICT(name) DO UPDATE SET 
		name = excluded.name, 
		password_hash = excluded.password_hash, 
		is_admin = excluded.is_admin
	`, 
		u.Name, 
		u.PasswordHash, 
		u.IsAdmin,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func GetSource(ctx context.Context, id string) (*domain.Source, error) {
	row := db.QueryRow("SELECT id, owner_id, name, url, feed_url, categories, disable_fetch, layout_id FROM sources WHERE ID = ?", id)

	var s storedSource
	err := row.Scan(&s.ID, &s.OwnerID, &s.Name, &s.URL, &s.FeedURL, &s.Categories, &s.DisableFetch, &s.LayoutID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching source found
			return nil, nil
		}
		return nil, err
	}

	source := domain.Source{
		ID:           s.ID,
		OwnerID:      s.OwnerID,
		Name:         s.Name,
		URL:          s.URL,
		FeedURL:      s.FeedURL,
		Categories:   strings.Split(s.Categories, ","),
		DisableFetch: s.DisableFetch,
		LayoutID:     s.LayoutID,
	}

	return &source, nil
}

func SetSource(ctx context.Context, s *domain.Source) error {
	if s.OwnerID == "" {
        s.OwnerID = "admin" // Default to "admin" if no OwnerID is provided
    }
	// if s.ID == "" {
	// 	s.ID = idgen.New("src")
	// }
	log.Printf("Processing source: %v", s)
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	categories := strings.Join(s.Categories, ",")
	// log.Printf("Inserting into sources: owner_id=%s, name=%s, url=%s, feed_url=%s, categories=%s, disable_fetch=%t", 
		// s.OwnerID, s.Name, s.URL, s.FeedURL, categories, s.DisableFetch)

	_, err = tx.Exec(`
		INSERT INTO sources (owner_id, name, url, feed_url, categories, disable_fetch) 
		VALUES (?, ?, ?, ?, ?, ?) 
		ON CONFLICT(owner_id, url) DO UPDATE SET 
		owner_id = excluded.owner_id, 
		name = excluded.name, 
		url = excluded.url, 
		feed_url = excluded.feed_url, 
		categories = excluded.categories, 
		disable_fetch = excluded.disable_fetch
	`, s.OwnerID, s.Name, s.URL, s.FeedURL, categories, s.DisableFetch)
	if err != nil {
		log.Printf("Error inserting into sources: %v", err)
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
	rows, err := db.Query("SELECT id, owner_id, name, url, feed_url, categories, disable_fetch FROM sources WHERE owner_id = ?", ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []domain.Source{}
	for rows.Next() {
		var s storedSource
		err = rows.Scan(&s.ID, &s.OwnerID, &s.Name, &s.URL, &s.FeedURL, &s.Categories, &s.DisableFetch)
		if err != nil {
			return nil, err
		}

		source := domain.Source{
			ID:          s.ID,
			OwnerID:     s.OwnerID,
			Name:        s.Name,
			URL:         s.URL,
			FeedURL:     s.FeedURL,
			Categories:  strings.Split(s.Categories, ","),
			DisableFetch: s.DisableFetch,
		}

		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}

func GetAllSources(ctx context.Context) ([]domain.Source, error) {
	rows, err := db.Query("SELECT id, owner_id, name, url, feed_url, categories, disable_fetch, last_fetch_time FROM sources")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []domain.Source{}
	for rows.Next() {
		var s storedSource
		err = rows.Scan(&s.ID, &s.OwnerID, &s.Name, &s.URL, &s.FeedURL, &s.Categories, &s.DisableFetch, &s.LastFetchTime)
		if err != nil {
			return nil, err
		}

		source := domain.Source{
			ID:            s.ID,
			OwnerID:       s.OwnerID,
			Name:          s.Name,
			URL:           s.URL,
			FeedURL:       s.FeedURL,
			Categories:    strings.Split(s.Categories, ","),
			DisableFetch:  s.DisableFetch,
			LastFetchTime: s.LastFetchTime,
		}

		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}
func GetAllSourcesForOwner(ctx context.Context, ownerID string) ([]domain.Source, error) {
	query := "SELECT id, owner_id, name, url, feed_url, categories, disable_fetch, last_fetch_time FROM sources WHERE owner_id = ?"
	rows, err := db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []domain.Source{}
	for rows.Next() {
		var s storedSource
		err = rows.Scan(&s.ID, &s.OwnerID, &s.Name, &s.URL, &s.FeedURL, &s.Categories, &s.DisableFetch, &s.LastFetchTime)
		if err != nil {
			return nil, err
		}

		source := domain.Source{
			ID:            s.ID,
			OwnerID:       s.OwnerID,
			Name:          s.Name,
			URL:           s.URL,
			FeedURL:       s.FeedURL,
			Categories:    strings.Split(s.Categories, ","),
			DisableFetch:  s.DisableFetch,
			LastFetchTime: s.LastFetchTime,
		}

		sources = append(sources, source)
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
		sqlStatement := fmt.Sprintf("SELECT id, title, description, link, image_url, source_id, timestamp FROM articles WHERE source_id = '%s' AND timestamp > '%s' AND timestamp < '%s' ORDER BY timestamp", s.ID, start.Format(time.RFC3339), end.Format(time.RFC3339))
		// log.Println(sqlStatement)
	
		rows, err := db.Query(sqlStatement)
		if err != nil {
			return nil, nil, err
		}
        defer rows.Close()

		for rows.Next() {
			var a domain.Article
			err = rows.Scan(&a.ID, &a.Title, &a.Description, &a.Link, &a.ImageURL, &a.SourceID, &a.Timestamp)
			if err != nil {
				return nil, nil, err
			}
			a.Source = s

			out = append(out, a)
		}

		if err = rows.Err(); err != nil {
			return nil, nil, err
		}
	}

	return out, sources, nil
}
// GetLastFetchTimeForSource fetches the last fetch time for a given source from the database
func GetLastFetchTimeForSource(ctx context.Context, sourceID int) (time.Time, error) {
	// Prepare a query to select the last fetch time for the given source
	query := "SELECT last_fetch_time FROM sources WHERE id = ?"

	// Execute the query
	row := db.QueryRowContext(ctx, query, sourceID)

	// Parse the result
	var lastFetchTime time.Time
	if err := row.Scan(&lastFetchTime); err != nil {
		return time.Time{}, err
	}

	// Return the last fetch time
	return lastFetchTime, nil
}

// SetLastFetchTimeForSource updates the last fetch time for a given source in the database
func SetLastFetchTimeForSource(ctx context.Context, sourceID int, lastFetchTime time.Time) error {
	// Prepare a query to update the last fetch time for the given source
	query := "UPDATE sources SET last_fetch_time = ? WHERE id = ?"

	// Execute the query
	_, err := db.ExecContext(ctx, query, lastFetchTime, sourceID)

	// Return any error
	return err
}

// func GetAllArticlesForOwner(ctx context.Context, ownerID string) ([]domain.Article, []domain.Source, error) {
// 	var (
// 		sources []domain.Source
// 		err     error
// 	)
// 	if ownerID != "" {
// 		sources, err = GetSources(ctx, ownerID)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 	} else {
// 		sources, err = GetAllSources(ctx)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 	}

// 	out := []domain.Article{}
// 	for _, s := range sources {
// 		rows, err := db.Query("SELECT id, title, description, link, image_url, source_id, timestamp FROM articles WHERE link = ? ORDER BY timestamp", s.URL)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		defer rows.Close()

// 		for rows.Next() {
// 			var a domain.Article
// 			err = rows.Scan(&a.ID, &a.Title, &a.Description, &a.Link, &a.ImageURL, &a.Source, &a.Timestamp)
// 			if err != nil {
// 				return nil, nil, err
// 			}

// 			out = append(out, a)
// 		}

// 		if err = rows.Err(); err != nil {
// 			return nil, nil, err
// 		}
// 	}

// 	return out, sources, nil
// }
type feedCache struct {
	ttl time.Duration
}

func (c *feedCache) Get(url string) ([]domain.Article, bool) {
	row := db.QueryRow("SELECT Data, Expiry FROM feed_cache WHERE URL = ?", url)

	var data string
	var expiry time.Time
	err := row.Scan(&data, &expiry)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching cache found
			return nil, false
		}
		return nil, false
	}

	// Check if the cache has expired
	if time.Now().After(expiry) {
		return nil, false
	}

	var articles []domain.Article
	err = json.Unmarshal([]byte(data), &articles)
	if err != nil {
		return nil, false
	}

	return articles, true
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

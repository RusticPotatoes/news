package articles

import (
	"context"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"

	"github.com/RusticPotatoes/news/dao"
	"github.com/RusticPotatoes/news/domain"
	"github.com/mmcdole/gofeed"
	"github.com/monzo/slog"
)

// FetchArticles fetches articles from all sources for a user
func FetchArticles(ctx context.Context, ownerID string) {
	sources, err := dao.GetAllSourcesForOwner(ctx, ownerID)
	if err != nil {
		slog.Critical(ctx, "Error getting sources: %s", err)
		return
	}

	fp := gofeed.NewParser()
	for _, source := range sources {
		feed, err := fp.ParseURL(source.FeedURL)
		if err != nil {
			slog.Critical(ctx, "Error getting feed: %s", err)
			continue
		}

		// Get the articles that were published by the source in the last 24 hours
		old_articles, err := dao.GetArticlesBySourceAndTime(ctx, source.ID, time.Now().Add(-24*time.Hour), time.Now())
		if err != nil {
			log.Printf("Error getting articles: %s", err)
			return
		}
		for _, item := range feed.Items {
			// Check if the article has already been fetched
			existingArticle := findArticleByLink(old_articles, item.Link)
			if existingArticle != nil {
				// If the article already exists, skip it
				continue
			}
			sourceID, err := strconv.Atoi(source.ID)
			if err != nil {
				slog.Critical(ctx, "Error converting source ID to int: %s", err)
				continue
			}
			var authorName string
			if item.Author != nil {
				authorName = item.Author.Name
			}
			var published time.Time
			if item.PublishedParsed != nil {
				published = *item.PublishedParsed
			}

			// Skip the item if it was not published in the last 24 hours
			if time.Since(published) > 24*time.Hour {
				continue
			}

			read_article, err := readability.FromURL(item.Link, 15*time.Second)
			if err != nil {
				if !strings.Contains(err.Error(), "failed to parse date") {
					log.Printf("failed to parse %s, %v\n", item.Link, err)
					continue
				}
				// If it's a date parsing error, ignore it and continue
				log.Printf("failed to parse date in %s, ignoring: %v\n", item.Link, err)
			}

			compressedContent, err := domain.CompressContent(read_article)
			if err != nil {
				compressedContent = []byte("")
			}

			// Create an Article from the feed item
			article := &domain.Article{
				Title:       removeHTMLTag(item.Title),
				Description: removeHTMLTag(read_article.Excerpt),
				Link:        item.Link,
				Author:      authorName, // This assumes that the item's Author field is not nil
				Source:    	 source, // This assumes that the source.Name is a string
				SourceID:    int64(sourceID), // This assumes that the source.Name is a string
				Timestamp:   published, // This assumes that the item's PublishedParsed field is not nil
				// Fill in the other Article fields as needed
				Content: read_article,
				CompressedContent: compressedContent,
				ImageURL: read_article.Image,
				TS:       published.Format("Mon Jan 2 15:04"),
			}

			// article.SetHTMLContent(body_text.Body)

			// sa, err := swan.FromHTML(article.Link, []byte(body_text.Body))
			// if err != nil {
			// 	return
			// }
			// if sa.Img != nil {
			// 	article.ImageURL = sa.Img.Src
			// }

			// Save the Article to the database
			err = dao.SetArticle(ctx, article)
			if err != nil {
				slog.Critical(ctx, "Error saving article: %s", err)
				continue
			}
		}
	}
}

func findArticleByLink(articles []domain.Article, link string) *domain.Article {
    for _, article := range articles {
        if article.Link == link {
            return &article
        }
    }
    return nil
}

func removeHTMLTag(in string) string {
	// regex to match html tag
	const pattern = `(<\/?[a-zA-A]+?[^>]*\/?>)*`
	r := regexp.MustCompile(pattern)
	groups := r.FindAllString(in, -1)
	// should replace long string first
	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i]) > len(groups[j])
	})
	for _, group := range groups {
		if strings.TrimSpace(group) != "" {
			in = strings.ReplaceAll(in, group, "")
		}
	}
	return in
}

// func compressContent(content string) ([]byte, error) {
// 	buf := new(bytes.Buffer)
// 	w := gzip.NewWriter(buf)
// 	_, err := w.Write([]byte(content))
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = w.Close()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

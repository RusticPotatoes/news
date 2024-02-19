package articles

import (
	"context"
	"strconv"
	"time"

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

		for _, item := range feed.Items {
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
			// Create an Article from the feed item
			article := &domain.Article{
				Title:       item.Title,
				Description: item.Description,
				Link:        item.Link,
				Author:      authorName, // This assumes that the item's Author field is not nil
				// Source:    	 domain.Source{Name: source.ID}, // This assumes that the source.Name is a string
				SourceID:    int64(sourceID), // This assumes that the source.Name is a string
				Timestamp:   published, // This assumes that the item's PublishedParsed field is not nil
				// Fill in the other Article fields as needed
				Content: []domain.Element{
					{
						Type:  "text",
						Value: item.Description,
					},
				},
			}

			// Save the Article to the database
			err = dao.SetArticle(ctx, article)
			if err != nil {
				slog.Critical(ctx, "Error saving article: %s", err)
				continue
			}
		}
	}
}
package articles

import (
	"bytes"
	"compress/gzip"
	"context"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
	"github.com/thatguystone/swan"

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

			read_article, err := readability.FromURL(item.Link, 30*time.Second)
			if err != nil {
				log.Printf("failed to parse %s, %v\n", item.Link, err)
				continue
			}
			
			var body_text = struct {
				Body     string `json:"body"`
				BodyText string `json:"body_text"`
			}{}

			cleaned_content := removeHTMLTag(read_article.TextContent)
			compressedContent, err := compressContent(cleaned_content)
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
				Content: []domain.Element{
					{
						Type:  "text",
						Value: removeHTMLTag(read_article.TextContent),
					},
				},
				CompressedContent: compressedContent,
				ImageURL: read_article.Image,
				TS:       published.Format("Mon Jan 2 15:04"),
			}

			// article.SetHTMLContent(body_text.Body)

			sa, err := swan.FromHTML(article.Link, []byte(body_text.Body))
			if err != nil {
				return
			}
			if sa.Img != nil {
				article.ImageURL = sa.Img.Src
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

func compressContent(content string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	_, err := w.Write([]byte(content))
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
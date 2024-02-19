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
	"golang.org/x/sync/semaphore"
)

var (
	s = semaphore.NewWeighted(20)
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
				log.Fatalf("failed to parse %s, %v\n", item.Link, err)
			}

			// Create an Article from the feed item
			article := &domain.Article{
				Title:       item.Title,
				Description: removeHTMLTag(read_article.Excerpt),
				Link:        item.Link,
				Author:      authorName, // This assumes that the item's Author field is not nil
				// Source:    	 domain.Source{Name: source.ID}, // This assumes that the source.Name is a string
				SourceID:    int64(sourceID), // This assumes that the source.Name is a string
				Timestamp:   published, // This assumes that the item's PublishedParsed field is not nil
				// Fill in the other Article fields as needed
				Content: []domain.Element{
					{
						Type:  "text",
						Value: removeHTMLTag(read_article.TextContent),
					},
				},
				ImageURL: read_article.Image,
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

// func findFirstImage(n *html.Node, base ...*url.URL) string {
//     if n.Type == html.ElementNode && n.Data == "img" {
//         for _, a := range n.Attr {
//             if a.Key == "src" {
//                 ext := strings.ToLower(path.Ext(a.Val))
//                 if ext == ".svg" {
//                     continue
//                 }
//                 if len(base) > 0 && base[0] != nil {
//                     imgURL, err := url.Parse(a.Val)
//                     if err != nil {
//                         continue
//                     }
//                     absURL := base[0].ResolveReference(imgURL)
//                     return absURL.String()
//                 }
//                 return a.Val
//             }
//         }
//     }
//     for c := n.FirstChild; c != nil; c = c.NextSibling {
//         if img := findFirstImage(c, base...); img != "" {
//             return img
//         }
//     }
//     return ""
// }

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
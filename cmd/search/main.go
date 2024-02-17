package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/RusticPotatoes/news/pkg/util"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/monzo/slog"
)

func main() {
	ctx := context.Background()

	var logger slog.Logger
	logger = util.ContextParamLogger{Logger: &util.StackDriverLogger{}}
	logger = util.ColourLogger{Writer: os.Stdout}
	slog.SetDefaultLogger(logger)

	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	query := flag.String("q", "", "query")
	flag.Parse()

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf(`{
		"query": {
			"multi_match" : {
				"query":    "%s", 
				"fields": [ "title", "content" ] 
			}
		}
	}`, *query))

	req := esapi.SearchRequest{
		Index: []string{"news"},
		Body:  strings.NewReader(buf.String()),
	}

	res, err := req.Do(ctx, es)
	if err != nil {
		slog.Critical(ctx, "Error getting response: %s", err)
		return
	}
	defer res.Body.Close()

	// TODO: Parse the response body to get the search results.
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		slog.Critical(ctx, "Error decoding response body: %s", err)
		return
	}

	hits, ok := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		slog.Critical(ctx, "Invalid response format")
		return
	}

	for _, hit := range hits {
		source, ok := hit.(map[string]interface{})["_source"].(map[string]interface{})
		if !ok {
			slog.Critical(ctx, "Invalid response format")
			return
		}

		// Access the search result fields
		title := source["title"].(string)
		content := source["content"].(string)

		// Process the search result
		fmt.Printf("Title: %s\nContent: %s\n\n", title, content)
	}
}
package main

import (
	"context"
	"os"

	"github.com/RusticPotatoes/news/dao"
	"github.com/RusticPotatoes/news/pkg/util"
	"github.com/monzo/slog"
)
func main() {
	ctx := context.Background()

	var logger slog.Logger
	logger = util.ContextParamLogger{Logger: &util.StackDriverLogger{}}
	logger = util.ColourLogger{Writer: os.Stdout}
	slog.SetDefaultLogger(logger)

	err := dao.Init(ctx)
	if err != nil {
		slog.Critical(ctx, "Error setting up dao: %s", err)
		return
	}

	articles, err := dao.GetArticles(ctx) // Fix the function call

	if err != nil {
		slog.Critical(ctx, "Error getting articles: %s", err)
		return
	}

	for _, a := range articles {
		a.RawHTML()
		contentStr := ""
		for _, e := range a.Content {
			if e.Type != "text" {
				continue
			}
			contentStr = contentStr + e.Value + ""
		}
		slog.Info(ctx, "Article: %s %s", a.Link)
	}
}
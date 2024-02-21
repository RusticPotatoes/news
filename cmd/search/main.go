package main

import (
	"context"
	"os"
	"time"

	"github.com/RusticPotatoes/news/dao"
	"github.com/RusticPotatoes/news/pkg/util"
	"github.com/monzo/slog"
	// "github.com/pacedotdev/firesearch-sdk/clients/go/firesearch"
	// secrets "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
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

	articles, err := dao.GetArticlesByTime(ctx, time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		slog.Critical(ctx, "Error getting articles: %s", err)
		return
	}

	for _, a := range articles {
		a.RawHTML()
		// contentStr := a.Content.Content
		slog.Info(ctx, "Article: %s", a.Link)
	}
}

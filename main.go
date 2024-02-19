package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/monzo/slog"

	"github.com/RusticPotatoes/news/cmd/articles"
	"github.com/RusticPotatoes/news/dao"
	"github.com/RusticPotatoes/news/handler"
	"github.com/RusticPotatoes/news/pkg/util"
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

	// Create a new scheduler
	s := gocron.NewScheduler(time.UTC)

	ownerID := "admin" 

	// Schedule fetchArticles to run every day at 9am
	_, err = s.Every(1).Day().At("9:00").Do(func() {
		articles.FetchArticles(ctx, ownerID)
	})
	if err != nil {
		slog.Critical(ctx, "Error scheduling task: %s", err)
		return
	}

    articles.FetchArticles(ctx, ownerID)

	// Start the scheduler (runs in its own goroutine)
	s.StartAsync()

	var addr string
	if os.Getenv("NEWS_ENV") == "debug" {
		addr = ":8081"
	} else {
		addr = ":8080"
	}

	slog.Info(ctx, "ready, listening on addr: %s", addr)
	slog.Error(ctx, "serving: %s", http.ListenAndServe(addr, handler.Init(ctx)))
	// Keep the main function running
	
	select {}
}
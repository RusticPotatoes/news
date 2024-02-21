package main

import (
	"context"
	"log"
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


    // Initialize the database
    err := dao.Init(context.Background())
    if err != nil {
        log.Fatalf("failed to initialize database: %v", err)
    }

	if err != nil {
		slog.Critical(ctx, "Error setting up dao: %s", err)
		return
	}
	ownerID := "admin" 

	// Create a new scheduler
	s := gocron.NewScheduler(time.UTC)


	_, err = s.Every(1).Day().At("2:00").Do(articles.FetchArticles, ctx, ownerID)
	if err != nil {
		slog.Critical(ctx, "Error scheduling task: %s", err)
		return
	}

	_, err = s.Every(1).Day().At("10:00").Do(articles.FetchArticles, ctx, ownerID)
	if err != nil {
		slog.Critical(ctx, "Error scheduling task: %s", err)
		return
	}

	_, err = s.Every(1).Day().At("17:00").Do(articles.FetchArticles, ctx, ownerID)
	if err != nil {
		slog.Critical(ctx, "Error scheduling task: %s", err)
		return
	}

	// Run tasks immediately
	go articles.FetchArticles(ctx, ownerID)

	go func() {
		s.StartBlocking()
	}()

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
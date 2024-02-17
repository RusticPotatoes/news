package main

import (
	"context"
	"net/http"
	"os"

	"github.com/monzo/slog"

	"github.com/RusticPotatoes/news/dao"
	"github.com/RusticPotatoes/news/handler"
	"github.com/RusticPotatoes/news/idgen"
	"github.com/RusticPotatoes/news/pkg/util"
)

func main() {
	ctx := context.Background()

	err := idgen.Init(ctx)
	if err != nil {
		slog.Error(ctx, "Error initialising idgen: %s", err)
		os.Exit(1)
	}

	var logger slog.Logger
	logger = util.ContextParamLogger{Logger: &util.StackDriverLogger{}}

	if os.Getenv("USER") == "alexrussell-saw" {
		logger = util.ColourLogger{Writer: os.Stdout}
		handler.Prefix = "dev-"
	}

	slog.SetDefaultLogger(logger)

	// err = dao.Init(ctx)
	// if err != nil {
	// 	slog.Error(ctx, "error initialising dao: %s", err)
	// 	os.Exit(1)
	// }

	var addr string
	if os.Getenv("NEWS_ENV") == "debug" {
		addr = ":8081"
	} else {
		addr = ":8080"
	}

	slog.Info(ctx, "ready, listening on addr: %s", addr)
	slog.Error(ctx, "serving: %s", http.ListenAndServe(addr, handler.Init(ctx)))
}

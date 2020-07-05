package handler

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"

	"time"

	requestlogger "github.com/kataras/iris/v12/middleware/logger"
)

var app *iris.Application

func StartListening(logger log.Logger, producer sarama.SyncProducer) error {
	app = newServer(logger, producer)
	err := app.Run(iris.Addr(":5000"), iris.WithOptimizations, iris.WithoutBanner, iris.WithoutStartupLog, iris.WithoutBanner, iris.WithoutServerError(iris.ErrServerClosed))
	if err != nil {
		return fmt.Errorf("error starting the web server : %s", err.Error())
	}
	return nil
}

func StopListening(logger log.Logger) {
	level.Info(logger).Log("msg", "quitting the application")
	err := app.Shutdown(context.Background())
	if err != nil {
		level.Error(logger).Log("msg", "error on stopping the application.", "err", err)
	}
}

func newServer(logger log.Logger, producer sarama.SyncProducer) *iris.Application {
	app := iris.New()
	app.Use(recover.New())
	app.Use(requestlogger.New(requestlogger.Config{
		LogFuncCtx: func(ctx iris.Context, latency time.Duration) {
			statusCode := ctx.GetStatusCode()
			url := ctx.Path()
			method := ctx.Method()
			log.With(logger, "statusCode", statusCode, "url", url, "action", method, "latency", latency).
				Log("Request completed in %v", latency)
		},
	}))

	h := newHandler(logger, producer)
	setupRoutes(app, h)

	return app
}

package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vincoll/vigie/foundation/web"
	"github.com/vincoll/vigie/internal/api/dbpgx"
	v0 "github.com/vincoll/vigie/internal/api/handlers/v0"
	"github.com/vincoll/vigie/internal/api/handlers/v0/testgrp"
	"github.com/vincoll/vigie/pkg/business/core/probemgmt"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type APIMuxConfig struct {
	Log    *zap.SugaredLogger
	DB     *dbpgx.Client
	Tracer trace.Tracer
}

func APIMux(cfg APIMuxConfig) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	var app *web.App

	// Accept CORS 'OPTIONS' preflight requests if config has been provided.
	// Don't forget to apply the CORS middleware to the routes that need it.
	// Example Config: `conf:"default:https://MY_DOMAIN.COM"`
	h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return nil
	}
	app.Handle(http.MethodOptions, "", "/*", h, nil)

	if app == nil {
		app = web.NewApp(
			context.Background(), nil,
		)
	}

	// Load the v1 routes.
	v0.Routes(app, v0.Config{
		Log: cfg.Log,
		//	Auth: cfg.Auth,
		DB: cfg.DB,
	})

	return app

}

// AddMux add routes to Gin
func AddMuxTests(rt *gin.Engine, logger *zap.SugaredLogger, db *dbpgx.Client) {

	tgrpHandler := testgrp.Handlers{Test: probemgmt.NewCore(logger, db)}
	//pas init
	rt.GET("api/test/create", tgrpHandler.Create)
	rt.GET("api/test/create2", tgrpHandler.QueryByID)

	rt.GET("api/test/get/:id", tgrpHandler.QueryByID)
}

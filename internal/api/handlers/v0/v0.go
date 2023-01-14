package v0

import (
	"github.com/vincoll/vigie/foundation/web"
	"github.com/vincoll/vigie/internal/api/dbpgx"
	"go.uber.org/zap"
)

type Config struct {
	Log *zap.SugaredLogger
	//Auth *auth.Auth
	DB *dbpgx.Client
}

// Routes binds all the version 0 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v0"
	//	x := probe.NewTest(cfg.Log, cfg.DB)

	// Tests Handlers

	//	tgrpHandler := testgrp.Handlers{Test: probe.NewCore(cfg.Log, cfg.DB)}

	//	tgrpHandler.GET("/test/create", tgrpHandler.Create)

	// Users Handlers
	// ...
}

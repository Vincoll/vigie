package webapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vincoll/vigie/internal/api/dbpgx"
	"go.uber.org/zap"
)

type APIServerConfig struct {
	ApiPort  string `toml:"ApiPort"`
	TechPort string `toml:"TechPort"`
	Pprof    string `toml:"pprof"`
	Env      string
}

type WebServer struct {

	// httpServerAPI exposes business parts of the goobs
	httpServerAPI *http.Server
	logger        *zap.SugaredLogger
	db            *dbpgx.Client
	status        string
}

// NewHTTPServer runs api business endpoint.
func NewHTTPServer(ctx context.Context, cfg APIServerConfig, env string, logger *zap.SugaredLogger, db *dbpgx.Client) (*WebServer, error) {

	ws := WebServer{logger: logger, db: db}

	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Exposes business routes
	// Note for Cloud Run (or others product who do not have HC and relies on Port opening)
	// Traffic will be sent as soon as the port is being open.
	// Therefore, every other internal components of the app must be Init and Ready.
	go ws.startAPIEndpoint(ctx, cfg.ApiPort, cfg.Env)

	return &ws, nil

}

// startAPIEndpoint exposes business routes
func (ws *WebServer) startAPIEndpoint(ctx context.Context, port, env string) {

}

// GracefulShutdown is controlled by AHS it will stop the API Web Server
func (ws *WebServer) GracefulShutdown() error {

	// Set app UnHealthy
	ws.status = "ShuttingDown"

	// From now the HealthCheck endpoint will return =! 200
	// Wait to clear this instance from the LB or any networking cache
	// If dealing with long-lived connection like WS : be prudent -> Graceful them
	time.Sleep(2 * time.Second)

	// Close the HTTP

	ctxTO, cancelTO := context.WithTimeout(context.Background(), 3*time.Second)
	defer func() { cancelTO() }()

	if err := ws.httpServerAPI.Shutdown(ctxTO); err != nil {
		return fmt.Errorf("HTTP Service failed to shutdown properly: %v", err)
	} else {
		zap.S().Infow("HTTP Service gracefully stopped")
	}

	return nil

}

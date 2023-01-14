package webapi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"

	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/internal/api/handlers"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

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

	_, httpSpan := otel.Tracer("vigie-boot").Start(ctx, "api-start")

	// App Routes ------------------------------------------

	// Register the HTTP handler and starts
	ws.logger.Infow(fmt.Sprintf("Expose /api/* routes on :"+port),
		"component", "api")

	router := gin.New()
	// Log
	router.Use(ginzap.Ginzap(ws.logger.Desugar(), time.RFC3339, true))
	// Trace
	router.Use(otelgin.Middleware("vigie-api"))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	// Add routes
	handlers.AddMuxTests(router, ws.logger, ws.db)

	ws.httpServerAPI = &http.Server{
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	// Run server

	// Listen : Open Socket (this operation is not blocking)
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		zap.S().Fatalf("%v", err)
	}

	// Server is ready to received requests
	ws.status = "ok"
	httpSpan.End()

	// Serve will consume any data on the socket
	if err := ws.httpServerAPI.Serve(l); err != http.ErrServerClosed {
		ws.status = "nok"

		if err.Error() != "http: Server closed" {
			zap.S().Fatalf("HTTP API ListenAndServe: %v", err)
			return
		}

	}

}

func (ws *WebServer) GracefulShutdown() error {

	// Set app UnHealthy
	ws.status = "ShuttingDown"

	// From now the HealthCheck endpoint will return =! 200
	// Wait to clear this instance from the LB or any networking cache
	// If dealing with long lived connection like WS : be prudent -> Graceful them
	time.Sleep(2 * time.Second)

	// Close the HTTP

	ctxTO, cancelTO := context.WithTimeout(context.Background(), 3*time.Second)
	defer func() { cancelTO() }()

	if err := ws.httpServerAPI.Shutdown(ctxTO); err != nil {
		return fmt.Errorf("HTTP Service failed to shutdown properly: %v", err)
	} else {
		zap.S().Warn("HTTP Service gracefully stopped\n")
	}

	return nil

}

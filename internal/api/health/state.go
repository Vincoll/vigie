package health

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vincoll/vigie/internal/api/dbsqlx"
	"github.com/vincoll/vigie/internal/api/webapi"
	"go.uber.org/zap"
)

// The AppHealthState register each component of the goobs where a failure can occur.
// We can assume that the LB or any equivalent service will HealthCheck the goobs.
// The goal is to declare the goobs unhealthy ASAP,
// and de-registered the faulty instance from the Load Balancer
type AppHealthState struct {
	mu sync.RWMutex

	// httpServerTech exposes technical parts of the goobs
	// such as debug, metrics, healthcheck endpoints, ...
	// This server is running on another port than the HTTP
	// to avoid any accidental access.
	httpServerTech *http.Server

	webServer      *webapi.WebServer
	db             *dbsqlx.Client
	askForShutdown bool
	status         Status
	log            *zap.SugaredLogger
}

func NewAHS(cfg webapi.APIServerConfig, ws *webapi.WebServer, dbc *dbsqlx.Client, log *zap.SugaredLogger) *AppHealthState {

	ahs := AppHealthState{
		mu:             sync.RWMutex{},
		webServer:      ws,
		db:             dbc,
		askForShutdown: false,
		status:         0,
		log:            log,
	}

	go ahs.startTechnicalEndpoint(cfg.TechPort, cfg.Pprof)
	go ahs.shutdownHandler()

	return &ahs

}

func (ahs *AppHealthState) HealthCheck() {
	ahs.mu.Lock()
	defer ahs.mu.Unlock()
	ahs.status = NotReady
}

func (ahs *AppHealthState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is my home page"))
}

func (ahs *AppHealthState) HTTPReady(c *gin.Context) {
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

func (ahs *AppHealthState) HTTPLiveness(c *gin.Context) {
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

type Status int

const (
	NotReady Status = iota
	Ready
	ShuttingDown
)

func (d Status) String() string {
	return [...]string{"NotReady", "Ready", "ShuttingDown"}[d]
}

// startTechnicalEndpoint exposes non-business routes.
// Technical routes are used for healthchecking, debugging, monitoring ...
// LBs, K8s, readiness probes should point to the port and /ready exposed by this http.Server
// Technical routes will be exposed on a different port.
// This allows : No extra LB conf to hide Technical services
// 				 A different Access Log, (it's useless to log health-checks)
func (ahs *AppHealthState) startTechnicalEndpoint(port, pprofEnabled string) {

	if port == "0" {
		// Do not expose tech routes
		return
	}

	// Technical Routes -----------------------------------

	ahs.log.Infow(fmt.Sprintf("Expose technical routes on :"+port),
		"component", "ahs")
	// routerTechnical will not be exposed publicly
	// But access by internal infrastructure things (HC, /metrics)
	routerTechnical := gin.Default()
	routerTechnical.GET("/metrics", gin.WrapH(promhttp.Handler()))
	routerTechnical.GET("/ready", ahs.HTTPReady)
	routerTechnical.GET("/health", ahs.HTTPReady)
	routerTechnical.GET("/live", ahs.HTTPLiveness)

	if pprofEnabled == "1" {
		// Add pprof routes
		pprof.Register(routerTechnical)
	}

	ahs.httpServerTech = &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      routerTechnical,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
		//BaseContext: func(_ net.Listener) context.Context { return ctxGS },
	}

	// Run server
	if err := ahs.httpServerTech.ListenAndServe(); err != http.ErrServerClosed {
		zap.S().Fatalf("HTTP TechEndpoint ListenAndServe: %v", err)
	}

}

func (ahs *AppHealthState) shutdownHandler() {

	// signChan channel is used to transmit signal notifications.
	signChan := make(chan os.Signal, 1)

	// Catch and relay certainF signal(s) to signChan channel.
	// GCP Cloud Run terminates the container with SIGTERM when down-scaling.
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Blocking until a signal is sent over signChan channel. Progress to
	// next line after signal
	sig := <-signChan

	zap.S().Warnf("Signal has been caught: %q. The goobs will now Shutdown Gracefully", sig)

	// Gracefully Shutdown in a precise order :
	// 0 - Set App "NotReady"
	// 1 - Vigie HTTP HTTP
	// 2 - DB Connection
	// 3 - HTTP Tech

	// Set app ShuttingDown
	ahs.askForShutdown = true
	ahs.status = ShuttingDown

	_ = ahs.webServer.GracefulShutdown()
	_ = ahs.db.GracefulShutdown()
	ahs.GracefulShutdown()

}

func (ahs *AppHealthState) GracefulShutdown() {

	// Close the Technical part after,
	// It might be interesting to grasp the last metrics of the app.

	ctxTOTech, cancelTOTech := context.WithTimeout(context.Background(), 2*time.Second)
	defer func() { cancelTOTech() }()

	if err := ahs.httpServerTech.Shutdown(ctxTOTech); err != nil {
		zap.S().Errorf("HTTP Technical Service failed to shutdown properly: %v", err)
		defer os.Exit(1)
		return
	} else {
		zap.S().Warn("HTTP Technical Service gracefully stopped\n")
	}
}

// Getter and Setter Like
// Because we must deal with the Mutex

func (ahs *AppHealthState) WebServerOK() {
	ahs.mu.Lock()
	defer ahs.mu.Unlock()
}

func (ahs *AppHealthState) WebServerNOK() {
	ahs.mu.Lock()
	defer ahs.mu.Unlock()
}

func (ahs *AppHealthState) DbOK() {
	ahs.mu.Lock()
	defer ahs.mu.Unlock()
}

func (ahs *AppHealthState) DbNOK() {
	ahs.mu.Lock()
	defer ahs.mu.Unlock()
}

func (ahs *AppHealthState) IsOK() bool {
	ahs.mu.Lock()
	defer ahs.mu.Unlock()

	if ahs.askForShutdown == false {
		return false
	}
	return true
}

func (ahs *AppHealthState) NotReady() {
	ahs.mu.Lock()
	defer ahs.mu.Unlock()
	ahs.status = NotReady
}

func (ahs *AppHealthState) Ready() {
	ahs.mu.Lock()
	defer ahs.mu.Unlock()
	ahs.status = Ready
}

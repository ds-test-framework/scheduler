package apiserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/ds-test-framework/scheduler/util"
	"github.com/gin-gonic/gin"
)

const DefaultAddr = "0.0.0.0:7074"

type APIServer struct {
	logger *log.Logger
	router *gin.Engine
	ctx    *types.Context
	gen    *util.IDGenerator

	server *http.Server

	addr string
}

func NewAPIServer(ctx *types.Context) *APIServer {
	config := ctx.Config("transport")
	config.SetDefault("addr", DefaultAddr)

	server := &APIServer{
		logger: ctx.Logger.With(log.LogParams{
			"service": "api-server",
		}),
		gen:  ctx.IDGen,
		ctx:  ctx,
		addr: config.GetString("addr"),
	}
	router := gin.New()
	router.Use(server.logMiddleware)

	router.POST("/message", server.HandleMessage)
	router.POST("/timeout", server.HandleTimeout)
	router.POST("/log", server.HandleLog)
	router.POST("/event", server.HandleEvent)
	router.POST("/replica", server.HandleReplica)

	router.GET("/runs", server.handleRun)
	router.GET("/run/:run/logs", server.handleRunLog)

	server.router = router
	server.server = &http.Server{
		Addr:    server.addr,
		Handler: router,
	}

	return server
}

func (a *APIServer) logMiddleware(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	// Process request
	c.Next()

	end := time.Now()
	if raw != "" {
		path = path + "?" + raw
	}
	a.logger.With(log.LogParams{
		"timestamp":   end,
		"latency":     end.Sub(start).String(),
		"client_ip":   c.ClientIP(),
		"method":      c.Request.Method,
		"status_code": c.Writer.Status(),
		"error":       c.Errors.ByType(gin.ErrorTypePrivate).String(),
		"body_size":   c.Writer.Size(),
		"path":        path,
	}).Info("Handled request")
}

func (a *APIServer) Start() {
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.With(log.LogParams{
				"addr": a.addr,
			}).Info("API server started!")
		}
	}()
}

func (a *APIServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("API server focefully shutdown")
	}
	a.logger.Info("API server stopped!")
}

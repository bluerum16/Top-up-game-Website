package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/irham/topup-backend/config"
	"github.com/irham/topup-backend/middleware"
)

type Deps struct {
	Cfg config.Config
	// Handlers will be added here as they are implemented:
	// Game    *handler.GameHandler
	// Order   *handler.OrderHandler
	// Payment *handler.PaymentHandler
}

func New(deps Deps) *gin.Engine {
	if deps.Cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(allowedOrigins(deps.Cfg)))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	registerPublicRoutes(v1, deps)
	registerProtectedRoutes(v1, deps)

	return r
}

func registerPublicRoutes(rg *gin.RouterGroup, deps Deps) {
	// rg.GET("/games", deps.Game.ListGames)
	// rg.GET("/games/:slug", deps.Game.GetGame)
	_ = rg
	_ = deps
}

func registerProtectedRoutes(rg *gin.RouterGroup, deps Deps) {
	auth := rg.Group("")
	auth.Use(middleware.RequireAuth(deps.Cfg.JWT))
	// auth.POST("/orders", deps.Order.Create)
	_ = auth
}

func allowedOrigins(cfg config.Config) []string {
	if cfg.App.Env == "production" {
		// TODO: ganti dengan domain production saat sudah siap deploy
		return []string{}
	}
	return []string{
		"http://localhost:3000",
		"http://localhost:3001",
	}
}

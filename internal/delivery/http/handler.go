package httpdelivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"API-techme/internal/token"
	"API-techme/internal/usecase"
)

// Handler menampung semua service + util yang dibutuhkan endpoint.
// main.go yang mengisinya (dependency injection).
type Handler struct {
	Auth   usecase.AuthService
	Tokens token.Manager
}

// Router mendaftarkan semua endpoint + Swagger UI.
func Router(h Handler) *gin.Engine {
	r := gin.Default()

	// Health check sederhana (tanpa auth).
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// --- Publik (tanpa token) ---
	v1.POST("/auth/login", h.Login)
	v1.POST("/auth/refresh", h.Refresh)

	// --- Butuh login (access token) ---
	authed := v1.Group("")
	authed.Use(RequireAuth(h.Tokens))
	authed.GET("/me", h.Me)

	// Swagger UI: http://localhost:8080/swagger/index.html
	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	return r
}

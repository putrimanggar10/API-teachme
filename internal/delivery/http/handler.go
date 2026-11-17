package httpdelivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"API-techme/internal/domain"
	"API-techme/internal/token"
	"API-techme/internal/usecase"
)

// Handler menampung semua service + util yang dibutuhkan endpoint.
// main.go yang mengisinya (dependency injection).
type Handler struct {
	Auth     usecase.AuthService
	Classes  usecase.ClassService
	Sessions usecase.SessionService
	Subjects usecase.SubjectService
	Tokens   token.Manager
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
	// Layar login: daftar guru & kelas untuk dropdown, lalu login.
	v1.GET("/accounts", h.ListAccounts)
	v1.POST("/accounts", h.CreateAccount) // halaman "Kelola Akun": tambah guru
	v1.GET("/accounts/:id/classes", h.AccountClasses)
	v1.POST("/accounts/:id/classes", h.AssignClass) // halaman "Manage accounts & classes": tugaskan guru (mapel + kelas)
	v1.GET("/school-classes", h.ListAllClasses)     // halaman "Manage accounts & classes": semua kelas
	v1.POST("/school-classes", h.CreateClass)       // halaman "Manage accounts & classes": tambah kelas
	v1.GET("/subjects", h.ListSubjects)             // halaman "Manage accounts & classes": semua mapel
	v1.POST("/subjects", h.CreateSubject)           // halaman "Manage accounts & classes": tambah mapel
	v1.GET("/languages", h.ListLanguages)           // layar "Choose board language"
	v1.POST("/auth/login", h.Login)
	v1.POST("/auth/refresh", h.Refresh)

	// --- Butuh login (access token) ---
	authed := v1.Group("")
	authed.Use(RequireAuth(h.Tokens))
	authed.GET("/me", h.Me)
	authed.PUT("/me/language", h.SetMyLanguage) // simpan bahasa papan pilihan

	// --- Khusus GURU (policy: RequireRole TEACHER) ---
	teacher := authed.Group("")
	teacher.Use(RequireRole(domain.RoleTeacher))
	teacher.GET("/classes", h.ListClasses)
	teacher.GET("/classes/:id/roster", h.ClassRoster)

	// Sesi kelas + join code + QR
	teacher.POST("/sessions", h.StartSession)
	teacher.GET("/sessions/:id", h.GetSession)
	teacher.GET("/sessions/:id/qr", h.SessionQR)
	teacher.POST("/sessions/:id/end", h.EndSession)

	// Swagger UI: http://localhost:8080/swagger/index.html
	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	return r
}

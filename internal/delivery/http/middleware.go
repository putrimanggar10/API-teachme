package httpdelivery

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"API-techme/internal/token"
)

// Kunci penyimpanan data di dalam gin.Context (hasil verifikasi token).
const (
	ctxUserID   = "auth_user_id"
	ctxRole     = "auth_role"
	ctxSchoolID = "auth_school_id"
)

// RequireAuth = middleware "penjaga pintu". Memastikan request membawa
// access token yang sah di header: Authorization: Bearer <token>.
// Kalau sah, identitas (user id, role, school) disimpan ke context untuk
// dipakai handler berikutnya.
func RequireAuth(tokens token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			fail(c, http.StatusUnauthorized, "UNAUTHENTICATED", "header Authorization: Bearer <token> wajib")
			return
		}

		claims, err := tokens.Parse(parts[1])
		if err != nil {
			fail(c, http.StatusUnauthorized, "UNAUTHENTICATED", "token tidak valid atau kedaluwarsa")
			return
		}
		// Token refresh TIDAK boleh dipakai untuk akses API biasa.
		if claims.Type != token.TypeAccess {
			fail(c, http.StatusUnauthorized, "UNAUTHENTICATED", "butuh access token, bukan refresh token")
			return
		}

		c.Set(ctxUserID, claims.Subject)
		c.Set(ctxRole, claims.Role)
		c.Set(ctxSchoolID, claims.SchoolID)
		c.Next()
	}
}

// RequireRole = middleware policy. Dipasang SETELAH RequireAuth.
// Hanya meneruskan request jika role pengguna cocok (mis. "TEACHER").
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(ctxRole) != role {
			fail(c, http.StatusForbidden, "FORBIDDEN", "aksi ini khusus role "+role)
			return
		}
		c.Next()
	}
}

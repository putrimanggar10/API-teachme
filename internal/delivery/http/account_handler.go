package httpdelivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListAccounts godoc
// @Summary  Daftar guru untuk dropdown "pilih nama" di layar login (publik)
// @Tags     Accounts
// @Produce  json
// @Success  200  {array}  map[string]interface{}
// @Router   /api/v1/accounts [get]
func (h Handler) ListAccounts(c *gin.Context) {
	teachers, err := h.Auth.ListTeachers(c.Request.Context())
	if err != nil {
		failErr(c, err)
		return
	}
	// Kirim hanya yang perlu untuk dropdown: id + nama.
	out := make([]gin.H, 0, len(teachers))
	for _, t := range teachers {
		out = append(out, gin.H{"id": t.ID, "name": t.Name})
	}
	c.JSON(http.StatusOK, out)
}

// AccountClasses godoc
// @Summary  Kelas yang diajar seorang guru (untuk dropdown "pilih kelas" saat login, publik)
// @Tags     Accounts
// @Produce  json
// @Param    id   path  string  true  "ID guru (UUID)"
// @Success  200  {array}  map[string]interface{}
// @Router   /api/v1/accounts/{id}/classes [get]
func (h Handler) AccountClasses(c *gin.Context) {
	teacherID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	classes, err := h.Classes.ListForTeacher(c.Request.Context(), teacherID)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, classes)
}

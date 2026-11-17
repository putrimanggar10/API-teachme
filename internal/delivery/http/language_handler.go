package httpdelivery

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"API-techme/internal/domain"
)

// ListLanguages godoc
// @Summary  Daftar bahasa papan (layar "Choose board language", publik)
// @Description Urut seperti di layar. Tampilkan `name` sebagai judul dan `english_name` sebagai keterangan kecil.
// @Tags     Languages
// @Produce  json
// @Success  200  {array}  map[string]interface{}
// @Router   /api/v1/languages [get]
func (h Handler) ListLanguages(c *gin.Context) {
	c.JSON(http.StatusOK, domain.Languages)
}

// setLanguageInput = isian saat guru menekan "Continue" di layar pilih bahasa.
type setLanguageInput struct {
	Language string `json:"language" example:"en"`
}

// SetMyLanguage godoc
// @Summary  Simpan bahasa papan pilihan pengguna yang sedang login
// @Description Dipanggil saat menekan "Continue". Pilihan ini ikut terkirim saat login berikutnya (user.preferred_language).
// @Tags     Languages
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      setLanguageInput  true  "kode bahasa: id, en, zh, es, ar"
// @Success  200   {object}  map[string]interface{}
// @Failure  400   {object}  map[string]interface{}
// @Failure  401   {object}  map[string]interface{}
// @Router   /api/v1/me/language [put]
func (h Handler) SetMyLanguage(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var in setLanguageInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "INVALID", "body JSON tidak valid (butuh language)")
		return
	}
	lang, err := h.Auth.SetLanguage(c.Request.Context(), userID, in.Language)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"preferred_language": lang.Code, "language": lang})
}

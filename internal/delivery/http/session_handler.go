package httpdelivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"

	"API-techme/internal/usecase"
)

type startSessionInput struct {
	ClassID string `json:"class_id" binding:"required"`
}

// StartSession godoc
// @Summary  Mulai sesi kelas -> hasilkan join_code + qr_payload
// @Tags     Sessions
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body  startSessionInput  true  "ID kelas"
// @Success  201   {object}  map[string]interface{}
// @Failure  403   {object}  map[string]interface{}
// @Router   /api/v1/sessions [post]
func (h Handler) StartSession(c *gin.Context) {
	teacherID, ok := currentUserID(c)
	if !ok {
		return
	}
	var in startSessionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "INVALID", err.Error())
		return
	}
	classID, ok := parseUUID(c, in.ClassID)
	if !ok {
		return
	}

	sess, err := h.Sessions.Start(c.Request.Context(), teacherID, classID)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"session_id":  sess.ID,
		"class_id":    sess.ClassID,
		"join_code":   sess.JoinCode,              // ditampilkan board ("type the code instead")
		"qr_payload":  usecase.QRPayload(sess.ID), // isi QR
		"qr_image_url": "/api/v1/sessions/" + sess.ID.String() + "/qr",
		"status":      sess.Status,
	})
}

// GetSession godoc
// @Summary  Info sesi
// @Tags     Sessions
// @Produce  json
// @Security BearerAuth
// @Param    id   path  string  true  "ID sesi (UUID)"
// @Success  200  {object}  map[string]interface{}
// @Failure  404  {object}  map[string]interface{}
// @Router   /api/v1/sessions/{id} [get]
func (h Handler) GetSession(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	sess, err := h.Sessions.Get(c.Request.Context(), id)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sess)
}

// SessionQR godoc
// @Summary  Gambar QR untuk bergabung ke sesi (PNG)
// @Tags     Sessions
// @Produce  png
// @Security BearerAuth
// @Param    id   path  string  true  "ID sesi (UUID)"
// @Success  200  {file}    binary
// @Failure  404  {object}  map[string]interface{}
// @Router   /api/v1/sessions/{id}/qr [get]
func (h Handler) SessionQR(c *gin.Context) {
	teacherID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	sess, err := h.Sessions.Get(c.Request.Context(), id)
	if err != nil {
		failErr(c, err)
		return
	}
	// Hanya guru pemilik yang boleh mengambil QR sesi ini.
	if sess.TeacherID != teacherID {
		fail(c, http.StatusForbidden, "FORBIDDEN", "bukan pemilik sesi")
		return
	}

	png, err := qrcode.Encode(usecase.QRPayload(sess.ID), qrcode.Medium, 256)
	if err != nil {
		fail(c, http.StatusInternalServerError, "INTERNAL", "gagal membuat QR")
		return
	}
	c.Data(http.StatusOK, "image/png", png)
}

// EndSession godoc
// @Summary  Akhiri sesi (kode jadi tidak berlaku)
// @Tags     Sessions
// @Produce  json
// @Security BearerAuth
// @Param    id   path  string  true  "ID sesi (UUID)"
// @Success  200  {object}  map[string]interface{}
// @Failure  404  {object}  map[string]interface{}
// @Router   /api/v1/sessions/{id}/end [post]
func (h Handler) EndSession(c *gin.Context) {
	teacherID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.Sessions.End(c.Request.Context(), teacherID, id); err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ENDED"})
}

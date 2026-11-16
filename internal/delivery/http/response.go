// Package httpdelivery mengurus HTTP: baca JSON, tentukan status code, routing.
// Tidak ada aturan bisnis atau SQL di sini.
package httpdelivery

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"API-techme/internal/domain"
)

// fail menulis error dengan amplop standar sesuai PAD:
//
//	{"error":{"code":"...","message":"..."},"request_id":"..."}
func fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error":      gin.H{"code": code, "message": message},
		"request_id": uuid.NewString(),
	})
}

// currentUserID mengambil ID pengguna (dari token) yang sudah disimpan middleware.
func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString(ctxUserID))
	if err != nil {
		fail(c, http.StatusUnauthorized, "UNAUTHENTICATED", "token rusak")
		return uuid.Nil, false
	}
	return id, true
}

// parseUUIDParam membaca parameter path (mis. :id) dan memastikan bentuk UUID.
func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID", name+" harus UUID valid")
		return uuid.Nil, false
	}
	return id, true
}

// parseUUID mengubah string (mis. dari body JSON) menjadi UUID.
func parseUUID(c *gin.Context, s string) (uuid.UUID, bool) {
	id, err := uuid.Parse(s)
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID", "nilai bukan UUID valid")
		return uuid.Nil, false
	}
	return id, true
}

// failErr memetakan error bisnis domain ke status + kode HTTP yang tepat.
func failErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
	case errors.Is(err, domain.ErrForbidden):
		fail(c, http.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, domain.ErrNotFound):
		fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrConflict):
		fail(c, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, domain.ErrInvalid):
		fail(c, http.StatusBadRequest, "INVALID", err.Error())
	default:
		fail(c, http.StatusInternalServerError, "INTERNAL", "internal server error")
	}
}

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

package httpdelivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"API-techme/internal/usecase"
)

// ---------------------------------------------------------------------------
// Bentuk input (request body)
// ---------------------------------------------------------------------------

// loginInput: guru pilih namanya (account_id), ketik password, pilih "Mapel · Kelas"
// (dari dropdown: class_id + subject_id).
type loginInput struct {
	AccountID string `json:"account_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
	ClassID   string `json:"class_id" binding:"required"`
	SubjectID string `json:"subject_id" binding:"required"`
}

type refreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// authResponse = bentuk balasan login/refresh.
func authResponse(c *gin.Context, status int, r *usecase.AuthResult) {
	c.JSON(status, gin.H{
		"access_token":   r.AccessToken,
		"refresh_token":  r.RefreshToken,
		"token_type":     "Bearer",
		"expires_in":     r.ExpiresIn,
		"user":           r.User,
		"active_class":   r.ActiveClass,   // kelas yang dipilih saat login (null saat refresh)
		"active_subject": r.ActiveSubject, // mapel yang dipilih saat login (null saat refresh)
		"roles":          []string{r.User.Role},
	})
}

// Login godoc
// @Summary  Login (guru/siswa) -> access + refresh token
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    body  body  loginInput  true  "username & password"
// @Success  200   {object}  map[string]interface{}
// @Failure  401   {object}  map[string]interface{}
// @Router   /api/v1/auth/login [post]
func (h Handler) Login(c *gin.Context) {
	var in loginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "INVALID", err.Error())
		return
	}
	accountID, ok := parseUUID(c, in.AccountID)
	if !ok {
		return
	}
	classID, ok := parseUUID(c, in.ClassID)
	if !ok {
		return
	}
	subjectID, ok := parseUUID(c, in.SubjectID)
	if !ok {
		return
	}
	res, err := h.Auth.Login(c.Request.Context(), accountID, in.Password, classID, subjectID)
	if err != nil {
		failErr(c, err)
		return
	}
	authResponse(c, http.StatusOK, res)
}

// Refresh godoc
// @Summary  Tukar refresh token dengan access token baru
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    body  body  refreshInput  true  "refresh_token"
// @Success  200   {object}  map[string]interface{}
// @Failure  401   {object}  map[string]interface{}
// @Router   /api/v1/auth/refresh [post]
func (h Handler) Refresh(c *gin.Context) {
	var in refreshInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "INVALID", err.Error())
		return
	}
	res, err := h.Auth.Refresh(c.Request.Context(), in.RefreshToken)
	if err != nil {
		failErr(c, err)
		return
	}
	authResponse(c, http.StatusOK, res)
}

// Me godoc
// @Summary  Profil pengguna yang sedang login
// @Tags     Auth
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  map[string]interface{}
// @Failure  401  {object}  map[string]interface{}
// @Router   /api/v1/me [get]
func (h Handler) Me(c *gin.Context) {
	// ctxUserID diisi oleh middleware RequireAuth (dari token).
	id, err := uuid.Parse(c.GetString(ctxUserID))
	if err != nil {
		fail(c, http.StatusUnauthorized, "UNAUTHENTICATED", "token rusak")
		return
	}
	user, err := h.Auth.Me(c.Request.Context(), id)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

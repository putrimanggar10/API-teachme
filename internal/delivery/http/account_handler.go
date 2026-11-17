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

// createAccountInput = isian form "Tambah Guru" (hanya nama + password).
type createAccountInput struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// CreateAccount godoc
// @Summary  Tambah guru baru (halaman "Kelola Akun", publik)
// @Tags     Accounts
// @Accept   json
// @Produce  json
// @Param    body  body      createAccountInput  true  "nama + password guru baru"
// @Success  201   {object}  map[string]interface{}
// @Failure  400   {object}  map[string]interface{}
// @Router   /api/v1/accounts [post]
func (h Handler) CreateAccount(c *gin.Context) {
	var in createAccountInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "INVALID", "body JSON tidak valid (butuh name & password)")
		return
	}
	teacher, err := h.Auth.AddTeacher(c.Request.Context(), in.Name, in.Password)
	if err != nil {
		failErr(c, err)
		return
	}
	// Username dikirim balik sebagai info; untuk login nanti tetap pilih NAMA (bukan ketik username).
	c.JSON(http.StatusCreated, gin.H{
		"id":       teacher.ID,
		"name":     teacher.Name,
		"username": teacher.Username,
	})
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

// assignClassInput = isian "tugaskan guru ke kelas".
type assignClassInput struct {
	ClassID string `json:"class_id"`
}

// AssignClass godoc
// @Summary  Tugaskan guru ke kelas (halaman "Kelola Akun & Kelas", publik)
// @Description Setelah ini guru bisa login dengan memilih kelas tsb. Aman dipanggil berulang (tidak dobel).
// @Tags     Accounts
// @Accept   json
// @Produce  json
// @Param    id    path      string            true  "ID guru (UUID)"
// @Param    body  body      assignClassInput  true  "kelas yang ditugaskan"
// @Success  201   {object}  map[string]interface{}
// @Failure  400   {object}  map[string]interface{}
// @Failure  404   {object}  map[string]interface{}
// @Router   /api/v1/accounts/{id}/classes [post]
func (h Handler) AssignClass(c *gin.Context) {
	teacherID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var in assignClassInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "INVALID", "body JSON tidak valid (butuh class_id)")
		return
	}
	classID, ok := parseUUID(c, in.ClassID)
	if !ok {
		return
	}
	class, err := h.Classes.AssignTeacher(c.Request.Context(), teacherID, classID)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"teacher_id": teacherID,
		"class":      class,
	})
}

// ListAllClasses godoc
// @Summary  Semua kelas di sekolah (halaman "Kelola Akun & Kelas", publik)
// @Tags     Accounts
// @Produce  json
// @Success  200  {array}  map[string]interface{}
// @Router   /api/v1/school-classes [get]
func (h Handler) ListAllClasses(c *gin.Context) {
	classes, err := h.Classes.ListAll(c.Request.Context())
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, classes)
}

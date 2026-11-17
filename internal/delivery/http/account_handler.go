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
// @Summary  Pilihan "Mapel · Kelas" seorang guru (dropdown "Class" saat login, publik)
// @Description Tiap item = 1 pasangan mapel+kelas, mis. label "Math · Class 7B". Kirim class_id + subject_id-nya saat login.
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
	options, err := h.Classes.ListOptionsForTeacher(c.Request.Context(), teacherID)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, options)
}

// assignClassInput = isian "tugaskan guru mengajar mapel di kelas".
type assignClassInput struct {
	ClassID   string `json:"class_id"`
	SubjectID string `json:"subject_id"`
}

// AssignClass godoc
// @Summary  Tugaskan guru mengajar mapel di kelas (halaman "Manage accounts & classes", publik)
// @Description Setelah ini guru bisa login memilih "Mapel · Kelas" tsb. Aman dipanggil berulang (tidak dobel).
// @Tags     Accounts
// @Accept   json
// @Produce  json
// @Param    id    path      string            true  "ID guru (UUID)"
// @Param    body  body      assignClassInput  true  "kelas + mapel yang ditugaskan"
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
	if err := c.ShouldBindJSON(&in); err != nil || in.ClassID == "" || in.SubjectID == "" {
		fail(c, http.StatusBadRequest, "INVALID", "body JSON tidak valid (butuh class_id & subject_id)")
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
	a, err := h.Classes.AssignTeacher(c.Request.Context(), teacherID, classID, subjectID)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"teacher_id": teacherID,
		"class":      a.Class,
		"subject":    a.Subject,
		"label":      a.Subject.Name + " · " + a.Class.Name,
	})
}

// createClassInput = isian "Tambah Kelas". grade & academic_year opsional.
type createClassInput struct {
	Name         string `json:"name" example:"Class 8C"`
	Grade        string `json:"grade" example:"8"`
	AcademicYear string `json:"academic_year" example:"2026/2027"`
}

// CreateClass godoc
// @Summary  Tambah kelas baru (halaman "Manage accounts & classes", publik)
// @Tags     Accounts
// @Accept   json
// @Produce  json
// @Param    body  body      createClassInput  true  "nama kelas (wajib), tingkat & tahun ajaran (opsional)"
// @Success  201   {object}  map[string]interface{}
// @Failure  400   {object}  map[string]interface{}
// @Failure  409   {object}  map[string]interface{}  "nama kelas sudah ada"
// @Router   /api/v1/school-classes [post]
func (h Handler) CreateClass(c *gin.Context) {
	var in createClassInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "INVALID", "body JSON tidak valid (butuh name)")
		return
	}
	class, err := h.Classes.CreateClass(c.Request.Context(), in.Name, in.Grade, in.AcademicYear)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, class)
}

// ListAllClasses godoc
// @Summary  Semua kelas di sekolah (halaman "Manage accounts & classes", publik)
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

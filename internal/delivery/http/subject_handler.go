package httpdelivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListSubjects godoc
// @Summary  Daftar mata pelajaran (halaman "Manage accounts & classes", publik)
// @Tags     Subjects
// @Produce  json
// @Success  200  {array}  map[string]interface{}
// @Router   /api/v1/subjects [get]
func (h Handler) ListSubjects(c *gin.Context) {
	subjects, err := h.Subjects.List(c.Request.Context())
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, subjects)
}

// createSubjectInput = isian "Tambah Mata Pelajaran".
type createSubjectInput struct {
	Name string `json:"name"`
}

// CreateSubject godoc
// @Summary  Tambah mata pelajaran (halaman "Manage accounts & classes", publik)
// @Tags     Subjects
// @Accept   json
// @Produce  json
// @Param    body  body      createSubjectInput  true  "nama mapel, mis. Math"
// @Success  201   {object}  map[string]interface{}
// @Failure  400   {object}  map[string]interface{}
// @Failure  409   {object}  map[string]interface{}  "nama mapel sudah ada"
// @Router   /api/v1/subjects [post]
func (h Handler) CreateSubject(c *gin.Context) {
	var in createSubjectInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "INVALID", "body JSON tidak valid (butuh name)")
		return
	}
	sub, err := h.Subjects.Create(c.Request.Context(), in.Name)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, sub)
}

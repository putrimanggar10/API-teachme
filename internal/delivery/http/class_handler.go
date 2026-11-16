package httpdelivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListClasses godoc
// @Summary  Daftar kelas yang boleh diajar guru (untuk layar "pilih kelas")
// @Tags     Classes
// @Produce  json
// @Security BearerAuth
// @Success  200  {array}   map[string]interface{}
// @Failure  401  {object}  map[string]interface{}
// @Failure  403  {object}  map[string]interface{}
// @Router   /api/v1/classes [get]
func (h Handler) ListClasses(c *gin.Context) {
	teacherID, ok := currentUserID(c)
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

// ClassRoster godoc
// @Summary  Daftar siswa pada sebuah kelas (khusus guru pengampu)
// @Tags     Classes
// @Produce  json
// @Security BearerAuth
// @Param    id   path      string  true  "ID kelas (UUID)"
// @Success  200  {array}   map[string]interface{}
// @Failure  403  {object}  map[string]interface{}
// @Router   /api/v1/classes/{id}/roster [get]
func (h Handler) ClassRoster(c *gin.Context) {
	teacherID, ok := currentUserID(c)
	if !ok {
		return
	}
	classID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	students, err := h.Classes.Roster(c.Request.Context(), teacherID, classID)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, students)
}

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"API-techme/internal/domain"
)

// ClassRepo membaca tabel classes + relasinya.
type ClassRepo struct {
	DB *sql.DB
}

// nullStr membantu ubah kolom nullable (grade, academic_year) jadi *string.
func nullStr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}

// ListForTeacher: JOIN classes dengan teacher_class_assignments agar hanya
// kelas yang di-assign ke guru tsb (dan masih aktif) yang muncul.
func (r ClassRepo) ListForTeacher(ctx context.Context, teacherID uuid.UUID) ([]domain.Class, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT c.id, c.school_id, c.name, c.grade, c.academic_year, c.created_at
		FROM classes c
		JOIN teacher_class_assignments t ON t.class_id = c.id
		WHERE t.teacher_user_id = $1 AND t.active = TRUE
		ORDER BY c.name`, teacherID)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	result := make([]domain.Class, 0)
	for rows.Next() {
		var c domain.Class
		var grade, year sql.NullString
		if err := rows.Scan(&c.ID, &c.SchoolID, &c.Name, &grade, &year, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Grade = nullStr(grade)
		c.AcademicYear = nullStr(year)
		result = append(result, c)
	}
	return result, rows.Err()
}

// Roster: daftar siswa (role STUDENT) yang terdaftar aktif di kelas.
func (r ClassRepo) Roster(ctx context.Context, classID uuid.UUID) ([]domain.User, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT u.id, u.school_id, u.name, u.username, u.role, u.status, u.created_at
		FROM users u
		JOIN class_enrollments e ON e.student_user_id = u.id
		WHERE e.class_id = $1 AND e.status = 'ACTIVE'
		ORDER BY u.name`, classID)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	result := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.SchoolID, &u.Name, &u.Username, &u.Role, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
}

// IsTeacherAssigned: cek apakah guru berwenang atas kelas (untuk policy).
func (r ClassRepo) IsTeacherAssigned(ctx context.Context, classID, teacherID uuid.UUID) (bool, error) {
	var exists bool
	err := r.DB.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM teacher_class_assignments
			WHERE class_id = $1 AND teacher_user_id = $2 AND active = TRUE
		)`, classID, teacherID).Scan(&exists)
	return exists, translate(err)
}

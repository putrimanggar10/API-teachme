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
	return r.queryClasses(ctx, `
		SELECT c.id, c.school_id, c.name, c.grade, c.academic_year, c.created_at
		FROM classes c
		JOIN teacher_class_assignments t ON t.class_id = c.id
		WHERE t.teacher_user_id = $1 AND t.active = TRUE
		ORDER BY c.name`, teacherID)
}

// ListAll: semua kelas (untuk halaman "Kelola Akun & Kelas").
func (r ClassRepo) ListAll(ctx context.Context) ([]domain.Class, error) {
	return r.queryClasses(ctx, `
		SELECT id, school_id, name, grade, academic_year, created_at
		FROM classes
		ORDER BY name`)
}

// Create: tambah kelas baru di sekolah pertama (demo ini hanya 1 sekolah).
// Tabel classes (schema resmi) tidak punya UNIQUE untuk nama, jadi pengecekan nama
// kembar dilakukan di query: INSERT hanya terjadi kalau nama belum ada.
// Kalau nama sudah ada, tidak ada baris yang dikembalikan -> ErrConflict.
func (r ClassRepo) Create(ctx context.Context, name string, grade, academicYear *string) (*domain.Class, error) {
	var c domain.Class
	var g, y sql.NullString
	err := r.DB.QueryRowContext(ctx, `
		WITH sch AS (SELECT id FROM schools ORDER BY created_at LIMIT 1)
		INSERT INTO classes (school_id, name, grade, academic_year)
		SELECT sch.id, $1, $2, $3 FROM sch
		WHERE NOT EXISTS (
			SELECT 1 FROM classes x WHERE x.school_id = sch.id AND lower(x.name) = lower($1)
		)
		RETURNING id, school_id, name, grade, academic_year, created_at`,
		name, grade, academicYear,
	).Scan(&c.ID, &c.SchoolID, &c.Name, &g, &y, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrConflict
	}
	if err != nil {
		return nil, translate(err)
	}
	c.Grade = nullStr(g)
	c.AcademicYear = nullStr(y)
	return &c, nil
}

// AssignTeacher: simpan penugasan guru mengajar mapel di kelas.
// Dua tabel diisi dalam 1 transaksi:
//   - teacher_class_assignments: guru berwenang atas kelas (dipakai policy roster/sesi)
//   - teacher_class_subjects   : guru mengajar mapel apa di kelas itu (dipakai login)
//
// ON CONFLICT: kalau sudah ada, cukup set active = TRUE lagi (tidak dobel).
func (r ClassRepo) AssignTeacher(ctx context.Context, classID, teacherID, subjectID uuid.UUID) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO teacher_class_assignments (class_id, teacher_user_id)
		VALUES ($1, $2)
		ON CONFLICT (class_id, teacher_user_id) DO UPDATE SET active = TRUE`,
		classID, teacherID); err != nil {
		return translate(err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO teacher_class_subjects (teacher_user_id, class_id, subject_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (teacher_user_id, class_id, subject_id) DO UPDATE SET active = TRUE`,
		teacherID, classID, subjectID); err != nil {
		return translate(err)
	}
	return tx.Commit()
}

// ListOptionsForTeacher: pasangan mapel+kelas yang diajar guru, untuk dropdown login
// ("Math · Class 7B"). Urut nama kelas lalu nama mapel.
func (r ClassRepo) ListOptionsForTeacher(ctx context.Context, teacherID uuid.UUID) ([]domain.TeachingOption, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT c.id, c.name, s.id, s.name
		FROM teacher_class_subjects t
		JOIN classes  c ON c.id = t.class_id
		JOIN subjects s ON s.id = t.subject_id
		WHERE t.teacher_user_id = $1 AND t.active = TRUE
		ORDER BY c.name, lower(s.name)`, teacherID)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	result := make([]domain.TeachingOption, 0)
	for rows.Next() {
		var o domain.TeachingOption
		if err := rows.Scan(&o.ClassID, &o.ClassName, &o.SubjectID, &o.SubjectName); err != nil {
			return nil, err
		}
		o.Label = o.SubjectName + " · " + o.ClassName
		result = append(result, o)
	}
	return result, rows.Err()
}

// IsTeachingSubject: cek guru mengajar mapel tsb di kelas tsb (masih aktif).
func (r ClassRepo) IsTeachingSubject(ctx context.Context, teacherID, classID, subjectID uuid.UUID) (bool, error) {
	var exists bool
	err := r.DB.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM teacher_class_subjects
			WHERE teacher_user_id = $1 AND class_id = $2 AND subject_id = $3 AND active = TRUE
		)`, teacherID, classID, subjectID).Scan(&exists)
	return exists, translate(err)
}

// queryClasses menjalankan query SELECT kelas lalu membaca hasilnya jadi []domain.Class.
// Kolom query wajib urut: id, school_id, name, grade, academic_year, created_at.
func (r ClassRepo) queryClasses(ctx context.Context, query string, args ...any) ([]domain.Class, error) {
	rows, err := r.DB.QueryContext(ctx, query, args...)
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

// GetByID: ambil 1 kelas berdasarkan ID.
func (r ClassRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	var c domain.Class
	var grade, year sql.NullString
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, school_id, name, grade, academic_year, created_at
		FROM classes WHERE id = $1`, id,
	).Scan(&c.ID, &c.SchoolID, &c.Name, &grade, &year, &c.CreatedAt)
	if err != nil {
		return nil, translate(err)
	}
	c.Grade = nullStr(grade)
	c.AcademicYear = nullStr(year)
	return &c, nil
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

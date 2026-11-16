// Package domain berisi bentuk data (entity) dan kontrak (interface).
// Layer ini TIDAK boleh mengimpor Gin atau driver PostgreSQL.
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Error bisnis. Repository & handler memetakan ini ke status HTTP yang tepat.
var (
	ErrNotFound           = errors.New("data tidak ditemukan")
	ErrConflict           = errors.New("data duplikat atau bentrok")
	ErrInvalid            = errors.New("data tidak valid")
	ErrInvalidCredentials = errors.New("nama guru atau password salah")
	ErrForbidden          = errors.New("tidak punya hak akses")
)

// Role = peran pengguna. Dipakai untuk policy (teacher/student).
const (
	RoleTeacher = "TEACHER"
	RoleStudent = "STUDENT"
	RoleAdmin   = "ADMIN"
)

// ---------------------------------------------------------------------------
// User (guru / siswa / admin)
// ---------------------------------------------------------------------------

// User adalah identitas pengguna.
// PasswordHash diberi tag `json:"-"` supaya TIDAK PERNAH ikut terkirim ke klien.
type User struct {
	ID        uuid.UUID `json:"id"`
	SchoolID  uuid.UUID `json:"school_id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	// Catatan: schema resmi `users` TIDAK punya kolom subject/mapel.
	// "Mr. Axel · Math" di gambar belum ada tempat simpannya — gap yang dicatat.

	PasswordHash string `json:"-"` // rahasia: hanya dipakai internal saat verifikasi login
}

// UserRepository adalah kontrak penyimpanan User (dimiliki domain, bukan PostgreSQL).
type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	// ListTeachers: daftar guru (untuk dropdown "pilih nama" di layar login).
	ListTeachers(ctx context.Context) ([]User, error)
}

// ---------------------------------------------------------------------------
// Class (kelas) + roster
// ---------------------------------------------------------------------------

// Class adalah entity kelas akademik (mis. "Class 7B").
type Class struct {
	ID           uuid.UUID `json:"id"`
	SchoolID     uuid.UUID `json:"school_id"`
	Name         string    `json:"name"`
	Grade        *string   `json:"grade,omitempty"`
	AcademicYear *string   `json:"academic_year,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ClassRepository = kontrak penyimpanan kelas.
type ClassRepository interface {
	// ListForTeacher: kelas yang BOLEH diajar guru tsb (via teacher_class_assignments).
	ListForTeacher(ctx context.Context, teacherID uuid.UUID) ([]Class, error)
	// Roster: daftar siswa pada satu kelas.
	Roster(ctx context.Context, classID uuid.UUID) ([]User, error)
	// IsTeacherAssigned: apakah guru berwenang atas kelas ini (untuk policy).
	IsTeacherAssigned(ctx context.Context, classID, teacherID uuid.UUID) (bool, error)
	// GetByID: ambil 1 kelas berdasarkan ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Class, error)
}

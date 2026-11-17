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
	// PreferredLanguage = bahasa papan pilihan terakhir (kode: id, en, zh, es, ar).
	// Kolom tambahan dari migrations/003_user_language.sql.
	PreferredLanguage string `json:"preferred_language,omitempty"`

	PasswordHash string `json:"-"` // rahasia: hanya dipakai internal saat verifikasi login
}

// UserRepository adalah kontrak penyimpanan User (dimiliki domain, bukan PostgreSQL).
type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	// ListTeachers: daftar guru (untuk dropdown "pilih nama" di layar login).
	ListTeachers(ctx context.Context) ([]User, error)
	// CreateTeacher: tambah guru baru dari halaman "Kelola Akun".
	// Butuh nama + username (dibuat dari nama) + hash password. school_id diisi otomatis.
	CreateTeacher(ctx context.Context, name, username, passwordHash string) (*User, error)
	// UpdateLanguage: simpan bahasa papan pilihan user.
	UpdateLanguage(ctx context.Context, id uuid.UUID, code string) error
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
	// ListAll: semua kelas di sekolah (untuk halaman "Manage accounts & classes").
	ListAll(ctx context.Context) ([]Class, error)
	// Create: tambah kelas baru. Nama sudah ada (tanpa peduli huruf besar/kecil) -> ErrConflict.
	Create(ctx context.Context, name string, grade, academicYear *string) (*Class, error)
	// AssignTeacher: tugaskan guru mengajar MAPEL di KELAS.
	// Mengisi teacher_class_assignments + teacher_class_subjects.
	// Kalau sudah pernah ditugaskan, cukup diaktifkan lagi (tidak dobel).
	AssignTeacher(ctx context.Context, classID, teacherID, subjectID uuid.UUID) error
	// ListOptionsForTeacher: pasangan mapel+kelas yang diajar guru (dropdown login).
	ListOptionsForTeacher(ctx context.Context, teacherID uuid.UUID) ([]TeachingOption, error)
	// IsTeachingSubject: apakah guru mengajar mapel tsb di kelas tsb (untuk login).
	IsTeachingSubject(ctx context.Context, teacherID, classID, subjectID uuid.UUID) (bool, error)
}

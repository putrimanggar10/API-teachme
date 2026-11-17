package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"API-techme/internal/domain"
)

// ClassService mengurus kelas: daftar kelas guru & roster siswa.
type ClassService struct {
	Repo     domain.ClassRepository
	Users    domain.UserRepository    // untuk cek akun yang ditugaskan memang guru
	Subjects domain.SubjectRepository // untuk cek mapel yang ditugaskan memang ada
}

// Assignment = hasil penugasan: guru mengajar mapel apa di kelas mana.
type Assignment struct {
	Class   *domain.Class
	Subject *domain.Subject
}

// ListOptionsForTeacher: pilihan dropdown "Class" di layar login ("Math · Class 7B").
func (s ClassService) ListOptionsForTeacher(ctx context.Context, teacherID uuid.UUID) ([]domain.TeachingOption, error) {
	return s.Repo.ListOptionsForTeacher(ctx, teacherID)
}

// CreateClass: tambah kelas baru (halaman "Manage accounts & classes").
// Nama wajib & tidak boleh kembar. grade & academicYear opsional (kosong = tidak diisi).
func (s ClassService) CreateClass(ctx context.Context, name, grade, academicYear string) (*domain.Class, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: nama kelas wajib diisi", domain.ErrInvalid)
	}
	if len([]rune(name)) > 50 {
		return nil, fmt.Errorf("%w: nama kelas maksimal 50 karakter", domain.ErrInvalid)
	}
	c, err := s.Repo.Create(ctx, name, optional(grade), optional(academicYear))
	if err == domain.ErrConflict {
		return nil, fmt.Errorf("%w: kelas %q sudah ada", domain.ErrConflict, name)
	}
	return c, err
}

// optional: string kosong (setelah dirapikan) -> nil, supaya tersimpan NULL di DB.
func optional(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

// ListAll: semua kelas (untuk halaman "Manage accounts & classes").
func (s ClassService) ListAll(ctx context.Context) ([]domain.Class, error) {
	return s.Repo.ListAll(ctx)
}

// AssignTeacher: tugaskan guru mengajar mapel di kelas, supaya guru itu bisa login
// memilih "Mapel · Kelas" tsb.
// Aturan: akun harus ada & ber-role TEACHER, kelas dan mapel harus ada.
func (s ClassService) AssignTeacher(ctx context.Context, teacherID, classID, subjectID uuid.UUID) (*Assignment, error) {
	user, err := s.Users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, err // ErrNotFound -> 404
	}
	if user.Role != domain.RoleTeacher {
		return nil, fmt.Errorf("%w: akun ini bukan guru", domain.ErrInvalid)
	}
	class, err := s.Repo.GetByID(ctx, classID)
	if err != nil {
		return nil, err // ErrNotFound -> 404
	}
	subject, err := s.Subjects.GetByID(ctx, subjectID)
	if err != nil {
		return nil, err // ErrNotFound -> 404
	}
	if err := s.Repo.AssignTeacher(ctx, classID, teacherID, subjectID); err != nil {
		return nil, err
	}
	return &Assignment{Class: class, Subject: subject}, nil
}

// ListForTeacher: kelas-kelas yang boleh diajar guru (dipakai layar "pilih kelas").
func (s ClassService) ListForTeacher(ctx context.Context, teacherID uuid.UUID) ([]domain.Class, error) {
	return s.Repo.ListForTeacher(ctx, teacherID)
}

// Roster: daftar siswa pada satu kelas.
// POLICY: guru harus berwenang atas kelas itu, kalau tidak -> ErrForbidden.
func (s ClassService) Roster(ctx context.Context, teacherID, classID uuid.UUID) ([]domain.User, error) {
	ok, err := s.Repo.IsTeacherAssigned(ctx, classID, teacherID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}
	return s.Repo.Roster(ctx, classID)
}

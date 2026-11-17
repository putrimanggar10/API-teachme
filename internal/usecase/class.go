package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"API-techme/internal/domain"
)

// ClassService mengurus kelas: daftar kelas guru & roster siswa.
type ClassService struct {
	Repo  domain.ClassRepository
	Users domain.UserRepository // untuk cek akun yang ditugaskan memang guru
}

// ListAll: semua kelas (untuk halaman "Kelola Akun & Kelas").
func (s ClassService) ListAll(ctx context.Context) ([]domain.Class, error) {
	return s.Repo.ListAll(ctx)
}

// AssignTeacher: tugaskan guru ke kelas, supaya guru itu bisa login memilih kelas tsb.
// Aturan: akun harus ada & ber-role TEACHER, kelas harus ada.
func (s ClassService) AssignTeacher(ctx context.Context, teacherID, classID uuid.UUID) (*domain.Class, error) {
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
	if err := s.Repo.AssignTeacher(ctx, classID, teacherID); err != nil {
		return nil, err
	}
	return class, nil
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

package usecase

import (
	"context"

	"github.com/google/uuid"

	"API-techme/internal/domain"
)

// ClassService mengurus kelas: daftar kelas guru & roster siswa.
type ClassService struct {
	Repo domain.ClassRepository
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

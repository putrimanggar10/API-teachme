package usecase

import (
	"context"
	"fmt"
	"strings"

	"API-techme/internal/domain"
)

// SubjectService mengurus mata pelajaran (halaman "Manage accounts & classes").
type SubjectService struct {
	Repo domain.SubjectRepository
}

// List: semua mapel.
func (s SubjectService) List(ctx context.Context) ([]domain.Subject, error) {
	return s.Repo.List(ctx)
}

// Create: tambah mapel baru. Nama wajib diisi dan tidak boleh kembar.
func (s SubjectService) Create(ctx context.Context, name string) (*domain.Subject, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: nama mata pelajaran wajib diisi", domain.ErrInvalid)
	}
	if len([]rune(name)) > 50 {
		return nil, fmt.Errorf("%w: nama mata pelajaran maksimal 50 karakter", domain.ErrInvalid)
	}
	sub, err := s.Repo.Create(ctx, name)
	if err == domain.ErrConflict {
		return nil, fmt.Errorf("%w: mata pelajaran %q sudah ada", domain.ErrConflict, name)
	}
	return sub, err
}

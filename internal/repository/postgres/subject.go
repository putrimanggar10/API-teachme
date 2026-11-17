package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"API-techme/internal/domain"
)

// SubjectRepo membaca tabel subjects (mata pelajaran).
type SubjectRepo struct {
	DB *sql.DB
}

func (r SubjectRepo) List(ctx context.Context) ([]domain.Subject, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, school_id, name, created_at
		FROM subjects
		ORDER BY lower(name)`)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	result := make([]domain.Subject, 0)
	for rows.Next() {
		var s domain.Subject
		if err := rows.Scan(&s.ID, &s.SchoolID, &s.Name, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// Create menyimpan mapel baru di sekolah pertama (demo ini hanya 1 sekolah).
func (r SubjectRepo) Create(ctx context.Context, name string) (*domain.Subject, error) {
	var s domain.Subject
	err := r.DB.QueryRowContext(ctx, `
		INSERT INTO subjects (school_id, name)
		VALUES ((SELECT id FROM schools ORDER BY created_at LIMIT 1), $1)
		RETURNING id, school_id, name, created_at`, name,
	).Scan(&s.ID, &s.SchoolID, &s.Name, &s.CreatedAt)
	if err != nil {
		// 23505 = unique_violation: nama mapel sudah ada.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrConflict
		}
		return nil, translate(err)
	}
	return &s, nil
}

func (r SubjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subject, error) {
	var s domain.Subject
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, school_id, name, created_at FROM subjects WHERE id = $1`, id,
	).Scan(&s.ID, &s.SchoolID, &s.Name, &s.CreatedAt)
	if err != nil {
		return nil, translate(err)
	}
	return &s, nil
}

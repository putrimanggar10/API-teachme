// Package postgres mengimplementasikan interface repository domain memakai
// database/sql + pgx. Query selalu pakai placeholder $1,$2,... (anti SQL injection).
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"API-techme/internal/domain"
)

// translate mengubah error teknis DB menjadi error bisnis domain.
func translate(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

// UserRepo membaca tabel users.
type UserRepo struct {
	DB *sql.DB
}

// kolom yang selalu diambil (urutannya harus sama dengan Scan di bawah).
const userCols = `id, school_id, name, username, COALESCE(password_hash,''), role, status, created_at`

// scanUser membaca 1 baris hasil query ke struct domain.User.
func scanUser(row interface{ Scan(...any) error }) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.SchoolID, &u.Name, &u.Username, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)
	if err != nil {
		return nil, translate(err)
	}
	return &u, nil
}

func (r UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	// Catatan multi-tenant: username unik PER sekolah. Demo ini 1 sekolah,
	// jadi cukup cari by username. LIMIT 1 untuk aman.
	row := r.DB.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE username=$1 LIMIT 1`, username)
	return scanUser(row)
}

func (r UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.DB.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE id=$1`, id)
	return scanUser(row)
}

// ListTeachers: semua guru aktif (untuk dropdown pilih nama di layar login).
func (r UserRepo) ListTeachers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT `+userCols+` FROM users WHERE role='TEACHER' AND status='ACTIVE' ORDER BY name`)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()

	result := make([]domain.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *u)
	}
	return result, rows.Err()
}

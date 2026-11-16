package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"API-techme/internal/domain"
)

// SessionRepo membaca/menulis tabel class_sessions.
type SessionRepo struct {
	DB *sql.DB
}

// nullTime ubah kolom timestamptz nullable jadi *time.Time.
func nullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time
	return &t
}

func (r SessionRepo) Create(ctx context.Context, s *domain.Session, joinCodeHash string, codeExpiresAt time.Time) error {
	err := r.DB.QueryRowContext(ctx, `
		INSERT INTO class_sessions(class_id, teacher_user_id, join_code_hash, join_code_expires_at, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, version, created_at`,
		s.ClassID, s.TeacherID, joinCodeHash, codeExpiresAt, s.Status,
	).Scan(&s.ID, &s.Version, &s.CreatedAt)
	return translate(err)
}

func (r SessionRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	var s domain.Session
	var starts, ends sql.NullTime
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, class_id, teacher_user_id, status, version, starts_at, ends_at, created_at
		FROM class_sessions WHERE id = $1`, id,
	).Scan(&s.ID, &s.ClassID, &s.TeacherID, &s.Status, &s.Version, &starts, &ends, &s.CreatedAt)
	if err != nil {
		return nil, translate(err)
	}
	s.StartsAt = nullTime(starts)
	s.EndsAt = nullTime(ends)
	return &s, nil
}

func (r SessionRepo) End(ctx context.Context, id, teacherID uuid.UUID) error {
	// Hanya guru PEMILIK yang boleh mengakhiri, dan hanya sesi yang masih hidup.
	res, err := r.DB.ExecContext(ctx, `
		UPDATE class_sessions
		SET status='ENDED', ends_at=now(), version=version+1, updated_at=now()
		WHERE id=$1 AND teacher_user_id=$2 AND status IN ('OPEN','ACTIVE')`,
		id, teacherID)
	if err != nil {
		return translate(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// Tidak ada baris yang berubah = sesi tak ada, bukan milik guru ini, atau sudah berakhir.
		return domain.ErrNotFound
	}
	return nil
}

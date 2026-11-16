package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Status sesi (sesuai schema: DRAFT -> OPEN -> ACTIVE -> ENDED | CANCELLED).
const (
	SessionOpen   = "OPEN"
	SessionActive = "ACTIVE"
	SessionEnded  = "ENDED"
)

// Session = sesi kelas aktif yang dibuat guru. Sumber dari "join code" & QR.
type Session struct {
	ID        uuid.UUID  `json:"id"`
	ClassID   uuid.UUID  `json:"class_id"`
	TeacherID uuid.UUID  `json:"teacher_user_id"`
	Status    string     `json:"status"`
	Version   int64      `json:"version"`
	StartsAt  *time.Time `json:"starts_at,omitempty"`
	EndsAt    *time.Time `json:"ends_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	// JoinCode = kode 6 karakter (mis. "8XW2J1"). HANYA diisi saat sesi dibuat,
	// supaya board bisa menampilkannya. TIDAK pernah dibaca balik dari DB
	// (yang disimpan di DB hanya hash-nya).
	JoinCode string `json:"join_code,omitempty"`
}

// SessionRepository = kontrak penyimpanan sesi.
type SessionRepository interface {
	// Create menyimpan sesi baru. joinCodeHash = hash dari kode; codeExpiresAt = kadaluarsa kode.
	Create(ctx context.Context, s *Session, joinCodeHash string, codeExpiresAt time.Time) error
	// Get mengambil sesi berdasarkan ID.
	Get(ctx context.Context, id uuid.UUID) (*Session, error)
	// End mengakhiri sesi (hanya guru pemilik). Mengembalikan ErrNotFound bila tak ada/tak berhak.
	End(ctx context.Context, id, teacherID uuid.UUID) error
}

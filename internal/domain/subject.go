package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Subject = mata pelajaran (mis. "Math"). Milik sekolah, bisa ditambah sendiri.
type Subject struct {
	ID        uuid.UUID `json:"id"`
	SchoolID  uuid.UUID `json:"school_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// SubjectRepository = kontrak penyimpanan mapel.
type SubjectRepository interface {
	// List: semua mapel, urut nama.
	List(ctx context.Context) ([]Subject, error)
	// Create: tambah mapel baru. Nama sudah ada (tanpa peduli huruf besar/kecil) -> ErrConflict.
	Create(ctx context.Context, name string) (*Subject, error)
	// GetByID: ambil 1 mapel.
	GetByID(ctx context.Context, id uuid.UUID) (*Subject, error)
}

// TeachingOption = satu pilihan di dropdown "Class" layar login, mis. "Math · Class 7B".
// Artinya: guru ini mengajar mapel tsb di kelas tsb.
type TeachingOption struct {
	ClassID     uuid.UUID `json:"class_id"`
	ClassName   string    `json:"class_name"`
	SubjectID   uuid.UUID `json:"subject_id"`
	SubjectName string    `json:"subject_name"`
	Label       string    `json:"label"` // teks siap tampil: "Math · Class 7B"
}

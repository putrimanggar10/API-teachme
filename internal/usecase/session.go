package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"time"

	"github.com/google/uuid"

	"API-techme/internal/domain"
)

// Charset kode kelas: huruf besar + angka, TANPA karakter ambigu (0,O,1,I)
// supaya tidak tertukar saat siswa membaca dari papan.
const joinCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// Lama kode berlaku (untuk 1 sesi). Setelah ini, kode kedaluwarsa.
const joinCodeTTL = 8 * time.Hour

// SessionService mengurus pembuatan sesi + kode + QR, dan mengakhiri sesi.
type SessionService struct {
	Sessions domain.SessionRepository
	Classes  domain.ClassRepository // untuk cek guru berwenang atas kelas
}

// Start membuat sesi baru untuk sebuah kelas (layar "Mulai kelas / QR code").
// POLICY: guru harus berwenang atas kelas itu.
func (s SessionService) Start(ctx context.Context, teacherID, classID uuid.UUID) (*domain.Session, error) {
	ok, err := s.Classes.IsTeacherAssigned(ctx, classID, teacherID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	code, err := generateJoinCode(6)
	if err != nil {
		return nil, err
	}

	sess := &domain.Session{
		ClassID:   classID,
		TeacherID: teacherID,
		Status:    domain.SessionOpen,
		JoinCode:  code, // dikembalikan ke board, tidak disimpan plaintext
	}
	if err := s.Sessions.Create(ctx, sess, hashCode(code), time.Now().Add(joinCodeTTL)); err != nil {
		return nil, err
	}
	return sess, nil
}

// Get mengambil info sesi.
func (s SessionService) Get(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	return s.Sessions.Get(ctx, id)
}

// End mengakhiri sesi (kode jadi tidak berlaku).
func (s SessionService) End(ctx context.Context, teacherID, sessionID uuid.UUID) error {
	return s.Sessions.End(ctx, sessionID, teacherID)
}

// QRPayload = isi yang dikodekan ke dalam gambar QR.
// Siswa yang scan akan mendapat ID sesi untuk bergabung (lewat aplikasi lain).
func QRPayload(sessionID uuid.UUID) string {
	return "teachme://join?s=" + sessionID.String()
}

// ---- util ----

// generateJoinCode membuat kode acak aman (crypto/rand) sepanjang n karakter.
func generateJoinCode(n int) (string, error) {
	buf := make([]byte, n)
	max := big.NewInt(int64(len(joinCodeCharset)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = joinCodeCharset[idx.Int64()]
	}
	return string(buf), nil
}

// hashCode mengubah kode -> SHA-256 hex. Yang ini yang disimpan di DB
// (bukan kode asli), agar kode tidak bocor & tahan tebakan.
func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

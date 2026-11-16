// Package token mengurus pembuatan & verifikasi JWT (JSON Web Token).
//
// Analogi: setelah login, server memberi "tiket" (token) yang sudah DITANDATANGANI
// pakai kunci rahasia. Tiap request berikutnya, klien menunjukkan tiket ini.
// Server cukup memeriksa tanda tangannya — tak perlu cek password lagi.
//
// Ada 2 jenis tiket:
//   - access  : masa berlaku PENDEK (mis. 15 menit). Dipakai di tiap request.
//   - refresh : masa berlaku PANJANG (mis. 30 hari). Hanya untuk minta access baru.
package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Jenis token.
const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

var ErrInvalidToken = errors.New("token tidak valid atau kedaluwarsa")

// Claims = isi tiket. RegisteredClaims berisi field standar JWT
// (Subject=pemilik, ExpiresAt=kedaluwarsa, IssuedAt=kapan dibuat).
// Kita tambah field sendiri: Role, SchoolID, Type.
type Claims struct {
	Role     string `json:"role"`
	SchoolID string `json:"school_id"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

// Manager membuat & memverifikasi token. Dibuat sekali di main.go.
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager membuat Manager. secret = kunci rahasia (dari env, JANGAN di-hardcode di produksi).
func NewManager(secret string, accessTTL, refreshTTL time.Duration) Manager {
	return Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Access membuat token access untuk user tertentu.
func (m Manager) Access(userID, role, schoolID string) (string, error) {
	return m.issue(userID, role, schoolID, TypeAccess, m.accessTTL)
}

// Refresh membuat token refresh.
func (m Manager) Refresh(userID, role, schoolID string) (string, error) {
	return m.issue(userID, role, schoolID, TypeRefresh, m.refreshTTL)
}

// AccessTTL dipakai handler untuk memberi tahu klien berapa detik token berlaku.
func (m Manager) AccessTTL() time.Duration { return m.accessTTL }

// issue adalah inti pembuatan token (dipakai Access & Refresh).
func (m Manager) issue(userID, role, schoolID, typ string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Role:     role,
		SchoolID: schoolID,
		Type:     typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	// HS256 = tanda tangan simetris (pakai 1 kunci rahasia yang sama untuk sign & verify).
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

// Parse memverifikasi token & mengembalikan isinya.
// Gagal bila: tanda tangan salah, kedaluwarsa, atau metode tanda tangan tak sesuai.
func (m Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		// Pastikan metode tanda tangan memang HS256 (cegah serangan "alg=none").
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

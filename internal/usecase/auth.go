// Package usecase berisi aturan bisnis. Bergantung HANYA pada interface domain
// + util (token, bcrypt), bukan pada SQL/driver maupun Gin.
package usecase

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"API-techme/internal/domain"
	"API-techme/internal/token"
)

// AuthResult = hasil login: data user + sepasang token.
type AuthResult struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // detik, masa berlaku access token
}

// AuthService mengurus login, refresh token, dan ambil profil.
type AuthService struct {
	Users  domain.UserRepository
	Tokens token.Manager
}

// Login memverifikasi username+password, lalu menerbitkan access & refresh token.
func (s AuthService) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	user, err := s.Users.GetByUsername(ctx, username)
	if err != nil {
		// Supaya tidak membocorkan "username ada tapi password salah" vs "username tak ada",
		// keduanya dipetakan ke error yang SAMA.
		if err == domain.ErrNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if user.Status != "ACTIVE" {
		return nil, domain.ErrInvalidCredentials
	}

	// bcrypt.CompareHashAndPassword membandingkan hash tersimpan dgn password yang diketik.
	// Aman dari timing attack & tidak pernah mengembalikan password asli.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.issueTokens(user)
}

// Refresh menukar refresh token yang sah dengan access token baru.
func (s AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	claims, err := s.Tokens.Parse(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	// Pastikan ini benar-benar token REFRESH, bukan access yang disalahgunakan.
	if claims.Type != token.TypeRefresh {
		return nil, domain.ErrInvalidCredentials
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	user, err := s.Users.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user.Status != "ACTIVE" {
		return nil, domain.ErrInvalidCredentials
	}
	return s.issueTokens(user)
}

// Me mengembalikan profil user berdasarkan ID (dari token).
func (s AuthService) Me(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.Users.GetByID(ctx, id)
}

// issueTokens membuat sepasang token untuk user (dipakai Login & Refresh).
func (s AuthService) issueTokens(user *domain.User) (*AuthResult, error) {
	uid := user.ID.String()
	sid := user.SchoolID.String()

	access, err := s.Tokens.Access(uid, user.Role, sid)
	if err != nil {
		return nil, err
	}
	refresh, err := s.Tokens.Refresh(uid, user.Role, sid)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		User:         user,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(s.Tokens.AccessTTL().Seconds()),
	}, nil
}

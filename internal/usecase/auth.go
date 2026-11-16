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

// AuthResult = hasil login: data user + sepasang token + kelas yang dipilih.
type AuthResult struct {
	User         *domain.User
	ActiveClass  *domain.Class // kelas yang dipilih saat login
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // detik, masa berlaku access token
}

// AuthService mengurus login, refresh token, dan ambil profil.
type AuthService struct {
	Users   domain.UserRepository
	Classes domain.ClassRepository // untuk cek guru mengajar kelas yang dipilih
	Tokens  token.Manager
}

// ListTeachers: daftar guru untuk dropdown "pilih nama" di layar login.
func (s AuthService) ListTeachers(ctx context.Context) ([]domain.User, error) {
	return s.Users.ListTeachers(ctx)
}

// Login: GURU pilih namanya (accountID) + password + kelas yang diajar.
// Validasi: password benar, role TEACHER, dan guru memang mengajar kelas itu.
func (s AuthService) Login(ctx context.Context, accountID uuid.UUID, password string, classID uuid.UUID) (*AuthResult, error) {
	user, err := s.Users.GetByID(ctx, accountID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if user.Status != "ACTIVE" {
		return nil, domain.ErrInvalidCredentials
	}

	// Cocokkan password dengan hash (bcrypt). Aman & tidak mengembalikan password asli.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Layar login papan ini khusus GURU.
	if user.Role != domain.RoleTeacher {
		return nil, domain.ErrForbidden
	}

	// Pastikan guru ini memang mengajar kelas yang dipilih.
	ok, err := s.Classes.IsTeacherAssigned(ctx, classID, accountID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	class, err := s.Classes.GetByID(ctx, classID)
	if err != nil {
		return nil, err
	}

	res, err := s.issueTokens(user)
	if err != nil {
		return nil, err
	}
	res.ActiveClass = class
	return res, nil
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

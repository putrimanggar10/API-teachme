// Package usecase berisi aturan bisnis. Bergantung HANYA pada interface domain
// + util (token, bcrypt), bukan pada SQL/driver maupun Gin.
package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"API-techme/internal/domain"
	"API-techme/internal/token"
)

// AuthResult = hasil login: data user + sepasang token + kelas yang dipilih.
type AuthResult struct {
	User          *domain.User
	ActiveClass   *domain.Class   // kelas yang dipilih saat login
	ActiveSubject *domain.Subject // mapel yang dipilih saat login
	AccessToken   string
	RefreshToken  string
	ExpiresIn     int // detik, masa berlaku access token
}

// AuthService mengurus login, refresh token, dan ambil profil.
type AuthService struct {
	Users    domain.UserRepository
	Classes  domain.ClassRepository   // untuk cek guru mengajar mapel+kelas yang dipilih
	Subjects domain.SubjectRepository // untuk data mapel aktif
	Tokens   token.Manager
}

// ListTeachers: daftar guru untuk dropdown "pilih nama" di layar login.
func (s AuthService) ListTeachers(ctx context.Context) ([]domain.User, error) {
	return s.Users.ListTeachers(ctx)
}

// AddTeacher: tambah guru baru dari halaman "Kelola Akun".
// Cukup nama + password. Password itulah yang nanti dipakai login.
func (s AuthService) AddTeacher(ctx context.Context, name, password string) (*domain.User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: nama guru wajib diisi", domain.ErrInvalid)
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("%w: password minimal 6 karakter", domain.ErrInvalid)
	}

	// Simpan password sebagai hash bcrypt, bukan teks asli.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.Users.CreateTeacher(ctx, name, usernameFromName(name), string(hash))
}

// usernameFromName membuat username sederhana dari nama:
// huruf kecil, karakter non-alfanumerik jadi titik ("Bu Rina!" -> "bu.rina").
func usernameFromName(name string) string {
	var b strings.Builder
	lastDot := false
	for _, r := range strings.ToLower(name) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDot = false
		default:
			if b.Len() > 0 && !lastDot {
				b.WriteRune('.')
				lastDot = true
			}
		}
	}
	u := strings.Trim(b.String(), ".")
	if u == "" {
		u = "guru"
	}
	return u
}

// Login: GURU pilih namanya (accountID) + password + "Mapel · Kelas" yang diajar.
// Validasi: password benar, role TEACHER, dan guru memang mengajar mapel itu di kelas itu.
func (s AuthService) Login(ctx context.Context, accountID uuid.UUID, password string, classID, subjectID uuid.UUID) (*AuthResult, error) {
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

	// Pastikan guru ini memang mengajar mapel yang dipilih di kelas yang dipilih.
	ok, err := s.Classes.IsTeachingSubject(ctx, accountID, classID, subjectID)
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
	subject, err := s.Subjects.GetByID(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	res, err := s.issueTokens(user)
	if err != nil {
		return nil, err
	}
	res.ActiveClass = class
	res.ActiveSubject = subject
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

// SetLanguage: simpan bahasa papan pilihan user (layar "Choose board language").
// Kode harus salah satu dari domain.Languages, kalau tidak -> ErrInvalid.
func (s AuthService) SetLanguage(ctx context.Context, userID uuid.UUID, code string) (*domain.Language, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	var lang *domain.Language
	for i := range domain.Languages {
		if domain.Languages[i].Code == code {
			lang = &domain.Languages[i]
			break
		}
	}
	if lang == nil {
		return nil, fmt.Errorf("%w: bahasa %q tidak didukung (pilih: id, en, zh, es, ar)", domain.ErrInvalid, code)
	}
	if err := s.Users.UpdateLanguage(ctx, userID, lang.Code); err != nil {
		return nil, err
	}
	return lang, nil
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

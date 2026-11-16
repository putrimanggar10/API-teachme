// Composition root: tempat semua komponen dirangkai (dependency injection).
package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // daftarkan driver "pgx" ke database/sql

	_ "API-techme/docs" // dokumentasi Swagger (di-generate "swag init")
	httpdelivery "API-techme/internal/delivery/http"
	"API-techme/internal/repository/postgres"
	"API-techme/internal/token"
	"API-techme/internal/usecase"
)

// @title           TeachMe API
// @version         0.1
// @description     Backend TeachMe (class board). Milestone F01: Auth, Pilih Kelas, Sesi + Code/QR. Gin + PostgreSQL, Clean Architecture.
// @BasePath        /
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL wajib diisi")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Kunci rahasia untuk tanda tangan JWT. WAJIB di-set di produksi.
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me" // hanya untuk latihan lokal
		log.Println("PERINGATAN: JWT_SECRET belum di-set, memakai secret default (JANGAN untuk produksi)")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("gagal terhubung ke PostgreSQL: %v", err)
	}

	// Pembuat token: access 15 menit, refresh 30 hari.
	tokens := token.NewManager(secret, 15*time.Minute, 30*24*time.Hour)

	// Rangkai: repository -> service -> handler.
	h := httpdelivery.Handler{
		Auth: usecase.AuthService{
			Users:   postgres.UserRepo{DB: db},
			Classes: postgres.ClassRepo{DB: db},
			Tokens:  tokens,
		},
		Classes: usecase.ClassService{
			Repo: postgres.ClassRepo{DB: db},
		},
		Sessions: usecase.SessionService{
			Sessions: postgres.SessionRepo{DB: db},
			Classes:  postgres.ClassRepo{DB: db},
		},
		Tokens: tokens,
	}

	log.Printf("server aktif di http://localhost:%s  (Swagger: /swagger/index.html)", port)
	if err := httpdelivery.Router(h).Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

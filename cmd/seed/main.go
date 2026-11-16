// Program seeder: mengisi data contoh TeachMe supaya bisa langsung dites.
//
// Jalankan (PowerShell):
//   $env:DATABASE_URL="postgres://postgres:putrinadhif08%23@localhost:5432/teachme?sslmode=disable"
//   go run ./cmd/seed
//
// CATATAN: karakter '#' pada password HARUS ditulis %23 di dalam URL.
//
// Seeder ini AMAN diulang: ia menghapus isi tabel dulu (TRUNCATE) lalu mengisi ulang.
package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // driver "pgx" untuk database/sql
	"golang.org/x/crypto/bcrypt"
)

// hash membungkus bcrypt supaya singkat dipanggil.
func hash(pw string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL wajib diisi")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("gagal konek DB: %v", err)
	}

	// Semua ditulis dalam SATU transaksi: kalau ada yang gagal, batal semua.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	// 1) Kosongkan semua tabel (urutan tak masalah karena CASCADE).
	if _, err := tx.ExecContext(ctx, `
		TRUNCATE session_invite_tokens, session_participants, class_sessions,
		         class_enrollments, teacher_class_assignments, classes, users, schools
		RESTART IDENTITY CASCADE`); err != nil {
		log.Fatalf("truncate: %v", err)
	}

	// 2) Sekolah (tenant).
	schoolID := uuid.New()
	mustExec(ctx, tx,
		`INSERT INTO schools(id, name) VALUES ($1, $2)`,
		schoolID, "Sekolah Demo TeachMe")

	// 3) Guru (role TEACHER). Password semua: "guru123".
	//    (Schema resmi tak punya kolom subject; mapel di gambar belum disimpan.)
	teacherPw := hash("guru123")
	axel := uuid.New()
	nova := uuid.New()
	sabo := uuid.New()
	vision := uuid.New()
	insertUser(ctx, tx, axel, schoolID, "Mr. Axel", "mr.axel", teacherPw, "TEACHER")
	insertUser(ctx, tx, nova, schoolID, "Nova", "nova", teacherPw, "TEACHER")
	insertUser(ctx, tx, sabo, schoolID, "Sabo", "sabo", teacherPw, "TEACHER")
	insertUser(ctx, tx, vision, schoolID, "Vision", "vision", teacherPw, "TEACHER")

	// 4) Siswa (role STUDENT). Password semua: "siswa123".
	studentPw := hash("siswa123")
	students := []struct{ name, username string }{
		{"Rifa", "rifa"},
		{"Dimas", "dimas"},
		{"Hana", "hana"},
		{"Maya", "maya"},
		{"Celine", "celine"},
	}
	studentIDs := make([]uuid.UUID, 0, len(students))
	for _, s := range students {
		id := uuid.New()
		insertUser(ctx, tx, id, schoolID, s.name, s.username, studentPw, "STUDENT")
		studentIDs = append(studentIDs, id)
	}

	// 5) Kelas.
	class7B := uuid.New()
	class7A := uuid.New()
	classVision := uuid.New()
	classPublic := uuid.New()
	insertClass(ctx, tx, class7B, schoolID, "Class 7B", "7")
	insertClass(ctx, tx, class7A, schoolID, "Class 7A", "7")
	insertClass(ctx, tx, classVision, schoolID, "Vision class", "")
	insertClass(ctx, tx, classPublic, schoolID, "Public class", "")

	// 6) Penugasan guru -> kelas (siapa boleh mengajar apa).
	assign(ctx, tx, class7B, axel)   // Mr. Axel pegang 7B (Math · Class 7B di gambar)
	assign(ctx, tx, class7A, axel)   // dan 7A
	assign(ctx, tx, class7A, nova)   // Nova juga di 7A
	assign(ctx, tx, classVision, vision)

	// 7) Roster siswa: semua siswa contoh dimasukkan ke Class 7B.
	for _, sid := range studentIDs {
		enroll(ctx, tx, class7B, sid)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit: %v", err)
	}

	log.Println("✅ Seed selesai.")
	log.Println("   Login guru  -> username: mr.axel | nova | sabo | vision   (password: guru123)")
	log.Println("   Login siswa -> username: rifa | dimas | hana | maya | celine (password: siswa123)")
}

// ---- helper kecil supaya main() tetap ringkas ----

func mustExec(ctx context.Context, tx *sql.Tx, q string, args ...any) {
	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		log.Fatalf("query gagal: %v\n  SQL: %s", err, q)
	}
}

func insertUser(ctx context.Context, tx *sql.Tx, id, schoolID uuid.UUID, name, username, pwHash, role string) {
	mustExec(ctx, tx,
		`INSERT INTO users(id, school_id, name, username, password_hash, role)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		id, schoolID, name, username, pwHash, role)
}

func insertClass(ctx context.Context, tx *sql.Tx, id, schoolID uuid.UUID, name, grade string) {
	var g any
	if grade == "" {
		g = nil
	} else {
		g = grade
	}
	mustExec(ctx, tx,
		`INSERT INTO classes(id, school_id, name, grade, academic_year)
		 VALUES ($1,$2,$3,$4,$5)`,
		id, schoolID, name, g, "2026/2027")
}

func assign(ctx context.Context, tx *sql.Tx, classID, teacherID uuid.UUID) {
	mustExec(ctx, tx,
		`INSERT INTO teacher_class_assignments(id, class_id, teacher_user_id)
		 VALUES ($1,$2,$3)`,
		uuid.New(), classID, teacherID)
}

func enroll(ctx context.Context, tx *sql.Tx, classID, studentID uuid.UUID) {
	mustExec(ctx, tx,
		`INSERT INTO class_enrollments(id, class_id, student_user_id)
		 VALUES ($1,$2,$3)`,
		uuid.New(), classID, studentID)
}

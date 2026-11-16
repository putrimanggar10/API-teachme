# TeachMe API

Backend **TeachMe** — aplikasi *class board* interaktif untuk **guru & siswa** (mirip IWB/Kahoot).
Dibangun dengan **Go + Gin + PostgreSQL**, pola **Clean Architecture**.

> **Status:** proyek bertahap. Repo ini baru menyelesaikan **Milestone F01** —
> *Authentication, Pilih Kelas, Sesi + Join Code/QR*. Fitur lain (quiz, whiteboard,
> screen share, points, voting, realtime WebSocket) menyusul sesuai PAD TeachMe v2.

---

## ✨ Yang sudah jalan (F01)

| Bagian | Status |
|--------|--------|
| Login guru/siswa (JWT access + refresh) | ✅ |
| Policy **teacher / student** (middleware role) | ✅ |
| Database 45 tabel (schema resmi PAD) + seeder | ✅ |
| Swagger UI | ✅ |
| Guru pilih kelas, Generate Code + QR, Join/Leave | ⏳ (berikutnya) |

## 🧱 Teknologi

- **Go** 1.27, **Gin** (HTTP framework)
- **PostgreSQL** (driver `pgx`)
- **JWT** (`golang-jwt/jwt/v5`) — token access + refresh
- **bcrypt** — hashing password
- **swaggo** — Swagger UI
- UUID sebagai Primary Key

## 📂 Struktur

```
API-techme/
├── cmd/
│   ├── api/main.go          # composition root (rangkai semua)
│   └── seed/main.go         # isi data contoh (sekolah, guru, siswa, kelas)
├── internal/
│   ├── domain/models.go     # entity + interface (kontrak), tanpa Gin/SQL
│   ├── token/jwt.go         # buat & verifikasi JWT
│   ├── usecase/auth.go      # aturan bisnis: login, refresh, me
│   ├── repository/postgres/ # implementasi SQL
│   └── delivery/http/       # handler Gin, middleware (policy), routing
├── migrations/001_init.sql  # schema PostgreSQL (45 tabel)
├── docs/                    # Swagger (di-generate swag init)
└── .env.example
```

Arah dependency: `handler → usecase → interface domain ← repository postgres`.

## 🚀 Cara menjalankan

### 1. Prasyarat
- Go 1.27+, PostgreSQL 13+
- Buat database kosong bernama `teachme`

### 2. Buat tabel & isi data contoh

```powershell
# jalankan migration (via psql) lalu seeder
$env:DATABASE_URL="postgres://postgres:PASSWORD%23@localhost:5432/teachme?sslmode=disable"
go run ./cmd/seed
```
> Karakter `#` pada password ditulis `%23` di dalam URL.

### 3. Nyalakan server (PowerShell)

```powershell
$env:DATABASE_URL="postgres://postgres:PASSWORD%23@localhost:5432/teachme?sslmode=disable"
$env:PORT="8099"
$env:JWT_SECRET="ganti-dengan-secret-acak"
go run ./cmd/api
```

Server: `http://localhost:8099` — Swagger UI: `http://localhost:8099/swagger/index.html`

## 🔌 Endpoint (F01)

| Metode | Path | Akses | Keterangan |
|--------|------|-------|------------|
| GET  | `/health` | publik | cek server hidup |
| POST | `/api/v1/auth/login` | publik | login → access + refresh token |
| POST | `/api/v1/auth/refresh` | publik | tukar refresh → access baru |
| GET  | `/api/v1/me` | login | profil pengguna |

## 👤 Akun contoh (hasil seeder)

| Peran | Username | Password |
|-------|----------|----------|
| Guru  | `mr.axel`, `nova`, `sabo`, `vision` | `guru123` |
| Siswa | `rifa`, `dimas`, `hana`, `maya`, `celine` | `siswa123` |

## 🔐 Catatan keamanan

- Password disimpan sebagai **hash bcrypt**, tidak pernah plaintext.
- `join_code` & token QR nanti disimpan sebagai **hash**, bukan nilai asli.
- File `.env` (berisi password asli) **tidak** ikut di-commit (lihat `.gitignore`).
- Proyek untuk pembelajaran; sebelum produksi tambahkan HTTPS, rate limit, isolasi tenant (school_id), dll.

---

*Bagian dari sistem TeachMe (IWB Teacher APK · Student APK · Student Windows EXE). Backend by Sukma.*

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
| Login guru (pilih nama + password + pilih kelas) → JWT access + refresh | ✅ |
| **Kelola Akun** → Tambah Guru (nama + password untuk login) | ✅ |
| Policy **teacher / student** (middleware role + cek pengampu) | ✅ |
| Database 45 tabel (schema resmi PAD) + seeder | ✅ |
| Guru pilih kelas + daftar siswa (roster) | ✅ |
| Mulai sesi → **generate join code + QR (PNG)** → akhiri sesi | ✅ |
| Swagger UI + Postman collection | ✅ |
| Join/Leave siswa (pakai aplikasi lain) | ⏳ (berikutnya) |

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
├── migrations/001_init.sql  # schema PostgreSQL (45 tabel resmi)
├── migrations/002_subjects.sql  # tambahan: mata pelajaran + guru mengajar mapel di kelas
├── migrations/003_user_language.sql  # tambahan: bahasa papan pilihan guru
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
# 1) jalankan migration berurutan (via psql)
& "C:\Program Files\PostgreSQL\13\bin\psql.exe" -U postgres -d teachme -f migrations/001_init.sql
& "C:\Program Files\PostgreSQL\13\bin\psql.exe" -U postgres -d teachme -f migrations/002_subjects.sql
& "C:\Program Files\PostgreSQL\13\bin\psql.exe" -U postgres -d teachme -f migrations/003_user_language.sql

# 2) isi data contoh
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
| GET  | `/api/v1/accounts` | publik | daftar guru (dropdown "pilih nama") |
| POST | `/api/v1/accounts` | publik | **Kelola Akun → Tambah Guru** (`name` + `password`) |
| GET  | `/api/v1/accounts/{id}/classes` | publik | pilihan **"Mapel · Kelas"** guru (dropdown "Class" di login) |
| POST | `/api/v1/accounts/{id}/classes` | publik | **Kelola → Tugaskan guru** mengajar mapel di kelas (`class_id` + `subject_id`) |
| GET  | `/api/v1/school-classes` | publik | semua kelas (halaman Manage accounts & classes) |
| POST | `/api/v1/school-classes` | publik | **Kelola → Tambah kelas** (`name`, opsional `grade` & `academic_year`) |
| GET  | `/api/v1/subjects` | publik | semua mata pelajaran |
| POST | `/api/v1/subjects` | publik | **Kelola → Tambah mata pelajaran** (`name`) |
| POST | `/api/v1/auth/login` | publik | login: `account_id` + `password` + `class_id` + `subject_id` → token |
| POST | `/api/v1/auth/refresh` | publik | tukar refresh → access baru |
| GET  | `/api/v1/me` | login | profil pengguna (termasuk `preferred_language`) |
| GET  | `/api/v1/languages` | publik | daftar bahasa papan (layar "Choose board language") |
| PUT  | `/api/v1/me/language` | login | simpan bahasa papan pilihan (`{ "language": "en" }`) |
| GET  | `/api/v1/classes` | guru | kelas yang diajar |
| GET  | `/api/v1/classes/{id}/roster` | guru | daftar siswa kelas |
| POST | `/api/v1/sessions` | guru | mulai sesi → `join_code` + `qr_payload` |
| GET  | `/api/v1/sessions/{id}` | guru | info sesi |
| GET  | `/api/v1/sessions/{id}/qr` | guru | gambar QR (PNG) |
| POST | `/api/v1/sessions/{id}/end` | guru | akhiri sesi |

### Alur login (sesuai layar Sign in)
1. `GET /accounts` → tampilkan dropdown nama guru
2. Guru pilih namanya → `GET /accounts/{id}/classes` → tampilkan dropdown **Class** berisi `label` seperti `"Math · Class 7B"`
3. Guru isi password + pilih "Mapel · Kelas" → `POST /auth/login { account_id, password, class_id, subject_id }` (ambil `class_id` & `subject_id` dari item dropdown yang dipilih)
4. Dapat `access_token` + `active_class` + `active_subject`. `user.preferred_language` = bahasa terakhir guru (default `id`).
5. Layar **Choose board language**: `GET /languages` untuk daftar pilihan, tandai yang sama dengan `preferred_language`. Saat **Continue** → `PUT /me/language { "language": "en" }`. Terjemahan teks tetap di aplikasi; backend hanya menyimpan pilihannya.
6. `POST /sessions` untuk memunculkan QR + kode.

### Manage accounts & classes (link di layar login)
Di layar login ada link **Manage accounts & classes** → kelola guru, mapel, dan kelas.
- **Tambah Guru:** `POST /accounts { "name": "...", "password": "..." }` → guru baru langsung muncul di dropdown login, dan password-nya dipakai untuk login. `username` dibuat otomatis dari nama (`Bu Rina` → `bu.rina`, bentrok → `bu.rina2`).
- **Tambah Kelas:** `POST /school-classes { "name": "Class 8C", "grade": "8", "academic_year": "2026/2027" }` (hanya `name` yang wajib). Nama kelas tidak boleh kembar → `409`.
- **Tambah Mata Pelajaran:** `POST /subjects { "name": "Math" }`. Nama tidak boleh kembar (tanpa peduli huruf besar/kecil) → `409`.
- **Tugaskan Guru:** pilih kelas (`GET /school-classes`) + mapel (`GET /subjects`), lalu `POST /accounts/{id}/classes { "class_id": "...", "subject_id": "..." }`. Setelah itu guru bisa login memilih "Mapel · Kelas" tsb. Satu guru boleh mengajar beberapa mapel di kelas yang sama. Aman dipanggil berulang (tidak dobel).

### Tes di Postman
Import `docs/teachme.postman_collection.json` → klik kanan collection → **Run collection**. Berisi 55 request berurutan (tambah kelas → tambah mapel → tambah guru → tugaskan mapel + kelas → login → pilih bahasa → kelas → sesi/QR → akhiri) lengkap dengan pengecekan otomatis, termasuk tes yang memang harus ditolak (bertanda `[GAGAL]`). Semua ID & token tersimpan otomatis.

## 👤 Akun contoh (hasil seeder)

| Peran | Nama (muncul di dropdown `/accounts`) | Password |
|-------|----------|----------|
| Guru  | Mr. Axel, Nova, Sabo, Vision | `guru123` |
| Siswa | Rifa, Dimas, Hana, Maya, Celine | `siswa123` |

> Login pakai `account_id` (didapat dari `GET /accounts`), bukan ketik nama manual.
> Pilihan "Mapel · Kelas" hasil seeder:
> - **Mr. Axel**: Math · Class 7A, Math · Class 7B, Science · Class 7B
> - **Nova**: English · Class 7A · **Sabo**: Bahasa Indonesia · Class 7A · **Vision**: Science · Vision class
>
> Class 7B berisi 5 siswa contoh.

## 🔐 Catatan keamanan

- Password disimpan sebagai **hash bcrypt**, tidak pernah plaintext.
- `join_code` & token QR nanti disimpan sebagai **hash**, bukan nilai asli.
- File `.env` (berisi password asli) **tidak** ikut di-commit (lihat `.gitignore`).
- Proyek untuk pembelajaran; sebelum produksi tambahkan HTTPS, rate limit, isolasi tenant (school_id), dll.

---

*Bagian dari sistem TeachMe (IWB Teacher APK · Student APK · Student Windows EXE). Backend by Sukma.*

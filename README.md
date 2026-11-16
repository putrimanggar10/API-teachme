# Belajar Go CRUD — Kasir Mini

REST API kasir mini dengan **Gin + PostgreSQL**, pola **Clean Architecture**.
Tiga entity: **Produk** (CRUD), **Pelanggan** (CRUD), dan **Penjualan** (checkout + baca).
Dilengkapi **Swagger UI**.

## Struktur

```
belajar-go-crud/
├── cmd/api/main.go                         # composition root (rangkai semua)
├── internal/
│   ├── domain/models.go                    # entity + interface (kontrak)
│   ├── usecase/services.go                 # aturan bisnis
│   ├── usecase/services_test.go            # unit test tanpa DB
│   ├── repository/postgres/repository.go   # implementasi SQL + transaksi checkout
│   └── delivery/http/handlers.go           # routing + handler Gin + Swagger
├── docs/openapi.json                       # kontrak OpenAPI (dibaca Swagger UI)
├── migrations/001_init.sql                 # CREATE TABLE (4 tabel)
├── .env.example
└── go.mod
```

Arah dependency: `handler → usecase → interface domain ← repository postgres`.
`main.go` yang merangkai semuanya.

## 1. Siapkan database (sekali saja)

Masuk ke psql sebagai admin lalu buat user + database:

```sql
CREATE USER belajar WITH PASSWORD 'belajar123';
CREATE DATABASE belajar_db OWNER belajar;
```

Jalankan migration dari folder root proyek (PowerShell):

```powershell
$env:PGPASSWORD="belajar123"
psql -h localhost -U belajar -d belajar_db -v ON_ERROR_STOP=1 -f migrations/001_init.sql
psql -h localhost -U belajar -d belajar_db -c "\dt"
```

Harus muncul 4 tabel: `products`, `customers`, `sales`, `sale_items`.

## 2. Install dependency & jalankan

```powershell
go mod tidy
$env:DATABASE_URL="postgres://belajar:belajar123@localhost:5432/belajar_db?sslmode=disable"
$env:PORT="8080"
go run ./cmd/api
```

Biarkan terminal ini tetap jalan.

## 3. Endpoint

| Metode | Path                     | Hasil                 |
|--------|--------------------------|-----------------------|
| GET    | /health                  | 200                   |
| POST   | /api/v1/products         | 201 / 400 / 409       |
| GET    | /api/v1/products         | 200                   |
| GET    | /api/v1/products/{id}    | 200 / 404             |
| PUT    | /api/v1/products/{id}    | 200 / 404 / 409       |
| DELETE | /api/v1/products/{id}    | 204 / 404 / 409       |
| POST   | /api/v1/customers        | 201 / 400 / 409       |
| GET    | /api/v1/customers        | 200                   |
| GET    | /api/v1/customers/{id}   | 200 / 404             |
| PUT    | /api/v1/customers/{id}   | 200 / 404 / 409       |
| DELETE | /api/v1/customers/{id}   | 204 / 404 / 409       |
| POST   | /api/v1/sales (checkout) | 201 / 400 / 404 / 409 |
| GET    | /api/v1/sales            | 200                   |
| GET    | /api/v1/sales/{id}       | 200 / 404             |

Penjualan sengaja **tidak punya** PUT/DELETE supaya riwayat transaksi tetap utuh.

## 4. Swagger UI

Buka di browser (butuh internet untuk aset CDN):

```
http://localhost:8080/swagger/index.html
```

Kontrak mentah: `http://localhost:8080/openapi.json`.

## 5. Coba cepat (terminal lain)

```powershell
# buat produk
curl -i -X POST http://localhost:8080/api/v1/products -H "Content-Type: application/json" -d "{\"name\":\"Kopi Susu\",\"sku\":\"KOPI-001\",\"price\":18000,\"stock\":20}"

# buat pelanggan
curl -i -X POST http://localhost:8080/api/v1/customers -H "Content-Type: application/json" -d "{\"name\":\"Budi\",\"email\":\"budi@example.com\"}"

# checkout: beli 2 produk id 1 (stok berkurang jadi 18)
curl -i -X POST http://localhost:8080/api/v1/sales -H "Content-Type: application/json" -d "{\"customer_id\":1,\"items\":[{\"product_id\":1,\"qty\":2}]}"
```

## 6. Test (tidak perlu DB)

```powershell
go test ./...
```

## Catatan

Proyek ini untuk belajar, **belum siap produksi**. Sebelum produksi tambahkan:
autentikasi (JWT), TLS, rate limit, CORS, structured logging, pagination,
dan migration versioning. Perubahan stok manual sebaiknya pakai ledger
(stock movement), bukan langsung mengganti angka lewat PUT.

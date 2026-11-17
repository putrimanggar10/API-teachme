-- 003_user_language.sql
-- Bahasa papan pilihan guru (layar "Choose board language" setelah login).
-- Disimpan per guru, supaya saat login di IWB mana pun bahasanya otomatis sama.
--
-- Hanya menyimpan PILIHAN bahasa. Terjemahan teks (tombol, menu) tetap di aplikasi.
-- Kode bahasa: id = Bahasa Indonesia, en = English, zh = 中文, es = Español, ar = العربية.
-- Tambahan di luar schema resmi 45 tabel. Aman dijalankan berulang.

BEGIN;

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS preferred_language text NOT NULL DEFAULT 'id';

ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_preferred_language;
ALTER TABLE users
  ADD CONSTRAINT chk_users_preferred_language
  CHECK (preferred_language IN ('id', 'en', 'zh', 'es', 'ar'));

COMMIT;

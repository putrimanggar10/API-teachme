-- 002_subjects.sql
-- Mata pelajaran (mapel) + penugasan guru: mengajar MAPEL apa di KELAS mana.
-- Dipakai layar Sign in: dropdown "Class" berisi pasangan "Math · Class 7B".
--
-- Tambahan di luar schema resmi 45 tabel (schema resmi belum punya tempat mapel).
-- Aman dijalankan berulang (IF NOT EXISTS).

BEGIN;

-- Daftar mapel milik sekolah. Guru bisa menambah sendiri dari "Manage accounts & classes".
CREATE TABLE IF NOT EXISTS subjects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id uuid NOT NULL REFERENCES schools(id),
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
-- Nama mapel unik per sekolah, tidak peduli huruf besar/kecil ("Math" = "math").
CREATE UNIQUE INDEX IF NOT EXISTS uq_subjects_school_name ON subjects(school_id, lower(name));

-- Guru X mengajar mapel Y di kelas Z.
-- Satu guru boleh mengajar beberapa mapel di kelas yang sama.
CREATE TABLE IF NOT EXISTS teacher_class_subjects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  teacher_user_id uuid NOT NULL REFERENCES users(id),
  class_id uuid NOT NULL REFERENCES classes(id),
  subject_id uuid NOT NULL REFERENCES subjects(id),
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(teacher_user_id, class_id, subject_id)
);
CREATE INDEX IF NOT EXISTS ix_tcs_teacher ON teacher_class_subjects(teacher_user_id, active);

COMMIT;

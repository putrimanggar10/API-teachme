-- =====================================================================
-- TeachMe — Schema PostgreSQL LENGKAP (45 tabel)
-- Sumber resmi: TeachMe_PostgreSQL_Full_Schema_45_Tabel_v4.sql (PAD TeachMe v2).
-- Ini acuan database. Milestone F01 hanya MEMAKAI sebagian tabel
-- (schools, users, classes, class_enrollments, teacher_class_assignments,
--  devices, class_sessions, session_participants, session_invite_tokens),
-- sisanya dibuat agar relasi & FK lengkap untuk fitur berikutnya.
--
-- Jalankan sekali (DB kosong):
--   psql -h localhost -U postgres -d teachme -v ON_ERROR_STOP=1 -f migrations/001_init.sql
-- PostgreSQL 13+ (gen_random_uuid tersedia; extension pgcrypto tetap dibuat).
-- =====================================================================
BEGIN;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- schools
CREATE TABLE schools (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text NOT NULL, external_woowedu_id text UNIQUE, status text NOT NULL DEFAULT 'ACTIVE',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- users
CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), school_id uuid NOT NULL REFERENCES schools(id), external_id text, name text NOT NULL, username text NOT NULL, password_hash text, role text NOT NULL CHECK(role IN ('TEACHER','STUDENT','ADMIN')), status text NOT NULL DEFAULT 'ACTIVE',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(school_id,username)
);

-- classes
CREATE TABLE classes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), school_id uuid NOT NULL REFERENCES schools(id), external_woowedu_id text, name text NOT NULL, grade text, academic_year text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(school_id,external_woowedu_id)
);

-- class_enrollments
CREATE TABLE class_enrollments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), class_id uuid NOT NULL REFERENCES classes(id), student_user_id uuid NOT NULL REFERENCES users(id), status text NOT NULL DEFAULT 'ACTIVE',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(class_id,student_user_id)
);

-- devices
CREATE TABLE devices (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), user_id uuid NOT NULL REFERENCES users(id), platform text NOT NULL CHECK(platform IN ('IWB_ANDROID','STUDENT_ANDROID','STUDENT_WINDOWS')), installation_id text NOT NULL UNIQUE, capabilities_json jsonb NOT NULL DEFAULT '{}'::jsonb, app_version text, last_seen_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- class_sessions
CREATE TABLE class_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), class_id uuid NOT NULL REFERENCES classes(id), teacher_user_id uuid NOT NULL REFERENCES users(id), join_code_hash text NOT NULL, join_code_expires_at timestamptz, status text NOT NULL DEFAULT 'OPEN' CHECK(status IN ('DRAFT','OPEN','ACTIVE','ENDED','CANCELLED')), version bigint NOT NULL DEFAULT 1, starts_at timestamptz, ends_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- session_participants
CREATE TABLE session_participants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), user_id uuid NOT NULL REFERENCES users(id), device_id uuid NOT NULL REFERENCES devices(id), state text NOT NULL DEFAULT 'JOINED' CHECK(state IN ('INVITED','JOINED','CONNECTED','DISCONNECTED','LEFT')), joined_at timestamptz NOT NULL DEFAULT now(), left_at timestamptz, last_seen_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(session_id,user_id,device_id)
);

-- session_events
CREATE TABLE session_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), sequence bigint NOT NULL CHECK(sequence > 0), type text NOT NULL, actor_id uuid REFERENCES users(id), aggregate_id uuid, payload_json jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(session_id,sequence)
);

-- outbox_events
CREATE TABLE outbox_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), aggregate_type text NOT NULL, aggregate_id uuid NOT NULL, event_type text NOT NULL, session_id uuid REFERENCES class_sessions(id), payload_json jsonb NOT NULL, published_at timestamptz, attempts integer NOT NULL DEFAULT 0, last_error text,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- idempotency_keys
CREATE TABLE idempotency_keys (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), actor_id uuid NOT NULL REFERENCES users(id), route text NOT NULL, key text NOT NULL, request_hash text NOT NULL, response_json jsonb, expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(actor_id,route,key)
);

-- boards
CREATE TABLE boards (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL UNIQUE REFERENCES class_sessions(id), active_page_id uuid, version bigint NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- board_pages
CREATE TABLE board_pages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), board_id uuid NOT NULL REFERENCES boards(id), page_no integer NOT NULL CHECK(page_no > 0), metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(board_id,page_no)
);

-- board_operations
CREATE TABLE board_operations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), board_id uuid NOT NULL REFERENCES boards(id), page_id uuid NOT NULL REFERENCES board_pages(id), actor_id uuid NOT NULL REFERENCES users(id), op_id uuid NOT NULL UNIQUE, seq bigint NOT NULL CHECK(seq>0), kind text NOT NULL, payload_json jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(board_id,seq)
);

-- board_snapshots
CREATE TABLE board_snapshots (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), board_id uuid NOT NULL REFERENCES boards(id), page_id uuid NOT NULL REFERENCES board_pages(id), last_seq bigint NOT NULL, blob_key text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(page_id,last_seq)
);

-- broadcasts
CREATE TABLE broadcasts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), type text NOT NULL CHECK(type IN ('NATIVE_BOARD','FULL_SCREEN','CAMERA')), target_json jsonb NOT NULL, started_at timestamptz NOT NULL DEFAULT now(), stopped_at timestamptz, status text NOT NULL DEFAULT 'ACTIVE',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- activities
CREATE TABLE activities (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), type text NOT NULL CHECK(type IN ('WOOWEDU_QUIZ','QUICK_QUIZ','DOCUMENT_QUIZ','HANDWRITING','PHOTO','RACE','VOTE')), status text NOT NULL DEFAULT 'DRAFT' CHECK(status IN ('DRAFT','PUBLISHED','RUNNING','PAUSED','FINISHED','RESULTS_PUBLISHED')), mode text, target_json jsonb NOT NULL DEFAULT '{}'::jsonb, settings_json jsonb NOT NULL DEFAULT '{}'::jsonb, version bigint NOT NULL DEFAULT 1, current_question_index integer, started_at timestamptz, deadline_at timestamptz, paused_remaining_sec integer, ended_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- question_banks
CREATE TABLE question_banks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), school_id uuid NOT NULL REFERENCES schools(id), external_woowedu_id text, title text NOT NULL, source text NOT NULL DEFAULT 'WOOWEDU',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(school_id,external_woowedu_id)
);

-- questions
CREATE TABLE questions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), bank_id uuid NOT NULL REFERENCES question_banks(id), external_id text, type text NOT NULL, body_json jsonb NOT NULL, answer_key_encrypted bytea, version integer NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(bank_id,external_id,version)
);

-- activity_questions
CREATE TABLE activity_questions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), activity_id uuid NOT NULL REFERENCES activities(id), question_id uuid REFERENCES questions(id), position integer NOT NULL CHECK(position > 0), content_snapshot_json jsonb NOT NULL, answer_key_snapshot_encrypted bytea,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(activity_id,position)
);

-- quiz_attempts
CREATE TABLE quiz_attempts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), activity_id uuid NOT NULL REFERENCES activities(id), participant_id uuid NOT NULL REFERENCES session_participants(id), status text NOT NULL DEFAULT 'STARTED' CHECK(status IN ('STARTED','SUBMITTED','EXPIRED','GRADED')), started_at timestamptz NOT NULL DEFAULT now(), submitted_at timestamptz, score numeric(9,3),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(activity_id,participant_id)
);

-- quiz_answers
CREATE TABLE quiz_answers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), attempt_id uuid NOT NULL REFERENCES quiz_attempts(id), activity_question_id uuid NOT NULL REFERENCES activity_questions(id), revision integer NOT NULL DEFAULT 1, answer_json jsonb NOT NULL, submitted_at timestamptz NOT NULL DEFAULT now(), is_final boolean NOT NULL DEFAULT false,
  UNIQUE(attempt_id,activity_question_id,revision)
);

-- results
CREATE TABLE results (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), activity_id uuid NOT NULL REFERENCES activities(id), participant_id uuid NOT NULL REFERENCES session_participants(id), score numeric(9,3), feedback_json jsonb NOT NULL DEFAULT '{}'::jsonb, published_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(activity_id,participant_id)
);

-- assets
CREATE TABLE assets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), school_id uuid NOT NULL REFERENCES schools(id), owner_id uuid NOT NULL REFERENCES users(id), object_key text NOT NULL UNIQUE, filename text NOT NULL, mime text NOT NULL, size bigint NOT NULL CHECK(size>=0), checksum text NOT NULL, scan_status text NOT NULL DEFAULT 'PENDING',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- file_deliveries
CREATE TABLE file_deliveries (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), asset_id uuid NOT NULL REFERENCES assets(id), target_json jsonb NOT NULL, sent_at timestamptz, status text NOT NULL DEFAULT 'QUEUED',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- answer_submissions
CREATE TABLE answer_submissions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), activity_id uuid NOT NULL REFERENCES activities(id), participant_id uuid NOT NULL REFERENCES session_participants(id), type text NOT NULL, payload_json jsonb NOT NULL DEFAULT '{}'::jsonb, asset_id uuid REFERENCES assets(id), submitted_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(activity_id,participant_id)
);

-- groups
CREATE TABLE groups (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), name text NOT NULL, version bigint NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(session_id,name)
);

-- group_members
CREATE TABLE group_members (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), group_id uuid NOT NULL REFERENCES groups(id), participant_id uuid NOT NULL REFERENCES session_participants(id), joined_at timestamptz NOT NULL DEFAULT now(), left_at timestamptz
);

-- point_transactions
CREATE TABLE point_transactions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), participant_id uuid NOT NULL REFERENCES session_participants(id), delta integer NOT NULL CHECK(delta BETWEEN -100 AND 100), reason text, actor_id uuid NOT NULL REFERENCES users(id), idem_key text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(actor_id,idem_key)
);

-- hand_queue
CREATE TABLE hand_queue (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), participant_id uuid NOT NULL REFERENCES session_participants(id), status text NOT NULL CHECK(status IN ('RAISED','CALLED','CLEARED')), raised_at timestamptz NOT NULL DEFAULT now(), called_at timestamptz, cleared_at timestamptz
);

-- lottery_draws
CREATE TABLE lottery_draws (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), eligible_json jsonb NOT NULL, selected_participant_id uuid NOT NULL REFERENCES session_participants(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

-- votes
CREATE TABLE votes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), activity_id uuid NOT NULL REFERENCES activities(id), status text NOT NULL DEFAULT 'OPEN' CHECK(status IN ('OPEN','CLOSED')), target_json jsonb NOT NULL DEFAULT '{}'::jsonb, ends_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- vote_options
CREATE TABLE vote_options (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), vote_id uuid NOT NULL REFERENCES votes(id), source_submission_id uuid REFERENCES answer_submissions(id), caption text NOT NULL
);

-- vote_ballots
CREATE TABLE vote_ballots (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), vote_id uuid NOT NULL REFERENCES votes(id), participant_id uuid NOT NULL REFERENCES session_participants(id), option_id uuid NOT NULL REFERENCES vote_options(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(vote_id,participant_id)
);

-- timers
CREATE TABLE timers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), duration_sec integer NOT NULL CHECK(duration_sec>0), deadline_at timestamptz, paused_remaining_sec integer, status text NOT NULL DEFAULT 'READY' CHECK(status IN ('READY','RUNNING','PAUSED','FINISHED')), show_on_students boolean NOT NULL DEFAULT false, target_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- device_commands
CREATE TABLE device_commands (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), actor_id uuid NOT NULL REFERENCES users(id), device_id uuid NOT NULL REFERENCES devices(id), command_type text NOT NULL, status text NOT NULL DEFAULT 'PENDING', ack_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- screen_share_sessions
CREATE TABLE screen_share_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), mode text NOT NULL CHECK(mode IN ('DEMO','PROJECTION')), source_device_id uuid NOT NULL REFERENCES devices(id), target_json jsonb NOT NULL, status text NOT NULL DEFAULT 'REQUESTED', started_at timestamptz, ended_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- recordings
CREATE TABLE recordings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid NOT NULL REFERENCES class_sessions(id), asset_id uuid REFERENCES assets(id), status text NOT NULL, source text NOT NULL, fps integer, quality text, started_at timestamptz, ended_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- audit_logs
CREATE TABLE audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), session_id uuid REFERENCES class_sessions(id), actor_id uuid REFERENCES users(id), action text NOT NULL, resource_type text NOT NULL, resource_id uuid, metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Deferred relation: boards active_page_id references board_pages
ALTER TABLE boards ADD CONSTRAINT fk_board_active_page FOREIGN KEY (active_page_id) REFERENCES board_pages(id);

CREATE UNIQUE INDEX uq_group_active_participant ON group_members(participant_id) WHERE left_at IS NULL;
CREATE UNIQUE INDEX uq_hand_active ON hand_queue(session_id,participant_id) WHERE status IN ('RAISED','CALLED');
CREATE INDEX ix_participant_presence ON session_participants(session_id,state,last_seen_at);
CREATE INDEX ix_session_events_cursor ON session_events(session_id,sequence);
CREATE INDEX ix_outbox_pending ON outbox_events(created_at) WHERE published_at IS NULL;
CREATE INDEX ix_board_ops_replay ON board_operations(page_id,seq);
CREATE INDEX ix_activity_session_state ON activities(session_id,status);
CREATE INDEX ix_attempt_activity ON quiz_attempts(activity_id,status);
CREATE INDEX ix_answer_attempt ON quiz_answers(attempt_id,activity_question_id,revision DESC);
CREATE INDEX ix_files_session ON file_deliveries(session_id,status);
CREATE INDEX ix_points_student ON point_transactions(participant_id,created_at DESC);
CREATE INDEX ix_audit_session ON audit_logs(session_id,created_at DESC);
CREATE INDEX ix_devices_user ON devices(user_id,platform);
CREATE INDEX ix_timer_session ON timers(session_id,status);
CREATE INDEX ix_share_session ON screen_share_sessions(session_id,status);
CREATE INDEX ix_result_activity ON results(activity_id,published_at);
COMMIT;

-- =====================================================================
-- v4 additive: tabel pendukung (teacher_class_assignments, session_invite_tokens,
-- participant_live_state, dll). Run setelah blok di atas.
-- =====================================================================
BEGIN;
CREATE TABLE IF NOT EXISTS teacher_class_assignments (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 class_id uuid NOT NULL REFERENCES classes(id),
 teacher_user_id uuid NOT NULL REFERENCES users(id),
 role text NOT NULL DEFAULT 'OWNER' CHECK (role IN ('OWNER','CO_TEACHER')),
 active boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(class_id,teacher_user_id)
);
CREATE TABLE IF NOT EXISTS participant_activity_targets (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 activity_id uuid NOT NULL REFERENCES activities(id),
 participant_id uuid NOT NULL REFERENCES session_participants(id),
 assigned_at timestamptz NOT NULL DEFAULT now(),
 status text NOT NULL DEFAULT 'ASSIGNED' CHECK(status IN ('ASSIGNED','DELIVERED','STARTED','COMPLETED','SKIPPED')),
 UNIQUE(activity_id,participant_id)
);
CREATE TABLE IF NOT EXISTS broadcast_recipients (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 broadcast_id uuid NOT NULL REFERENCES broadcasts(id),
 participant_id uuid NOT NULL REFERENCES session_participants(id),
 delivery_state text NOT NULL DEFAULT 'QUEUED',
 delivered_at timestamptz,
 UNIQUE(broadcast_id,participant_id)
);
CREATE TABLE IF NOT EXISTS device_command_receipts (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 command_id uuid NOT NULL REFERENCES device_commands(id),
 device_id uuid NOT NULL REFERENCES devices(id),
 delivery_status text NOT NULL CHECK(delivery_status IN ('SENT','RECEIVED','APPLIED','FAILED')),
 error_code text,
 received_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(command_id,device_id,delivery_status)
);
CREATE TABLE IF NOT EXISTS session_invite_tokens (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 session_id uuid NOT NULL REFERENCES class_sessions(id),
 token_hash text NOT NULL UNIQUE,
 expires_at timestamptz NOT NULL,
 revoked_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS event_delivery_receipts (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 session_event_id uuid NOT NULL REFERENCES session_events(id),
 device_id uuid NOT NULL REFERENCES devices(id),
 state text NOT NULL CHECK(state IN ('SENT','RECEIVED','APPLIED','FAILED')),
 acknowledged_at timestamptz,
 error_code text,
 UNIQUE(session_event_id,device_id)
);
CREATE TABLE IF NOT EXISTS participant_live_state (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 session_id uuid NOT NULL REFERENCES class_sessions(id),
 participant_id uuid NOT NULL REFERENCES session_participants(id),
 foreground_activity_id uuid REFERENCES activities(id),
 current_broadcast_id uuid REFERENCES broadcasts(id),
 is_locked boolean NOT NULL DEFAULT false,
 state_version bigint NOT NULL DEFAULT 1,
 updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(session_id,participant_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_activity_question_pair ON activity_questions(id,activity_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_quiz_attempt_pair ON quiz_attempts(id,activity_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_vote_option_pair ON vote_options(id,vote_id);
CREATE INDEX IF NOT EXISTS idx_participant_targets_activity ON participant_activity_targets(activity_id,status);
CREATE INDEX IF NOT EXISTS idx_live_state_session ON participant_live_state(session_id,state_version);
CREATE INDEX IF NOT EXISTS idx_delivery_event ON event_delivery_receipts(session_event_id,state);
COMMIT;

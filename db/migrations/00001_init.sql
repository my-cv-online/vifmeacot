-- Skema Tahap 1 untuk sistem PFD / PFMEA / Control Plan.
-- Target: PostgreSQL 18 (memakai uuidv7() bawaan). Alat migrasi: goose (pressly/goose v3).
-- Penjelasan ada di docs/04-data-model.md. Jangan diubah setelah pernah dijalankan di mana pun;
-- buat migrasi baru dengan nomor berikutnya.

-- +goose Up

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ---------------------------------------------------------------------------
-- Tipe enum. Nilai untuk tahap berikutnya sudah ada agar Tahap 2 tidak perlu
-- migrasi enum; kode Tahap 1 tidak boleh membuat baris yang memakai nilai tersebut.
-- ---------------------------------------------------------------------------
CREATE TYPE user_role               AS ENUM ('admin', 'approver', 'reviewer', 'author', 'viewer');
CREATE TYPE auth_source             AS ENUM ('local', 'ldap');
CREATE TYPE package_kind            AS ENUM ('general', 'model');
CREATE TYPE package_status          AS ENUM ('draft', 'in_review', 'approved', 'released', 'superseded');
CREATE TYPE doc_type                AS ENUM ('PFD', 'PFMEA', 'CP');
CREATE TYPE methodology             AS ENUM ('PFD', 'AIAG4', 'AIAGVDA1', 'APQP2', 'CP1');
CREATE TYPE step_symbol             AS ENUM ('operation', 'inspection', 'operation_inspection', 'transport', 'storage', 'delay', 'decision');
CREATE TYPE step_kind               AS ENUM ('normal', 'rework');
CREATE TYPE general_mode            AS ENUM ('mandatory', 'optional');
CREATE TYPE row_origin              AS ENUM ('local', 'general');
CREATE TYPE sync_status             AS ENUM ('in_sync', 'override', 'review', 'detached');
CREATE TYPE flow_kind               AS ENUM ('normal', 'ng', 'rework', 'scrap', 'return');
CREATE TYPE characteristic_kind     AS ENUM ('product', 'process');
CREATE TYPE effect_level            AS ENUM ('your_plant', 'ship_to_plant', 'end_user', 'unspecified');
CREATE TYPE control_kind            AS ENUM ('prevention', 'detection');
CREATE TYPE action_kind             AS ENUM ('prevention', 'detection', 'design_change', 'other');
CREATE TYPE action_status           AS ENUM ('open', 'in_progress', 'done', 'cancelled');
CREATE TYPE cp_phase                AS ENUM ('prototype', 'pre_launch', 'safe_launch', 'production');
CREATE TYPE freq_basis              AS ENUM ('pct100', 'volume', 'time', 'event', 'other');
CREATE TYPE control_method_category AS ENUM ('visual_100', 'visual_sampling', 'gauge_attribute', 'gauge_variable',
                                             'automated_inspection', 'functional_test', 'spc', 'error_proofing',
                                             'procedure', 'maintenance', 'other');
CREATE TYPE rating_dimension        AS ENUM ('S', 'O', 'D');
CREATE TYPE finding_level           AS ENUM ('error', 'warning', 'info');
CREATE TYPE finding_status          AS ENUM ('open', 'fixed', 'waived');
CREATE TYPE check_trigger           AS ENUM ('save', 'manual', 'nightly', 'sync', 'import');
CREATE TYPE link_state              AS ENUM ('in_sync', 'pending', 'syncing', 'failed', 'awaiting_customer');
CREATE TYPE sync_run_status         AS ENUM ('queued', 'running', 'done', 'failed');
CREATE TYPE sync_action             AS ENUM ('update', 'insert', 'delete', 'skip_override', 'conflict');
CREATE TYPE export_status           AS ENUM ('queued', 'running', 'done', 'failed');

-- ---------------------------------------------------------------------------
-- Fungsi trigger umum
-- ---------------------------------------------------------------------------

-- BEFORE UPDATE: perbarui updated_at dan naikkan version hanya jika ada kolom
-- bermakna bagi pengguna yang berubah. TG_ARGV berisi kolom turunan yang diabaikan.
-- +goose StatementBegin
CREATE FUNCTION trg_touch_row() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
  ignored text[] := ARRAY['updated_at', 'version'] || TG_ARGV::text[];
BEGIN
  NEW.updated_at := now();
  IF (to_jsonb(NEW) - ignored) IS DISTINCT FROM (to_jsonb(OLD) - ignored) THEN
    NEW.version := OLD.version + 1;
  ELSE
    NEW.version := OLD.version;
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- BEFORE UPDATE untuk tabel yang punya updated_at tetapi tidak punya kolom version.
-- +goose StatementBegin
CREATE FUNCTION trg_touch_updated_at() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- Trigger audit tingkat statement. TG_ARGV[0] = 'content' juga menaikkan
-- packages.content_version satu kali per paket yang tersentuh; TG_ARGV[1..] adalah kolom
-- turunan tabel itu yang tidak dihitung sebagai perubahan (cache, penghitung).
-- API mengisi app.user_id, app.request_id dan app.source dengan SET LOCAL.
-- +goose StatementBegin
CREATE FUNCTION trg_audit_insert() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  IF TG_NARGS > 0 AND TG_ARGV[0] = 'content' THEN
    UPDATE packages p SET content_version = p.content_version + 1
    WHERE p.id IN (SELECT DISTINCT (to_jsonb(_n) ->> 'package_id')::uuid FROM new_rows _n);
  END IF;
  INSERT INTO audit_log (user_id, package_id, table_name, row_id, action, old_row, new_row, request_id, source)
  SELECT nullif(current_setting('app.user_id', true), '')::uuid,
         coalesce((to_jsonb(_n) ->> 'package_id')::uuid, CASE WHEN TG_TABLE_NAME = 'packages' THEN (to_jsonb(_n) ->> 'id')::uuid END),
         TG_TABLE_NAME, (to_jsonb(_n) ->> 'id')::uuid, 'insert', NULL, to_jsonb(_n),
         nullif(current_setting('app.request_id', true), ''),
         coalesce(nullif(current_setting('app.source', true), ''), 'system')
  FROM new_rows _n;
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- Trigger AFTER UPDATE tingkat statement: hanya menyimpan kolom yang berubah dan
-- melewati baris tanpa perubahan yang bermakna (cache turunan, penghitung).
-- +goose StatementBegin
CREATE FUNCTION trg_audit_update() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
  ignored      text[] := ARRAY['updated_at', 'version'] || TG_ARGV[1:TG_NARGS];
  changed_pkgs uuid[];
BEGIN
  WITH d AS (
    SELECT coalesce((to_jsonb(_n) ->> 'package_id')::uuid,
                    CASE WHEN TG_TABLE_NAME = 'packages' THEN (to_jsonb(_n) ->> 'id')::uuid END) AS package_id,
           (to_jsonb(_n) ->> 'id')::uuid AS row_id,
           to_jsonb(_o) - ignored AS o_j,
           to_jsonb(_n) - ignored AS n_j
    FROM new_rows _n
    JOIN old_rows _o ON (to_jsonb(_o) ->> 'id') = (to_jsonb(_n) ->> 'id')
  ), c AS (
    SELECT d.package_id, d.row_id,
           (SELECT jsonb_object_agg(k, d.o_j -> k) FROM jsonb_object_keys(d.n_j) AS k WHERE d.n_j -> k IS DISTINCT FROM d.o_j -> k) AS old_diff,
           (SELECT jsonb_object_agg(k, d.n_j -> k) FROM jsonb_object_keys(d.n_j) AS k WHERE d.n_j -> k IS DISTINCT FROM d.o_j -> k) AS new_diff
    FROM d
  ), ins AS (
    INSERT INTO audit_log (user_id, package_id, table_name, row_id, action, old_row, new_row, request_id, source)
    SELECT nullif(current_setting('app.user_id', true), '')::uuid, c.package_id, TG_TABLE_NAME, c.row_id, 'update',
           c.old_diff, c.new_diff,
           nullif(current_setting('app.request_id', true), ''),
           coalesce(nullif(current_setting('app.source', true), ''), 'system')
    FROM c
    WHERE c.new_diff IS NOT NULL
    RETURNING package_id
  )
  SELECT array_agg(DISTINCT package_id) INTO changed_pkgs FROM ins;

  IF TG_NARGS > 0 AND TG_ARGV[0] = 'content' AND changed_pkgs IS NOT NULL THEN
    UPDATE packages p SET content_version = p.content_version + 1 WHERE p.id = ANY (changed_pkgs);
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION trg_audit_delete() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  IF TG_NARGS > 0 AND TG_ARGV[0] = 'content' THEN
    UPDATE packages p SET content_version = p.content_version + 1
    WHERE p.id IN (SELECT DISTINCT (to_jsonb(_o) ->> 'package_id')::uuid FROM old_rows _o);
  END IF;
  INSERT INTO audit_log (user_id, package_id, table_name, row_id, action, old_row, new_row, request_id, source)
  SELECT nullif(current_setting('app.user_id', true), '')::uuid,
         coalesce((to_jsonb(_o) ->> 'package_id')::uuid, CASE WHEN TG_TABLE_NAME = 'packages' THEN (to_jsonb(_o) ->> 'id')::uuid END),
         TG_TABLE_NAME, (to_jsonb(_o) ->> 'id')::uuid, 'delete', to_jsonb(_o), NULL,
         nullif(current_setting('app.request_id', true), ''),
         coalesce(nullif(current_setting('app.source', true), ''), 'system')
  FROM old_rows _o;
  RETURN NULL;
END $$;
-- +goose StatementEnd

-- Pasang tiga trigger audit tingkat statement (dan kenaikan content_version) ke sebuah tabel.
-- +goose StatementBegin
CREATE FUNCTION attach_audit(tbl regclass, is_content boolean, derived text[] DEFAULT '{}') RETURNS void
LANGUAGE plpgsql AS $$
DECLARE
  t   text := tbl::text;
  arg text := (SELECT string_agg(quote_literal(x), ', ')
               FROM unnest(ARRAY[CASE WHEN is_content THEN 'content' ELSE 'plain' END] || derived) AS x);
BEGIN
  EXECUTE format('CREATE TRIGGER %s_audit_ins AFTER INSERT ON %s REFERENCING NEW TABLE AS new_rows FOR EACH STATEMENT EXECUTE FUNCTION trg_audit_insert(%s)', t, t, arg);
  EXECUTE format('CREATE TRIGGER %s_audit_upd AFTER UPDATE ON %s REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows FOR EACH STATEMENT EXECUTE FUNCTION trg_audit_update(%s)', t, t, arg);
  EXECUTE format('CREATE TRIGGER %s_audit_del AFTER DELETE ON %s REFERENCING OLD TABLE AS old_rows FOR EACH STATEMENT EXECUTE FUNCTION trg_audit_delete(%s)', t, t, arg);
END $$;
-- +goose StatementEnd

-- ---------------------------------------------------------------------------
-- Pengguna dan sesi
-- ---------------------------------------------------------------------------
CREATE TABLE users (
  id            uuid PRIMARY KEY DEFAULT uuidv7(),
  username      text NOT NULL UNIQUE CHECK (username ~ '^[a-z0-9._-]{3,64}$'),
  display_name  text NOT NULL,
  email         text,
  department    text,
  role          user_role   NOT NULL DEFAULT 'viewer',
  auth_source   auth_source NOT NULL DEFAULT 'local',
  password_hash text,                                   -- string PHC argon2id; NULL untuk pengguna LDAP
  is_active     boolean NOT NULL DEFAULT true,
  last_login_at timestamptz,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  version       integer NOT NULL DEFAULT 1,
  CHECK (auth_source = 'ldap' OR password_hash IS NOT NULL)
);

CREATE TABLE sessions (
  token_hash   bytea PRIMARY KEY,                       -- sha256 dari token cookie acak
  user_id      uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at   timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  expires_at   timestamptz NOT NULL,
  ip           inet,
  user_agent   text
);
CREATE INDEX sessions_user_idx ON sessions (user_id);
CREATE INDEX sessions_expires_idx ON sessions (expires_at);

-- ---------------------------------------------------------------------------
-- Master data
-- ---------------------------------------------------------------------------
CREATE TABLE customers (
  id                       uuid PRIMARY KEY DEFAULT uuidv7(),
  code                     text NOT NULL UNIQUE CHECK (code ~ '^[A-Z0-9]{2,8}$'),
  name                     text NOT NULL,
  rpn_action_threshold     integer  CHECK (rpn_action_threshold BETWEEN 1 AND 1000),  -- F08; NULL = nonaktif (bawaan)
  cc_min_severity          smallint CHECK (cc_min_severity BETWEEN 1 AND 10),        -- S04; NULL = nonaktif
  requires_change_approval boolean NOT NULL DEFAULT false,                            -- dipakai mulai Tahap 2
  notes                    text,
  is_active                boolean NOT NULL DEFAULT true,
  created_at               timestamptz NOT NULL DEFAULT now(),
  updated_at               timestamptz NOT NULL DEFAULT now(),
  version                  integer NOT NULL DEFAULT 1
);

-- Kelas special characteristic internal (tingkat perusahaan), mis. CC dan SC.
CREATE TABLE sc_symbols (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  code        text NOT NULL UNIQUE CHECK (code ~ '^[A-Z0-9]{1,6}$'),
  name        text NOT NULL,
  is_critical boolean NOT NULL DEFAULT false,            -- kelas keselamatan/regulasi; dipakai S04
  description text,
  sort_order  integer NOT NULL DEFAULT 0,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  version     integer NOT NULL DEFAULT 1
);

-- Tabel konversi simbol customer (S02): simbol yang dicetak untuk tiap customer.
CREATE TABLE customer_sc_symbols (
  id              uuid NOT NULL DEFAULT uuidv7() UNIQUE,  -- kunci pengganti untuk trigger audit
  customer_id     uuid NOT NULL REFERENCES customers (id) ON DELETE CASCADE,
  sc_symbol_id    uuid NOT NULL REFERENCES sc_symbols (id) ON DELETE RESTRICT,
  customer_symbol text NOT NULL,
  customer_label  text,
  PRIMARY KEY (customer_id, sc_symbol_id)
);

CREATE TABLE parts (
  id             uuid PRIMARY KEY DEFAULT uuidv7(),
  customer_id    uuid NOT NULL REFERENCES customers (id),
  part_no        text NOT NULL,
  part_name      text NOT NULL,
  change_level   text,
  product_family text,
  is_active      boolean NOT NULL DEFAULT true,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  version        integer NOT NULL DEFAULT 1,
  UNIQUE (customer_id, part_no)
);

CREATE TABLE control_library (
  id                uuid PRIMARY KEY DEFAULT uuidv7(),
  name              text NOT NULL,
  kind              control_kind NOT NULL,
  method_category   control_method_category NOT NULL,
  d_min             smallint CHECK (d_min BETWEEN 1 AND 10),   -- D terbaik (terendah) yang bisa dibenarkan metode ini (R04)
  d_max             smallint CHECK (d_max BETWEEN 1 AND 10),
  is_error_proofing boolean NOT NULL DEFAULT false,
  is_strong         boolean NOT NULL DEFAULT false,            -- SPC / 100% otomatis / error-proofing (S03)
  description       text,
  is_active         boolean NOT NULL DEFAULT true,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  version           integer NOT NULL DEFAULT 1,
  UNIQUE (kind, name),
  CHECK (d_min IS NULL OR d_max IS NULL OR d_min <= d_max),
  CHECK (kind = 'detection' OR (d_min IS NULL AND d_max IS NULL))
);

-- Teks kriteria S/O/D per metodologi. Diisi perusahaan dari manual AIAG
-- berlisensi miliknya; aplikasi tidak pernah mengisi teks manual tersebut.
CREATE TABLE rating_criteria (
  id          uuid NOT NULL DEFAULT uuidv7() UNIQUE,  -- kunci pengganti untuk trigger audit
  methodology methodology NOT NULL CHECK (methodology IN ('AIAG4', 'AIAGVDA1')),
  dimension   rating_dimension NOT NULL,
  rating      smallint NOT NULL CHECK (rating BETWEEN 1 AND 10),
  criteria    text NOT NULL,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  updated_by  uuid REFERENCES users (id),
  PRIMARY KEY (methodology, dimension, rating)
);

-- Kamus untuk F12 (istilah yang tidak boleh dipakai, mis. "operator error" sebagai cause).
CREATE TABLE banned_terms (
  id         uuid PRIMARY KEY DEFAULT uuidv7(),
  term       text NOT NULL CHECK (length(btrim(term)) >= 2),
  applies_to text NOT NULL DEFAULT 'any'
             CHECK (applies_to IN ('any', 'cause', 'control', 'failure_mode', 'effect', 'action')),
  reason     text NOT NULL,
  suggestion text,
  is_active  boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  version    integer NOT NULL DEFAULT 1,
  UNIQUE (term, applies_to)
);

CREATE TABLE app_settings (
  id         uuid NOT NULL DEFAULT uuidv7() UNIQUE,  -- kunci pengganti untuk trigger audit
  key        text PRIMARY KEY,
  value      jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES users (id)
);

-- ---------------------------------------------------------------------------
-- Paket, dokumen, revisi
-- ---------------------------------------------------------------------------
CREATE TABLE packages (
  id               uuid PRIMARY KEY DEFAULT uuidv7(),
  kind             package_kind NOT NULL,
  code             text NOT NULL UNIQUE CHECK (code ~ '^[A-Z0-9][A-Z0-9-]{1,23}$'),
  name             text NOT NULL,
  customer_id      uuid REFERENCES customers (id),
  part_id          uuid REFERENCES parts (id),
  pfmea_method     methodology NOT NULL DEFAULT 'AIAG4' CHECK (pfmea_method IN ('AIAG4', 'AIAGVDA1')),
  cp_format        methodology NOT NULL DEFAULT 'APQP2' CHECK (cp_format IN ('APQP2', 'CP1')),
  status           package_status NOT NULL DEFAULT 'draft',
  revision         integer NOT NULL DEFAULT 0,          -- revisi rilis terakhir (paket general: revisi template)
  doc_seq          integer,                             -- nomor urut per customer untuk nomor dokumen
  owner_id         uuid NOT NULL REFERENCES users (id),
  plant            text,
  core_team        text,
  override_policy  jsonb NOT NULL DEFAULT '{}'::jsonb,  -- hanya paket general: {"tabel": ["kolom", ...]}
  last_reviewed_at timestamptz NOT NULL DEFAULT now(),  -- W04
  content_version  bigint NOT NULL DEFAULT 0,           -- dinaikkan trigger pada setiap perubahan konten
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  version          integer NOT NULL DEFAULT 1,
  CHECK ((kind = 'general' AND customer_id IS NULL AND part_id IS NULL AND doc_seq IS NULL)
      OR (kind = 'model' AND customer_id IS NOT NULL AND part_id IS NOT NULL AND doc_seq IS NOT NULL)),
  UNIQUE (customer_id, doc_seq)
);
-- Tahap 1 hanya punya satu Template General. Hapus indeks ini saat template per family hadir.
CREATE UNIQUE INDEX packages_single_general_idx ON packages (kind) WHERE kind = 'general';
CREATE INDEX packages_customer_idx ON packages (customer_id);
CREATE INDEX packages_search_idx ON packages USING gin ((code || ' ' || name) gin_trgm_ops);

CREATE TABLE package_members (
  id         uuid NOT NULL DEFAULT uuidv7() UNIQUE,  -- kunci pengganti untuk trigger audit
  package_id uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  can_edit   boolean NOT NULL DEFAULT true,
  PRIMARY KEY (package_id, user_id)
);

CREATE TABLE documents (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id  uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  doc_type    doc_type NOT NULL,
  methodology methodology NOT NULL,
  doc_no      text NOT NULL UNIQUE,
  revision    integer NOT NULL DEFAULT 0,
  header      jsonb NOT NULL DEFAULT '{}'::jsonb,       -- isian header form, kuncinya ada di docs/04-data-model.md
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  version     integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, doc_type),
  CHECK ((doc_type = 'PFD' AND methodology = 'PFD')
      OR (doc_type = 'PFMEA' AND methodology IN ('AIAG4', 'AIAGVDA1'))
      OR (doc_type = 'CP' AND methodology IN ('APQP2', 'CP1')))
);

-- Salinan beku semua baris konten sebuah paket (Tahap 1: rilis Template General).
CREATE TABLE package_revisions (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id  uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  rev_no      integer NOT NULL CHECK (rev_no >= 1),
  snapshot    jsonb NOT NULL,
  content_version bigint NOT NULL,                     -- content_version paket general saat dirilis
  change_note text NOT NULL CHECK (length(btrim(change_note)) >= 3),
  created_by  uuid NOT NULL REFERENCES users (id),
  created_at  timestamptz NOT NULL DEFAULT now(),
  UNIQUE (package_id, rev_no)
);

-- ---------------------------------------------------------------------------
-- Konten proses. Setiap tabel konten membawa kolom tautan template yang sama:
--   origin, source_id, source_rev, sync_status, overrides, detach_reason
-- origin = 'general' menandai salinan tertaut dari baris Template General (lihat docs/07-template-general.md).
-- Foreign key gabungan (package_id, x_id) menjaga setiap referensi tetap di dalam satu paket.
-- ---------------------------------------------------------------------------
CREATE TABLE process_steps (
  id                  uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id          uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  op_no               text NOT NULL CHECK (op_no ~ '^[0-9]{1,4}[A-Z]?$'),
  seq                 integer NOT NULL,
  name                text NOT NULL CHECK (length(btrim(name)) BETWEEN 1 AND 200),
  function            text,                               -- fungsi / tujuan step (PFMEA)
  symbol              step_symbol NOT NULL DEFAULT 'operation',
  kind                step_kind NOT NULL DEFAULT 'normal',
  is_optional         boolean NOT NULL DEFAULT false,     -- opsional dalam alur (mis. rework)
  machines            text[] NOT NULL DEFAULT '{}',       -- mesin / perangkat / jig / alat
  inputs              text,
  outputs             text,
  department          text,                               -- PIC / departemen
  wi_ref              text,                               -- nomor instruksi kerja + revisi (metadata)
  not_analyzed        boolean NOT NULL DEFAULT false,     -- pengecualian K01
  not_analyzed_reason text,
  general_mode        general_mode,                       -- hanya diisi di paket general
  origin              row_origin NOT NULL DEFAULT 'local',
  source_id           uuid,
  source_rev          integer,
  sync_status         sync_status,
  overrides           jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason       text,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  version             integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  CONSTRAINT process_steps_op_no_uq UNIQUE (package_id, op_no) DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT process_steps_seq_uq UNIQUE (package_id, seq) DEFERRABLE INITIALLY DEFERRED,
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL)),
  CHECK (NOT not_analyzed OR coalesce(btrim(not_analyzed_reason), '') <> '')
);
CREATE INDEX process_steps_source_idx ON process_steps (package_id, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX process_steps_search_idx ON process_steps USING gin (name gin_trgm_ops);

CREATE TABLE step_flows (
  id            uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id    uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  from_step_id  uuid NOT NULL,
  to_step_id    uuid,                                     -- NULL = keluar dari alur (scrap, hold, kembali ke supplier)
  kind          flow_kind NOT NULL,
  disposition   text,                                     -- wajib untuk ng / rework / scrap (W02)
  label         text,
  origin        row_origin NOT NULL DEFAULT 'local',
  source_id     uuid,
  source_rev    integer,
  sync_status   sync_status,
  overrides     jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  version       integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  FOREIGN KEY (package_id, from_step_id) REFERENCES process_steps (package_id, id) ON DELETE CASCADE,
  FOREIGN KEY (package_id, to_step_id) REFERENCES process_steps (package_id, id) ON DELETE CASCADE,
  CHECK (kind <> 'normal' OR to_step_id IS NOT NULL),
  CHECK (to_step_id IS NULL OR from_step_id <> to_step_id),
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL))
);
CREATE INDEX step_flows_from_idx ON step_flows (package_id, from_step_id);
CREATE INDEX step_flows_source_idx ON step_flows (package_id, source_id) WHERE source_id IS NOT NULL;

CREATE TABLE characteristics (
  id            uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id    uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  step_id       uuid NOT NULL,
  char_no       text NOT NULL CHECK (char_no ~ '^[0-9]{1,4}[A-Z]?-[0-9]{2,3}$'),   -- mis. 30-01
  kind          characteristic_kind NOT NULL,
  name          text NOT NULL CHECK (length(btrim(name)) BETWEEN 1 AND 200),
  spec          text,                                    -- spesifikasi / toleransi produk atau proses
  lsl           numeric,
  usl           numeric,
  unit          text,
  sc_symbol_id  uuid REFERENCES sc_symbols (id),
  seq           integer NOT NULL DEFAULT 0,
  origin        row_origin NOT NULL DEFAULT 'local',
  source_id     uuid,
  source_rev    integer,
  sync_status   sync_status,
  overrides     jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  version       integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  CONSTRAINT characteristics_char_no_uq UNIQUE (package_id, char_no) DEFERRABLE INITIALLY DEFERRED,
  FOREIGN KEY (package_id, step_id) REFERENCES process_steps (package_id, id) ON DELETE CASCADE,
  CHECK (lsl IS NULL OR usl IS NULL OR lsl <= usl),
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL))
);
CREATE INDEX characteristics_step_idx ON characteristics (package_id, step_id);
CREATE INDEX characteristics_source_idx ON characteristics (package_id, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX characteristics_search_idx ON characteristics USING gin (name gin_trgm_ops);

CREATE TABLE failure_modes (
  id                uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id        uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  step_id           uuid NOT NULL,
  characteristic_id uuid NOT NULL,                       -- requirement yang dilanggar failure mode ini
  text              text NOT NULL DEFAULT '',
  seq               integer NOT NULL DEFAULT 0,
  origin            row_origin NOT NULL DEFAULT 'local',
  source_id         uuid,
  source_rev        integer,
  sync_status       sync_status,
  overrides         jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason     text,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  version           integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  -- menghapus step atau karakteristik yang masih punya baris PFMEA ditolak (docs/05-api.md §7)
  FOREIGN KEY (package_id, step_id) REFERENCES process_steps (package_id, id) ON DELETE RESTRICT,
  FOREIGN KEY (package_id, characteristic_id) REFERENCES characteristics (package_id, id) ON DELETE RESTRICT,
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL))
);
CREATE INDEX failure_modes_step_idx ON failure_modes (package_id, step_id);
CREATE INDEX failure_modes_char_idx ON failure_modes (characteristic_id);
CREATE INDEX failure_modes_source_idx ON failure_modes (package_id, source_id) WHERE source_id IS NOT NULL;

CREATE TABLE failure_effects (
  id              uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id      uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  failure_mode_id uuid NOT NULL,
  level           effect_level NOT NULL DEFAULT 'unspecified',
  text            text NOT NULL DEFAULT '',
  s               smallint CHECK (s BETWEEN 1 AND 10),
  seq             integer NOT NULL DEFAULT 0,
  origin          row_origin NOT NULL DEFAULT 'local',
  source_id       uuid,
  source_rev      integer,
  sync_status     sync_status,
  overrides       jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason   text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  version         integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  FOREIGN KEY (package_id, failure_mode_id) REFERENCES failure_modes (package_id, id) ON DELETE CASCADE,
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL))
);
CREATE INDEX failure_effects_fm_idx ON failure_effects (failure_mode_id);
CREATE INDEX failure_effects_source_idx ON failure_effects (package_id, source_id) WHERE source_id IS NOT NULL;

CREATE TABLE failure_causes (
  id              uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id      uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  failure_mode_id uuid NOT NULL,
  text            text NOT NULL DEFAULT '',
  seq             integer NOT NULL DEFAULT 0,
  origin          row_origin NOT NULL DEFAULT 'local',
  source_id       uuid,
  source_rev      integer,
  sync_status     sync_status,
  overrides       jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason   text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  version         integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  UNIQUE (id, failure_mode_id),
  FOREIGN KEY (package_id, failure_mode_id) REFERENCES failure_modes (package_id, id) ON DELETE CASCADE,
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL))
);
CREATE INDEX failure_causes_fm_idx ON failure_causes (failure_mode_id);
CREATE INDEX failure_causes_source_idx ON failure_causes (package_id, source_id) WHERE source_id IS NOT NULL;

-- Satu baris per pasangan failure mode + cause (satu baris worksheet AIAG 4th).
-- s adalah cache max(failure_effects.s) dari failure mode-nya, dijaga oleh trigger.
CREATE TABLE failure_chains (
  id               uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id       uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  failure_mode_id  uuid NOT NULL,
  failure_cause_id uuid NOT NULL UNIQUE,                  -- Tahap 1: satu chain per cause
  s                smallint CHECK (s BETWEEN 1 AND 10),
  o                smallint CHECK (o BETWEEN 1 AND 10),
  d                smallint CHECK (d BETWEEN 1 AND 10),
  rpn              integer GENERATED ALWAYS AS (s * o * d) STORED,
  ap               char(1) CHECK (ap IN ('H', 'M', 'L')),   -- AIAG-VDA, Tahap 2
  justification    text,                                  -- alasan tidak perlu aksi (lanjutan) (F07, F08)
  o_evidence       text,                                  -- data histori yang mendasari O rendah (R05)
  seq              integer NOT NULL DEFAULT 0,
  origin           row_origin NOT NULL DEFAULT 'local',
  source_id        uuid,
  source_rev       integer,
  sync_status      sync_status,
  overrides        jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason    text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  version          integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  FOREIGN KEY (package_id, failure_mode_id) REFERENCES failure_modes (package_id, id) ON DELETE CASCADE,
  FOREIGN KEY (failure_cause_id, failure_mode_id) REFERENCES failure_causes (id, failure_mode_id) ON DELETE CASCADE,
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL))
);
CREATE INDEX failure_chains_fm_idx ON failure_chains (failure_mode_id);
CREATE INDEX failure_chains_pkg_rpn_idx ON failure_chains (package_id, rpn DESC NULLS LAST);
CREATE INDEX failure_chains_source_idx ON failure_chains (package_id, source_id) WHERE source_id IS NOT NULL;

CREATE TABLE controls (
  id                    uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id            uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  failure_chain_id      uuid NOT NULL,
  kind                  control_kind NOT NULL,
  text                  text NOT NULL DEFAULT '',
  control_library_id    uuid REFERENCES control_library (id),
  is_system_control     boolean NOT NULL DEFAULT false,    -- pengecualian R02, mis. interlock ERP/MES
  system_control_reason text,
  seq                   integer NOT NULL DEFAULT 0,
  origin                row_origin NOT NULL DEFAULT 'local',
  source_id             uuid,
  source_rev            integer,
  sync_status           sync_status,
  overrides             jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason         text,
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now(),
  version               integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  FOREIGN KEY (package_id, failure_chain_id) REFERENCES failure_chains (package_id, id) ON DELETE CASCADE,
  CHECK (NOT is_system_control OR coalesce(btrim(system_control_reason), '') <> ''),
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL))
);
CREATE INDEX controls_chain_idx ON controls (failure_chain_id);
CREATE INDEX controls_pkg_kind_idx ON controls (package_id, kind);
CREATE INDEX controls_source_idx ON controls (package_id, source_id) WHERE source_id IS NOT NULL;

-- Recommended action (AIAG 4th). Khusus per paket: tidak pernah disalin dari Template General.
CREATE TABLE actions (
  id                  uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id          uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  failure_chain_id    uuid NOT NULL,
  kind                action_kind NOT NULL DEFAULT 'prevention',
  text                text NOT NULL DEFAULT '',
  responsible_user_id uuid REFERENCES users (id),
  responsible_text    text,                                -- PIC yang bukan pengguna sistem
  target_date         date,
  status              action_status NOT NULL DEFAULT 'open',
  action_taken        text,
  completed_on        date,
  new_s               smallint CHECK (new_s BETWEEN 1 AND 10),
  new_o               smallint CHECK (new_o BETWEEN 1 AND 10),
  new_d               smallint CHECK (new_d BETWEEN 1 AND 10),
  new_rpn             integer GENERATED ALWAYS AS (new_s * new_o * new_d) STORED,
  design_change_note  text,                                -- wajib jika new_s < s (F10)
  seq                 integer NOT NULL DEFAULT 0,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  version             integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  FOREIGN KEY (package_id, failure_chain_id) REFERENCES failure_chains (package_id, id) ON DELETE CASCADE,
  CHECK (status <> 'done' OR completed_on IS NOT NULL)
);
CREATE INDEX actions_chain_idx ON actions (failure_chain_id);
CREATE INDEX actions_open_idx ON actions (package_id, target_date) WHERE status IN ('open', 'in_progress');

CREATE TABLE cp_lines (
  id                uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id        uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  phase             cp_phase NOT NULL DEFAULT 'production',
  step_id           uuid NOT NULL,
  characteristic_id uuid NOT NULL,
  control_id        uuid,                                  -- kontrol PFMEA yang diterapkan baris ini (R01, R02)
  failure_mode_id   uuid,                                  -- failure mode PFMEA yang dicakup (R03)
  machines          text[] NOT NULL DEFAULT '{}',          -- bagian dari daftar mesin step (K05)
  eval_technique    text,                                  -- teknik evaluasi / pengukuran
  gauge             text,
  sample_size       text,
  sample_freq       text,
  freq_basis        freq_basis,
  control_method    text,
  is_error_proofing boolean NOT NULL DEFAULT false,
  ep_verify_freq    text,                                  -- frekuensi verifikasi error-proofing (R06)
  seq               integer NOT NULL DEFAULT 0,
  origin            row_origin NOT NULL DEFAULT 'local',
  source_id         uuid,
  source_rev        integer,
  sync_status       sync_status,
  overrides         jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason     text,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  version           integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  FOREIGN KEY (package_id, step_id) REFERENCES process_steps (package_id, id) ON DELETE RESTRICT,
  FOREIGN KEY (package_id, characteristic_id) REFERENCES characteristics (package_id, id) ON DELETE RESTRICT,
  FOREIGN KEY (package_id, control_id) REFERENCES controls (package_id, id) ON DELETE SET NULL (control_id),
  FOREIGN KEY (package_id, failure_mode_id) REFERENCES failure_modes (package_id, id) ON DELETE SET NULL (failure_mode_id),
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL))
);
CREATE INDEX cp_lines_step_idx ON cp_lines (package_id, phase, step_id);
CREATE INDEX cp_lines_char_idx ON cp_lines (characteristic_id);
CREATE INDEX cp_lines_control_idx ON cp_lines (control_id) WHERE control_id IS NOT NULL;
CREATE INDEX cp_lines_source_idx ON cp_lines (package_id, source_id) WHERE source_id IS NOT NULL;

-- 1:1 dengan cp_lines. Template A (APQP 2nd) memakai text; Template B (CP-1, Tahap 2)
-- memakai kolom terstruktur.
CREATE TABLE reaction_plans (
  id              uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id      uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  cp_line_id      uuid NOT NULL UNIQUE,
  text            text NOT NULL DEFAULT '',
  isolation       text,
  stop_process    text,
  recovery        text,
  owner_user_id   uuid REFERENCES users (id),
  owner_text      text,
  instruction_ref text,
  origin          row_origin NOT NULL DEFAULT 'local',
  source_id       uuid,
  source_rev      integer,
  sync_status     sync_status,
  overrides       jsonb NOT NULL DEFAULT '{}'::jsonb,
  detach_reason   text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  version         integer NOT NULL DEFAULT 1,
  UNIQUE (package_id, id),
  FOREIGN KEY (package_id, cp_line_id) REFERENCES cp_lines (package_id, id) ON DELETE CASCADE,
  CHECK ((origin = 'local' AND source_id IS NULL AND sync_status IS NULL)
      OR (origin = 'general' AND source_id IS NOT NULL AND sync_status IS NOT NULL))
);
CREATE INDEX reaction_plans_source_idx ON reaction_plans (package_id, source_id) WHERE source_id IS NOT NULL;

-- Jaga agar failure_chains.s selalu sama dengan severity effect tertinggi dari failure mode-nya.
-- +goose StatementBegin
CREATE FUNCTION trg_chain_s_on_insert() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  NEW.s := (SELECT max(e.s) FROM failure_effects e WHERE e.failure_mode_id = NEW.failure_mode_id);
  RETURN NEW;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION trg_effects_refresh_chain_s() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP <> 'INSERT' THEN
    UPDATE failure_chains c
       SET s = (SELECT max(e.s) FROM failure_effects e WHERE e.failure_mode_id = OLD.failure_mode_id)
     WHERE c.failure_mode_id = OLD.failure_mode_id;
  END IF;
  IF TG_OP = 'INSERT' OR (TG_OP = 'UPDATE' AND NEW.failure_mode_id IS DISTINCT FROM OLD.failure_mode_id) THEN
    UPDATE failure_chains c
       SET s = (SELECT max(e.s) FROM failure_effects e WHERE e.failure_mode_id = NEW.failure_mode_id)
     WHERE c.failure_mode_id = NEW.failure_mode_id;
  END IF;
  RETURN NULL;
END $$;
-- +goose StatementEnd

CREATE TRIGGER failure_chains_s_ins BEFORE INSERT ON failure_chains
  FOR EACH ROW EXECUTE FUNCTION trg_chain_s_on_insert();
CREATE TRIGGER failure_effects_chain_s AFTER INSERT OR DELETE OR UPDATE OF s, failure_mode_id ON failure_effects
  FOR EACH ROW EXECUTE FUNCTION trg_effects_refresh_chain_s();

-- ---------------------------------------------------------------------------
-- Aturan konsistensi dan temuan
-- ---------------------------------------------------------------------------
CREATE TABLE rules (
  id            uuid NOT NULL DEFAULT uuidv7() UNIQUE,  -- kunci pengganti untuk trigger audit
  code          text PRIMARY KEY CHECK (code ~ '^[KSRFCWT][0-9]{2}$'),
  title         text NOT NULL,                            -- bahasa Inggris, tampil di UI
  description   text NOT NULL,
  default_level finding_level NOT NULL,
  applies_to    text NOT NULL DEFAULT 'all' CHECK (applies_to IN ('all', 'AIAG4', 'AIAGVDA1', 'CP1')),
  phase         smallint NOT NULL CHECK (phase IN (1, 2)),
  is_enabled    boolean NOT NULL DEFAULT true
);

CREATE TABLE rule_overrides (
  id          uuid NOT NULL DEFAULT uuidv7() UNIQUE,  -- kunci pengganti untuk trigger audit
  rule_code   text NOT NULL REFERENCES rules (code) ON DELETE CASCADE,
  customer_id uuid NOT NULL REFERENCES customers (id) ON DELETE CASCADE,
  level       finding_level,                              -- NULL = bawaan aturan
  is_enabled  boolean,                                    -- NULL = bawaan aturan
  PRIMARY KEY (rule_code, customer_id)
);

CREATE TABLE check_runs (
  id           uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id   uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  trigger      check_trigger NOT NULL,
  rule_codes   text[] NOT NULL,
  started_at   timestamptz NOT NULL DEFAULT now(),
  finished_at  timestamptz,
  duration_ms  integer,
  opened       integer,
  fixed        integer,
  still_open   integer,
  error        text,
  triggered_by uuid REFERENCES users (id)
);
CREATE INDEX check_runs_pkg_idx ON check_runs (package_id, started_at DESC);

CREATE TABLE findings (
  id            uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id    uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  rule_code     text NOT NULL REFERENCES rules (code),
  level         finding_level NOT NULL,                   -- level efektif (override customer sudah diterapkan)
  object_type   text NOT NULL,                            -- nama tabel, mis. 'failure_chains'
  object_id     uuid,
  field         text,                                     -- nama field API, untuk deep link "Open cell"
  message       text NOT NULL,                            -- pesan (bahasa Inggris) yang sudah dirender
  params        jsonb NOT NULL DEFAULT '{}'::jsonb,
  fingerprint   text NOT NULL,                            -- sha256(rule|object_type|object_id|field|key)
  status        finding_status NOT NULL DEFAULT 'open',
  first_seen_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at  timestamptz NOT NULL DEFAULT now(),
  fixed_at      timestamptz,
  waived_by     uuid REFERENCES users (id),
  waived_at     timestamptz,
  waiver_reason text,
  UNIQUE (package_id, fingerprint),
  CHECK (status <> 'waived'
         OR (level <> 'error' AND waived_by IS NOT NULL AND coalesce(btrim(waiver_reason), '') <> ''))
);
CREATE INDEX findings_open_idx ON findings (package_id, level) WHERE status = 'open';
CREATE INDEX findings_rule_idx ON findings (rule_code) WHERE status = 'open';
CREATE INDEX findings_object_idx ON findings (object_id) WHERE status = 'open';

-- Penghitung untuk dashboard, diperbarui setiap check run selesai dan saat waive/unwaive.
CREATE TABLE package_stats (
  package_id       uuid PRIMARY KEY REFERENCES packages (id) ON DELETE CASCADE,
  errors_open      integer NOT NULL DEFAULT 0,
  warnings_open    integer NOT NULL DEFAULT 0,
  infos_open       integer NOT NULL DEFAULT 0,
  waived           integer NOT NULL DEFAULT 0,
  steps_total      integer NOT NULL DEFAULT 0,
  steps_in_pfmea   integer NOT NULL DEFAULT 0,
  steps_in_cp      integer NOT NULL DEFAULT 0,
  chains_total     integer NOT NULL DEFAULT 0,
  chains_s9_10     integer NOT NULL DEFAULT 0,
  max_rpn          integer,
  actions_open     integer NOT NULL DEFAULT 0,
  actions_overdue  integer NOT NULL DEFAULT 0,
  last_check_at    timestamptz,
  updated_at       timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- Tautan dan sinkronisasi Template General
-- ---------------------------------------------------------------------------
CREATE TABLE package_links (
  model_package_id   uuid PRIMARY KEY REFERENCES packages (id) ON DELETE CASCADE,
  general_package_id uuid NOT NULL REFERENCES packages (id),
  synced_rev         integer NOT NULL,                    -- revisi template yang sedang diikuti paket model
  state              link_state NOT NULL DEFAULT 'in_sync',
  last_sync_run_id   uuid,
  updated_at         timestamptz NOT NULL DEFAULT now(),
  CHECK (model_package_id <> general_package_id)
);
CREATE INDEX package_links_general_idx ON package_links (general_package_id);

CREATE TABLE sync_runs (
  id                 uuid PRIMARY KEY DEFAULT uuidv7(),
  general_package_id uuid NOT NULL REFERENCES packages (id),
  from_rev           integer NOT NULL,
  to_rev             integer NOT NULL CHECK (to_rev > from_rev),
  status             sync_run_status NOT NULL DEFAULT 'queued',
  packages_total     integer NOT NULL DEFAULT 0,
  packages_done      integer NOT NULL DEFAULT 0,
  packages_failed    integer NOT NULL DEFAULT 0,
  summary            jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_by         uuid NOT NULL REFERENCES users (id),
  created_at         timestamptz NOT NULL DEFAULT now(),
  started_at         timestamptz,
  finished_at        timestamptz
);

-- Satu baris per sel berubah / baris baru / baris terhapus / konflik, per paket model.
CREATE TABLE sync_changes (
  id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  sync_run_id uuid NOT NULL REFERENCES sync_runs (id) ON DELETE CASCADE,
  package_id  uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  table_name  text NOT NULL,
  row_id      uuid NOT NULL,
  column_name text,
  old_value   jsonb,
  new_value   jsonb,
  action      sync_action NOT NULL,
  resolved_at timestamptz,                                -- hanya untuk konflik (T03)
  resolved_by uuid REFERENCES users (id),
  resolution  text CHECK (resolution IN ('follow_template', 'keep_local')),
  created_at  timestamptz NOT NULL DEFAULT now(),
  CHECK (action = 'conflict' OR resolved_at IS NULL)
);
CREATE INDEX sync_changes_run_pkg_idx ON sync_changes (sync_run_id, package_id);
CREATE INDEX sync_changes_open_conflicts_idx ON sync_changes (package_id) WHERE action = 'conflict' AND resolved_at IS NULL;

-- ---------------------------------------------------------------------------
-- Ekspor dan audit
-- ---------------------------------------------------------------------------
CREATE TABLE export_files (
  id              uuid PRIMARY KEY DEFAULT uuidv7(),
  package_id      uuid NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
  doc_type        doc_type NOT NULL,
  phase           cp_phase,                               -- hanya CP
  format          text NOT NULL DEFAULT 'xlsx' CHECK (format IN ('xlsx', 'pdf')),
  content_version bigint NOT NULL,                        -- packages.content_version saat file dibuat
  status          export_status NOT NULL DEFAULT 'queued',
  file_path       text,
  file_size       integer,
  error           text,
  created_by      uuid NOT NULL REFERENCES users (id),
  created_at      timestamptz NOT NULL DEFAULT now(),
  finished_at     timestamptz,
  CHECK ((doc_type = 'CP') = (phase IS NOT NULL))
);
CREATE UNIQUE INDEX export_files_cache_idx
  ON export_files (package_id, doc_type, phase, format, content_version) NULLS NOT DISTINCT
  WHERE status IN ('queued', 'running', 'done');

CREATE TABLE audit_log (
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  at         timestamptz NOT NULL DEFAULT now(),
  user_id    uuid,
  package_id uuid,
  table_name text NOT NULL,
  row_id     uuid,
  action     text NOT NULL CHECK (action IN ('insert', 'update', 'delete')),
  old_row    jsonb,                                       -- update: hanya kolom yang berubah
  new_row    jsonb,
  request_id text,
  source     text NOT NULL DEFAULT 'system'               -- api | sync | system | seed
);
CREATE INDEX audit_log_package_at_idx ON audit_log (package_id, at DESC);
CREATE INDEX audit_log_at_brin_idx ON audit_log USING brin (at);

-- ---------------------------------------------------------------------------
-- Trigger baris (updated_at + version) dan trigger statement (audit + content_version)
-- ---------------------------------------------------------------------------
CREATE TRIGGER users_touch           BEFORE UPDATE ON users           FOR EACH ROW EXECUTE FUNCTION trg_touch_row('last_login_at');
CREATE TRIGGER customers_touch       BEFORE UPDATE ON customers       FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER sc_symbols_touch      BEFORE UPDATE ON sc_symbols      FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER parts_touch           BEFORE UPDATE ON parts           FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER control_library_touch BEFORE UPDATE ON control_library FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER banned_terms_touch    BEFORE UPDATE ON banned_terms    FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER packages_touch        BEFORE UPDATE ON packages        FOR EACH ROW EXECUTE FUNCTION trg_touch_row('content_version');
CREATE TRIGGER documents_touch       BEFORE UPDATE ON documents       FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER process_steps_touch   BEFORE UPDATE ON process_steps   FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER step_flows_touch      BEFORE UPDATE ON step_flows      FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER characteristics_touch BEFORE UPDATE ON characteristics FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER failure_modes_touch   BEFORE UPDATE ON failure_modes   FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER failure_effects_touch BEFORE UPDATE ON failure_effects FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER failure_causes_touch  BEFORE UPDATE ON failure_causes  FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER failure_chains_touch  BEFORE UPDATE ON failure_chains  FOR EACH ROW EXECUTE FUNCTION trg_touch_row('s', 'rpn');
CREATE TRIGGER controls_touch        BEFORE UPDATE ON controls        FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER actions_touch         BEFORE UPDATE ON actions         FOR EACH ROW EXECUTE FUNCTION trg_touch_row('new_rpn');
CREATE TRIGGER cp_lines_touch        BEFORE UPDATE ON cp_lines        FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER reaction_plans_touch  BEFORE UPDATE ON reaction_plans  FOR EACH ROW EXECUTE FUNCTION trg_touch_row();
CREATE TRIGGER package_links_touch   BEFORE UPDATE ON package_links   FOR EACH ROW EXECUTE FUNCTION trg_touch_updated_at();
CREATE TRIGGER package_stats_touch   BEFORE UPDATE ON package_stats   FOR EACH ROW EXECUTE FUNCTION trg_touch_updated_at();
CREATE TRIGGER rating_criteria_touch BEFORE UPDATE ON rating_criteria FOR EACH ROW EXECUTE FUNCTION trg_touch_updated_at();
CREATE TRIGGER app_settings_touch    BEFORE UPDATE ON app_settings    FOR EACH ROW EXECUTE FUNCTION trg_touch_updated_at();

-- +goose StatementBegin
DO $$
BEGIN
  -- Tabel konten: audit + kenaikan content_version.
  PERFORM attach_audit('documents', true);
  PERFORM attach_audit('process_steps', true);
  PERFORM attach_audit('step_flows', true);
  PERFORM attach_audit('characteristics', true);
  PERFORM attach_audit('failure_modes', true);
  PERFORM attach_audit('failure_effects', true);
  PERFORM attach_audit('failure_causes', true);
  PERFORM attach_audit('failure_chains', true, ARRAY['s', 'rpn']);
  PERFORM attach_audit('controls', true);
  PERFORM attach_audit('actions', true, ARRAY['new_rpn']);
  PERFORM attach_audit('cp_lines', true);
  PERFORM attach_audit('reaction_plans', true);
  -- Tabel lain yang punya kolom id: hanya audit.
  PERFORM attach_audit('packages', false, ARRAY['content_version']);
  PERFORM attach_audit('users', false, ARRAY['last_login_at']);
  PERFORM attach_audit('customers', false);
  PERFORM attach_audit('sc_symbols', false);
  PERFORM attach_audit('parts', false);
  PERFORM attach_audit('control_library', false);
  PERFORM attach_audit('banned_terms', false);
  PERFORM attach_audit('customer_sc_symbols', false);
  PERFORM attach_audit('rating_criteria', false);
  PERFORM attach_audit('app_settings', false);
  PERFORM attach_audit('rules', false);
  PERFORM attach_audit('rule_overrides', false);
  PERFORM attach_audit('package_members', false);
END $$;
-- +goose StatementEnd

-- Nama kolom database snake_case -> nama field API camelCase. Dipakai oleh query view
-- (kunci override) dan oleh aturan yang nama field-nya berasal dari data (T03).
-- +goose StatementBegin
CREATE FUNCTION snake_to_camel(s text) RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE AS $$
  SELECT string_agg(CASE WHEN i = 1 THEN w ELSE upper(left(w, 1)) || substr(w, 2) END, '' ORDER BY i)
  FROM unnest(string_to_array(s, '_')) WITH ORDINALITY AS t(w, i)
$$;
-- +goose StatementEnd

-- Salinan beku setiap baris konten paket, dipakai untuk package_revisions.snapshot.
-- Kuncinya nama tabel; setiap baris berupa to_jsonb(row) dengan nama kolom database.
-- +goose StatementBegin
CREATE FUNCTION package_snapshot(p_package_id uuid) RETURNS jsonb
LANGUAGE sql STABLE AS $$
  SELECT jsonb_build_object(
    'schema', 1,
    'packageId', p_package_id,
    'takenAt', now(),
    'tables', jsonb_build_object(
      'documents',       (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.doc_type), '[]'::jsonb) FROM documents t WHERE t.package_id = p_package_id),
      'process_steps',   (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.seq), '[]'::jsonb) FROM process_steps t WHERE t.package_id = p_package_id),
      'step_flows',      (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.id), '[]'::jsonb) FROM step_flows t WHERE t.package_id = p_package_id),
      'characteristics', (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.char_no), '[]'::jsonb) FROM characteristics t WHERE t.package_id = p_package_id),
      'failure_modes',   (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.id), '[]'::jsonb) FROM failure_modes t WHERE t.package_id = p_package_id),
      'failure_effects', (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.id), '[]'::jsonb) FROM failure_effects t WHERE t.package_id = p_package_id),
      'failure_causes',  (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.id), '[]'::jsonb) FROM failure_causes t WHERE t.package_id = p_package_id),
      'failure_chains',  (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.id), '[]'::jsonb) FROM failure_chains t WHERE t.package_id = p_package_id),
      'controls',        (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.id), '[]'::jsonb) FROM controls t WHERE t.package_id = p_package_id),
      'actions',         (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.id), '[]'::jsonb) FROM actions t WHERE t.package_id = p_package_id),
      'cp_lines',        (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.id), '[]'::jsonb) FROM cp_lines t WHERE t.package_id = p_package_id),
      'reaction_plans',  (SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.id), '[]'::jsonb) FROM reaction_plans t WHERE t.package_id = p_package_id)
    ));
$$;
-- +goose StatementEnd

-- ---------------------------------------------------------------------------
-- Data referensi
-- ---------------------------------------------------------------------------
INSERT INTO app_settings (key, value) VALUES
  ('doc_number_pattern', '{"PFD": "SH-{customer}-{seq:03}", "PFMEA": "SF-{customer}-{seq:03}", "CP": "SC-{customer}-{seq:03}", "generalSuffix": "GEN"}'),
  ('review_interval_months', '12'),
  ('plant_timezone', '"Asia/Jakarta"'),
  ('nightly_check_time', '"02:00"');

INSERT INTO rules (code, title, description, default_level, applies_to, phase) VALUES
  ('K01', 'PFD step missing in PFMEA', 'A PFD step is not analysed in the PFMEA, unless it is marked "not analysed" with a reason.', 'error', 'all', 1),
  ('K02', 'Characteristic missing in Control Plan', 'A (non-special) characteristic of a PFD step has no Control Plan line.', 'error', 'all', 1),
  ('K03', 'PFMEA/CP row does not match PFD', 'A PFMEA or CP row references a characteristic of another step (orphan).', 'error', 'all', 1),
  ('K04', 'Step data differs between documents (import)', 'Step number, name or order differs between documents in imported data.', 'error', 'all', 2),
  ('K05', 'CP machine not listed in PFD', 'A machine or tool in the CP is not listed on the PFD step.', 'warning', 'all', 1),
  ('K06', 'Part number/change level mismatch', 'Part number or change level in a document header differs from the package part data.', 'error', 'all', 1),
  ('S01', 'Special characteristic not cascaded', 'A special characteristic does not appear in the PFMEA and the Control Plan.', 'error', 'all', 1),
  ('S02', 'Symbol missing in customer table', 'The special characteristic symbol has no entry in the customer symbol conversion table.', 'error', 'all', 1),
  ('S03', 'Special characteristic without strong control', 'A CP line of a special characteristic has no strong control (SPC, 100% automated, error-proofing).', 'warning', 'all', 1),
  ('S04', 'Classification does not match S rule', 'Critical classification, but the highest S is below the customer requirement.', 'warning', 'all', 1),
  ('R01', 'Detection control missing in CP', 'A PFMEA detection control does not appear in the CP on the same step and characteristic.', 'error', 'all', 1),
  ('R02', 'Prevention control missing in CP', 'A PFMEA prevention control is not in the CP and not marked as a system control with a reason.', 'warning', 'all', 1),
  ('R03', 'CP line without PFMEA link', 'A CP line has no link to a failure mode or control in the PFMEA.', 'warning', 'all', 1),
  ('R04', 'D does not match detection method', 'D is better than the detection method in the control library can justify.', 'warning', 'all', 1),
  ('R05', 'Low O without basis', 'O <= 3 without a prevention control or history data.', 'warning', 'all', 1),
  ('R06', 'Error-proofing without verification', 'Error-proofing in the CP without a verification frequency.', 'warning', 'all', 1),
  ('F01', 'Required PFMEA field empty', 'Failure mode, effect, cause, S, O or D is empty.', 'error', 'all', 1),
  ('F02', 'More than one failure mode per cell', 'One cell contains more than one failure mode.', 'warning', 'all', 1),
  ('F03', 'Same effect, different S', 'The same failure effect is rated with different S values.', 'warning', 'all', 1),
  ('F04', '4M work elements incomplete', 'Step without 4M work elements, or a failure cause not linked to a work element.', 'error', 'AIAGVDA1', 2),
  ('F05', 'AP H without action', 'AP H without action and without justification.', 'error', 'AIAGVDA1', 2),
  ('F06', 'AP M without action', 'AP M without action and without justification.', 'warning', 'AIAGVDA1', 2),
  ('F07', 'S 9-10 without action', 'S 9-10 without a recommended action or justification.', 'warning', 'AIAG4', 1),
  ('F08', 'RPN above customer threshold', 'RPN above the customer CSR threshold without action or justification (only when the customer sets a threshold).', 'warning', 'AIAG4', 1),
  ('F09', 'Completed action without re-rating', 'Action completed but the new S/O/D is not filled in.', 'error', 'all', 1),
  ('F10', 'S lowered without design change', 'S lowered after an action without a design change note.', 'warning', 'all', 1),
  ('F11', 'Action overdue', 'Action past its target date.', 'warning', 'all', 1),
  ('F12', 'Term from the banned list', 'Text contains a term from the banned-terms list.', 'info', 'all', 1),
  ('C01', 'Required CP field empty', 'Specification, evaluation method, sample size, frequency, control method or reaction plan is empty.', 'error', 'all', 1),
  ('C02', 'Reaction plan without owner', 'Reaction plan without an owner/PIC.', 'error', 'CP1', 2),
  ('C03', 'Reaction plan incomplete', 'Reaction plan does not cover isolation of suspect product, process stop and process recovery.', 'error', 'CP1', 2),
  ('C04', 'Frequency not volume-based', 'Frequency is not 100% and not volume-based.', 'warning', 'CP1', 2),
  ('C05', 'Safe Launch without exit criteria', 'Safe Launch phase without exit criteria.', 'error', 'CP1', 2),
  ('C06', 'SC detection only visual', 'Detection of a special characteristic is only 100% visual.', 'warning', 'CP1', 2),
  ('W01', 'Rework step without PFMEA', 'A rework/repair step in the PFD has no PFMEA rows.', 'error', 'all', 1),
  ('W02', 'NG branch without disposition', 'An inspection step has no NG branch with a disposition (rework, scrap, hold).', 'error', 'all', 1),
  ('W03', 'PFD/WI changed after release', 'PFD or WI changed after release; PFMEA/CP not reviewed yet.', 'warning', 'all', 2),
  ('W04', 'Package not reviewed for too long', 'Package not reviewed for longer than the month limit in settings (default 12).', 'info', 'all', 1),
  ('T01', 'Mandatory general process missing', 'A mandatory general process is missing in the model package.', 'error', 'all', 1),
  ('T02', 'Template General not up to date', 'The package still uses an older Template General revision (sync pending or failed).', 'warning', 'all', 1),
  ('T03', 'Override conflicts with template', 'A local override on a general row conflicts with the latest template change.', 'warning', 'all', 1),
  ('T04', 'General process detached without reason', 'A general row was detached from the template without a reason.', 'warning', 'all', 1);

-- +goose Down
DROP TABLE IF EXISTS audit_log, export_files, sync_changes, sync_runs, package_links, package_stats,
  findings, check_runs, rule_overrides, rules, reaction_plans, cp_lines, actions, controls, failure_chains,
  failure_causes, failure_effects, failure_modes, characteristics, step_flows, process_steps,
  package_revisions, documents, package_members, packages, app_settings, banned_terms, rating_criteria,
  control_library, parts, customer_sc_symbols, sc_symbols, customers, sessions, users CASCADE;
DROP FUNCTION IF EXISTS package_snapshot(uuid), snake_to_camel(text), attach_audit(regclass, boolean, text[]), trg_audit_insert(), trg_audit_update(),
  trg_audit_delete(), trg_touch_row(), trg_touch_updated_at(), trg_chain_s_on_insert(), trg_effects_refresh_chain_s();
DROP TYPE IF EXISTS user_role, auth_source, package_kind, package_status, doc_type, methodology, step_symbol,
  step_kind, general_mode, row_origin, sync_status, flow_kind, characteristic_kind, effect_level, control_kind,
  action_kind, action_status, cp_phase, freq_basis, control_method_category, rating_dimension, finding_level,
  finding_status, check_trigger, link_state, sync_run_status, sync_action, export_status;

-- 007_audit_files.sql — Audit log (partitioned by month) + file uploads.

-- ── audit_log (range-partitioned by occurred_at) ───────────────────────────
CREATE TABLE IF NOT EXISTS audit_log (
  id               UUID NOT NULL DEFAULT gen_random_uuid(),
  occurred_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  correlation_id   TEXT,
  tenant_id        UUID,
  organization_id  UUID,
  user_id          UUID,
  user_email       CITEXT,
  method           TEXT,
  path             TEXT,
  route            TEXT,
  status_code      INT,
  latency_ms       BIGINT,
  ip               INET,
  user_agent       TEXT,
  action           TEXT,
  target_type      TEXT,
  target_id        UUID,
  error_code       TEXT,
  request_headers  JSONB,
  request_body     JSONB,
  response_headers JSONB,
  response_body    JSONB,
  metadata         JSONB,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by       UUID,
  PRIMARY KEY (id, occurred_at)
) PARTITION BY RANGE (occurred_at);

CREATE INDEX IF NOT EXISTS idx_audit_log_tenant_time ON audit_log (tenant_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_org_time ON audit_log (organization_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_user_time ON audit_log (user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_correlation ON audit_log (correlation_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log (action);
CREATE INDEX IF NOT EXISTS idx_audit_log_status ON audit_log (status_code);

-- Default catch-all partition. The worker promotes data to monthly partitions
-- on a cron schedule and creates next month's partition each month.
CREATE TABLE IF NOT EXISTS audit_log_default PARTITION OF audit_log DEFAULT;

-- Initial monthly partitions (current + next 2). Worker keeps extending.
CREATE TABLE IF NOT EXISTS audit_log_2026_06 PARTITION OF audit_log
  FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_07 PARTITION OF audit_log
  FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_08 PARTITION OF audit_log
  FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_09 PARTITION OF audit_log
  FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');


-- ── file_uploads ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS file_uploads (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  organization_id  UUID REFERENCES organizations(id) ON DELETE SET NULL,
  uploaded_by      UUID REFERENCES users(id) ON DELETE SET NULL,
  bucket           TEXT NOT NULL,
  storage_key      TEXT NOT NULL UNIQUE,
  original_name    TEXT,
  mime             TEXT,
  size_bytes       BIGINT NOT NULL DEFAULT 0,
  checksum_sha256  TEXT,
  width            INT,
  height           INT,
  duration_ms      BIGINT,
  status           TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'uploaded', 'scanned', 'failed', 'archived')),
  scanned_at       TIMESTAMPTZ,
  scan_result      TEXT,
  visibility       TEXT NOT NULL DEFAULT 'private'
                   CHECK (visibility IN ('public', 'private', 'tenant', 'organization')),
  purpose          TEXT,
  related_type     TEXT,
  related_id       UUID,
  expires_at       TIMESTAMPTZ,
  public_url       TEXT,
  metadata         JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at       TIMESTAMPTZ,
  created_by       UUID,
  updated_by       UUID,
  deleted_by       UUID
);

CREATE INDEX IF NOT EXISTS idx_file_uploads_tenant ON file_uploads (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_file_uploads_org ON file_uploads (organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_file_uploads_uploader ON file_uploads (uploaded_by) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_file_uploads_status ON file_uploads (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_file_uploads_related ON file_uploads (related_type, related_id) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_file_uploads_updated_at ON file_uploads;
CREATE TRIGGER trg_file_uploads_updated_at BEFORE UPDATE ON file_uploads FOR EACH ROW EXECUTE FUNCTION set_updated_at();

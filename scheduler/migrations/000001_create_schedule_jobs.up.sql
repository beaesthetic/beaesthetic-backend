CREATE TABLE IF NOT EXISTS schedule_jobs (
  id UUID PRIMARY KEY,
  scheduler_name TEXT NOT NULL,
  route TEXT NOT NULL,
  payload JSONB NOT NULL,
  schedule_at TIMESTAMPTZ NOT NULL,
  leased_until TIMESTAMPTZ,
  lease_owner TEXT,
  attempts INT NOT NULL DEFAULT 0,
  last_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS schedule_jobs_due_idx
ON schedule_jobs (scheduler_name, schedule_at)
WHERE leased_until IS NULL;

CREATE INDEX IF NOT EXISTS schedule_jobs_lease_idx
ON schedule_jobs (scheduler_name, leased_until);

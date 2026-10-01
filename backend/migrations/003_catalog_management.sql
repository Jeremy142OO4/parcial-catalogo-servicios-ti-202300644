ALTER TABLE services_level2
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS services_level2_search_idx
    ON services_level2 (service_level1_id, is_active);

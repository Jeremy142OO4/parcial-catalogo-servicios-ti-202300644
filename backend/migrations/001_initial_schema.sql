CREATE TABLE IF NOT EXISTS companies (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS areas (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (company_id, code)
);

CREATE TABLE IF NOT EXISTS departments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    area_id BIGINT NOT NULL REFERENCES areas(id),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (area_id, code)
);

CREATE TABLE IF NOT EXISTS sections (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(id),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (department_id, code)
);

CREATE TABLE IF NOT EXISTS positions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    section_id BIGINT NOT NULL REFERENCES sections(id),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (section_id, code)
);

CREATE TABLE IF NOT EXISTS app_users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    position_id BIGINT NOT NULL REFERENCES positions(id),
    full_name TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'consulta')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES app_users(id),
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS service_classes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS criticalities (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS service_types (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS services_level1 (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    canonical_name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    review_required BOOLEAN NOT NULL DEFAULT FALSE,
    source_sheet TEXT,
    source_row INTEGER,
    source_range TEXT
);

CREATE TABLE IF NOT EXISTS services_level1_source_names (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    service_level1_id BIGINT NOT NULL REFERENCES services_level1(id),
    original_name TEXT NOT NULL,
    source_sheet TEXT NOT NULL,
    source_row INTEGER NOT NULL,
    source_range TEXT NOT NULL,
    is_canonical BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS services_level2 (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    service_level1_id BIGINT NOT NULL REFERENCES services_level1(id),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    active_value TEXT,
    service_class_id BIGINT REFERENCES service_classes(id),
    criticality_id BIGINT REFERENCES criticalities(id),
    service_type_id BIGINT REFERENCES service_types(id),
    description TEXT,
    metric TEXT,
    minimum NUMERIC,
    maximum NUMERIC,
    review_required BOOLEAN NOT NULL DEFAULT FALSE,
    source_sheet TEXT,
    source_row INTEGER,
    source_range TEXT,
    CONSTRAINT valid_threshold_range CHECK (minimum IS NULL OR maximum IS NULL OR minimum <= maximum)
);

CREATE TABLE IF NOT EXISTS service_assignments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    service_level2_id BIGINT NOT NULL UNIQUE REFERENCES services_level2(id),
    section_id BIGINT NOT NULL REFERENCES sections(id),
    responsible_user_id BIGINT REFERENCES app_users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS import_runs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    source_file TEXT NOT NULL,
    source_sha256 TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    status TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed')),
    created_count INTEGER NOT NULL DEFAULT 0,
    updated_count INTEGER NOT NULL DEFAULT 0,
    omitted_count INTEGER NOT NULL DEFAULT 0,
    observed_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS import_observations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    import_run_id BIGINT NOT NULL REFERENCES import_runs(id),
    code TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('info', 'warning', 'error')),
    message TEXT NOT NULL,
    sheet_name TEXT,
    row_start INTEGER,
    row_end INTEGER,
    source_range TEXT,
    details JSONB NOT NULL DEFAULT '{}'::JSONB
);

INSERT INTO service_classes (name) VALUES ('A DEMANDA'), ('RECURRENTE') ON CONFLICT (name) DO NOTHING;
INSERT INTO criticalities (name) VALUES ('Very Low'), ('Low'), ('Normal'), ('High'), ('Very High') ON CONFLICT (name) DO NOTHING;
INSERT INTO service_types (name) VALUES
    ('Back End'), ('Demostration'), ('End User Service'), ('Front End'),
    ('IT Management'), ('IT Operational'), ('Other'), ('Project'),
    ('Reporting'), ('Training'), ('Underpinning Contract')
ON CONFLICT (name) DO NOTHING;

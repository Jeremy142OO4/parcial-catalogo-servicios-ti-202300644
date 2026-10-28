CREATE TABLE ticket_attachments (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ticket_id bigint NOT NULL REFERENCES tickets(id),
    uploader_id bigint NOT NULL REFERENCES app_users(id),
    original_name text NOT NULL,
    storage_key text NOT NULL UNIQUE CHECK (storage_key ~ '^[a-f0-9]{64}$'),
    mime_type text NOT NULL CHECK (mime_type IN ('application/pdf','image/png','image/jpeg','text/plain')),
    byte_size bigint NOT NULL CHECK (byte_size BETWEEN 1 AND 10485760),
    sha256 text NOT NULL CHECK (sha256 ~ '^[a-f0-9]{64}$'),
    visibility text NOT NULL CHECK (visibility IN ('public','internal')),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ticket_attachments_ticket_idx ON ticket_attachments(ticket_id, id);

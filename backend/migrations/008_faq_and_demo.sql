CREATE TABLE faq_articles (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 slug text NOT NULL UNIQUE CHECK(slug ~ '^[a-z0-9][a-z0-9-]{0,99}$'),
 title text NOT NULL CHECK(length(btrim(title)) BETWEEN 1 AND 200),
 body text NOT NULL CHECK(length(btrim(body)) BETWEEN 1 AND 20000),
 process_code text REFERENCES processes(code),
 service_level2_id bigint REFERENCES services_level2(id),
 is_published boolean NOT NULL DEFAULT false,
 is_active boolean NOT NULL DEFAULT true,
 version bigint NOT NULL DEFAULT 1,
 created_by bigint NOT NULL REFERENCES app_users(id),
 updated_by bigint NOT NULL REFERENCES app_users(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE faq_history (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 article_id bigint NOT NULL REFERENCES faq_articles(id),
 actor_id bigint NOT NULL REFERENCES app_users(id),
 occurred_at timestamptz NOT NULL DEFAULT now(),
 before_value jsonb, after_value jsonb NOT NULL
);
CREATE TRIGGER faq_audit_immutable BEFORE UPDATE OR DELETE ON faq_history FOR EACH ROW EXECUTE FUNCTION protect_audit();
CREATE TABLE demo_tickets (
 dataset_key text NOT NULL, case_key text NOT NULL,
 ticket_id bigint NOT NULL UNIQUE REFERENCES tickets(id),
 prepared_at timestamptz,
 PRIMARY KEY(dataset_key,case_key)
);

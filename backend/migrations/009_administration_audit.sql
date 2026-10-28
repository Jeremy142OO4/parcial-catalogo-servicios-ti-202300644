-- Technical identity survives edits to the visible classification code.
ALTER TABLE ci_classes ADD COLUMN technical_profile TEXT UNIQUE
 CHECK (technical_profile IN ('servidor','usuario','red','impresion','almacenamiento','energia'));
UPDATE ci_classes SET technical_profile=code
 WHERE code IN ('servidor','usuario','red','impresion','almacenamiento','energia');
CREATE FUNCTION protect_technical_profile() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.technical_profile IS DISTINCT FROM OLD.technical_profile THEN
  RAISE EXCEPTION 'el perfil técnico es inmutable';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER immutable_technical_profile BEFORE UPDATE ON ci_classes
 FOR EACH ROW EXECUTE FUNCTION protect_technical_profile();

-- No foreign key to the affected entity: evidence survives physical deletion.
CREATE TABLE administration_history (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 actor_id BIGINT REFERENCES app_users(id),
 origin TEXT NOT NULL,
 occurred_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 entity_table TEXT NOT NULL,
 action TEXT NOT NULL,
 before_value JSONB,
 after_value JSONB
);
CREATE INDEX administration_history_entity ON administration_history(entity_table,id);
CREATE FUNCTION capture_administration_change() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
 previous JSONB;
 following JSONB;
 actor BIGINT;
 ci BIGINT;
BEGIN
 IF TG_OP <> 'INSERT' THEN previous=to_jsonb(OLD); END IF;
 IF TG_OP <> 'DELETE' THEN following=to_jsonb(NEW); END IF;
 IF previous IS NOT DISTINCT FROM following THEN RETURN NULL; END IF;
 actor=nullif(current_setting('app.actor_id',true),'')::bigint;
 INSERT INTO administration_history(actor_id,origin,entity_table,action,before_value,after_value)
 VALUES(actor,CASE WHEN actor IS NULL THEN 'system' ELSE 'application' END,
        TG_TABLE_NAME,TG_OP,previous,following);
 IF TG_TABLE_NAME IN ('ci_relations','ci_services') THEN
  FOR ci IN SELECT DISTINCT value::bigint FROM (
   VALUES(previous->>'source_ci_id'),(previous->>'target_ci_id'),
         (following->>'source_ci_id'),(following->>'target_ci_id'),
         (previous->>'ci_id'),(following->>'ci_id')
  ) refs(value) WHERE value IS NOT NULL LOOP
   INSERT INTO ci_changes(ci_id,actor_id,description,before_value,after_value)
   VALUES(ci,actor,TG_TABLE_NAME||':'||TG_OP,previous,following);
  END LOOP;
 END IF;
 RETURN NULL;
END $$;
DO $$
DECLARE entity TEXT;
BEGIN
 FOREACH entity IN ARRAY ARRAY['ci_classes','ci_relations','ci_services','locations','queues','queue_grants','user_roles'] LOOP
  EXECUTE format('CREATE TRIGGER administration_change AFTER INSERT OR UPDATE OR DELETE ON %I FOR EACH ROW EXECUTE FUNCTION capture_administration_change()',entity);
 END LOOP;
END $$;
CREATE FUNCTION protect_administration_history() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'historial administrativo inmutable';
END $$;
CREATE TRIGGER administration_history_immutable BEFORE UPDATE OR DELETE OR TRUNCATE
 ON administration_history FOR EACH STATEMENT EXECUTE FUNCTION protect_administration_history();

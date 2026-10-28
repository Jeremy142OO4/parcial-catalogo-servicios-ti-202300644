-- User snapshots must never contain password hashes or session tokens.
CREATE FUNCTION capture_user_administration_change() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE previous JSONB; following JSONB; actor BIGINT;
BEGIN
 IF TG_OP <> 'INSERT' THEN previous=to_jsonb(OLD)-'password_hash'; END IF;
 IF TG_OP <> 'DELETE' THEN following=to_jsonb(NEW)-'password_hash'; END IF;
 IF previous IS NOT DISTINCT FROM following THEN RETURN NULL; END IF;
 actor=nullif(current_setting('app.actor_id',true),'')::bigint;
 INSERT INTO administration_history(actor_id,origin,entity_table,action,before_value,after_value)
 VALUES(actor,CASE WHEN actor IS NULL THEN 'system' ELSE 'application' END,
        TG_TABLE_NAME,TG_OP,previous,following);
 RETURN NULL;
END $$;
CREATE TRIGGER user_administration_change AFTER INSERT OR UPDATE OR DELETE ON app_users
 FOR EACH ROW EXECUTE FUNCTION capture_user_administration_change();

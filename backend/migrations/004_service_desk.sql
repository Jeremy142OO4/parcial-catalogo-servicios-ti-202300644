-- Incremental extension. Existing catalog data and migration history are retained.
CREATE TABLE roles (code TEXT PRIMARY KEY, name TEXT NOT NULL);
INSERT INTO roles VALUES ('admin','Administrador'),('consulta','Consulta'),('solicitante','Solicitante'),('agente','Agente'),('aprobador','Aprobador');
ALTER TABLE app_users DROP CONSTRAINT app_users_role_check;
ALTER TABLE app_users ADD CONSTRAINT app_users_role_check CHECK (role IN ('admin','consulta','solicitante','agente','aprobador'));
CREATE TABLE user_roles (user_id BIGINT REFERENCES app_users(id), role_code TEXT REFERENCES roles(code), PRIMARY KEY(user_id,role_code));
INSERT INTO user_roles SELECT id,role FROM app_users;
CREATE FUNCTION sync_primary_role() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='INSERT' OR NEW.role IS DISTINCT FROM OLD.role THEN
  DELETE FROM user_roles WHERE user_id=NEW.id;
  INSERT INTO user_roles VALUES(NEW.id,NEW.role);
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER user_primary_role AFTER INSERT OR UPDATE OF role ON app_users FOR EACH ROW EXECUTE FUNCTION sync_primary_role();

CREATE TABLE processes (code TEXT PRIMARY KEY, name TEXT NOT NULL, ticket_type TEXT NOT NULL);
INSERT INTO processes VALUES ('nuevos','Nuevos servicios','requisicion'),('cambios','Cambios','cambio'),('incidentes','Incidentes','incidente'),('eventos','Eventos','evento'),('problemas','Problemas','problema'),('accesos','Accesos','acceso'),('servicios','Servicios existentes','requisicion');
CREATE TABLE support_levels (level INTEGER PRIMARY KEY CHECK(level BETWEEN 0 AND 4), name TEXT NOT NULL);
INSERT INTO support_levels VALUES(0,'Autoservicio'),(1,'Help Desk'),(2,'Soporte Técnico'),(3,'Ingenieros'),(4,'Externo');
CREATE TABLE queues (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, code TEXT NOT NULL UNIQUE CHECK(btrim(code)<>''), name TEXT NOT NULL CHECK(btrim(name)<>''), description TEXT NOT NULL DEFAULT '',
 process_code TEXT NOT NULL REFERENCES processes(code), support_level INTEGER NOT NULL DEFAULT 1 REFERENCES support_levels(level),
 ordering_rule TEXT NOT NULL DEFAULT 'priority_created_id' CHECK(ordering_rule='priority_created_id'),
 first_response_minutes INTEGER CHECK(first_response_minutes>0), resolution_minutes INTEGER CHECK(resolution_minutes>0),
 is_active BOOLEAN NOT NULL DEFAULT TRUE, UNIQUE(id,process_code)
);
INSERT INTO queues(code,name,process_code) SELECT upper(code),name,code FROM processes;
CREATE TABLE queue_grants (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, queue_id BIGINT NOT NULL REFERENCES queues(id), user_id BIGINT REFERENCES app_users(id), role_code TEXT REFERENCES roles(code),
 can_read BOOLEAN NOT NULL DEFAULT FALSE, can_handle BOOLEAN NOT NULL DEFAULT FALSE, can_transfer BOOLEAN NOT NULL DEFAULT FALSE, can_approve BOOLEAN NOT NULL DEFAULT FALSE, can_admin BOOLEAN NOT NULL DEFAULT FALSE,
 CHECK ((user_id IS NULL) <> (role_code IS NULL))
);
CREATE UNIQUE INDEX queue_grant_user ON queue_grants(queue_id,user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX queue_grant_role ON queue_grants(queue_id,role_code) WHERE role_code IS NOT NULL;
CREATE TABLE workflow_states (process_code TEXT REFERENCES processes(code), code TEXT, PRIMARY KEY(process_code,code));
INSERT INTO workflow_states
SELECT p.code,s FROM processes p CROSS JOIN LATERAL unnest(CASE p.code
 WHEN 'nuevos' THEN ARRAY['solicitado','analisis','pendiente_aprobacion','aprobado','implementacion','validacion','resuelto','cerrado','rechazado']
 WHEN 'cambios' THEN ARRAY['solicitado','evaluacion','pendiente_aprobacion','aprobado','planificado','ejecucion','validacion','resuelto','cerrado','rechazado']
 WHEN 'incidentes' THEN ARRAY['registrado','clasificado','diagnostico','atencion','escalado','recuperado','resuelto','cerrado']
 WHEN 'eventos' THEN ARRAY['registrado','clasificado','evaluado','respuesta','resuelto','cerrado']
 WHEN 'problemas' THEN ARRAY['registrado','investigacion','causa_identificada','solucion_propuesta','solucion_aplicada','resuelto','cerrado']
 WHEN 'accesos' THEN ARRAY['solicitado','validacion','pendiente_aprobacion','aprobado','provision_o_revocacion','verificacion','resuelto','cerrado','rechazado']
 ELSE ARRAY['solicitado','validacion','asignado','cumplimiento','confirmacion','resuelto','cerrado'] END) s;

CREATE TABLE ci_classes (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, code TEXT NOT NULL UNIQUE CHECK(btrim(code)<>''), name TEXT NOT NULL CHECK(btrim(name)<>''), parent_id BIGINT REFERENCES ci_classes(id),
 is_device_class BOOLEAN NOT NULL DEFAULT FALSE, is_active BOOLEAN NOT NULL DEFAULT TRUE, CHECK(parent_id<>id)
);
CREATE FUNCTION validate_ci_class() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 PERFORM pg_advisory_xact_lock(202300645);
 IF NEW.parent_id IS NOT NULL THEN
  IF NOT EXISTS(SELECT 1 FROM ci_classes WHERE id=NEW.parent_id AND is_active) THEN RAISE EXCEPTION 'padre inexistente o inactivo'; END IF;
  IF EXISTS(WITH RECURSIVE parents AS (SELECT id,parent_id FROM ci_classes WHERE id=NEW.parent_id UNION SELECT c.id,c.parent_id FROM ci_classes c JOIN parents p ON c.id=p.parent_id) SELECT 1 FROM parents WHERE id=NEW.id) THEN RAISE EXCEPTION 'ciclo de clasificación'; END IF;
 END IF;
 IF NOT NEW.is_active AND EXISTS(SELECT 1 FROM ci_classes WHERE parent_id=NEW.id AND is_active) THEN RAISE EXCEPTION 'clasificación con hijos activos'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER ci_class_integrity BEFORE INSERT OR UPDATE ON ci_classes FOR EACH ROW EXECUTE FUNCTION validate_ci_class();
INSERT INTO ci_classes(code,name) VALUES ('hardware','Hardware'),('software','Software');
INSERT INTO ci_classes(code,name,parent_id,is_device_class) SELECT x.code,x.name,c.id,TRUE FROM ci_classes c CROSS JOIN (VALUES ('servidor','Servidor'),('usuario','Equipo de usuario'),('red','Red'),('impresion','Impresión'),('almacenamiento','Almacenamiento'),('energia','Energía')) x(code,name) WHERE c.code='hardware';
CREATE TABLE locations(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,code TEXT NOT NULL UNIQUE,name TEXT NOT NULL,is_active BOOLEAN NOT NULL DEFAULT TRUE);
CREATE TABLE configuration_items (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, code TEXT NOT NULL UNIQUE CHECK(btrim(code)<>''), name TEXT NOT NULL CHECK(btrim(name)<>''), class_id BIGINT NOT NULL REFERENCES ci_classes(id),
 description TEXT,serial_number TEXT,version TEXT,acquired_on DATE,status TEXT NOT NULL DEFAULT 'activo' CHECK(status IN ('activo','mantenimiento','retirado','desconocido')),
 location_id BIGINT NOT NULL REFERENCES locations(id),owner_user_id BIGINT NOT NULL REFERENCES app_users(id),security_level TEXT,compliance TEXT,configuration_status TEXT,license_number TEXT,license_expires_on DATE,
 technical_attributes JSONB NOT NULL DEFAULT '{}',is_device BOOLEAN NOT NULL DEFAULT FALSE,is_active BOOLEAN NOT NULL DEFAULT TRUE,review_required BOOLEAN NOT NULL DEFAULT FALSE,
 dataset_key TEXT,seed BIGINT,created_at TIMESTAMPTZ NOT NULL DEFAULT now(),updated_at TIMESTAMPTZ NOT NULL DEFAULT now(), CHECK(jsonb_typeof(technical_attributes)='object')
);
CREATE INDEX ci_filter ON configuration_items(class_id,is_active,id);
CREATE INDEX ci_dataset ON configuration_items(dataset_key,seed);
CREATE TABLE ci_changes(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,ci_id BIGINT NOT NULL REFERENCES configuration_items(id),actor_id BIGINT REFERENCES app_users(id),occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),description TEXT NOT NULL,before_value JSONB,after_value JSONB);
CREATE TABLE ci_relation_types(code TEXT PRIMARY KEY,allows_cycles BOOLEAN NOT NULL,propagates_impact BOOLEAN NOT NULL);
INSERT INTO ci_relation_types VALUES('depende_de',FALSE,TRUE),('alojado_en',FALSE,TRUE),('conectado_a',TRUE,FALSE);
CREATE TABLE ci_relations(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,source_ci_id BIGINT NOT NULL REFERENCES configuration_items(id),target_ci_id BIGINT NOT NULL REFERENCES configuration_items(id),kind TEXT NOT NULL REFERENCES ci_relation_types(code),CHECK(source_ci_id<>target_ci_id),UNIQUE(source_ci_id,target_ci_id,kind));
CREATE INDEX ci_relation_target ON ci_relations(target_ci_id);
CREATE TABLE ci_services(ci_id BIGINT REFERENCES configuration_items(id),service_level2_id BIGINT REFERENCES services_level2(id),PRIMARY KEY(ci_id,service_level2_id));
CREATE TABLE source_records(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,source_file TEXT NOT NULL,sha256 TEXT NOT NULL,sheet TEXT NOT NULL,source_row INTEGER NOT NULL,raw_values JSONB NOT NULL,observations JSONB NOT NULL DEFAULT '[]',UNIQUE(sha256,sheet,source_row));
CREATE TABLE generation_runs(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,dataset_key TEXT NOT NULL,seed BIGINT NOT NULL,generator_version TEXT NOT NULL,requested_count INTEGER NOT NULL,actual_count INTEGER NOT NULL,summary JSONB NOT NULL,finished_at TIMESTAMPTZ NOT NULL DEFAULT now());

CREATE TABLE tickets (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, number TEXT UNIQUE NOT NULL, title TEXT NOT NULL CHECK(btrim(title)<>''),description TEXT NOT NULL CHECK(btrim(description)<>''),requester_id BIGINT NOT NULL REFERENCES app_users(id),creator_id BIGINT NOT NULL REFERENCES app_users(id),
 service_level2_id BIGINT NOT NULL REFERENCES services_level2(id),process_code TEXT NOT NULL REFERENCES processes(code),queue_id BIGINT NOT NULL,assignee_id BIGINT REFERENCES app_users(id),state TEXT NOT NULL,
 priority INTEGER NOT NULL CHECK(priority BETWEEN 1 AND 3),impact TEXT NOT NULL CHECK(impact IN ('alto','medio','bajo')),urgency TEXT NOT NULL CHECK(urgency IN ('alto','medio','bajo')),channel TEXT NOT NULL CHECK(btrim(channel)<>''),
 details JSONB NOT NULL DEFAULT '{}',resolution TEXT,no_ci_reason TEXT,version BIGINT NOT NULL DEFAULT 1,cycle INTEGER NOT NULL DEFAULT 1,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),first_attended_at TIMESTAMPTZ,resolved_at TIMESTAMPTZ,closed_at TIMESTAMPTZ,
 FOREIGN KEY(queue_id,process_code) REFERENCES queues(id,process_code), FOREIGN KEY(process_code,state) REFERENCES workflow_states(process_code,code), CHECK(jsonb_typeof(details)='object'),CHECK(state<>'cerrado' OR nullif(btrim(resolution),'') IS NOT NULL)
);
CREATE SEQUENCE ticket_number_seq;
CREATE INDEX ticket_queue ON tickets(queue_id,state,priority,created_at,id);
CREATE INDEX ticket_requester ON tickets(requester_id,created_at);
CREATE TABLE ticket_ci(ticket_id BIGINT REFERENCES tickets(id),ci_id BIGINT REFERENCES configuration_items(id),PRIMARY KEY(ticket_id,ci_id));
CREATE TABLE ticket_history(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,ticket_id BIGINT NOT NULL REFERENCES tickets(id),actor_id BIGINT NOT NULL REFERENCES app_users(id),action TEXT NOT NULL,occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),before_value JSONB,after_value JSONB,reason TEXT NOT NULL DEFAULT '');
CREATE TABLE ticket_comments(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,ticket_id BIGINT NOT NULL REFERENCES tickets(id),author_id BIGINT NOT NULL REFERENCES app_users(id),body TEXT NOT NULL CHECK(btrim(body)<>''),visibility TEXT NOT NULL CHECK(visibility IN ('public','internal')),created_at TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE TABLE ticket_approvals(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,ticket_id BIGINT NOT NULL REFERENCES tickets(id),cycle INTEGER NOT NULL,approver_id BIGINT NOT NULL REFERENCES app_users(id),decision TEXT NOT NULL CHECK(decision IN ('approved','rejected')),reason TEXT NOT NULL,decided_at TIMESTAMPTZ NOT NULL DEFAULT now(),UNIQUE(ticket_id,cycle));
CREATE TABLE ticket_links(source_id BIGINT REFERENCES tickets(id),target_id BIGINT REFERENCES tickets(id),kind TEXT NOT NULL CHECK(kind IN ('incidente_problema','problema_cambio','relacionado')),PRIMARY KEY(source_id,target_id,kind),CHECK(source_id<>target_id));
CREATE TABLE event_incidents(event_ticket_id BIGINT PRIMARY KEY REFERENCES tickets(id),incident_ticket_id BIGINT NOT NULL REFERENCES tickets(id));
CREATE TABLE notifications(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,ticket_id BIGINT REFERENCES tickets(id),recipient_id BIGINT NOT NULL REFERENCES app_users(id),message TEXT NOT NULL,dedup_key TEXT NOT NULL UNIQUE,created_at TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE TABLE ticket_cycles(ticket_id BIGINT REFERENCES tickets(id),cycle INTEGER,started_at TIMESTAMPTZ NOT NULL DEFAULT now(),resolved_at TIMESTAMPTZ,closed_at TIMESTAMPTZ,resolution TEXT,PRIMARY KEY(ticket_id,cycle));
CREATE TABLE ticket_pauses(id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,ticket_id BIGINT NOT NULL REFERENCES tickets(id),started_at TIMESTAMPTZ NOT NULL DEFAULT now(),ended_at TIMESTAMPTZ,reason TEXT NOT NULL,CHECK(ended_at>=started_at));
CREATE UNIQUE INDEX one_ticket_pause ON ticket_pauses(ticket_id) WHERE ended_at IS NULL;

CREATE FUNCTION protect_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'historial inmutable'; END $$;
CREATE TRIGGER ticket_audit_immutable BEFORE UPDATE OR DELETE ON ticket_history FOR EACH ROW EXECUTE FUNCTION protect_audit();
CREATE TRIGGER ci_audit_immutable BEFORE UPDATE OR DELETE ON ci_changes FOR EACH ROW EXECUTE FUNCTION protect_audit();
CREATE FUNCTION protect_desk_dependencies() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_TABLE_NAME='app_users' AND (NOT NEW.is_active OR NEW.role IS DISTINCT FROM OLD.role) THEN
  IF EXISTS(SELECT 1 FROM tickets WHERE assignee_id=OLD.id AND state NOT IN ('cerrado','rechazado')) THEN RAISE EXCEPTION 'usuario con tickets pendientes, reasignar primero'; END IF;
 END IF;
 IF TG_TABLE_NAME='queues' AND NOT NEW.is_active THEN
  IF EXISTS(SELECT 1 FROM tickets WHERE queue_id=OLD.id AND state NOT IN ('cerrado','rechazado')) THEN RAISE EXCEPTION 'cola con tickets pendientes, transferir primero'; END IF;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER user_desk_dependencies BEFORE UPDATE ON app_users FOR EACH ROW EXECUTE FUNCTION protect_desk_dependencies();
CREATE TRIGGER queue_desk_dependencies BEFORE UPDATE ON queues FOR EACH ROW EXECUTE FUNCTION protect_desk_dependencies();

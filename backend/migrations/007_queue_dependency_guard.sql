-- A shared trigger must not resolve app_users.role against a queues record.
CREATE OR REPLACE FUNCTION protect_desk_dependencies() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_TABLE_NAME='app_users' THEN
  IF NOT NEW.is_active OR NEW.role IS DISTINCT FROM OLD.role THEN
   IF EXISTS(SELECT 1 FROM tickets WHERE assignee_id=OLD.id AND state NOT IN ('cerrado','rechazado')) THEN
    RAISE EXCEPTION 'usuario con tickets pendientes, reasignar primero';
   END IF;
  END IF;
 ELSIF TG_TABLE_NAME='queues' THEN
  IF NOT NEW.is_active THEN
   IF EXISTS(SELECT 1 FROM tickets WHERE queue_id=OLD.id AND state NOT IN ('cerrado','rechazado')) THEN
    RAISE EXCEPTION 'cola con tickets pendientes, transferir primero';
   END IF;
  END IF;
 END IF;
 RETURN NEW;
END $$;

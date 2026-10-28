-- Existing tickets have unknown historical objectives, never today's queue values.
CREATE TABLE ticket_objective_segments (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 ticket_id bigint NOT NULL REFERENCES tickets(id),
 cycle integer NOT NULL,
 queue_id bigint NOT NULL REFERENCES queues(id),
 started_at timestamptz NOT NULL,
 ended_at timestamptz,
 first_response_minutes integer CHECK(first_response_minutes>0),
 resolution_minutes integer CHECK(resolution_minutes>0),
 source text NOT NULL CHECK(source IN ('snapshot','historical_unknown')),
 FOREIGN KEY(ticket_id,cycle) REFERENCES ticket_cycles(ticket_id,cycle),
 CHECK(ended_at>=started_at)
);
CREATE INDEX objective_ticket ON ticket_objective_segments(ticket_id,cycle,started_at);
CREATE UNIQUE INDEX one_objective_segment ON ticket_objective_segments(ticket_id) WHERE ended_at IS NULL;
CREATE INDEX pause_ticket ON ticket_pauses(ticket_id,started_at);
INSERT INTO ticket_objective_segments(ticket_id,cycle,queue_id,started_at,ended_at,source)
 SELECT c.ticket_id,c.cycle,t.queue_id,c.started_at,
 coalesce(c.resolved_at,c.closed_at,(SELECT min(n.started_at) FROM ticket_cycles n WHERE n.ticket_id=c.ticket_id AND n.cycle>c.cycle),CASE WHEN t.state IN ('cerrado','rechazado') THEN t.updated_at END),
 'historical_unknown' FROM ticket_cycles c JOIN tickets t ON t.id=c.ticket_id;

CREATE FUNCTION snapshot_cycle_objectives() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 UPDATE ticket_objective_segments SET ended_at=NEW.started_at WHERE ticket_id=NEW.ticket_id AND ended_at IS NULL;
 UPDATE ticket_pauses SET ended_at=NEW.started_at WHERE ticket_id=NEW.ticket_id AND ended_at IS NULL;
 INSERT INTO ticket_objective_segments(ticket_id,cycle,queue_id,started_at,first_response_minutes,resolution_minutes,source)
 SELECT NEW.ticket_id,NEW.cycle,q.id,NEW.started_at,q.first_response_minutes,q.resolution_minutes,'snapshot'
 FROM tickets t JOIN queues q ON q.id=t.queue_id WHERE t.id=NEW.ticket_id;
 RETURN NEW;
END $$;
CREATE TRIGGER cycle_objective_snapshot AFTER INSERT ON ticket_cycles FOR EACH ROW EXECUTE FUNCTION snapshot_cycle_objectives();

CREATE FUNCTION transfer_objectives() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.queue_id IS DISTINCT FROM OLD.queue_id THEN
  UPDATE ticket_objective_segments SET ended_at=now() WHERE ticket_id=NEW.id AND ended_at IS NULL;
  IF NEW.state NOT IN ('resuelto','cerrado','rechazado') THEN
   INSERT INTO ticket_objective_segments(ticket_id,cycle,queue_id,started_at,first_response_minutes,resolution_minutes,source)
   SELECT NEW.id,NEW.cycle,id,now(),first_response_minutes,resolution_minutes,'snapshot' FROM queues WHERE id=NEW.queue_id;
  END IF;
 END IF;
 IF NEW.state IN ('resuelto','cerrado','rechazado') THEN
  UPDATE ticket_objective_segments SET ended_at=now() WHERE ticket_id=NEW.id AND ended_at IS NULL;
  IF NEW.state='rechazado' THEN
   UPDATE ticket_cycles SET closed_at=now() WHERE ticket_id=NEW.id AND cycle=NEW.cycle;
  END IF;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER ticket_objective_transfer AFTER UPDATE OF queue_id,state ON tickets FOR EACH ROW EXECUTE FUNCTION transfer_objectives();

#!/usr/bin/env bash
set -euo pipefail
# Read-only receipt for the independent 37-ticket demonstration.
docker compose -p catalogo-service-desk-demo -f compose.desk-demo.yaml exec -T db psql -v ON_ERROR_STOP=1 -U catalogo -d catalogo_desk_demo <<'SQL'
SELECT current_database(),now();
SELECT (SELECT count(*) FROM queues) queues,
       (SELECT count(*) FROM tickets) tickets,
       (SELECT count(*) FROM demo_tickets WHERE prepared_at IS NOT NULL) prepared,
       (SELECT count(*) FROM configuration_items WHERE is_device AND is_active) devices,
       (SELECT count(*) FROM faq_articles WHERE is_active AND is_published) faq;
SELECT process_code,count(*) FROM tickets GROUP BY 1 ORDER BY 1;
SELECT details->>'category' category,count(*) FROM tickets WHERE process_code='incidentes' GROUP BY 1 ORDER BY 1;
SELECT username,role,is_active FROM app_users WHERE username LIKE 'desk-%' ORDER BY username;
SELECT case_key,t.number,t.state,t.cycle FROM demo_tickets d JOIN tickets t ON t.id=d.ticket_id ORDER BY case_key;
SELECT md5(string_agg(row_to_json(t)::text,'' ORDER BY id)) ticket_digest FROM tickets t;
SELECT md5(string_agg(row_to_json(t)::text,'' ORDER BY id)) history_digest FROM ticket_history t;
DO $$
BEGIN
 IF (SELECT count(*) FROM tickets)<>37 OR (SELECT count(*) FROM demo_tickets WHERE prepared_at IS NOT NULL)<>37 THEN RAISE EXCEPTION 'expected 37 prepared tickets'; END IF;
 IF (SELECT count(*) FROM queues)<>7 THEN RAISE EXCEPTION 'expected seven queues'; END IF;
 IF EXISTS(SELECT 1 FROM tickets GROUP BY process_code HAVING count(*)<>CASE WHEN process_code='incidentes' THEN 7 ELSE 5 END) THEN RAISE EXCEPTION 'invalid process distribution'; END IF;
 IF (SELECT count(DISTINCT details->>'category') FROM tickets WHERE process_code='incidentes')<>7 THEN RAISE EXCEPTION 'missing I0-I6'; END IF;
 IF (SELECT count(*) FROM configuration_items WHERE is_device AND is_active)<>5000 THEN RAISE EXCEPTION 'expected 5000 devices'; END IF;
 IF (SELECT count(*) FROM faq_articles WHERE is_active AND is_published)<8 THEN RAISE EXCEPTION 'missing FAQ'; END IF;
END $$;
SQL
echo 'PASS demostración: 37 tickets, siete colas, I0–I6, 5000 dispositivos y FAQ'

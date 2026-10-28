#!/usr/bin/env bash
set -euo pipefail
PROJECT="${DESK_TEST_PROJECT:-catalogo-final-test}"
[[ "$PROJECT" == *-test ]] || { echo 'Solo se permite un proyecto -test' >&2; exit 1; }
dc() { docker compose -p "$PROJECT" -f compose.desk-test.yaml "$@"; }
snapshot() {
 dc exec -T db psql -U catalogo -d catalogo_desk_test -At <<'SQL'
SELECT jsonb_build_object(
 'catalogue_level1',(SELECT count(*) FROM services_level1),
 'catalogue_level2',(SELECT count(*) FROM services_level2),
 'catalogue_level1_digest',(SELECT md5(string_agg(row_to_json(t)::text,'' ORDER BY id)) FROM services_level1 t),
 'catalogue_level2_digest',(SELECT md5(string_agg(row_to_json(t)::text,'' ORDER BY id)) FROM services_level2 t),
 'tickets',(SELECT count(*) FROM tickets),
 'history',(SELECT count(*) FROM ticket_history),
 'ci',(SELECT count(*) FROM configuration_items),
 'classes',(SELECT count(*) FROM ci_classes),
 'attachments',(SELECT count(*) FROM ticket_attachments),
 'faq',(SELECT count(*) FROM faq_articles),
 'ticket_digest',(SELECT md5(string_agg(row_to_json(t)::text,'' ORDER BY id)) FROM tickets t),
 'history_digest',(SELECT md5(string_agg(row_to_json(t)::text,'' ORDER BY id)) FROM ticket_history t),
 'ci_digest',(SELECT md5(string_agg(row_to_json(t)::text,'' ORDER BY id)) FROM configuration_items t),
 'attachment_digest',(SELECT md5(string_agg(row_to_json(t)::text,'' ORDER BY id)) FROM ticket_attachments t));
SQL
}
check_attachment() {
 dc exec -T -e DESK_TEST_DATABASE_URL='postgres://catalogo:catalogo@db:5432/catalogo_desk_test?sslmode=disable' -e ATTACHMENT_PERSISTENCE_PASSWORD='TemporalDesk123!' api ./desk-tests -test.v -test.run '^TestAttachmentPersistenceCheck$'
}
workspace=$(mktemp -d)
trap 'rm -f -- "$workspace/before" "$workspace/after"; rmdir -- "$workspace"' EXIT
snapshot > "$workspace/before"
echo 'P18 ANTES'
cat "$workspace/before"
check_attachment
dc stop api
dc restart db
ready=false
for attempt in $(seq 1 60); do
 if dc exec -T db pg_isready -U catalogo -d catalogo_desk_test >/dev/null 2>&1; then ready=true; break; fi
 sleep 1
done
[[ "$ready" == true ]]
dc start api
ready=false
for attempt in $(seq 1 60); do
 if dc exec -T api wget -qO- http://127.0.0.1:8080/api/ready >/dev/null 2>&1; then ready=true; break; fi
 sleep 1
done
[[ "$ready" == true ]]
snapshot > "$workspace/after"
echo 'P18 DESPUÉS'
cat "$workspace/after"
cmp "$workspace/before" "$workspace/after"
check_attachment
echo 'PASS P18: conteos, digest de registros y descarga del adjunto conservados'

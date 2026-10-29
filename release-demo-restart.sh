#!/usr/bin/env bash
set -euo pipefail
cd /home/jeremy/service-desk-release-8a103ac
backup=$(cat deployment-backup-path.txt)
snapshot() { podman exec -i catalogo-service-desk-demo_db_1 psql -U catalogo -d catalogo_desk_demo -At -v ON_ERROR_STOP=1 < release-demo-snapshot.sql; }
snapshot > "$backup/restart-before.json"
echo 'DEMO BEFORE RESTART'
cat "$backup/restart-before.json"
podman stop catalogo-service-desk-demo_api_1
podman restart catalogo-service-desk-demo_db_1
ready=false
for attempt in $(seq 1 30); do
 if podman exec catalogo-service-desk-demo_db_1 pg_isready -U catalogo -d catalogo_desk_demo >/dev/null 2>&1; then ready=true; break; fi
 sleep 1
done
test "$ready" = true
podman start catalogo-service-desk-demo_api_1
podman restart catalogo-service-desk-demo_frontend_1
ready=false
for attempt in $(seq 1 30); do
 if curl --fail --silent http://127.0.0.1:18090/api/ready; then ready=true; break; fi
 sleep 1
done
test "$ready" = true
snapshot > "$backup/restart-after.json"
cmp "$backup/restart-before.json" "$backup/restart-after.json"
podman exec catalogo-service-desk-demo_db_1 psql -U catalogo -d catalogo_desk_demo -Atc "SELECT max(version),(SELECT count(*) FROM ci_classes WHERE technical_profile IS NOT NULL) FROM schema_migrations"
echo 'PASS DEMO RESTART: counts and data digests conserved'

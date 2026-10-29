#!/usr/bin/env bash
set -euo pipefail
umask 077
cd /home/jeremy/service-desk-release-8a103ac
test "$(git rev-parse HEAD)" = 8a103ac5aeef0b36ef4bbb6d3ed5e56efe97fff6
backup=$(mktemp -d /home/jeremy/service-desk-backup-8a103ac-XXXXXX)
printf '%s\n' "$backup" > deployment-backup-path.txt
echo "BACKUP=$backup"
podman inspect catalogo-service-desk-demo_api_1 --format '{{.Image}}' > "$backup/previous-api-image.txt"
podman inspect catalogo-service-desk-demo_frontend_1 --format '{{.Image}}' > "$backup/previous-frontend-image.txt"
podman tag "$(cat "$backup/previous-api-image.txt")" localhost/catalogo-demo-api:pre-8a103ac
podman tag "$(cat "$backup/previous-frontend-image.txt")" localhost/catalogo-demo-frontend:pre-8a103ac
podman stop catalogo-service-desk-demo_frontend_1 catalogo-service-desk-demo_api_1
podman exec catalogo-service-desk-demo_db_1 pg_dump -U catalogo -d catalogo_desk_demo -Fc > "$backup/database.dump"
podman exec -i catalogo-service-desk-demo_db_1 psql -U catalogo -d catalogo_desk_demo -At -v ON_ERROR_STOP=1 < release-demo-snapshot.sql > "$backup/before.json"
volume=/home/jeremy/.local/share/containers/storage/volumes/catalogo-service-desk-demo_desk_demo_attachments/_data
podman unshare tar -czf "$backup/attachments.tgz" -C "$volume" .
manifest() { podman unshare sh -c 'cd "$1" && find . -type f -print0 | sort -z | xargs -0 -r sha256sum' sh "$volume"; }
manifest > "$backup/attachments-before.sha256"
test -s "$backup/database.dump"
echo 'BACKUP OK'
cat "$backup/before.json"
# Promote precisely the images already built and tested, without reseeding data.
podman tag localhost/catalogo-release-test_api:latest localhost/catalogo-service-desk-demo_api:latest
podman tag localhost/catalogo-release-test_frontend:latest localhost/catalogo-service-desk-demo_frontend:latest
docker compose -p catalogo-service-desk-demo -f compose.desk-demo.yaml up -d --no-deps --force-recreate api frontend
ready=false
for attempt in $(seq 1 30); do
 if curl --fail --silent http://127.0.0.1:18090/api/ready; then ready=true; break; fi
 sleep 1
done
test "$ready" = true
podman exec -i catalogo-service-desk-demo_db_1 psql -U catalogo -d catalogo_desk_demo -At -v ON_ERROR_STOP=1 < release-demo-snapshot.sql > "$backup/after.json"
manifest > "$backup/attachments-after.sha256"
cmp "$backup/before.json" "$backup/after.json"
cmp "$backup/attachments-before.sha256" "$backup/attachments-after.sha256"
podman exec catalogo-service-desk-demo_db_1 psql -U catalogo -d catalogo_desk_demo -Atc 'SELECT max(version) FROM schema_migrations'
podman inspect catalogo-service-desk-demo_api_1 --format 'DEPLOYED API {{.Image}}'
podman inspect catalogo-service-desk-demo_frontend_1 --format 'DEPLOYED FRONTEND {{.Image}}'
echo 'PASS deployment: database digests, counts and attachment files conserved'

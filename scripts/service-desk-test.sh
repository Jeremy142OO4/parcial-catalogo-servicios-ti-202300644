#!/usr/bin/env bash
set -euo pipefail
# Dedicated Compose project, DB and volume. Never target the partial's Compose stack.
PROJECT="${DESK_TEST_PROJECT:-catalogo-service-desk-test}"
[[ "$PROJECT" == *-test ]] || { echo 'El proyecto aislado debe terminar en -test' >&2; exit 1; }
dc() { docker compose -p "$PROJECT" -f compose.desk-test.yaml "$@"; }
dc up --build --force-recreate -d
for attempt in $(seq 1 60); do
  if dc exec -T api /bin/sh -c 'wget -qO- http://localhost:8080/api/ready' >/dev/null 2>&1; then break; fi
  if [[ "$attempt" == 60 ]]; then echo 'API no disponible' >&2; exit 1; fi
  sleep 1
done
# Demonstration-only passwords for this isolated test stack.
dc exec -T -e SEED_ADMIN_PASSWORD='TemporalDesk123!' -e SEED_CONSULTA_PASSWORD='TemporalRead123!' api ./catalogo-seed
dc exec -T -e DESK_TEST_DATABASE_URL='postgres://catalogo:catalogo@db:5432/catalogo_desk_test?sslmode=disable' api ./desk-tests -test.v
dc exec -T frontend nginx -t
dc exec -T frontend wget -qO- http://127.0.0.1/api/ready

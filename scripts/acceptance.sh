#!/usr/bin/env bash
set -euo pipefail

# Pruebas P01-P12. Requiere curl, python3 y Docker/Podman Compose.
# Las contraseñas se reciben por variables de entorno y nunca se imprimen.

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_IDENTIFIER="${ADMIN_IDENTIFIER:-admin-demo}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:?Defina ADMIN_PASSWORD con la clave local del usuario administrador}"
CONSULTA_IDENTIFIER="${CONSULTA_IDENTIFIER:-consulta-demo}"
CONSULTA_PASSWORD="${CONSULTA_PASSWORD:?Defina CONSULTA_PASSWORD con la clave local del usuario de consulta}"
RUN_ID="${RUN_ID:-$(date +%Y%m%d%H%M%S)}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

PROJECT="${COMPOSE_PROJECT_NAME:-$(basename "$PWD")}"
container_id() {
  local service="$1" id name
  while read -r id name; do
    if [[ "$name" == "${PROJECT}_${service}_1" || "$name" == "${PROJECT}-${service}-1" ]]; then printf '%s' "$id"; return; fi
  done < <(docker ps -a --format '{{.ID}} {{.Names}}')
  printf 'Contenedor no encontrado: %s\n' "$service" >&2
  return 1
}
API_CONTAINER="$(container_id api)"
DB_CONTAINER="$(container_id db)"
FRONTEND_CONTAINER="$(container_id frontend)"

TOKEN=""
LAST_BODY=""
LAST_STATUS=""

pass() { printf 'PASS %s: %s\n' "$1" "$2"; }
fail() { printf 'FAIL %s: %s\n' "$1" "$2" >&2; exit 1; }

request() {
  local method="$1" path="$2" expected="$3" body="${4:-}"
  local output body_file status
  body_file="$TMP_DIR/body"
  if [[ -n "$body" ]]; then
    status="$(curl -sS -o "$body_file" -w '%{http_code}' -X "$method" \
      -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN" \
      --data "$body" "$BASE_URL$path")"
  else
    status="$(curl -sS -o "$body_file" -w '%{http_code}' -X "$method" \
      -H "Authorization: Bearer $TOKEN" "$BASE_URL$path")"
  fi
  LAST_STATUS="$status"
  LAST_BODY="$(<"$body_file")"
  [[ "$status" == "$expected" ]] || fail "$CURRENT_TEST" "$method $path devolvió HTTP $status; esperado $expected; respuesta: $LAST_BODY"
}

json_value() {
  local expression="$1"
  JSON_VALUE_EXPRESSION="$expression" python3 -c \
    'import json, os, sys; value=json.load(sys.stdin); print(eval(os.environ["JSON_VALUE_EXPRESSION"], {"__builtins__": {}}, {"value": value}))' \
    <<< "$LAST_BODY"
}

login() {
  local identifier="$1" password="$2" expected="$3"
  local previous_token="$TOKEN"
  TOKEN=""
  request POST /api/auth/login "$expected" "{\"identifier\":\"$identifier\",\"password\":\"$password\"}"
  if [[ "$expected" == "200" ]]; then
    TOKEN="$(json_value 'value["token"]')"
  else
    TOKEN="$previous_token"
  fi
}

compose_sql() {
  docker exec "$DB_CONTAINER" psql -U catalogo -d catalogo -Atc "$1" | tr -d '\r'
}

run_import() {
  docker exec "$API_CONTAINER" ./catalogo-importer \
    --input /app/data/CatalogoServicios.xlsx \
    --report /app/outputs/acceptance-import-report.json \
    --database-url 'postgres://catalogo:catalogo@db:5432/catalogo?sslmode=disable' >/dev/null
}

CURRENT_TEST=P01
login "$ADMIN_IDENTIFIER" 'clave-invalida-de-prueba' 401
login "$ADMIN_IDENTIFIER" "$ADMIN_PASSWORD" 200
request GET /api/me 200
pass P01 'inicio de sesión válido e inválido'

# Usuario desechable para P02. Nunca modifica las cuentas de evaluación.
PARENT_ID=""
for KIND in company area department section position; do
  PARENT_JSON=""
  if [[ -n "$PARENT_ID" ]]; then PARENT_JSON=",\"parent_id\":$PARENT_ID"; fi
  request POST /api/organization/units 201 "{\"type\":\"$KIND\",\"code\":\"P02-$RUN_ID\",\"name\":\"Prueba P02\"$PARENT_JSON}"
  PARENT_ID="$(json_value 'value["id"]')"
done
INACTIVE_IDENTIFIER="inactive-$RUN_ID"
INACTIVE_PASSWORD="Temporal-$RUN_ID!"
request POST /api/users 201 "{\"position_id\":$PARENT_ID,\"full_name\":\"Usuario aislado P02\",\"username\":\"$INACTIVE_IDENTIFIER\",\"email\":\"$INACTIVE_IDENTIFIER@example.local\",\"password\":\"$INACTIVE_PASSWORD\",\"role\":\"consulta\"}"
INACTIVE_ID="$(json_value 'value["id"]')"

CURRENT_TEST=P02
TOKEN=""
request GET /api/me 401
login "$ADMIN_IDENTIFIER" "$ADMIN_PASSWORD" 200
request POST /api/auth/logout 204
request GET /api/me 401
login "$ADMIN_IDENTIFIER" "$ADMIN_PASSWORD" 200
request PATCH "/api/users/$INACTIVE_ID/active" 204 '{"is_active":false}'
TOKEN=""
login "$INACTIVE_IDENTIFIER" "$INACTIVE_PASSWORD" 401
TOKEN="$(json_value 'value["token"]' 2>/dev/null || true)"
login "$ADMIN_IDENTIFIER" "$ADMIN_PASSWORD" 200
pass P02 'sesión ausente, logout y usuario inactivo rechazados'

CURRENT_TEST=P03
login "$CONSULTA_IDENTIFIER" "$CONSULTA_PASSWORD" 200
request GET /api/catalog/services 200
request POST /api/catalog/level1 403 '{"code":"P03.NO-AUTORIZADO","name":"No debe crearse"}'
pass P03 'consulta puede leer pero no modificar'

CURRENT_TEST=P04
login "$ADMIN_IDENTIFIER" "$ADMIN_PASSWORD" 200
COMPANY_CODE="ACPT-${RUN_ID}"
AREA_CODE="AR-${RUN_ID}"
DEPT_CODE="DP-${RUN_ID}"
SECTION_CODE="SC-${RUN_ID}"
POSITION_CODE="PS-${RUN_ID}"
SERVICE_CODE="ACPT.${RUN_ID}.1"
SECOND_SECTION_CODE="SC2-${RUN_ID}"
SECOND_SERVICE_CODE="ACPT.${RUN_ID}.2"

request POST /api/organization/units 201 "{\"type\":\"company\",\"code\":\"$COMPANY_CODE\",\"name\":\"Prueba $RUN_ID\"}"
COMPANY_ID="$(json_value 'value["id"]')"
request POST /api/organization/units 201 "{\"type\":\"area\",\"parent_id\":$COMPANY_ID,\"code\":\"$AREA_CODE\",\"name\":\"Área $RUN_ID\"}"
AREA_ID="$(json_value 'value["id"]')"
request POST /api/organization/units 201 "{\"type\":\"department\",\"parent_id\":$AREA_ID,\"code\":\"$DEPT_CODE\",\"name\":\"Departamento $RUN_ID\"}"
DEPT_ID="$(json_value 'value["id"]')"
request POST /api/organization/units 201 "{\"type\":\"section\",\"parent_id\":$DEPT_ID,\"code\":\"$SECTION_CODE\",\"name\":\"Sección $RUN_ID\"}"
SECTION_ID="$(json_value 'value["id"]')"
request POST /api/organization/units 201 "{\"type\":\"position\",\"parent_id\":$SECTION_ID,\"code\":\"$POSITION_CODE\",\"name\":\"Puesto $RUN_ID\"}"
POSITION_ID="$(json_value 'value["id"]')"
USER_NAME="acceptance-${RUN_ID}"
USER_EMAIL="${USER_NAME}@example.local"
request POST /api/users 201 "{\"position_id\":$POSITION_ID,\"full_name\":\"Usuario P04 $RUN_ID\",\"username\":\"$USER_NAME\",\"email\":\"$USER_EMAIL\",\"password\":\"PruebaP04-${RUN_ID}!\",\"role\":\"consulta\"}"
USER_ID="$(json_value 'value["id"]')"
request POST /api/catalog/services 201 "{\"level1_id\":1,\"code\":\"$SERVICE_CODE\",\"name\":\"Servicio de aceptación $RUN_ID\",\"active_value\":\"S\"}"
request GET "/api/catalog/services?q=$SERVICE_CODE" 200
SERVICE_ID="$(json_value 'value[0]["id"]')"
request POST /api/catalog/assignments 204 "{\"service_id\":$SERVICE_ID,\"section_id\":$SECTION_ID,\"responsible_user_id\":$USER_ID}"
request GET "/api/catalog/services?q=$SERVICE_CODE" 200
[[ "$(json_value 'value[0]["section_id"]')" == "$SECTION_ID" ]] || fail P04 'asignación de sección no recuperable'
[[ "$(json_value 'value[0]["responsible_user_id"]')" == "$USER_ID" ]] || fail P04 'asignación de usuario no recuperable'
request POST /api/organization/units 201 "{\"type\":\"section\",\"parent_id\":$DEPT_ID,\"code\":\"$SECOND_SECTION_CODE\",\"name\":\"Sección alterna $RUN_ID\"}"
SECOND_SECTION_ID="$(json_value 'value["id"]')"
request POST /api/catalog/services 201 "{\"level1_id\":1,\"code\":\"$SECOND_SERVICE_CODE\",\"name\":\"Servicio alterno $RUN_ID\"}"
request GET "/api/catalog/services?q=$SECOND_SERVICE_CODE" 200
SECOND_SERVICE_ID="$(json_value 'value[0]["id"]')"
pass P04 'jerarquía, usuario y asignación válidas recuperables'

CURRENT_TEST=P05
request POST /api/organization/units 409 "{\"type\":\"company\",\"code\":\"$COMPANY_CODE\",\"name\":\"Duplicada\"}"
request POST /api/organization/units 400 '{"type":"area","parent_id":999999999,"code":"NO-PARENT","name":"Sin padre"}'
pass P05 'duplicado y referencia inexistente rechazados'

CURRENT_TEST=P06
[[ -f data/CatalogoServicios.xlsx ]] || fail P06 'no existe el Excel original'
run_import
python3 - <<'PY' || fail P06 'el reporte de importación no contiene los conteos esperados'
import json
with open('outputs/acceptance-import-report.json', encoding='utf-8') as fh:
    counts = json.load(fh)['counts']
assert counts['level1'] == 12 and counts['level2'] == 46
assert counts['observed'] >= 4
PY
[[ "$(compose_sql 'SELECT count(*) FROM services_level1;')" -ge 12 ]] || fail P06 'la base no conserva los nivel 1 importados'
[[ "$(compose_sql 'SELECT count(*) FROM services_level2;')" -ge 46 ]] || fail P06 'la base no conserva los nivel 2 importados'
[[ "$(compose_sql 'SELECT count(*) FROM services_level1 WHERE source_sheet IS NOT NULL;')" == 12 ]] || fail P06 'conteo de origen nivel 1 incorrecto'
[[ "$(compose_sql 'SELECT count(*) FROM services_level2 WHERE source_sheet IS NOT NULL;')" == 46 ]] || fail P06 'conteo de origen nivel 2 incorrecto'
pass P06 'Excel original, reporte y conteos importados verificados'

CURRENT_TEST=P07
run_import
python3 - <<'PY' || fail P07 'la reimportación no reporta las actualizaciones reales'
import json
with open('outputs/acceptance-import-report.json', encoding='utf-8') as fh:
    counts = json.load(fh)['counts']
assert counts['created'] == 0 and counts['updated'] == 58
PY
[[ "$(compose_sql 'SELECT count(*) FROM services_level2;')" == "$(compose_sql 'SELECT count(DISTINCT code) FROM services_level2;')" ]] || fail P07 'hay códigos nivel 2 duplicados'
[[ "$(compose_sql 'SELECT count(*) FROM services_level1;')" == "$(compose_sql 'SELECT count(DISTINCT code) FROM services_level1;')" ]] || fail P07 'hay códigos nivel 1 duplicados'
[[ "$(compose_sql "SELECT count(*) FROM import_runs WHERE status = 'succeeded';")" -ge 1 ]] || fail P07 'no existe una importación trazable exitosa'
pass P07 'repetición idempotente y trazable verificada por restricciones y conteos'

CURRENT_TEST=P08
[[ "$(compose_sql "SELECT count(*) FROM services_level2 WHERE code IN ('SE.12.1','SE.12.2','SE.12.3');")" == "3" ]] || fail P08 'no están los tres códigos SE.12.x como texto'
[[ "$(compose_sql "SELECT count(*) FROM services_level2 WHERE code LIKE 'SE.12.%' AND review_required = TRUE AND active_value IS NULL AND service_class_id IS NULL AND criticality_id IS NULL AND service_type_id IS NULL;")" == "3" ]] || fail P08 'los atributos incompletos no se conservaron como desconocidos'
[[ "$(compose_sql "SELECT count(*) FROM services_level1_source_names n JOIN services_level1 l ON l.id = n.service_level1_id WHERE l.code = 'SE.12';")" -ge 2 ]] || fail P08 'no se conservaron ambos nombres originales de SE.12'
[[ "$(compose_sql "SELECT canonical_name FROM services_level1 WHERE code='SE.12';")" == 'Suministrar Analitica' ]] || fail P08 'nombre canónico incorrecto'
[[ "$(compose_sql "SELECT count(*) FROM services_level2 WHERE code IN ('SE.12.1','SE.12.2','SE.12.3') AND metric IS NULL AND minimum IS NULL AND maximum IS NULL;")" == 3 ]] || fail P08 'ausencias de métricas y umbrales alteradas'
pass P08 'SE.12, códigos de texto y atributos ausentes verificados'

CURRENT_TEST=P09
request POST /api/catalog/services 400 "{\"level1_id\":1,\"code\":\"BAD.${RUN_ID}\",\"name\":\"Rango inválido\",\"minimum\":10,\"maximum\":2}"
pass P09 'mínimo mayor que máximo rechazado'
request PATCH "/api/catalog/services/$SECOND_SERVICE_ID" 400 "{\"level1_id\":1,\"code\":\"$SECOND_SERVICE_CODE\",\"name\":\"Rango inválido editado\",\"minimum\":10,\"maximum\":2}"

CURRENT_TEST=P10
request GET "/api/catalog/services?q=$SERVICE_CODE&is_active=true" 200
python3 - "$SERVICE_CODE" "$TMP_DIR/body" <<'PY' || fail P10 'el filtro devolvió resultados incoherentes'
import json, sys
with open(sys.argv[2], encoding='utf-8') as fh:
    items = json.load(fh)
assert items and all(sys.argv[1] in (item['code'] + item['name']) and item['is_active'] for item in items)
PY
pass P10 'búsqueda y filtro de estado verificados'
for FILTER in 'service_class=A%20DEMANDA' 'criticality=High' 'service_type=Back%20End' 'level1_id=1'; do
  request GET "/api/catalog/services?$FILTER" 200
  python3 - "$FILTER" "$TMP_DIR/body" <<'PY' || fail P10 'filtro específico incoherente'
import json, sys, urllib.parse
key, value = urllib.parse.parse_qsl(sys.argv[1])[0]
with open(sys.argv[2]) as fh:
    items = json.load(fh)
assert all(str(item[key]) == value for item in items)
PY
done

CURRENT_TEST=P11
request POST /api/catalog/assignments 400 "{\"service_id\":$SECOND_SERVICE_ID,\"section_id\":$SECOND_SECTION_ID,\"responsible_user_id\":$USER_ID}"
pass P11 'responsable de otra sección rechazado'

CURRENT_TEST=EXTRA
LOOKUP_NAME="ACPT-$RUN_ID"
request POST /api/catalog/lookups/classes 204 "{\"name\":\"$LOOKUP_NAME\"}"
request PATCH /api/catalog/lookups/classes 204 "{\"original\":\"$LOOKUP_NAME\",\"name\":\"$LOOKUP_NAME-editado\"}"
request DELETE /api/catalog/lookups/classes 204 "{\"original\":\"$LOOKUP_NAME-editado\"}"
request POST /api/catalog/services 400 "{\"level1_id\":1,\"code\":\"BAD-LOOKUP-$RUN_ID\",\"name\":\"Referencia inválida\",\"service_class\":\"NO-EXISTE-$RUN_ID\"}"
request POST /api/catalog/level1 201 "{\"code\":\"L1-$RUN_ID\",\"name\":\"Nivel 1 aislado\"}"
TEST_L1_ID="$(json_value 'value["id"]')"
request PATCH "/api/catalog/level1/$TEST_L1_ID/active" 204 '{"is_active":false}'
request GET /api/catalog/level1 200
python3 - "$TEST_L1_ID" "$TMP_DIR/body" <<'PY' || fail EXTRA 'nivel 1 inactivo no recuperable'
import json, sys
with open(sys.argv[2]) as fh:
    assert any(item['id'] == int(sys.argv[1]) and not item['is_active'] for item in json.load(fh))
PY
request PATCH "/api/catalog/level1/$TEST_L1_ID/active" 204 '{"is_active":true}'
pass EXTRA 'CRUD de opciones, referencias controladas y estado nivel 1'

CURRENT_TEST=P12
docker stop "$FRONTEND_CONTAINER" "$API_CONTAINER" "$DB_CONTAINER" >/dev/null
docker start "$DB_CONTAINER" >/dev/null
for attempt in {1..20}; do
  if docker exec "$DB_CONTAINER" pg_isready -U catalogo -d catalogo >/dev/null 2>&1; then break; fi
  sleep 2
done
docker start "$API_CONTAINER" >/dev/null
for attempt in {1..20}; do
  if curl -fsS "$BASE_URL/api/ready" >/dev/null 2>&1; then break; fi
  sleep 2
done
docker start "$FRONTEND_CONTAINER" >/dev/null
curl -fsS "$BASE_URL/api/health" >/dev/null || fail P12 'la API no volvió después del reinicio'
curl -fsS "$BASE_URL/api/ready" >/dev/null || fail P12 'la base no volvió después del reinicio'
[[ "$(compose_sql 'SELECT count(*) FROM services_level1;')" -ge 12 ]] || fail P12 'los nivel 1 no persistieron'
[[ "$(compose_sql 'SELECT count(*) FROM services_level2;')" -ge 46 ]] || fail P12 'los nivel 2 no persistieron'
login "$ADMIN_IDENTIFIER" "$ADMIN_PASSWORD" 200
request GET "/api/catalog/services?q=$SERVICE_CODE" 200
[[ "$(json_value 'value[0]["id"]')" == "$SERVICE_ID" ]] || fail P12 'el registro creado en P04 no persistió'
pass P12 'reinicio sin eliminar volumen conserva los datos'

printf 'Todas las pruebas P01-P12 finalizaron correctamente.\n'

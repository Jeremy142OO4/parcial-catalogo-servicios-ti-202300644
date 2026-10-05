# Pruebas de aceptación

La suite `scripts/acceptance.sh` automatiza los doce escenarios solicitados. Se
ejecuta contra una instalación local ya levantada y recibe las contraseñas por
variables de entorno:

```bash
ADMIN_PASSWORD='clave-local-admin' \
CONSULTA_PASSWORD='clave-local-consulta' \
BASE_URL=http://localhost:8080 \
bash scripts/acceptance.sh
```

La prueba crea datos con un prefijo `ACPT-<marca de tiempo>` para no borrar ni
modificar registros de evaluación. P01–P05, P09 y P11 son pruebas de API; P06–P08
ejecutan dos veces el importador incluido en la imagen y validan el archivo
original, el reporte y la persistencia; P12 reinicia los contenedores sin
eliminar el volumen.

| ID | Cobertura | Verificación principal |
|---|---|---|
| P01 | Autenticación | Login válido e inválido |
| P02 | Sesiones | Sin sesión, logout y usuario inactivo |
| P03 | Autorización | Consulta lee, pero no modifica |
| P04 | Organización | Jerarquía, usuario y asignación válida |
| P05 | Validación | Código duplicado y padre inexistente |
| P06 | Importación | Excel original, 12 nivel 1 y 46 nivel 2 |
| P07 | Idempotencia | Códigos únicos e importación trazable |
| P08 | Incidencias | SE.12, códigos de texto y nulos conservados |
| P09 | Reglas de negocio | Mínimo no mayor que máximo |
| P10 | Consulta | Búsqueda y filtro de estado |
| P11 | Asignaciones | Responsable de otra sección rechazado |
| P12 | Persistencia | Reinicio sin eliminar el volumen |

La ejecución y sus resultados se documentan en
`docs/evidencias/ciclo-crud-y-pruebas.md`.

## Tipo de prueba

La suite actual crea un usuario aislado para P02 y nunca desactiva al usuario de consulta de evaluación. P06 lee `outputs/acceptance-import-report.json`, generado por esa ejecución. P07 exige cero creados y 58 actualizados después de repetir la importación. Los datos con prefijos ACPT, P02 y L1 se conservan para inspección y solo se eliminan al reiniciar explícitamente un entorno desechable.

El runner `tests/Dockerfile` incluye Bash, Python, curl y Docker Compose. El comando Docker Engine está en el README principal. En Fedora con Podman:

```bash
systemctl --user start podman.socket
docker build -f tests/Dockerfile -t catalogo-tests .
docker run --rm --network host --security-opt label=disable \
  -v /run/user/$(id -u)/podman/podman.sock:/var/run/docker.sock \
  -v "$PWD:$PWD" -w "$PWD" \
  -e ADMIN_PASSWORD='clave-local-admin' \
  -e CONSULTA_PASSWORD='clave-local-consulta' \
  catalogo-tests
```

El socket se monta solo en el runner de pruebas. No se monta en la aplicación. El contenedor puede administrar los servicios del entorno de evaluación para verificar P12.

- P01 a P05, P09 y P11 son pruebas de integración de API. Ejecutan la API real, PostgreSQL real y validan autenticación, permisos, relaciones y reglas de negocio.
- P06 a P08 son pruebas de integración del importador con PostgreSQL. Ejecutan el binario dentro del contenedor y verifican conteos, trazabilidad, códigos y valores desconocidos.
- P12 es una prueba de persistencia y recuperación del entorno. Reinicia los contenedores y consulta nuevamente la API y PostgreSQL.
- `go test ./...` contiene pruebas unitarias del paquete de importación y se ejecuta antes de las pruebas de integración.

La suite no depende de una interfaz gráfica para validar los contratos del servidor. Las capturas de la interfaz y del recorrido manual se conservan en `docs/evidencias/`.

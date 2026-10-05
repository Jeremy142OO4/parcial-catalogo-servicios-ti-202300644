# Sistema de gestión del catálogo de servicios de TI

Asignatura: Software Avanzado
Integrante: Jeremy Estuardo Orellana Aldana
Carné: 202300644
Repositorio: `parcial-catalogo-servicios-ti-202300644`
Estado: cierre de auditoría del 4 de octubre de 2026

## Objetivo

Construir una aplicación web para administrar el catálogo de servicios de TI, la estructura organizacional, los usuarios y los responsables de cada servicio.

## Tecnologías acordadas

- Backend: Go.
- Frontend: React, Vite y JSX.
- Base de datos: PostgreSQL.
- Arquitectura: monolito modular con MVC por capas, Service y Repository.
- Ejecución: Docker Compose.
- Verificación actual: laptop Fedora con Podman compatible con Docker.

## Estructura del repositorio

```text
backend/       Código Go, migraciones y pruebas del backend
frontend/      Interfaz React/Vite
data/          Excel original sin modificaciones
scripts/       Comandos reproducibles
tests/         Pruebas y controles de aceptación
docs/          Resolución, contexto, decisiones, prompts y evidencias
```

## Fuente de datos

El archivo original se conserva en:

```text
data/CatalogoServicios.xlsx
```

No debe modificarse. La importación debe conservar 12 códigos de nivel 1 y 46 servicios de nivel 2, documentando las incidencias del archivo.

## Estado actual

Ya están implementados el importador trazable, la persistencia PostgreSQL, la autenticación local, la organización, las asignaciones, el CRUD del catálogo, los filtros y la interfaz React. El reporte validado obtuvo:

```text
Nivel 1: 12
Nivel 2: 46
Registros creados en el análisis: 58
Filas omitidas: 51
Observaciones: 4
```

Reporte: `outputs/import-report.json`.

## Ejecución reproducible actual

Desde un clon limpio, el catedrático puede ejecutar:

```bash
docker compose up --build -d
```

La migración inicial se aplica automáticamente al arrancar la API. La verificación básica queda disponible en:

```text
http://localhost:8080/api/health
http://localhost:8080/api/ready
http://localhost:5173
```

En la laptop de pruebas, este flujo fue comprobado con PostgreSQL saludable, migraciones aplicadas, login local, roles, logout invalidante y el frontend servido por Nginx.

Para crear las cuentas locales de demostración, configurar contraseñas de al menos ocho caracteres solo en la terminal y ejecutar:

```bash
docker compose exec -e SEED_ADMIN_PASSWORD='clave-local-admin' \
  -e SEED_CONSULTA_PASSWORD='clave-local-consulta' \
  api ./catalogo-seed
```

El frontend remoto queda disponible en `http://IP_DEL_EQUIPO:5173`. El proxy `/api` evita configurar la IP manualmente en el navegador.

El seed importa primero el Excel original y después crea tres asignaciones.

### Cuentas de evaluación

| Rol | Usuario | Correo alternativo | Contraseña si se usa el comando de ejemplo |
|---|---|---|---|
| Administrador | `admin-demo` | `admin-demo@example.local` | `clave-local-admin` |
| Consulta | `consulta-demo` | `consulta-demo@example.local` | `clave-local-consulta` |

Se puede iniciar sesión con el usuario o el correo. Las contraseñas de la tabla son exclusivamente de demostración y corresponden al comando anterior. Si se proporcionan otros valores al ejecutar `catalogo-seed`, deben utilizarse esas contraseñas.

Acceso local: `http://localhost:5173`. En la laptop de pruebas: `http://192.168.1.23:5173`.

## Requisitos y operación de Docker

La solución se verificó con Docker Compose compatible con Compose v2. Las imágenes utilizan Go 1.24, Node 22, Nginx 1.27 y PostgreSQL 18. No es necesario instalar Go, Node ni PostgreSQL en el equipo que evalúe el proyecto.

Para validar y persistir el catálogo dentro del contenedor de la API:

```bash
docker compose run --rm api ./catalogo-importer \
  --input /app/data/CatalogoServicios.xlsx \
  --report /app/outputs/import-report.json \
  --database-url 'postgres://catalogo:catalogo@db:5432/catalogo?sslmode=disable'
```

En la prueba remota se verificaron 12 registros de nivel 1, 46 de nivel 2, 4 observaciones y una importación exitosa.

## Pruebas de aceptación

Con los contenedores levantados y las cuentas locales creadas, ejecutar:

```bash
ADMIN_PASSWORD='clave-local-admin' \
CONSULTA_PASSWORD='clave-local-consulta' \
BASE_URL=http://localhost:8080 \
bash scripts/acceptance.sh
```

La suite cubre P01–P12 y crea datos de prueba aislados. Ver [pruebas de aceptación](tests/README.md) y la [evidencia del ciclo CRUD](docs/evidencias/ciclo-crud-y-pruebas.md).

Para ejecutar las pruebas sin instalar Python, Bash o curl en el anfitrión, construir el runner:

```bash
docker build -f tests/Dockerfile -t catalogo-tests .
docker run --rm --network host \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$PWD:$PWD" -w "$PWD" \
  -e ADMIN_PASSWORD='clave-local-admin' \
  -e CONSULTA_PASSWORD='clave-local-consulta' \
  -e BASE_URL=http://localhost:8080 \
  catalogo-tests
```

Este runner se utiliza en Linux con Docker Engine. En Docker Desktop habilitar host networking o usar `BASE_URL=http://host.docker.internal:8080`. El socket permite que P12 reinicie los contenedores del proyecto de pruebas. Ejecutarlo únicamente sobre un entorno de evaluación, ya que también agrega registros aislados. Para Podman, consultar la adaptación en `tests/README.md`.

Comandos de operación habituales:

```bash
docker compose ps
docker compose logs -f api
docker compose stop
docker compose start
docker compose down
```

El comando `docker compose down -v` elimina el volumen y debe reservarse para un entorno de pruebas desechable. No se utiliza durante la suite normal.

## Verificación actual del importador

La prueba actual se ejecuta en la laptop remota mediante un contenedor de Go:

```bash
docker run --rm \
  -v /home/jeremy/parcial-catalogo-servicios-ti-202300644:/workspace:Z \
  -w /workspace \
  golang:1.24 \
  go test ./backend/internal/importer
```

El sufijo `:Z` es necesario en Fedora/Podman para el etiquetado SELinux del volumen. Esta instrucción corresponde a la verificación de desarrollo. El procedimiento final está documentado con Docker Compose.

## Documentación

- [Cierre de auditoría del 4 de octubre](docs/evidencias/2026-10-04-cierre-auditoria.md)
- [Diccionario de datos y mapeo Excel](docs/decisiones/diccionario-y-mapeo.md)

- [Resolución completa de la tarea](docs/RESOLUCION.md)
- [Índice de documentación](docs/README.md)
- [Decisiones de importación](docs/decisiones/001-reglas-importacion-catalogo.md)
- [Prompt del importador](docs/prompts/001-diseno-importador-trazabilidad.md)
- [Evidencia del ciclo de pruebas](docs/evidencias/2026-10-01-ciclo-importador.md)
- [Evidencia del ciclo Docker y PostgreSQL](docs/evidencias/2026-10-01-ciclo-compose.md)
- [Prompt de persistencia del catálogo](docs/prompts/004-persistencia-importacion.md)
- [Prompt de autenticación, organización e interfaz](docs/prompts/005-auth-organizacion-interfaz.md)
- [Evidencia de autenticación e interfaz](docs/evidencias/2026-10-01-ciclo-auth-ui.md)
- [Prompt de CRUD, filtros y pruebas](docs/prompts/006-crud-filtros-pruebas.md)
- [Evidencia de CRUD y pruebas P01–P12](docs/evidencias/ciclo-crud-y-pruebas.md)
- [Evidencias de consola, importación y persistencia](docs/evidencias/2026-10-02-evidencias-consola.md)

## Seguridad

- No subir archivos `.env` reales.
- No subir contraseñas, tokens ni secretos.
- Las cuentas de evaluación se crearán localmente mediante un procedimiento documentado.

## Entrega en GitHub

- Repositorio: `https://github.com/Jeremy142OO4/parcial-catalogo-servicios-ti-202300644`
- Rama de trabajo: `main`
- Etiqueta de entrega prevista: `parcial-v2.0`
- Usuario del catedrático: `maldanap-usac`

Antes de entregar, confirmar en la configuración del repositorio que `maldanap-usac` tenga acceso suficiente. El SHA final y la etiqueta deben actualizarse después del último commit que incluya el código y la documentación.

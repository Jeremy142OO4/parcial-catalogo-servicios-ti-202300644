# Sistema de gestión del catálogo de servicios de TI

Asignatura: Software Avanzado
Repositorio: `parcial-catalogo-servicios-ti-202300644`
Estado: desarrollo

## Objetivo

Construir una aplicación web para administrar el catálogo de servicios de TI, la estructura organizacional, los usuarios y los responsables de cada servicio.

## Tecnologías acordadas

- Backend: Go.
- Frontend: React, Vite y TypeScript.
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

Ya está implementado y probado el primer importador de lectura. El reporte validado obtuvo:

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
```

En la laptop de pruebas, este flujo fue comprobado con PostgreSQL saludable, una migración aplicada y ambos endpoints respondiendo correctamente.

Para validar y persistir el catálogo en un entorno con PostgreSQL disponible:

```bash
go run ./backend/cmd/importer \
  --input data/CatalogoServicios.xlsx \
  --report outputs/import-report.json \
  --database-url "$DATABASE_URL"
```

En la prueba remota se verificaron 12 registros de nivel 1, 46 de nivel 2, 4 observaciones y una importación exitosa.

## Verificación actual del importador

La prueba actual se ejecuta en la laptop remota mediante un contenedor de Go:

```bash
docker run --rm \
  -v /home/jeremy/parcial-catalogo-servicios-ti-202300644:/workspace:Z \
  -w /workspace \
  golang:1.24 \
  go test ./backend/internal/importer
```

El sufijo `:Z` es necesario en Fedora/Podman para el etiquetado SELinux del volumen. Esta instrucción es de verificación de desarrollo; el procedimiento final será documentado con Docker Compose.

## Documentación

- [Resolución completa de la tarea](docs/RESOLUCION.md)
- [Índice de documentación](docs/README.md)
- [Decisiones de importación](docs/decisiones/001-reglas-importacion-catalogo.md)
- [Prompt del importador](docs/prompts/001-diseno-importador-trazabilidad.md)
- [Evidencia del ciclo de pruebas](docs/evidencias/2026-10-01-ciclo-importador.md)
- [Evidencia del ciclo Docker y PostgreSQL](docs/evidencias/2026-10-01-ciclo-compose.md)
- [Prompt de persistencia del catálogo](docs/prompts/004-persistencia-importacion.md)

## Seguridad

- No subir archivos `.env` reales.
- No subir contraseñas, tokens ni secretos.
- Las cuentas de evaluación se crearán localmente mediante un procedimiento documentado.

# Prompt 003: Docker y persistencia PostgreSQL

Fecha de uso: 2026-10-01

Herramienta: Codex desktop.

Modelo/version: no expuesto por la interfaz de esta sesión.

## Prompt utilizado

> Prepara un entorno reproducible con `Dockerfile` y `compose.yaml` para una API Go y PostgreSQL 18. Debe esperar la disponibilidad de la base, conservar los datos mediante un volumen y permitir ejecutar `docker compose up --build -d`. Verifica el arranque real en la laptop remota y documenta cualquier fallo antes de corregirlo.

## Respuesta aplicada

- Se creó un `Dockerfile` multi-etapa para compilar el servidor Go.
- Se creó `compose.yaml` con API, PostgreSQL, healthcheck y volumen.
- La primera ejecución mostró que PostgreSQL 18 requiere montar el volumen en `/var/lib/postgresql`.
- Se corrigió la ruta, se repitió la ejecución desde un volumen de prueba limpio y ambos servicios arrancaron correctamente.
- La API respondió `/api/health` y `/api/ready`; después se comprobó la aplicación de dos migraciones.

## Criterio de aceptación

`docker compose up --build -d` terminó correctamente, la API arrancó después de la base y `/api/ready` respondió `200`.

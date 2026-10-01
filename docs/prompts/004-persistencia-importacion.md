# Prompt 004: persistencia del catálogo importado

Fecha de uso: 2026-10-01

Herramienta: Codex desktop.

Modelo/version: no expuesto por la interfaz de esta sesión.

## Prompt utilizado

> Conecta el importador validado con PostgreSQL mediante la arquitectura Service/Repository. Persiste los niveles 1 y 2, conserva los nombres fuente y las observaciones, respeta valores desconocidos como NULL, hace la operación repetible por código y valida el resultado en la laptop remota.

## Respuesta aplicada

- Se agregó un repositorio de persistencia para `services_level1`, `services_level2`, nombres fuente, observaciones e importaciones.
- Se agregó un servicio que rechaza reportes con errores de validación antes de persistirlos.
- Se agregaron migraciones para la evidencia de origen y la idempotencia de nombres fuente.
- El comando mantiene su reporte JSON y acepta `--database-url` para persistir el resultado.

## Resultado comprobado

```text
Persistencia PostgreSQL: OK
12 servicios nivel 1
46 servicios nivel 2
4 observaciones
1 importación exitosa
```

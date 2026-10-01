# Prompt 002: esquema y persistencia

Fecha de uso: 2026-10-01

Herramienta: Codex desktop.

Modelo/version: no expuesto por la interfaz de esta sesión.

## Prompt utilizado

> Diseña la primera migración PostgreSQL para el sistema de catálogo de servicios. Debe cubrir la jerarquía Empresa-Área-Departamento-Sección-Puesto-Usuario, servicios nivel 1 y nivel 2, catálogos controlados, asignaciones, sesiones e importaciones trazables. Incluye claves foráneas, unicidad, bajas lógicas y la validación mínimo <= máximo. No inventes atributos del Excel y conserva campos de origen.

## Respuesta aplicada

- Se creó una migración inicial normalizada.
- Los códigos de servicios son únicos.
- Los catálogos de clase, criticidad y tipo se cargan como opciones controladas.
- Los atributos faltantes permanecen nulos.
- Las importaciones tienen ejecución y observaciones trazables.
- Las sesiones almacenan el hash del token, no el token en texto plano.

## Criterio de aceptación

La migración debe ejecutarse en PostgreSQL, poder repetirse sin duplicar tablas ni catálogos y permitir que el servidor responda en `/api/ready`.

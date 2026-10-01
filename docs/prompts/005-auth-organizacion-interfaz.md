# Prompt 005: autenticación, organización e interfaz

Fecha de uso: 2026-10-01

Herramienta: Codex desktop.

Modelo/version: no expuesto por la interfaz de esta sesión.

## Prompt utilizado

> Implementa autenticación local con bcrypt, sesiones invalidables, roles administrador/consulta y protección en el servidor. Agrega mantenimiento de la jerarquía Empresa–Área–Departamento–Sección–Puesto, usuarios y asignaciones de servicios con validación de padres activos y responsables pertenecientes a la sección. Construye una interfaz React/Vite conectada a la API, ejecutable mediante Docker Compose y accesible desde otra computadora.

## Respuesta aplicada

- Se implementó el servicio de autenticación y el seed local de cuentas de demostración.
- Se agregaron endpoints protegidos para organización, usuarios, consulta de servicios y asignaciones.
- Se validan roles, padres activos y pertenencia del responsable a la sección.
- Se creó una interfaz React/Vite con login, búsqueda, organización, usuarios y asignaciones.
- Se agregó Nginx como proxy `/api` para que el frontend funcione mediante la IP del host remoto.

## Resultado comprobado

```text
API y PostgreSQL: activos
Frontend: HTTP 200
Login válido: HTTP 200
Login inválido: HTTP 401
Rol consulta intentando modificar: HTTP 403
Logout y reutilización de sesión: HTTP 204 y luego HTTP 401
Asignaciones de demostración: 3
```

# Evidencia: autenticación, organización e interfaz

Fecha: 2026-10-01

## Entorno

Se reconstruyó el entorno remoto con:

```bash
docker compose up --build -d
```

Quedaron activos PostgreSQL, la API Go y el frontend Nginx en los puertos `5432`, `8080` y `5173` respectivamente.

## Controles ejecutados

```text
GET http://127.0.0.1:5173/api/health -> 200 {"status":"ok"}
POST /api/auth/login con admin-demo -> 200
GET /api/me con token válido -> 200
POST /api/auth/login con contraseña incorrecta -> 401
POST /api/auth/login con consulta-demo -> 200
POST /api/organization/units con consulta-demo -> 403
POST /api/auth/logout -> 204
GET /api/me reutilizando sesión cerrada -> 401
Crear una sección con administrador -> 201
Asignar responsable de otra sección -> 400
Desactivar usuario y volver a iniciar sesión -> 204, 401; restauración -> 204
```

El seed creó la jerarquía de demostración, dos usuarios y tres asignaciones válidas. Las contraseñas se entregaron por variables de entorno durante la prueba y no se guardaron en el repositorio.

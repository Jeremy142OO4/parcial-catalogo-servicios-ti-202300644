# Scripts

Comandos reproducibles previstos:

1. `docker compose up --build -d` levanta PostgreSQL, la API y React/Nginx.
2. El importador valida el Excel y puede persistirlo con `--database-url`.
3. `docker compose exec -e SEED_ADMIN_PASSWORD=... -e SEED_CONSULTA_PASSWORD=... api ./catalogo-seed` crea los datos locales de demostración.
4. Las pruebas de aceptación se ejecutan contra los endpoints protegidos y se documentan en `docs/evidencias/`.

Las contraseñas de demostración solo se proporcionan por variables de entorno; nunca se escriben en estos archivos.

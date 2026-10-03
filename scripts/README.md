# Scripts

Comandos reproducibles previstos:

1. `docker compose up --build -d` levanta PostgreSQL, la API y React/Nginx.
2. Las migraciones se aplican automáticamente cuando inicia la API.
3. El importador valida el Excel y puede persistirlo dentro del contenedor:

   ```bash
   docker compose run --rm api ./catalogo-importer \
     --input /app/data/CatalogoServicios.xlsx \
     --report /app/outputs/import-report.json \
     --database-url 'postgres://catalogo:catalogo@db:5432/catalogo?sslmode=disable'
   ```

4. El reporte queda en `outputs/import-report.json` mediante el volumen de salida.
5. `docker compose exec -e SEED_ADMIN_PASSWORD=... -e SEED_CONSULTA_PASSWORD=... api ./catalogo-seed` crea los datos locales de demostración.
6. Las pruebas de aceptación se ejecutan contra los endpoints protegidos y se documentan en `docs/evidencias/`.

Las contraseñas de demostración solo se proporcionan por variables de entorno. Nunca se escriben en estos archivos.

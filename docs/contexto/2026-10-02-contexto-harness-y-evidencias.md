# Actualización de contexto: harness y evidencias

Fecha: 2026-10-02

## Motivo

Después de completar el primer ciclo de desarrollo se necesitó comprobar el entorno remoto, separar las pruebas de infraestructura de las pruebas funcionales y conservar capturas reproducibles para la revisión.

## Información agregada

- La verificación remota se ejecuta en `192.168.1.23` con Docker compatible con Podman.
- El entorno contiene PostgreSQL, la API Go y el frontend Nginx.
- Las pruebas P01 a P12 usan datos con prefijo propio y no eliminan el volumen de evaluación.
- Las capturas 26 a 34 documentan contenedores, endpoints, migraciones, importación, reimportación, pruebas y recuperación.
- Las capturas 24 y 25 documentan el primer fallo de `GetSheetIndex` y la corrección aplicada.
- Los secretos se mantienen fuera de Git y se pasan mediante variables de entorno locales.

## Regla de seguridad agregada

El asistente puede consultar el Excel, el código y los resultados de prueba dentro del proyecto. No debe publicar contraseñas, tokens ni valores secretos, modificar el Excel original ni ejecutar eliminación de volúmenes fuera de un entorno de prueba desechable.

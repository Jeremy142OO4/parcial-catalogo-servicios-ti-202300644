# Pruebas de aceptación

La suite `scripts/acceptance.sh` automatiza los doce escenarios solicitados. Se
ejecuta contra una instalación local ya levantada y recibe las contraseñas por
variables de entorno:

```bash
ADMIN_PASSWORD='clave-local-admin' \
CONSULTA_PASSWORD='clave-local-consulta' \
BASE_URL=http://localhost:8080 \
bash scripts/acceptance.sh
```

La prueba crea datos con un prefijo `ACPT-<marca de tiempo>` para no borrar ni
modificar registros de evaluación. P01–P05, P09 y P11 son pruebas de API; P06–P08
ejecutan dos veces el importador incluido en la imagen y validan el archivo
original, el reporte y la persistencia; P12 reinicia los contenedores sin
eliminar el volumen.

| ID | Cobertura | Verificación principal |
|---|---|---|
| P01 | Autenticación | Login válido e inválido |
| P02 | Sesiones | Sin sesión, logout y usuario inactivo |
| P03 | Autorización | Consulta lee, pero no modifica |
| P04 | Organización | Jerarquía, usuario y asignación válida |
| P05 | Validación | Código duplicado y padre inexistente |
| P06 | Importación | Excel original, 12 nivel 1 y 46 nivel 2 |
| P07 | Idempotencia | Códigos únicos e importación trazable |
| P08 | Incidencias | SE.12, códigos de texto y nulos conservados |
| P09 | Reglas de negocio | Mínimo no mayor que máximo |
| P10 | Consulta | Búsqueda y filtro de estado |
| P11 | Asignaciones | Responsable de otra sección rechazado |
| P12 | Persistencia | Reinicio sin eliminar el volumen |

La ejecución y sus resultados se documentan en
`docs/evidencias/ciclo-crud-y-pruebas.md`.

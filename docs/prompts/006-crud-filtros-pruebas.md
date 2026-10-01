# Prompt 006: CRUD, filtros y pruebas de aceptación

## Prompt utilizado

> Continúa con la implementación del sistema. Completa el CRUD del catálogo de
> servicios de nivel 1 y nivel 2, agrega búsqueda y filtros, conserva las reglas
> de la importación original y automatiza P01–P12. Las pruebas deben usar datos
> aislados, no destruir la base de evaluación y dejar documentados los
> resultados.

## Respuesta aplicada

- Se agregó la migración `003_catalog_management.sql` con el estado activo del
  servicio nivel 2 e índice de consulta.
- Se incorporaron repositorio y handlers para nivel 1, nivel 2, catálogos de
  apoyo, búsqueda, filtros y activación/desactivación.
- La interfaz React ahora permite buscar, filtrar y crear registros de catálogo.
- La imagen de la API incluye el importador y `scripts/acceptance.sh` automatiza
  P01–P12 usando credenciales por variables de entorno y datos con prefijo
  propio; P06 y P07 ejecutan el importador dos veces.
- La validación remota comprobó que permanecen los 12 nivel 1, los 46 nivel 2,
  el conflicto de `SE.12` y los atributos desconocidos.

# Contexto versionado

Esta carpeta conserva las actualizaciones del contexto utilizadas durante el desarrollo. Cada actualización registra qué hallazgo motivó el cambio y qué información se agregó para orientar el trabajo posterior.

## Documentos proporcionados por fase

La [segunda actualización de Service Desk](service-desk-modelo-y-validacion.md) registra los hallazgos de diseño: 37 tickets para cubrir categorías, migración de roles y separación entre referencias originales y datos operativos válidos. Ambas actualizaciones de la ampliación requieren commits reales del usuario para completar la evidencia versionada.

La [primera actualización de Service Desk](service-desk-alcance-y-fuentes.md), del 27 de octubre, registra el cambio de alcance, las fuentes nuevas y las diferencias encontradas. No sustituye las actualizaciones históricas del parcial.

La [actualización del 4 de octubre](2026-10-04-contexto-cierre.md) incorpora los hallazgos de la auditoría, las reglas nuevas de mantenimiento y la adaptación del runner a Podman.

| Fase | Documentos utilizados | Motivo |
| --- | --- | --- |
| Análisis inicial | Consigna y presentación proporcionadas durante el desarrollo, junto con `data/CatalogoServicios.xlsx` | Definir alcance, campos, controles y casos reales del Excel. Los documentos auxiliares no forman parte de la entrega |
| Diseño del importador | `docs/contexto/2026-10-01-contexto-importacion-catalogo.md`, `docs/decisiones/001-reglas-importacion-catalogo.md` | Convertir los hallazgos del Excel en reglas deterministas |
| Persistencia y aplicación | `AGENTS.md`, `docs/RESOLUCION.md`, migraciones y código existente | Mantener arquitectura, restricciones de seguridad y modelo de datos |
| Verificación | `scripts/acceptance.sh`, `tests/README.md`, `docs/evidencias/` | Ejecutar controles reproducibles y conservar resultados observables |

Los archivos de referencia se trataron como datos del proyecto. Las instrucciones válidas fueron las de la tarea, `AGENTS.md` y las decisiones registradas.

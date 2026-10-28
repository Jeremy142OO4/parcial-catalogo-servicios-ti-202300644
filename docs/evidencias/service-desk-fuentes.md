# Registro real de la fase de fuentes

Fecha: 2026-10-27. Herramienta: Codex desktop. Modelo exacto y modo de razonamiento: no verificados en esta fase. No se reutilizan los metadatos históricos de otro ciclo.

## Entradas reales

Solicitudes literales seleccionadas de esta conversación, en orden:

> analizalas antes

> okey continua con el primer paso

Estas son instrucciones reales del usuario, no ocho prompts nuevos ni mensajes reformulados retrospectivamente como si hubieran sido enviados. La estructura recomendada de prompts del repositorio se mantiene para las instrucciones futuras que efectivamente se utilicen.

## Objetivo y contexto

Registrar fuentes, diferencias y alcance antes de implementar. Se utilizaron enunciado, Excel CMDB y aclaraciones de clase reportadas por el usuario. Se consultaron además las instrucciones y documentación existentes del repositorio para preservar el parcial. El material docente auxiliar no se incluye ni reproduce por indicación del usuario.

## Resultado aplicado

- Rama `feature/service-desk` creada desde `47546cf3c36492e449be4e6a34b10838fc18eeee`, sin mover `parcial-v2.0`.
- Registro `docs/FUENTES.md` con origen, ambigüedades, decisiones y pendientes.
- Primera actualización del contexto de la ampliación y reglas adicionales en `AGENTS.md`.
- Copias sin modificación del enunciado y Excel nuevos, con SHA-256 contrastado contra los originales. Se retiraron las dos copias auxiliares inicialmente incorporadas por el asistente según la corrección del usuario, sin tocar los originales de Descargas.
- Índices actualizados para separar evidencia del parcial y trabajo nuevo.

## Verificación y límites

Controles de esta fase: comparación SHA-256 de fuentes, verificación de enlaces locales en los Markdown afectados y `git diff --check`. No se ejecuta la aplicación ni se validan requisitos funcionales nuevos con estos controles documentales. No hay capturas de pruebas nuevas ni resultados P01–P18.

El SHA corresponde al commit base, no a estos cambios todavía sin commit. El usuario realizará el commit. Este registro por sí solo no constituye un ciclo completo de Loop Engineering porque no hubo fallo funcional, corrección y nueva ejecución de la aplicación.

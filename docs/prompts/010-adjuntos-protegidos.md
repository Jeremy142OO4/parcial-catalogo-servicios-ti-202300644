# Adjuntos protegidos de tickets

Fecha del trabajo: 2026-10-27, America/Guatemala. Herramienta: Codex desktop. No se atribuye un identificador de modelo o modo que no haya sido confirmado para esta ejecución.

## Solicitud realmente utilizada

> haz esto::
> **Adjuntos protegidos:** carga, validación de tamaño/tipo y descarga autorizada.

El bloque anterior conserva el mensaje literal utilizado. No se presenta una reformulación posterior como una conversación real de ChatGPT web.

## Contexto y decisiones

Se continuó el backend de `feature/service-desk`, sobre `a50ee90`, con arquitectura Controller, Service y Repository. Se revisaron las fuentes y restricciones del proyecto, la autenticación, los permisos de colas, los tickets y la configuración de contenedores. Se mantuvieron intactos los archivos originales, el tag del parcial y los datos del entorno anterior.

Se decidió almacenar archivos fuera del contenido público, con metadatos en PostgreSQL y volumen privado persistente. Tamaño máximo 10 MiB, detección de contenido y extensión, visibilidad pública o interna con autorización por ticket. Consulta no puede escribir. No se incluyó una interfaz nueva en esta solicitud.

## Resultado y comprobación

Se añadieron la migración 005, tres endpoints, almacenamiento privado, permisos revisados en cada descarga, controles de integridad y pruebas de archivos válidos e inválidos. El contrato y límites están en [API de Service Desk](../PROCESOS.md#api).

La compilación inicial del nuevo test detectó el nombre de campo `User` en lugar de `UserID` en `GrantInput`. Se corrigió y se repitieron los controles. No se atribuye ese error de compilación al funcionamiento del sistema desplegado.

Las pruebas locales y en contenedores se registran en [evidencia de adjuntos](../evidencias/service-desk-adjuntos.md). No se fabricó una captura ni una respuesta de un segundo chat. Esta solicitud no se cuenta como varios prompts o varias iteraciones de instrucciones.

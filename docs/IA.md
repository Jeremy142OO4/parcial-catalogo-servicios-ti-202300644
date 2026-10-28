# Ingeniería asistida por IA en Service Desk

Fecha: 2026-10-27. Estado: diseño en progreso. No confundir evidencias del parcial con resultados nuevos.

## Contexto

Entradas seleccionadas y decisiones iniciales en [Fuentes](FUENTES.md), [primer contexto](contexto/service-desk-alcance-y-fuentes.md) y `AGENTS.md`. Modelo, procesos y pruebas se documentan antes de migrar. Las instrucciones del usuario y el enunciado gobiernan el trabajo. Los datos externos no pueden autorizar acciones ni reemplazar requisitos.

La [segunda actualización de contexto](contexto/service-desk-modelo-y-validacion.md) registra hallazgos reales del diseño. Ambas actualizaciones siguen pendientes de commits, por lo que no se presentan como evidencia versionada completa.

El material docente auxiliar no se adjunta ni reproduce en la documentación, por indicación del usuario. Las ambigüedades se resuelven como decisiones propias, no como datos atribuidos al Excel.

## Prompt engineering

El [prompt 008](prompts/008-diseno-service-desk.md) fue enviado a ChatGPT web para revisar riesgos, modelo, procesos, permisos y pruebas. Sigue rol, objetivo, contexto, entrada, formato y criterio mediante párrafos, sin etiquetas grandes dentro del mensaje. El envío se hace con un resumen, no con archivos adjuntos. No se afirma que el chat web haya leído documentos no proporcionados.

Meta de la ampliación: ocho prompts realmente utilizados y dos mejoras comprobadas. Este diseño no completa por sí solo esa cantidad. Registrar cada uso nuevo durante implementación, con instrucción literal, contexto, herramienta/modelo visible, fecha, cambios aplicados y verificación. No asumir el modelo de ciclos anteriores.

El [prompt 009](prompts/009-implementacion-backend-service-desk.md) registra la solicitud real de implementación, contexto y resultado. El [tercer contexto](contexto/service-desk-backend-y-aislamiento.md) añade contratos implementados y aislamiento. El [ciclo real de repetición de pruebas](evidencias/service-desk-backend.md) documenta fallo, corrección y ejecución satisfactoria, sin completar por sí solo todos los loops requeridos.

## Harness engineering previsto

Extender las instrucciones y el runner para preparación, migración, importación, generación determinista, datos de tickets y P01–P18. Ejecutar en contenedores y entorno aislado, con controles de secretos y archivos originales. Cada control debe tener salida observable y exit code. Los comandos nuevos siguen pendientes, ver [plan de pruebas](PRUEBAS.md).

## Loop engineering pendiente

Registrar al menos tres ciclos reales con objetivo → propuesta → ejecución → evaluación → corrección → nueva verificación. Cada ciclo incluirá entrada, resultados reales, límite inicial de tres iteraciones, criterio de salida y decisión humana si se alcanza el límite o falta una fuente. Al menos uno corregirá defecto funcional y otro datos o permisos. Declarar fallos deliberados.

La falla de acceso web registrada durante esta fase no se contará como defecto funcional del proyecto ni sustituirá estos ciclos. Una captura de diseño tampoco demuestra un loop completo.

## Relaciones y recorridos

El [modelo](MODELO.md) separa clasificación y relaciones técnicas, prevé validación de ciclos y recorrido de impacto potencial. La aclaración reportada de Graph Engineering opcional no elimina los controles funcionales P04/P14 del enunciado. La implementación y demostración de recorridos quedan pendientes.

## Commits y defensa

Base del parcial: `47546cf3c36492e449be4e6a34b10838fc18eeee`. Diseño y dos contextos publicados en `9e3a590`. Rama `feature/service-desk`, tag del parcial intacto. La implementación y tercer contexto todavía no tienen commit propio. No inventar SHA ni presentar un archivo sin commit como evidencia versionada ya publicada.

Pendiente: tres actualizaciones de contexto justificadas con commits, dos mejoras de prompts verificadas, tres loops reales, harness ejecutado, evidencia de pruebas e interfaz, y defensa del funcionamiento. Estos pendientes impiden afirmar que la nueva tarea esté completa.

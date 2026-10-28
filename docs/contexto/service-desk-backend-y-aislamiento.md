# Contexto de backend e integración aislada

Fecha: 2026-10-27. Tercera actualización de contenido de contexto, todavía pendiente de commit propio. Las dos anteriores están en `9e3a590`, no en commits separados inventados.

## Hallazgos

La suite pasó una vez y falló al repetir porque intentaba reconfigurar grants de colas con tickets pendientes. Se mantuvo la regla de dependencias y se aislaron colas/CI por corrida. La CMDB del generador conserva un dataset estable y no duplica 5,000 dispositivos. No usar el total de tickets acumulados de pruebas como conjunto de defensa preparado.

Las aprobaciones necesitan revisión vigente por ciclo. Se añadió reevaluate para modificar alcance sin reciclar decisiones anteriores. Las siete rutas se comprobaron contra PostgreSQL, pero eso no demuestra interfaz, adjuntos ni indicadores.

## Contratos implementados

Consultar [API-SERVICE-DESK](../API-SERVICE-DESK.md) antes de editar: `tickets.details` JSONB reemplaza siete tablas de detalle propuestas, las rutas están en Go y roles usan code como PK. Mantener la API y documentación consistentes, no generar código contra nombres de tablas solo planeadas.

CI con campos operativos válidos se distingue de filas originales de Excel capturadas para revisión. No completar referencias ausentes con datos sintéticos como si fueran originales. El generador se identifica como sintético y registra resumen por clase/ancestro.

## Ejecución

La [evidencia](../evidencias/service-desk-backend.md) y salida real distinguen pruebas locales, integración en laptop y persistencia. El Compose de pruebas es independiente y usa base con sufijo `_desk_test`. El puerto 18081 debe estar libre. No sobrescribir la carpeta anterior de la laptop ni aplicar limpieza destructiva a sus volúmenes.

La siguiente fase debe completar interfaz y pendientes del enunciado. No afirmar P01–P18 completos por tener líneas PASS de cobertura parcial. Los resultados de esta fase aún no tienen SHA de implementación hasta que el usuario confirme un commit.

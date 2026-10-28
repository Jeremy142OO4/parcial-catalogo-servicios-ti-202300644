# Indicadores y recorrido de Service Desk

Implementación del 27 al 28 de octubre de 2026, America/Guatemala. Decisiones del proyecto para completar el enunciado, no SLA atribuidos a las presentaciones.

## Relojes y población

Los objetivos se configuran en minutos calendario. Las fechas se almacenan como instantes PostgreSQL y el cálculo utiliza UTC. La interfaz presenta fechas en la zona del navegador. No se implementa calendario de días hábiles.

Cada consulta aplica la autorización del usuario. El solicitante ve sus tickets, los agentes sus colas autorizadas y consulta los datos funcionales sin escritura ni notas internas. Los filtros del tablero son proceso y cola. Los contadores por proceso, cola, prioridad y estado usan tickets distintos, sin multiplicación por comentarios o CI.

| Indicador | Cálculo y tratamiento |
| --- | --- |
| Antigüedad abierta | Ahora menos creación para tickets no cerrados ni rechazados. Incluye pausas y no se reinicia al transferir |
| Primera atención | Primera actuación humana de atención menos creación, descontando la unión de pausas dentro del intervalo |
| Activo pendiente de primera atención | Mismo cálculo hasta ahora si no hay atención. El valor de primera atención completada permanece NULL |
| Resolución por ciclo | Resolución del ciclo menos inicio de ese ciclo, menos unión de pausas del intervalo |
| Activo del ciclo abierto | Ahora menos inicio del ciclo, menos pausas. No se presenta como tiempo de resolución completada |
| Pausas | Unión de intervalos recortados al período medido. Una pausa abierta termina provisionalmente en el instante de consulta. Un solapamiento no se resta dos veces |
| Reaperturas | Número de acciones `reopen` del historial, no número de ciclos menos uno. La reevaluación abre revisión, pero no cuenta como reapertura |

Tomar un ticket, una transición de atención, una aprobación/rechazo, procesar un evento y un comentario de personal con permiso de atención registran la primera atención si todavía no existe. La simple asignación de responsable, una transferencia o un comentario de un solicitante sin permiso de atención no cuentan. Un administrador que atiende su propio ticket sí actúa como personal de atención.

Resolver conserva la medición de ese ciclo. Cerrar no prolonga su tiempo de resolución. Reabrir inicia otro ciclo y mantiene los anteriores. La primera atención original no se borra. Los rechazados no se contabilizan como resoluciones completadas. En tickets antiguos sin cierre de ciclo registrado, el último cambio terminal limita el cálculo, sin inventar una resolución.

## Objetivos históricos

`ticket_objective_segments`, creada por migración 006, conserva ticket, ciclo, cola, entrada, salida y objetivos de primera atención y resolución. Al iniciar un ciclo se toma una copia de los objetivos configurados. Al transferir se cierra el tramo origen y empieza otro con los valores del destino. La antigüedad y el tiempo total del ciclo no se reinician.

Los objetivos de cada cola se evalúan sobre su tramo de permanencia, de manera prospectiva. No se suman arbitrariamente como un SLA global. La ficha muestra los tramos y sus objetivos para que una transferencia no oculte lo sucedido en origen.

Cambiar la configuración de una cola no actualiza snapshots existentes. Una reapertura toma los objetivos vigentes en su nuevo ciclo. Los objetivos vacíos son desconocidos, no cero ni un cumplimiento supuesto.

La migración inicializa ciclos anteriores como `historical_unknown`, con objetivos NULL. La cola conservada allí es la cola actualmente conocida del ticket, no una reconstrucción verificable de todas sus transferencias antiguas. No se atribuyen retrospectivamente los objetivos actuales a esos tickets.

| Estado | Regla |
| --- | --- |
| cumplido | Tramo completado con tiempo activo menor o igual al objetivo |
| incumplido | Tiempo activo mayor al objetivo, incluso si continúa pendiente |
| pendiente | Tramo en curso dentro de un objetivo conocido |
| sin_objetivo | Objetivo NULL o histórico desconocido |
| sin_resolucion | Ciclo finalizado sin resolución, por ejemplo una reevaluación o rechazo |

Para combinar tramos, un incumplimiento conocido prevalece. Si no lo hay y falta un objetivo, el resultado es sin objetivo. Con todos conocidos, un tramo pendiente mantiene el resultado pendiente. Solo todos los tramos cumplidos producen cumplimiento. El tablero separa las poblaciones, no mezcla desconocidos o abiertos con resoluciones de cero minutos. Los promedios de primera atención omiten NULL y la antigüedad promedio omite terminales.

## Pantallas y operaciones

El catálogo muestra totales independientes de los filtros para nivel 2 y nivel 1, además del número de coincidencias. Los servicios se enumeran con posiciones 1, 2, 3 y sucesivas, continuas entre páginas (13 inicia la segunda página de 12). Los filtros vuelven a enumerar sus resultados desde 1. También se numeran el listado de nivel 1 y los selectores de servicios de Service Desk. Estas posiciones son solo presentación, no cambian IDs, códigos originales ni el Excel. El total incluye registros activos e inactivos, mientras que las coincidencias respetan el filtro de estado.

- Mis solicitudes: búsqueda por número/título y estado, páginas de 20, alta por proceso con campos diferenciados, CI opcionales con justificación y notificaciones propias de eventos.
- Consola de atención: tickets de colas autorizadas, filtros por proceso, cola, estado, prioridad, responsable, servicio y fechas, orden estable por prioridad, creación e ID.
- Ficha: datos del servicio y nivel 1 derivado, acciones con versión, aprobación, transferencia/escalamiento, evidencias posteriores a aprobación, comentarios, archivos protegidos, historial, vínculos, ciclos y objetivos históricos.
- Colas y permisos: crear, editar y desactivar, niveles 0–4, objetivos positivos o desconocidos, permisos por rol o usuario y roles múltiples. El permiso de administración delegada sigue reservado, no habilita configuración fuera del administrador global.
- CMDB: búsqueda y páginas de 20, clasificación completa, CRUD con baja lógica de clases, CI y ubicaciones, atributos originales, JSON técnico validado, servicios, relaciones navegables, impacto potencial y tickets filtrados por autorización.

Las acciones visibles dependen de permisos y estado. El servidor vuelve a validar todas las solicitudes. El catálogo y la administración organizacional anteriores permanecen accesibles en las pestañas principales. La FAQ aún no está implementada en esta fase.

## API y reproducción

`GET /api/desk/indicators` devuelve población autorizada, agrupaciones y métricas por ticket. Acepta `process` y `queue`. `GET /api/desk/tickets/{id}/indicators` devuelve solo ese ticket autorizado. La respuesta incluye `clock: calendar_utc`, instante de cálculo, ciclos y snapshots.

Otros endpoints de interfaz: `GET /api/desk/my-tickets`, `GET /api/desk/intake-queues`, `GET /api/desk/queues/{id}/grants`, `GET /api/cmdb/items/{id}/tickets` y `GET /api/auth/me`. La lista de recepción muestra únicamente ID, código, nombre y proceso de colas activas. Conocerla no otorga permisos de atención.

```bash
docker compose up --build -d
# Interfaz normal en http://localhost:5173
# Preparar catálogo y cuentas según README antes de operar
bash scripts/service-desk-test.sh
# Entorno aislado en http://localhost:18080
```

La salida y el recorrido real están en [pruebas de indicadores e interfaz](evidencias/service-desk-indicadores-interfaz.md). No se requiere instalar Go o Node en el equipo que levanta Compose. La prueba frontend se ejecuta también dentro de la etapa de construcción.

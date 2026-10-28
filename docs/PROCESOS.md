# Procesos, colas y permisos

Fecha: 2026-10-27. Estado: especificación propuesta, pendiente de implementar y verificar. Referencia principal: enunciado Service Desk, apartados 4–7. Reglas complementarias indicadas como decisiones del proyecto.

Actualización: las siete rutas y controles principales ya se ejecutaron en el backend. El [alcance implementado](API-SERVICE-DESK.md) y la [evidencia real](evidencias/service-desk-backend.md) distinguen lo operativo de lo todavía planeado, como adjuntos, objetivos e indicadores. La interfaz nueva sigue pendiente.

## Tipos, colas y flujos diferenciados

Seis tipos y siete procesos. Cada cola tiene exactamente un proceso y puede haber más de una cola por proceso. Todas las rutas registran actor, fecha, transición y evidencia. Estados terminales: cerrado y rechazado cuando corresponda. Resuelto requiere una resolución escrita, cerrado requiere confirmación o motivo de cierre autorizado. Un rechazo conserva decisión y motivo y no se disfraza de resolución técnica.

| Proceso y cola de demostración | Tipo y subtipo | Ruta principal | Condiciones y evidencia |
| --- | --- | --- | --- |
| Nuevos servicios, NUEVOS | Requisición de servicio / nuevo | solicitado → analisis → pendiente_aprobacion → aprobado → implementacion → validacion → resuelto → cerrado. Alternativa pendiente_aprobacion → rechazado | Necesidad, viabilidad y propuesta antes de aprobar. Crear o vincular servicio una sola vez después de aprobación, UNIQUE code y result_service_id trazable. Servicio de contexto existente distinto del resultado |
| Cambios, CAMBIOS | Solicitud de cambio | solicitado → evaluacion → pendiente_aprobacion → aprobado → planificado → ejecucion → validacion → resuelto → cerrado. Alternativa pendiente_aprobacion → rechazado | Impacto, riesgo, CI o justificación sin CI, pruebas y reversión antes de planificar. Sin aprobación vigente no ejecutar. Registrar resultado y validación, fallo permite volver a planificado o ejecutar reversión con historial |
| Incidentes, INCIDENTES | Incidente | registrado → clasificado → diagnostico → atencion → recuperado → resuelto → cerrado. atencion → escalado → diagnostico | Interrupción/degradación, I0–I6, prioridad y redundancia conocida o desconocida. Solución y recuperación documentadas. Escalamiento con motivo y destino autorizado |
| Eventos, EVENTOS | Evento | registrado → clasificado → evaluado → respuesta → resuelto → cerrado. E0/E1 permiten evaluado → resuelto | E0 registro. E1 evaluación sin respuesta operativa. E2/E3 exigen incidente creado/vinculado y notificación verificable antes de resolver. Repetir procesamiento no duplica incidente ni notificación |
| Problemas, PROBLEMAS | Problema | registrado → investigacion → causa_identificada → solucion_propuesta → solucion_aplicada → resuelto → cerrado | Vincular incidentes, causa raíz y workaround si aplica. Marcar error conocido solo con causa y solución temporal documentadas. Corrección definitiva o justificación de salida y cambio relacionado cuando aplica, sin KEDB independiente |
| Accesos, ACCESOS | Solicitud de acceso | solicitado → validacion → pendiente_aprobacion → aprobado → provision_o_revocacion → verificacion → resuelto → cerrado. Alternativa pendiente_aprobacion → rechazado | Beneficiario, recurso/servicio, permiso y acción grant/revoke. Solicitante nunca se autoaprueba, incluso admin. Comprobar decisión vigente y evidencia de provisión/revocación |
| Servicios existentes, SERVICIOS | Requisición de servicio / existente | solicitado → validacion → asignado → cumplimiento → confirmacion → resuelto → cerrado | Servicio existente activo, solicitud concreta y evidencia de cumplimiento. No crear catálogo nuevo ni registrar interrupción como trámite estándar |

Cada estado y transición se vincula al proceso. No permitir saltos mediante edición genérica de state_id. Las aprobaciones son acciones separadas que aplican una transición y guardan decisión en la misma transacción. Cambiar alcance, CI o riesgo tras aprobación invalida esa aprobación y vuelve a evaluación. No permitir rechazos silenciosos ni omitir campos exigidos por una ruta.

## Roles y autorización efectiva

Decisión: permisos aditivos por rol y usuario en la cola, excepto consulta, que es un límite global de solo lectura. Una cuenta que tenga consulta y agente sigue sin escritura hasta retirar consulta explícitamente. Admin administra configuración y puede gestionar tickets, pero nunca omite guardas de aprobación, estado o autoaprobación. Todas las verificaciones se hacen en servidor por acción y objeto.

| Rol | Lectura | Escritura autorizada |
| --- | --- | --- |
| Solicitante | Sus tickets, comentarios públicos y adjuntos públicos de ellos | Crear solicitud y comentario público, confirmación o petición de reapertura de sus tickets según reglas |
| Agente | Colas con can_read y tickets en ellas | Tomar/asignar/atender con can_handle. Transferir con can_transfer. Notas internas solo en ámbito autorizado |
| Aprobador | Ticket en cola con permiso correspondiente | Aprobar/rechazar con can_approve, sin autoaprobación de accesos |
| Administrador | Configuración y datos funcionales | CRUD de configuración y acciones autorizadas, sin manipular auditoría ni saltarse reglas |
| Consulta | Datos funcionales, excluyendo hashes, secretos y contenido interno restringido | Ninguna, incluso por grant directo o membresía |

Los grants de cola pueden asignarse a usuario, rol o ambos mediante filas separadas. Ser responsable del catálogo no convierte automáticamente a alguien en agente. Notas internas visibles solo a agentes autorizados de la cola y administradores. Descarga de adjunto vuelve a comprobar acceso al ticket y visibilidad del adjunto, no basta con conocer su ID.

## Asignación, transferencia y bajas

Política inicial mixta: orden automático, reclamación y asignación manual. Orden `priority_rank ASC, created_at ASC, id ASC`, con Alta=1, Media=2 y Baja=3. Nunca confiar en orden alfabético del rótulo.

Tomar ticket: transacción con UPDATE condicional sobre assignee_id NULL, versión esperada y estado atendible, validar agente y cola activa. Solo un agente obtiene una fila modificada, el otro recibe conflicto 409. Auditoría solo para el ganador. Cualquier selección de siguiente ticket debe bloquear filas o usar SKIP LOCKED.

Transferencia requiere permiso en origen y agente receptor válido en destino o dejar ticket sin reclamar dentro de la cola activa. Destino debe tener el mismo proceso como regla inicial. Para pasar de evento a incidente o de incidente a problema se crea/vincula otro ticket, no se cambia arbitrariamente process_id. Guardar colas y responsables anteriores/nuevos y motivo.

Escalar usa una cola del mismo proceso con mayor nivel configurado y motivo obligatorio. No forzar porcentajes por nivel. Decisión de soporte: cinco niveles configurables 0 Autoservicio, 1 Help Desk, 2 Soporte Técnico, 3 Ingenieros, 4 Externo. Equivalencia del esquema alternativo de cuatro niveles: 0→0, 1→1, 2 Ingenieros→3 y 3 Proveedores→4. Nivel 2 de cinco niveles queda como especialización propia. Este mapeo se documenta como resolución del proyecto basada en la ambigüedad descrita por el enunciado, sin adjuntar material docente.

Desactivar cola con tickets no terminales se rechaza hasta transferirlos a destino activo. Desactivar agente con tickets no terminales asignados se rechaza hasta reasignar o liberar explícitamente en su cola activa. Revocar rol/grant con asignaciones sigue la misma regla. Bajas de CI, servicios o usuarios no eliminan vínculos históricos. Nuevas referencias exigen activos.

## Comentarios, adjuntos y trazabilidad

Decisión inicial: adjuntos hasta 10 MiB, PDF, PNG, JPEG y texto plano, MIME detectado y extensión coherente. Rechazar ejecutables, HTML/SVG activo y rutas proporcionadas por el cliente. Nombre interno aleatorio, almacenamiento fuera del directorio público en volumen persistente. Descarga con Content-Disposition attachment y nosniff, autorización antes de abrir el archivo. Historial no almacena contraseñas, tokens o bytes de adjuntos.

Cada mutación funcional registra before/after, actor, instante, acción y motivo cuando aplica. Actualización e historial son atómicos. Reintentos de solicitudes críticas usan clave de idempotencia. Para E2/E3 una relación única por evento y notificación deduplicada evitan duplicación concurrente. Se puede vincular manualmente un incidente válido existente con motivo.

## Reglas complementarias de incidentes

Decisión del proyecto, no transcripción ni SLA: si afecta seguridad, sugerir I6 antes de usar duración. Si no hay interrupción ni degradación, I0. Con interrupción/degradación y duración positiva conocida, elegir el menor límite que contiene la duración: I1 hasta 3 h, I2 más de 3 hasta 8 h, I3 más de 8 hasta 48 h, I4 más de 48 hasta 168 h, I5 desde 720 h. Usar 30 días para un mes en esta regla únicamente.

Entre más de 168 y menos de 720 h asignar I5 con indicador `duration_gap_override` y motivo obligatorio que advierta que está por debajo del rótulo de un mes. No presentar ese ajuste como valor docente. Si duración desconocida o degradación con cero minutos, permitir registro con categoría manual justificada y `classification_review_required`. Rechazar negativos. Guardar duración estimada y real separadas, ambas nullable y sin convertir NULL a cero. No reemplazar la clasificación histórica sin auditoría al conocer duración real.

Mapeo inicial exigido: I0 baja, I1–I2 media, I3–I6 alta. Referencia de redundancia I0 disponible e I1–I6 no disponible, pero registrar lo observado como true/false/unknown. Cualquier discrepancia con el mapeo exige motivo y revisión, no alterar hechos para encajar. Prioridad del ticket se obtiene de esta regla para incidentes con ajuste autorizado y auditado. Para otros procesos usar matriz impacto/urgencia documentada: alto/alto y alto/medio o medio/alto=Alta, bajo/bajo=Baja, demás combinaciones=Media. Valores desconocidos requieren clasificación de agente antes de atender, no relleno arbitrario.

## Pausas, cierre, reapertura e indicadores

Decisión inicial: objetivos en minutos calendario UTC. Objetivos de cola positivos o desconocidos, nunca inventar cumplimiento si no se configuraron. Al crear ticket y al transferir, guardar versión/snapshot del objetivo para que editar configuración no cambie mediciones históricas. Transferencia no reinicia el reloj original, objetivos nuevos aplican prospectivamente al segmento destino y se muestran separados.

Pausar solo desde estado atendible por espera justificada de solicitante/tercero, con estado previo guardado, motivo y una sola pausa abierta. Reanudar cierra pausa y retorna al estado previo. No pausar tickets terminales ni usar pausas para ocultar incumplimiento. Aprobación pendiente no pausa por defecto. Cerrar exige resuelto, resolución y confirmación autorizada. Rechazado exige motivo. Reabrir cerrado/resuelto por solicitante propio o agente autorizado exige motivo y abre nuevo ciclo en validación/diagnóstico según proceso. No borra resolución anterior y las aprobaciones de ejecución quedan sujetas a nueva revisión si cambia alcance. Rechazo requiere nueva solicitud en lugar de reapertura automática.

| Indicador | Cálculo propuesto y población |
| --- | --- |
| Tickets por cola/proceso/prioridad/estado | Conteo actual con filtros y ACL, sin multiplicar por comentarios o CI |
| Antigüedad abierta | now − created_at para no terminales, incluye pausas y excluye cerrados del agregado abierto |
| Primera atención | Primer evento humano de atención − creación − unión de pausas previas. Asignación automática o comentario del solicitante no cuentan |
| Resolución por ciclo | resolved_at − cycle.started_at − unión de pausas de ese ciclo. Tickets sin resolución no se mezclan como cero |
| Tiempo activo abierto | now − cycle.started_at − pausas, visible como pendiente, no como resolución completada |
| Cumplimiento | Dentro de objetivo / medibles con objetivo configurado. Mostrar por separado cumplidos, incumplidos, pendientes y sin objetivo |
| Reaperturas | Número de ciclos posteriores al inicial, conservar también primera resolución y última resolución |

Fixtures previstos: ticket de 120 min con pausa de 30 da 90 min activos. Ticket abierto de 60 con pausa de 10 da 50 activos, sin tiempo de resolución final. Una reapertura abre ciclo nuevo sin borrar el anterior. Probar fronteras de objetivos e intervalos solapados sin doble resta. No se reportan estos ejemplos como mediciones ejecutadas.

## Datos de demostración

Al menos siete colas activas, una por proceso. Al menos 35 tickets, cinco por proceso y distintos estados. Eventos incluyen E0–E3. Incidentes requieren I0–I6: preparar al menos siete incidentes y por tanto 37 tickets como conjunto inicial, cinco en cada uno de los otros seis procesos. Incluir caso de seguridad, aprobaciones/rechazos, reapertura, pausa, incidente→problema→cambio y caso sin CI justificado. Cuentas sintéticas de administrador, solicitante, dos agentes para concurrencia, aprobador y consulta. No reutilizar datos personales reales.

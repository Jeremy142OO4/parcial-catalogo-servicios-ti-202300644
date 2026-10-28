# Backend de Service Desk: uso y alcance

Inicio: 2026-10-27. Actualización: 2026-10-28. La API y la interfaz React incorporan tickets, colas, CMDB, adjuntos e indicadores. El catálogo del parcial se conserva. El alcance de cada fase y sus pendientes se distinguen en las evidencias.

## Preparación normal

Las migraciones 001–007 se aplican al arrancar la API. La 004 agrega roles, procesos, colas, clasificación, CI y tickets sin borrar el catálogo. La 005 agrega adjuntos, la 006 snapshots de objetivos y la 007 corrige la validación de dependencias al editar colas. Crear cuentas e importar el catálogo con el seed del README antes de generar dispositivos.

```bash
docker compose up --build -d
docker compose exec -e SEED_ADMIN_PASSWORD='clave-local-admin' \
  -e SEED_CONSULTA_PASSWORD='clave-local-consulta' api ./catalogo-seed
docker compose exec api ./cmdb-generator --dataset cmdb-demo --seed 202300644 --count 5000
```

El generador registra las dos filas del Excel en `source_records` con hash y observaciones, sin inventar referencias para convertirlas en CI válidos. Inserta dispositivos sintéticos con ubicaciones, responsables y servicios existentes. Repetir conjunto y semilla conserva códigos y reporta cero creados. No borra al reducir cantidad. Cada ejecución crea una fila de `generation_runs`, no se impone una fila única por todas las repeticiones.

## Verificación aislada reproducible

```bash
bash scripts/service-desk-test.sh
```

El script usa `compose.desk-test.yaml`, proyecto `catalogo-service-desk-test`, base `catalogo_desk_test`, volumen separado y puerto de API local 18081. No se combina como override con el Compose habitual. El puerto debe estar libre. El comando construye, espera disponibilidad, aplica migraciones, ejecuta seed y pruebas. No borra datos previos. Cada corrida crea usuarios, colas y tickets identificados, mientras el conjunto CMDB `integration` es estable. Las contraseñas del script son solo del entorno sintético de pruebas.

Alternativa desde un contenedor preparado, después de seed:

```bash
docker compose -p catalogo-service-desk-test -f compose.desk-test.yaml exec -T \
  -e DESK_TEST_DATABASE_URL='postgres://catalogo:catalogo@db:5432/catalogo_desk_test?sslmode=disable' \
  api ./desk-tests -test.v
```

Sin `DESK_TEST_DATABASE_URL`, la prueba de integración se omite de forma explícita. La URL debe terminar en una base con sufijo `_desk_test`. El resto de tests unitarios se ejecuta normalmente con `go test ./...` en desarrollo.

## Convenciones HTTP

Usar `Authorization: Bearer TOKEN` obtenido por `/api/auth/login`. Nuevas respuestas de lectura/guardado: JSON con HTTP 200. Operaciones sin cuerpo: 204. Permisos: 403. Estado o versión incompatible: 409. Referencias y datos inválidos: 400. Sesión ausente: 401. Bajas lógicas mediante `is_active:false` en PUT, no DELETE de entidades históricas. Los PUT envían el objeto completo, incluidos campos booleanos.

| Método y ruta | Uso |
| --- | --- |
| PUT `/api/users/{id}/roles` | Admin envía `roles` como lista de códigos |
| GET `/api/desk/processes` | Siete procesos y estados |
| GET, POST `/api/desk/queues` | Listar colas autorizadas y crear cola |
| PUT `/api/desk/queues/{id}` | Editar/desactivar, rechaza pendientes |
| PUT `/api/desk/queues/{id}/grants` | Un `user_id` o `role_code`, flags can_read/handle/transfer/approve/admin |
| GET, POST `/api/cmdb/classes` | Consultar/crear clasificación |
| PUT `/api/cmdb/classes/{id}` | Editar, mover o desactivar, sin ciclos |
| GET, POST `/api/cmdb/locations` | Ubicaciones referenciadas |
| PUT `/api/cmdb/locations/{id}` | Actualización completa de ubicación |
| GET, POST `/api/cmdb/items` | Listar y crear CI |
| GET, PUT `/api/cmdb/items/{id}` | Ficha con ancestros, relaciones, servicios e historial, o edición |
| GET `/api/cmdb/items/{id}/impact` | Recorrido de dependientes e impacto potencial, no interrupción confirmada |
| PUT `/api/cmdb/items/{id}/service` | Asociar `service_id` existente activo |
| POST `/api/cmdb/relations` | `source_ci_id`, `target_ci_id`, `kind` |
| PUT, DELETE `/api/cmdb/relations/{id}` | Editar/eliminar vínculo técnico |
| GET, POST `/api/desk/tickets` | Listar según ACL o crear |
| GET `/api/desk/tickets/{id}` | Ficha, comentarios permitidos, vínculos e historial filtrado |
| POST `/api/desk/tickets/{id}/actions` | Transiciones y acciones críticas con versión |
| POST `/api/desk/tickets/{id}/comments` | body y visibility public/internal |
| POST `/api/desk/tickets/{id}/links` | target_id y kind relacionado/incidente_problema/problema_cambio |
| GET `/api/desk/notifications` | Notificaciones del usuario autenticado |
| GET `/api/desk/tickets/{id}/attachments` | Metadatos de adjuntos visibles, sin ruta de almacenamiento |
| POST `/api/desk/tickets/{id}/attachments` | Multipart con un `file` y una `visibility`, devuelve 201 |
| GET `/api/desk/tickets/{id}/attachments/{attachment}/download` | Descarga autenticada y autorizada por ticket y visibilidad |

Filtros CI: q, class_id, status, limit, offset. Filtros tickets: q, process, state, queue, priority, assignee, service, from, to, limit, offset. Límite máximo 100, orden estable. Los códigos y nombres se consultan por sus endpoints. La ficha no incluye todos los tickets de un CI indiscriminadamente porque deben respetarse los permisos de cada ticket.

## Ejemplos de cuerpos

Crear CI (IDs deben existir y estar activos):

```json
{"code":"CI-MANUAL-1","name":"Servidor de prueba","class_id":3,"location_id":1,"owner_user_id":1,"status":"activo","is_device":true,"is_active":true,"technical_attributes":{"ram_gb":16,"storage_gb":512}}
```

No asumir que los IDs del ejemplo corresponden a una clase concreta. Consultarlos antes de enviar.

Crear solicitud de servicio:

```json
{"title":"Solicitud de respaldo","description":"Respaldar el entorno de evaluación","requester_id":1,"service_level2_id":1,"process_code":"servicios","queue_id":7,"priority":2,"impact":"medio","urgency":"medio","channel":"web","no_ci_reason":"Solicitud sobre el servicio, sin equipo identificado","ci_ids":[],"details":{}}
```

Acción con versión obtenida de la última ficha:

```json
{"action":"transition","version":1,"to":"validacion","reason":"Datos verificados"}
```

Acciones implementadas: claim, assign, transfer, escalate, details, approve, reject, transition, close, reopen, reevaluate, pause, resume, process_event y link_service. assign recibe assignee_id. transfer/escalate reciben queue_id y motivo. approve/reject requieren motivo y permiso. Una versión desactualizada no se reintenta a ciegas.

`reevaluate` abre una revisión nueva y conserva las aprobaciones previas, que no valen para el nuevo ciclo. Los detalles de alcance aprobado no pueden editarse directamente. `link_service` vincula un servicio activo existente después de aprobación, dentro de implementación. Se puede crear ese servicio mediante el catálogo existente y después vincularlo. No se duplica al repetir la asociación.

## Ajustes respecto al diseño inicial

- Los detalles diferenciados de procesos se almacenan en `tickets.details` JSONB con guardas específicas en Service, en lugar de siete tablas 1:1. Las claves utilizadas están en `desk.guard` y PROCESOS.md. No implica siete flujos idénticos.
- El estado es texto con FK compuesta a `workflow_states`. Las rutas y guardas están en Go, no en una tabla editable de transiciones.
- `roles.code` es la PK de roles y `user_roles.role_code` la referencia. `app_users.role` queda como proyección primaria compatible, con trigger de sincronización. Los endpoints protegidos usan roles efectivos y consulta deniega escritura.
- La administración de configuración/grants se reserva al administrador global. `can_admin` queda almacenado para evolución de administración delegada, pero no otorga hoy acceso a configuración. Los permisos read/handle/transfer/approve sí se aplican.
- Los ciclos registran resolución y revisión sin borrar historia. Para cambios aprobados se abre reevaluación con motivo, no se recicla una aprobación antigua.

## Pendientes fuera de estos cinco bloques

La confirmación de la taxonomía docente sigue siendo una limitación de P04. Los atributos técnicos dependientes de clase ya se validan y editan, un editor genérico de esquemas es una mejora opcional, no un requisito expreso. Las dos filas del Excel se conservan y consultan en origen, mientras sus referencias ambiguas requieren revisión antes de cualquier alta operativa. No se inventan CI para resolverlas. La FAQ, los 37 tickets y P18 tienen evidencia. Consultar [demostración y cuentas](DEMOSTRACION-SERVICE-DESK.md), [verificación final](evidencias/service-desk-final.md) y [contexto de cierre](contexto/service-desk-cierre-y-fuentes.md).

GET `/api/cmdb/sources` devuelve las filas originales con hash, coordenadas, valores y observaciones, solo al administrador sin rol consulta. No ejecuta escrituras. La preparación de demostración registra la fuente de forma repetible sin generar CI desde referencias no resueltas.

## FAQ y preparación de demostración, migración 008

| Método y ruta | Acceso |
| --- | --- |
| GET `/api/desk/faq?q=&process=&limit=10&offset=0` | Sesión válida, solo artículos publicados y activos |
| GET `/api/desk/faq?manage=true` | Administrador sin rol consulta, incluye borradores e inactivos |
| POST `/api/desk/faq` | Administrador, validación y auditoría |
| PUT `/api/desk/faq/{id}` | Administrador, requiere versión vigente y conserva historial |

Campos: `slug`, `title`, `body`, `process_code` y `service_level2_id` opcionales, `is_published`, `is_active` y `version` para actualizar. El cuerpo se representa como texto, no se ejecuta HTML. La baja es lógica. Los artículos no se borran físicamente ni se permite modificar su historial.

El comando `./desk-demo` agrega datos sintéticos únicamente con `DESK_DEMO_ALLOW_WRITE=1`, `DESK_DEMO_PASSWORD` y una base aislada permitida. Las tablas `demo_tickets` y `prepared_at` mantienen identidad y estado de preparación sin reiniciar tickets ya preparados. La guía contiene los comandos y advertencias del seed inicial.

## Adjuntos protegidos

Implementados mediante migración 005, Service y controlador HTTP. No se publica el directorio de archivos. La ficha React permite cargar y descargar con sesión y visibilidad autorizada.

Carga mediante `multipart/form-data` con exactamente un campo de archivo `file` y un campo `visibility` con valor `public` o `internal`. No hay asociación a un comentario ni reemplazo o eliminación de adjuntos en esta fase. Se conserva la evidencia histórica.

Tamaño por archivo: de 1 byte a 10 MiB inclusive. Solicitud multipart: máximo 11 MiB. El proxy Nginx permite este tamaño y la API aplica su propio límite. Respuestas: 400 para multipart, nombre, visibilidad o archivo vacío inválidos, 413 por tamaño y 415 por tipo o contenido incompatible.

Extensiones admitidas: `.pdf`, `.png`, `.jpg`, `.jpeg` y `.txt`. Se exige coincidencia con el tipo detectado a partir del contenido, sin confiar en el MIME declarado por el cliente. Es detección de firma, no un análisis antivirus ni una validación estructural completa de PDF o imágenes. Los archivos siempre se descargan como adjuntos con `nosniff` y `private, no-store`. HTML, SVG, ejecutables y archivos comprimidos no se admiten.

Puede cargar el solicitante del ticket o un administrador/agente con permiso de atención. El rol consulta nunca escribe, aunque tenga otros roles. Solo personal con permiso de atención, sin rol consulta, puede cargar y descargar archivos internos. Para archivos públicos, se requiere acceso de lectura al ticket, no conocer su URL. Público significa visible a lectores autorizados del ticket, no acceso anónimo. Los tickets cerrados o rechazados deben reabrirse antes de adjuntar.

Cada listado y descarga comprueba la sesión y el ámbito del ticket. La carga revalida permisos y estado en la transacción antes de registrar metadatos e historial. Se rechazan nombres con rutas, controles o longitud superior a 180 bytes. El almacenamiento usa nombres aleatorios de 256 bits, archivos con permisos 0600 y una carpeta privada. La descarga comprueba tamaño y SHA-256 antes de enviar contenido. Los metadatos públicos no exponen claves ni rutas físicas.

`ATTACHMENTS_DIR` define la carpeta privada. Por defecto local es `storage/attachments`, excluida de Git. Compose configura `/app/storage/attachments` sobre el volumen `catalogo_attachments`. Las pruebas usan su propio volumen `desk_test_attachments`. Para respaldar se necesitan tanto PostgreSQL como el volumen de adjuntos. Recrear un contenedor conserva el volumen, eliminarlo explícitamente con `down -v` elimina los archivos. No ejecutar ese comando para una demostración de persistencia.

Los rechazos normales eliminan el archivo provisional. Una interrupción abrupta del proceso entre escritura y commit puede dejar un archivo huérfano sin metadatos, que no es accesible por API. No se implementa todavía una limpieza automática de huérfanos, cuotas acumuladas ni antivirus.

Ejemplo en PowerShell, con token e ID obtenidos del entorno local:

```powershell
curl.exe -H "Authorization: Bearer $token" -F "file=@C:\ruta\evidencia.pdf" -F "visibility=public" "http://localhost:8080/api/desk/tickets/$ticketId/attachments"
curl.exe -H "Authorization: Bearer $token" -o evidencia-descargada.pdf "http://localhost:8080/api/desk/tickets/$ticketId/attachments/$attachmentId/download"
```

Pruebas y salida real en [evidencia de adjuntos](evidencias/service-desk-adjuntos.md).

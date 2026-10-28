# Backend de Service Desk: uso y alcance

Fecha: 2026-10-27. Esta fase implementa API, persistencia y controles. La interfaz React del parcial se conserva y todavía no incorpora las pantallas nuevas de tickets y CMDB.

## Preparación normal

Las migraciones 001–004 se aplican al arrancar la API. La 004 agrega roles, procesos, colas, clasificación, CI y tickets sin borrar el catálogo. Crear cuentas e importar el catálogo con el seed del README antes de generar dispositivos.

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

Interfaz nueva React, adjuntos protegidos, FAQ, indicadores completos y snapshots de objetivos, configuración avanzada de atributos, importación operativa de ejemplos tras revisión, datos de defensa de 37 tickets en estados preparados y cierre integral P01–P18. La API actual no equivale a toda la nueva tarea terminada. El histórico P01–P12 del parcial no se presenta como validación de estos pendientes.

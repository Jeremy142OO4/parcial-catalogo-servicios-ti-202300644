# Actor del panel antiguo y revocación concurrente

Fecha: 28 de octubre de 2026. Zona America/Guatemala, UTC−06:00. Base Git `a6c2de51381eeabb8e6505654964e86f9678e9f9`, más cambios locales sin commit. Las pruebas corresponden al worktree, no al contenido publicado de ese SHA.

## Solicitud real y alcance

Jeremy pidió corregir la pérdida del actor en el panel antiguo de usuarios y la posible acción con permisos recién revocados, con énfasis en verificar los permisos. Se trabajó con Codex desktop sobre código local y PostgreSQL aislado en la laptop. No se creó un nuevo chat web, captura o identificación de modelo que no se haya observado.

## Correcciones

Los endpoints antiguos de crear, editar y activar usuarios transmiten el actor autenticado a métodos transaccionales del servicio. La identidad no procede del JSON. Dentro de la transacción se bloquea al actor y, en una consulta posterior con snapshot nuevo, se verifica que esté activo, tenga admin y no tenga consulta. El mismo control se reutiliza al cambiar roles múltiples. El actor se configura únicamente para esa transacción.

La migración 010 registra cambios de `app_users` con los mismos snapshots y garantías del historial administrativo. Se excluye `password_hash` de ambos valores. La asignación de roles realizada por triggers conserva el actor de la operación HTTP. Las operaciones internas de preparación sin actor siguen identificadas como sistema.

Las acciones sobre tickets adquieren un bloqueo compartido de las colas afectadas antes de comprobar permisos y lo mantienen hasta finalizar la transacción. Las revocaciones usan el bloqueo exclusivo de esa misma cola. Se revisaron además creación de tickets, comentarios, vínculos, generación de incidentes desde eventos y adjuntos. Los tickets destino vinculados se mantienen estables mientras se valida su ámbito.

Si la revocación queda primero en el orden de bloqueo, la acción ve los permisos revocados y recibe 403. Si una acción ya obtuvo el bloqueo y fue autorizada, termina antes de que pueda confirmarse la revocación. No se promete deshacer operaciones previamente autorizadas. Consulta sigue sin escritura incluso al combinarse con admin o agente. La administración de cola no sustituye los permisos separados de atender, aprobar y transferir.

## Pruebas observadas

| Prueba | Resultado |
| --- | --- |
| Revocar atender por usuario y por rol mientras espera una pausa | Dos casos PASS, 403 y ticket/historial sin cambios |
| Revocar aprobar por usuario y por rol mientras espera una aprobación | Dos casos PASS, sin aprobación persistida |
| Revocar transferir por usuario y por rol mientras espera una transferencia | Dos casos PASS, cola y versión sin cambios |
| Restaurar solo el permiso de cada caso | La misma acción pasa sin conceder los otros permisos |
| Panel antiguo mediante HTTP | Crear, cambiar rol, editar nombre, desactivar y reactivar conservan el actor real |
| Intentar suplantar actor desde JSON | El historial conserva al administrador autenticado, no el actor enviado |
| Agente y combinación consulta/admin/agente | POST y PATCH de usuarios reciben 403 |
| Revocación de administrador mientras crear usuario espera el bloqueo | El servicio rechaza la operación y no crea al usuario |
| Historial y rechazos | Ningún hash de contraseña en snapshots, cambio con rol inválido sin nueva evidencia |

Las pruebas de concurrencia observan los bloqueos en `pg_stat_activity` antes de liberarlos. No dependen de asumir que un retraso fijo consiguió crear la carrera.

## Control negativo y loop real

Se preparó una copia temporal en `tmp/permission-negative`, excluida de Git. Solo en esa copia se devolvió la autorización de `Act` al orden previo al bloqueo. El caso `handle_user` falló con `revoked handle action accepted or wrong error: <nil>`, código de salida 1. Es un control negativo deliberado sobre una copia, no una falla del código corregido ni un resultado inventado. [Salida real del control](service-desk-permisos-control-negativo.txt).

Con el código corregido, la suite completa terminó en PASS, código de salida 0. Incluye los seis casos de concurrencia, HTTP de usuarios, regresiones administrativas anteriores, siete flujos, adjuntos e indicadores. [Salida completa final](service-desk-permisos-y-actor-salida.txt).

Entorno: `catalogo-audit-test-api`, `catalogo-audit-test-db` y base `catalogo_audit_desk_test`, versión de migraciones 10. El test binario Linux se compiló localmente y se ejecutó dentro del contenedor con `DESK_TEST_DATABASE_URL` apuntando solo a esa base:

```text
/app/tmp/audit-fixes/desk-tests -test.v
```

También pasaron `go test ./backend/...`, `go vet ./backend/...`, las pruebas de renderizado del frontend, su build y `git diff --check`. La integración local sin DSN se omite, no se cuenta como ejecución de PostgreSQL. La comprobación de persistencia por reinicio P18 no se repitió en este ciclo y aparece como SKIP explícito.

## Entrega y límites

No se modificó la demo original ni se publicaron commits. Sus conteos siguieron en 12/46 servicios, 37 tickets y 5,000 dispositivos, con migración 8 y respuesta `ready`. Para usar las correcciones hace falta desplegar la nueva versión conservando volúmenes. Los datos de prueba no representan los conteos de demostración.

Al terminar se detuvieron los contenedores aislados de esta ejecución, conservando sus datos y sin detener la demo.

No se afirma una auditoría exhaustiva de todas las intercalaciones posibles ni una protección frente a un propietario de PostgreSQL que deshabilite triggers. La corrección de edición de dependencias señalada en la segunda auditoría no fue solicitada en esta intervención y permanece fuera de estos resultados.

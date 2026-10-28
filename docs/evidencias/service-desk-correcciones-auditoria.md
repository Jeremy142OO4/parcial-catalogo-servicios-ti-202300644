# Correcciones de la auditoría administrativa

Fecha: 28 de octubre de 2026. Zona: America/Guatemala, UTC−06:00.

Base Git: `a6c2de51381eeabb8e6505654964e86f9678e9f9`. Las correcciones corresponden al worktree posterior, todavía sin commit ni push. No atribuir estas pruebas al contenido publicado de ese SHA. Se conservaron los cambios documentales existentes y el tag anterior.

## Solicitud y contexto reales

Jeremy solicitó corregir los tres hallazgos de la auditoría: renombrado de clases, trazabilidad administrativa y administración delegada de colas. Se utilizó Codex desktop sobre el repositorio local, con compilación local y pruebas de PostgreSQL en la laptop. No se realizó un nuevo intercambio en ChatGPT web ni se generaron capturas para aparentarlo. No se registra un modelo o nivel de razonamiento nuevo que no haya sido verificado.

Las instrucciones aplicadas fueron AGENTS.md, las fuentes y el contexto de alcance, la autorización de consulta siempre negativa para escritura y la conservación de los archivos originales. El diseño de estas correcciones es una decisión propia, no una regla atribuida al Excel o al docente.

## Solución y pruebas

| Hallazgo | Corrección | Verificación real |
| --- | --- | --- |
| Código de clase usado como identidad técnica | Migración 009 añade perfil técnico inmutable a las seis clases originales. Código y nombre siguen editables | Renombrar servidor, crear y editar CI con RAM, generar 5,000 y repetir con cero duplicados |
| Relaciones eliminadas sin evidencia | Triggers transaccionales conservan actor, fecha y JSON anterior/posterior. El historial de ambos CI conserva creación, edición y eliminación de relaciones | Verificar las tres entradas, snapshot de eliminación y rechazo de UPDATE, DELETE y TRUNCATE del historial administrativo |
| Otras operaciones administrativas sin historial | Registro de clases, ubicaciones, colas, grants, enlaces CI-servicio y roles múltiples | Comprobar evidencia en las siete tablas y actor delegado real en cambios de cola |
| `can_admin` sin aplicación | Autorizar edición y grants de la cola concreta, comprobando permiso dentro del bloqueo transaccional | HTTP permite edición y grants, rechaza otra cola, creación de colas, cambio de roles globales e historial global |
| Consulta y revocación | Consulta no escribe aunque reciba administración. Se consultan permisos vigentes | HTTP 403 para consulta y para permisos de usuario o rol revocados con la misma sesión |
| Interfaz reservada a administrador global | Mostrar colas autorizadas, habilitar administración y ocultar creación y roles globales a delegados | Pruebas de renderizado y compilación Vite satisfactorias |

La auditoría administrativa se consulta con `GET /api/desk/admin-history?limit=25&offset=0`, exclusivamente como administrador global. Los cambios en relaciones también están en `GET /api/cmdb/items/{id}`, campo `history`.

## Entorno y resultados

Entorno nuevo: red `catalogo-audit-test`, contenedores `catalogo-audit-test-db` y `catalogo-audit-test-api`, base `catalogo_audit_desk_test`. Binarios Linux compilados desde el worktree local y transferidos a `/home/jeremy/service-desk-audit-20261028`. No se reemplazó el código de los entornos anteriores.

Comando de la suite, ejecutado dentro del contenedor aislado con `DESK_TEST_DATABASE_URL` apuntando a esa base:

```text
/app/tmp/audit-fixes/desk-tests -test.v
```

[Salida completa real](service-desk-correcciones-auditoria-salida.txt), código de salida 0. Incluye `TestAdministrationRegressions`, `TestFinalAcceptance`, adjuntos, indicadores y siete flujos. La base de pruebas contiene dos datasets reproducibles de 5,000 dispositivos y fixtures adicionales. Su total no es el conteo de la demo. No se inventa un nuevo resultado P18: `TestAttachmentPersistenceCheck` fue omitida por necesitar su ejecución explícita de reinicio.

Comprobaciones locales, todas con resultado satisfactorio:

```text
go test ./backend/...
go vet ./backend/...
npm --prefix frontend test
npm --prefix frontend run build
```

Las pruebas locales sin DSN omiten integración. La evidencia de integración es la salida remota enlazada. La prueba de frontend es renderizado, no interacción real en navegador.

## Migración de una base existente

Se copió por streaming la base demo a `catalogo_upgrade_desk_test`, dentro del PostgreSQL aislado. Sobre esa copia se inició el nuevo backend y se aplicó la migración 009. No se ejecutó seed sobre la copia ni se modificó la base original.

Antes: versión 8, 12 servicios nivel 1, 46 nivel 2, 37 tickets y 5,000 dispositivos activos. Después: versión 9 con los mismos conteos y seis perfiles técnicos. La API respondió `ready`. La demo original continuó en versión 8 con sus conteos originales. [Salida posterior y comprobación de la demo](service-desk-correcciones-auditoria-migracion.txt).

En la preparación del primer entorno vacío se inició seed simultáneamente con la migración automática del servidor. Seed recibió un error de columna ya existente por competir con la migración. Tras esperar `ready`, la repetición del seed pasó. No fue un fallo de la migración sobre la copia existente y no se presenta como una corrección funcional adicional. El procedimiento reproducible debe esperar a que terminen las migraciones antes de ejecutar seed.

## Límites y publicación

El historial administrativo empieza al aplicar la migración. No reconstruye operaciones históricas anteriores. Operaciones internas sin actor se identifican como sistema. Un propietario de PostgreSQL con permisos para deshabilitar triggers puede eludir la protección y queda fuera de la garantía de la aplicación.

Las clases nuevas no adquieren un perfil reservado por usar un nombre parecido. Desactivar una de las seis clases del generador sigue produciendo un rechazo explícito. No se cambia esa regla ni se inventan atributos para categorías personalizadas.

La demo original no fue actualizada en esta intervención. Para que muestre estas correcciones debe reconstruirse desde la nueva versión, conservando sus volúmenes y esperando `ready`. No usar `down -v`. Los cambios de código y documentación todavía necesitan commit y push si Jeremy desea publicarlos.

Al terminar se detuvieron únicamente los tres contenedores creados para esta verificación, conservando sus datos. La demo original respondió nuevamente `ready`. No se dejaron procesos de prueba adicionales consumiendo recursos de la laptop.

# Pruebas reales de adjuntos protegidos

Fecha del trabajo: 2026-10-27, America/Guatemala. Algunas marcas del host remoto aparecen en UTC del 28 de octubre. Base de trabajo `a50ee90`, rama `feature/service-desk`. Los cambios de esta fase no tienen commit todavía y no pertenecen al tag del parcial.

## Entorno aislado

Desarrollo local en Windows y pruebas en la laptop Fedora con Podman y PostgreSQL 18. Se reutilizó exclusivamente el proyecto de pruebas `catalogo-service-desk-test`, base `catalogo_desk_test` y carpeta remota `service-desk-isolated-20261027/source`. La migración 005 se aplicó mediante el arranque normal de la API. Los archivos se guardaron en el nuevo volumen privado de pruebas `desk_test_attachments`, separado del volumen de la aplicación original.

No se modificaron el catálogo original ni los archivos fuente de la tarea. No se arrancaron ni reiniciaron los contenedores anteriores del parcial. Las escrituras de pruebas usan usuarios, colas, tickets y adjuntos sintéticos.

## Controles ejecutados

Localmente finalizaron con éxito `go test ./...`, `go vet ./...` y `git diff --check`. La integración local se omite expresamente sin la variable de base aislada. Ese SKIP no se cuenta como una integración aprobada.

En la laptop se reconstruyó la imagen desde el código modificado mediante:

```bash
bash scripts/service-desk-test.sh
```

El runner finalizó con código 0. [Salida real de build, pruebas y persistencia](service-desk-adjuntos-salida.txt). La suite anterior de backend también pasó, incluidos los siete flujos, autorización, ciclos de CMDB y 5,000 dispositivos sin duplicación al repetir la carga.

| Comprobación | Resultado observado |
| --- | --- |
| Firmas y extensiones PDF, PNG, JPEG y TXT | Aceptadas por el detector unitario. Son firmas sintéticas, no prueba de validez estructural completa |
| HTML, contenido incompatible, ejecutable y nombres con rutas o controles | Rechazados |
| Carga sin sesión, tercero sin acceso y consulta | HTTP 401 o 403 según corresponda |
| Solicitante intenta cargar un archivo interno | HTTP 403 |
| Archivo vacío | HTTP 400 |
| Archivo de 10 MiB más 1 byte | HTTP 413 |
| Solicitud superior a 11 MiB | HTTP 413 |
| Archivo de exactamente 10 MiB | HTTP 201 |
| Rechazos de carga | Sin nuevos archivos residuales en el volumen |
| Carga TXT y descarga por solicitante | Bytes originales conservados, Content-Disposition attachment y nosniff |
| Descarga por tercero y después de revocar permiso de cola | HTTP 403 |
| ID de adjunto asociado a otro ID de ticket | HTTP 404 |
| Consulta descarga archivo público autorizado | HTTP 200, sin escritura |
| Solicitante lista o intenta descargar archivo interno | No aparece en listado y descarga HTTP 403 |
| Clave física de almacenamiento | Ausente del listado público |

## Persistencia

Se ejecutó `TestAttachmentPersistenceCheck` explícitamente contra la API real antes y después de reiniciar los contenedores aislados. El test inicia una sesión con la cuenta sintética, busca el adjunto de prueba, lo descarga y compara su contenido. Las credenciales sintéticas de ejecución no se imprimen como tokens ni se consideran cuentas de producción.

El adjunto del ticket 62, ID 4, conservó sus 30 bytes antes y después. Un intento de reinicio conjunto produjo una advertencia del proveedor Podman por dependencia no iniciada. Para no dar ese intento como prueba suficiente, se repitió el reinicio en secuencia: base, espera de disponibilidad, API, espera de ready y descarga. La salida conserva ambos intentos y las marcas de inicio reales.

El test de persistencia se omite en el runner general si no se proporciona `ATTACHMENT_PERSISTENCE_PASSWORD`. Su ejecución explícita está al final del registro, no se interpreta el SKIP inicial como PASS. Esto comprueba archivos y metadatos de adjuntos, no acredita por sí solo todos los escenarios integrales P18.

## Límites y pendientes

El contrato completo está en [API de Service Desk](../PROCESOS.md#api). La interfaz React para utilizar estos endpoints sigue pendiente. La detección de firma no sustituye antivirus ni desarme de documentos. No hay cuotas acumuladas ni limpieza automática de archivos huérfanos tras caída abrupta. No se han creado capturas simuladas ni atribuido resultados al tag anterior.

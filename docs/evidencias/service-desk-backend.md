# Ejecución real del backend Service Desk

Fecha del trabajo: 2026-10-27, America/Guatemala. El host remoto imprime algunos instantes en UTC del 28 de octubre, sin atribuir esas horas UTC a la fecha local.

Base versionada: `9e3a590`, rama `feature/service-desk`. Los cambios funcionales de esta fase todavía no tienen commit. Los resultados corresponden al worktree modificado, no al tag del parcial. No se inventa un SHA final.

## Entorno y alcance

Laptop Fedora, Podman, PostgreSQL 18 en contenedor, CPU Intel i7-11850H, 16 hilos y 31 GiB de RAM reportados por el host. Se creó una carpeta nueva `service-desk-isolated-20261027`, red `desk-isolated-test`, volumen `desk-isolated-test-data`, contenedores `desk-isolated-test-db` y `desk-isolated-test-api`, y base `catalogo_desk_test`. No se sobrescribió la copia remota anterior ni se iniciaron los contenedores del parcial.

Primero se compilaron binarios Linux localmente y se verificaron contra PostgreSQL aislado. Después se construyó la imagen desde Dockerfile y fuentes nuevas, y se ejecutó la suite dentro de esa imagen. El [registro completo seleccionado de la ejecución](service-desk-backend-salida.txt) contiene salida real, no transcripción reconstruida ni captura simulada.

Finalmente se ejecutó el runner completo `bash scripts/service-desk-test.sh` desde la carpeta de fuentes aislada, con su propio proyecto Compose `catalogo-service-desk-test` y volumen separado. Finalizó con exit code 0. La última imagen comprobada fue `ff0f98ee40247ceab861213c40acd8724f3c40317a5e9249572eeaff670bb995`. La salida enlazada corresponde a esta ejecución completa y muestra build, preparación y pruebas, no solo al comando equivalente de abajo.

## Controles locales

```bash
go test ./...
go vet ./...
git diff --check
```

Los tres finalizaron con éxito. La integración local se omite explícitamente si no recibe una URL de pruebas aislada. No confundir un SKIP con una integración aprobada.

## Ejecución en contenedor

Comando equivalente al ejecutado después de seed y migraciones:

```bash
podman run --rm --network desk-isolated-test \
  -e DESK_TEST_DATABASE_URL='postgres://catalogo:catalogo@db:5432/catalogo_desk_test?sslmode=disable' \
  localhost/catalogo-service-desk:test ./desk-tests -test.v
```

La salida mostró controles de consulta incluso con admin/agente, conservación 12/46 por origen, 5,000 dispositivos y repetición, CRUD CI, rechazo de padres y ciclos, referencias de tickets, dos reclamaciones con un ganador, transferencia con proceso compatible, notas internas filtradas, las siete rutas completas, cierre/reapertura, autoaprobación admin rechazada, eventos E0–E3 sin duplicación, vínculos incidente→problema→cambio y autorización HTTP.

La comparación de tiempos de CI registra una consulta paginada real, no una garantía de rendimiento ni un percentil estadístico. El entorno y población se identifican para permitir repetirla.

## Fallo real y corrección

Objetivo: repetir la suite sin depender de la ejecución anterior. Primera corrida pasó, pero repetirla falló al editar grants de colas que ya tenían tickets de prueba pendientes:

```text
integration_test.go:81: reasignar tickets antes de editar permisos
--- FAIL: TestServiceDeskIntegration (0.22s)
FAIL
```

No se debilitó la regla de permisos para pasar la prueba. La corrección fue crear colas y CI propios por corrida, conservando el conjunto de 5,000 dispositivos con semilla estable. Se repitió la suite satisfactoriamente. Es un ciclo real de datos/aislamiento de pruebas, no un fallo introducido a propósito.

Revisión adicional: se detectó que editar detalles durante validación podía conservar una aprobación previa de otro alcance. Se bloqueó la edición de detalles con aprobación vigente y se agregó reevaluate. La prueba ahora intenta modificar el riesgo aprobado y exige rechazo. La detección inicial fue revisión de código, no se afirma que una ejecución fallida lo haya demostrado.

Durante la actualización de imagen, Podman Compose mantuvo un contenedor anterior pese a construir una imagen nueva. Se observó la discrepancia en las líneas de código de la salida. El runner se corrigió con `up --build --force-recreate -d`. La nueva ejecución recreó solo el proyecto de pruebas y confirmó la versión actual, incluida la denegación de roles revocados entre autenticación y mutación. No se alteraron los contenedores del parcial.

La intervención humana solicitó mantener intactos los datos existentes. Criterio de salida de esta fase: pruebas locales y de integración aprobadas, generación repetible y persistencia demostrada. Límite operativo de tres iteraciones antes de detenerse y escalar un mismo bloqueo. Esta evidencia no constituye por sí sola los tres loops de toda la tarea.

## Generador y conservación

El comando del contenedor `./cmdb-generator --dataset integration --seed 202300644 --count 5000` produjo 5,000 dispositivos y cero creados en repetición. Distribución: Almacenamiento 834, Energía 834, Impresión 833, Red 833, Servidor 833 y Equipo de usuario 833. Ancestro Hardware: 5,000, no se suma como otro grupo de dispositivos.

Las dos filas CMDB se conservaron en `source_records` con hash, coordenadas y observaciones. No se crearon registros ficticios para Base de Datos1 ni propietarios inexistentes.

En un punto de control, antes y después de detener API y reiniciar PostgreSQL sin borrar volúmenes:

```text
devices | tickets | history | source_rows
5000    | 45      | 261     | 2
```

Los conteos coincidieron. Son datos acumulados de pruebas aisladas, no el conjunto de defensa de 37 tickets ya preparado. Corridas posteriores agregan sus propios tickets. Se verificaron migraciones 001–004 registradas.

## Límites

La cobertura de P02 aquí verifica conteos y conserva las pruebas unitarias del importador, no constituye por sí sola una ejecución completa de la nueva P02. P16 solo tiene una medición puntual. El reinicio comprueba persistencia de catálogo/CMDB/tickets/historial, todavía no adjuntos. P17 no está implementada. Los IDs de escenarios mencionados indican cobertura parcial concreta y no una declaración de toda la rúbrica P01–P18 terminada.

La interfaz nueva, adjuntos, FAQ e indicadores siguen pendientes. Las fuentes originales permanecen sin modificar. No se hizo commit, push ni cambio de tag de estos cambios funcionales.

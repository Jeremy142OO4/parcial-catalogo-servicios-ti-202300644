# Despliegue y verificación final en la laptop

Fecha: 28 de octubre de 2026. Zona America/Guatemala, UTC−06:00.

Versión publicada y desplegada: `8a103ac5aeef0b36ef4bbb6d3ed5e56efe97fff6`, merge de Service Desk en main. Se clonó esa versión en `/home/jeremy/service-desk-release-8a103ac`, sin sobrescribir la copia remota anterior. No se cambió código durante esta intervención. Este documento y sus evidencias son posteriores al commit probado.

## Estado disponible

Demo: [192.168.1.23:18090](http://192.168.1.23:18090/). En la laptop: [localhost:18090](http://localhost:18090/).

Proyecto `catalogo-service-desk-demo`, base `catalogo_desk_demo`. Migraciones aplicadas hasta la 010. Se conservaron los mismos volúmenes de PostgreSQL y adjuntos, 12 servicios nivel 1, 46 nivel 2, siete colas, 37 tickets preparados, 5,000 dispositivos y ocho FAQ publicados. No se ejecutó seed ni se regeneraron los datos de la demo.

Las imágenes promovidas a la demo son exactamente las construidas y probadas en el proyecto aislado. Sus identificadores están en [metadatos](service-desk-despliegue-metadatos.txt).

## Respaldo y actualización

Se detuvieron brevemente API y frontend para respaldar PostgreSQL y el volumen de adjuntos con un punto consistente. Respaldo privado en la laptop: `/home/jeremy/service-desk-backup-8a103ac-Ku4iLo`. No se copió a Git porque contiene datos y hashes de cuentas.

Se guardaron las imágenes anteriores con etiquetas locales `catalogo-demo-api:pre-8a103ac` y `catalogo-demo-frontend:pre-8a103ac`. Se recrearon exclusivamente API y frontend del proyecto demo desde el clon nuevo, usando las imágenes verificadas y conservando volúmenes. No se utilizó `down -v`.

Se compararon conteos y digests de catálogo, tickets, historial, CI, usuarios, roles, colas, grants, clasificación, relaciones y FAQ. Para clasificación se excluyó solamente la columna técnica nueva al comparar contra el esquema anterior. Se compararon además los manifiestos de archivos de adjuntos. La demo no tenía adjuntos ni relaciones técnicas almacenadas, por lo que esos dos controles estaban vacíos. La persistencia con archivos reales se probó en el entorno aislado.

[Salida de actualización](service-desk-despliegue-salida.txt). Resultado: PASS, exit code 0.

Se restauró además el dump anterior en una base nueva aislada `catalogo_restore_desk_test`. La restauración terminó sin errores y conservó versión 8, 12/46 servicios, 37 tickets y 5,000 dispositivos. También se validó la lectura del archivo comprimido de adjuntos. No se restauró encima de la demo actualizada. [Resultado de restauración y versión activa](service-desk-despliegue-metadatos.txt).

## Pruebas completas en entorno separado

Proyecto nuevo `catalogo-release-test`, base `catalogo_desk_test`, puertos 18380 y 18381. La construcción desde el clon ejecutó las pruebas del frontend, su build, integridad de originales y revisión heurística de secretos del árbol de fuentes.

```bash
DESK_TEST_PROJECT=catalogo-release-test DESK_TEST_WEB_PORT=18380 DESK_TEST_API_PORT=18381 bash scripts/service-desk-test.sh
```

La suite de integración pasó desde cero y volvió a ejecutarse para conservar una salida completa de la versión desplegada. Incluye siete flujos, CRUD de CI, importación repetida, generación, indicadores, FAQ, adjuntos, auditoría, consulta con roles elevados, administración delegada y seis revocaciones concurrentes por usuario o rol. [Salida completa final](service-desk-despliegue-pruebas.txt), exit code 0.

Los conteos mayores de la base de pruebas corresponden a dos datasets de 5,000 dispositivos y fixtures de ejecuciones repetidas. No se confundieron con los 5,000 dispositivos y 37 tickets de la demo.

## P18 y reinicio de la demo

```bash
DESK_TEST_PROJECT=catalogo-release-test DESK_TEST_WEB_PORT=18380 DESK_TEST_API_PORT=18381 bash scripts/service-desk-persistence.sh
```

P18 pasó con comparación de conteos y digests antes y después de reiniciar API y PostgreSQL. La descarga protegida del adjunto real devolvió sus 30 bytes originales antes y después. El SKIP de esta prueba en la suite general no se cuenta como PASS, se reemplaza por esta ejecución explícita. [Salida P18](service-desk-despliegue-persistencia.txt).

También se reiniciaron API, PostgreSQL y frontend de la demo actualizada. Su snapshot posterior fue idéntico, la API respondió `ready` y se conservaron migración 10 y seis perfiles técnicos. [Reinicio de la demo](service-desk-despliegue-reinicio-demo.txt). La [verificación de datos preparados](service-desk-despliegue-demo.txt) conservó 37 casos, I0–I6, siete procesos, siete colas, 5,000 dispositivos y ocho FAQ.

## Comprobación real de navegador

Se comprobó inicio de sesión con las seis cuentas `desk-admin`, `desk-solicitante`, `desk-agente1`, `desk-agente2`, `desk-aprobador` y `desk-consulta`, usando las credenciales originales, sin restablecerlas. La contraseña se recuperó del registro de preparación de esta conversación, no de un gestor del navegador ni de la base de datos. No se incorpora a Git.

Consulta no mostró crear tickets ni administración global o de colas. Las otras cuentas no administrativas tampoco mostraron administración global. Como administrador cargaron sin alertas las pantallas de FAQ, CMDB, consola, catálogo, organización, administración e indicadores. En colas se comprobó que `Administrar esta cola` es editable, sin modificar ni guardar grants. Los indicadores mostraron los 37 tickets conservados.

Las operaciones con escritura de estas funciones se probaron mediante la suite aislada, no sobre los casos preparados. Después de los accesos de navegador, los digests de datos de la demo seguían coincidiendo con los originales. Los cambios de sesiones por login y logout no forman parte de esa comparación.

![Demo actualizada con 37 tickets visibles](../Imagenes/service-desk-demo-actualizada.png)

Al terminar se detuvieron exclusivamente los tres contenedores nuevos de pruebas, conservando sus volúmenes. La demo quedó encendida y abierta como administrador en indicadores.

## Límites y pendiente conocido

Las pruebas ejecutadas pasaron, pero no equivalen a verificar manualmente toda combinación posible de estados y permisos. Sigue pendiente el falso positivo al editar determinadas dependencias entre CI, identificado en la segunda auditoría y no incluido en las correcciones publicadas. No se modificó código para ocultarlo ni se declara completa toda la tarea por tener una suite satisfactoria.

No se creó un tag nuevo ni se hizo commit o push de las evidencias de este despliegue. El tag anterior `parcial-v2.0` se conserva. La entrega en Wedy y el acceso del catedrático requieren comprobación del estudiante.

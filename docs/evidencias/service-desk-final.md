# Verificación final de Service Desk

Fecha: 28 de octubre de 2026, zona America/Guatemala. Estas son pruebas nuevas de Service Desk, no la suite histórica P01–P12 del catálogo.

## Resultado y alcance

La suite ejecutada en PostgreSQL pasó, la interfaz compiló y P18 pasó con reinicio real de base y API. La demostración separada conserva exactamente 37 tickets, siete colas, ocho artículos FAQ y 5,000 dispositivos. Esto no equivale a certificar la calificación al 100 %. P04 tenía una reserva documental al ejecutar esta fase. La aclaración posterior de Jeremy permite su [cierre documental en el alcance disponible, con salvedad interpretativa](../decisiones/002-interpretacion-cmdb-p04.md). No se cambian las salidas originales ni se atribuyen las clases propias al docente. Las pruebas de navegador descritas abajo son recorridos concretos, no una suite E2E de todos los formularios.

La matriz registra lo observado en cada prueba. No convierte una comprobación parcial en evidencia de todos los casos posibles del requisito.

| ID | Resultado observado | Evidencia ejecutada |
| --- | --- | --- |
| P01 | Verificado en el alcance indicado | HTTP sin sesión y con sesión cerrada retorna 401, consulta con roles elevados no escribe, cuenta inactiva no inicia sesión ni reutiliza sesión |
| P02 | Verificado | Dos importaciones consecutivas mantienen 12/46 por origen, con regresiones unitarias del importador para SE.12, combinaciones, códigos y ausencias |
| P03 | Verificado | 5,000 dispositivos del dataset `integration`, semilla `202300644`, repetición con cero altas, códigos deterministas únicos |
| P04 | Verificado en el alcance disponible, con salvedad interpretativa | Padre inválido, ciclo indirecto y carrera A→B/B→A rechazados. Correspondencia de campos y valores documentada. El pendiente de obtener un árbol adicional queda cerrado por la decisión posterior, sin atribuir las seis subclases propias al material docente |
| P05 | Verificado en el alcance indicado | Referencia inexistente, servicio/CI/cola inactivos y ausencia injustificada de CI rechazados. Nivel 1 derivado desde nivel 2 |
| P06 | Verificado en el alcance indicado | Permisos de cola por rol y usuario, notas internas ocultas, revocación de descarga, consulta sin escrituras aun con roles superiores |
| P07 | Verificado | Empate de prioridad y fecha ordenado por ID, dos reclamaciones sincronizadas con un éxito y exactamente una auditoría |
| P08 | Verificado en el alcance indicado | Transferencia válida y rechazo de destino de otro proceso. La suite conserva el historial generado, no simula una transferencia solo cambiando la UI |
| P09 | Verificado | Ruta completa de los siete procesos hasta cerrado, saltos inválidos rechazados. La demostración incluye además un acceso rechazado |
| P10 | Verificado en el alcance indicado | Aprobación requerida en las rutas, autoaprobación del acceso solicitado por el administrador rechazada y alcance aprobado protegido contra cambios posteriores |
| P11 | Verificado | E0–E3, repetición E2/E3 y carrera concurrente con un vínculo y una notificación por evento |
| P12 | Verificado según regla documentada | Fronteras 3/8/48/168/720 h, hueco I5, cero, NULL, negativos, seguridad y redundancia contradictoria o desconocida. La regla complementaria del proyecto no se atribuye falsamente a ITIL |
| P13 | Verificado en el alcance indicado | Vínculos tipados incidente→problema→cambio sobre tickets de rutas completas, con detalles de causa y solución conservados |
| P14 | Verificado en el alcance indicado | Dependientes y servicio asociado recuperados, ciclo indirecto `depende_de` rechazado, triángulo `conectado_a` permitido y marca de límite de profundidad alcanzada |
| P15 | Verificado en el alcance indicado | Cierre y reapertura con nuevo ciclo. Fallo de auditoría inyectado produce rollback de reclamación, versión y responsable. Borrado del historial rechazado |
| P16 | Verificado en el alcance indicado | Filtros combinados por texto, cola, proceso, estado, prioridad, servicio, responsable y fechas. Página estable y plan SQL con al menos 5,000 dispositivos |
| P17 | Verificado | PostgreSQL controlado: primera atención 30 min, resolución neta 90 min, objetivo histórico 90, nuevo ciclo 200, pausas y objetivo desconocido |
| P18 | Verificado | Conteos y digest iguales antes/después de reiniciar PostgreSQL y API. Adjunto descargado por HTTP con contenido idéntico en ambos momentos |

Las precisiones de alcance son intencionales. Por ejemplo, P14 devuelve rutas de impacto potencial y no promete una interrupción confirmada o una lista global de servicios sin caminos alternativos. La prueba no intenta agotar un grafo de más de 1,000 resultados. P16 registra mediciones reales, no un SLA ni una prueba de carga concurrente.

## Entorno y versión

- Desarrollo local Windows, Go y React/Vite.
- Ejecución remota Fedora, Podman 5.8.4 con podman-compose 1.5.0, PostgreSQL 18 Alpine.
- Laptop Intel Core i7-11850H, 16 hilos lógicos y 31 GiB de memoria, según la [salida de demostración](service-desk-demo-salida.txt).
- Rama `feature/service-desk`, HEAD de partida `55404d16765d18963bf7d9df348d0cbc26d87b7c`, más cambios locales de indicadores, interfaz, numeración, FAQ y pruebas. **No se probó un commit nuevo publicado.**
- Tag histórico sin mover: `parcial-v2.0` → `47546cf3c36492e449be4e6a34b10838fc18eeee`.
- Copia de fuentes en `/home/jeremy/service-desk-isolated-20261027/source`, enviada sin dependencias locales ni build del frontend. Las imágenes se construyeron desde esas fuentes.
- Coincidencia local/remota comprobada: SHA-256 de `final_integration_test.go` = `FE04F666AC52D3988333346A76A372866E28A328DE3B5DB1C3AFB62B4EF449CB` y de `demo.go` = `C9697A964DBE05B2760982C34A83C103C91001029F93D18FE7562A49CA532146`.
- Proyecto de pruebas `catalogo-final-test`, base `catalogo_desk_test`, frontend 18180 y API en loopback 18181. Demostración `catalogo-service-desk-demo`, base diferente y frontend 18090. Se conservó el entorno previo de 18080.

## Comandos ejecutados

Desde las fuentes de la laptop:

```bash
DESK_TEST_PROJECT=catalogo-final-test DESK_TEST_WEB_PORT=18180 DESK_TEST_API_PORT=18181 bash scripts/service-desk-test.sh
DESK_TEST_PROJECT=catalogo-final-test DESK_TEST_WEB_PORT=18180 DESK_TEST_API_PORT=18181 bash scripts/service-desk-persistence.sh
bash scripts/service-desk-demo-verify.sh
```

Exit code final de los tres comandos: **0**. El primer runner construye contenedores, espera readiness, aplica seed/migraciones, ejecuta el binario de pruebas, verifica Nginx y comprueba readiness a través del frontend. No ejecuta `down -v`. El segundo limita su destino a un proyecto terminado en `-test`, compara registros y descarga el adjunto antes y después del reinicio.

En la suite general `TestAttachmentPersistenceCheck` aparece omitida deliberadamente cuando falta su variable explícita. Eso no se registra como aprobación: el segundo comando la ejecutó antes y después del reinicio, con resultado PASS real en ambas ocasiones.

Comprobaciones locales adicionales: `go test ./backend/...`, `npm --prefix frontend test` y `npm --prefix frontend run build`, con exit code 0. Las pruebas que requieren PostgreSQL se ejecutaron en la laptop, no se dan por aprobadas por la omisión local.

## Mediciones y persistencia

La ejecución final registró 2.480418 ms para una consulta paginada de CI, y 0.137 ms de ejecución SQL en el `EXPLAIN ANALYZE` de una página ordenada. El plan usa `configuration_items_code_key` e incremental sort. Son observaciones de esa ejecución en esa laptop, sin garantía de rendimiento para otro entorno.

Último recibo P18 conservado:

| Dato | Antes | Después |
| --- | --- | --- |
| Tickets de pruebas | 120 | 120 |
| Entradas de historial | 559 | 559 |
| CI, incluidos fixtures no dispositivo | 5,044 | 5,044 |
| Clases | 27 | 27 |
| Adjuntos | 15 | 15 |
| Artículos FAQ de pruebas | 5 | 5 |

Los digest de tickets, historial, CI y metadatos de adjuntos coincidieron. Se descargó el mismo `ticket_id=96`, `attachment_id=13`, con los mismos 30 bytes. Las sesiones no forman parte del digest porque el propio control inicia sesión.

Estos 120 tickets no son los 37 de defensa. Son residuos identificados de varias ejecuciones en la base desechable. No se borró la base para producir una salida más bonita. El [recibo separado de demostración](service-desk-demo-salida.txt) comprueba el total exacto de 37 y los mismos digest tras repetir la preparación.

## Loops y evidencia de fallos

1. El [primer intento](service-desk-final-primer-intento.txt) pasó, pero el plan SQL se ejecutó sobre cero dispositivos porque la carga ocurría después. Se agregó una precondición de 5,000 dispositivos y se repitió la medición. Ese primer plan no se utiliza como evidencia de rendimiento a escala.
2. Al repetir P14 sobre la misma base, un fixture reutilizó dispositivos y relaciones existentes. La [salida del fallo](service-desk-final-fallo-p14.txt) conserva la restricción única rechazada y el exit code fallido en la ejecución. Se corrigió creando 34 CI exclusivos por ejecución, de clase no dispositivo, y se volvió a ejecutar la suite. Una repetición aislada adicional de P14 también pasó sobre la misma base sin borrar datos. Esa repetición posterior agrega sus propios fixtures y no cambia el recibo P18 ya capturado.
3. El reinicio de persistencia se ejecuta por etapas, esperando PostgreSQL antes de arrancar la API, para no confundir un fallo de orden de arranque con pérdida de datos.

## Evidencias completas

- [Construcción, tests y readiness](service-desk-final-salida.txt).
- [Persistencia antes y después](service-desk-persistencia-salida.txt).
- [Conteos, cuentas y repetición de la demostración](service-desk-demo-salida.txt).
- [Prompt y contexto de la implementación](../prompts/012-faq-demostracion-verificacion.md).
- [Guía de demostración](../../README.md#demo-completa-seis-cuentas-y-37-tickets).

En navegador se inició sesión como solicitante, se abrió una respuesta publicada y se verificó la búsqueda `acceso`. Como administrador se abrió la edición de un artículo y se comprobó que cargara sus datos, cancelando sin modificar el contenido. También se abrió la consola con tickets `[DEMO]`. Las capturas son reales, no reconstrucciones de consola ni maquetas.

![FAQ publicada en el navegador](../Imagenes/service-desk-faq.png)

## Integridad de las fuentes

SHA-256 comprobados localmente, sin cambios respecto de los originales:

| Archivo | SHA-256 |
| --- | --- |
| `data/CatalogoServicios.xlsx` | `DE3B478A5FAEEEAEBCE1AA7726E0E3321188A68E41BBB656E1D17B0C5B74DCF0` |
| `data/fuentes-service-desk/CMDB.xlsx` | `D182796BB40B492BE2CB87E2C5FE19A8F910F2E754183DC48B5E89BD360464DC` |
| `data/fuentes-service-desk/Tarea_Service_Desk.pdf` | `0A36D768F0EF305334B4F56AD65E8DDA1BE0410F6F6E636254BA0090EC29396E` |

La revisión de resultados no certifica la entrega en Wedy, fecha de push, defensa oral o commits pendientes. Esas condiciones requieren su evidencia propia.

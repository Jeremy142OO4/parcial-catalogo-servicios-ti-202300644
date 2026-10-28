# Guía breve para la defensa

Jeremy Estuardo Orellana Aldana, 202300644. Preparación del 28 de octubre de 2026.

## Antes del turno

Dejar encendida la laptop y abrir [la demostración](http://192.168.1.23:18090/). Usar las [cuentas y casos preparados](DEMOSTRACION-SERVICE-DESK.md). No ejecutar la suite sobre esa base ni crear tickets antes de mostrar el total inicial de 37.

Desde un clon del repositorio, en Bash:

```bash
git log -6 --oneline
docker compose -p catalogo-service-desk-demo -f compose.desk-demo.yaml up -d
bash scripts/service-desk-demo-verify.sh
```

El último comando muestra siete colas, 37 tickets, 5,000 dispositivos, ocho artículos, distribución por proceso y las cuentas. Es de solo lectura. Si aún no se preparó la base, seguir la guía de demostración en lugar de asumir que el arranque crea usuarios y tickets.

## Si pide el harness

Mostrar el contenido de `scripts/service-desk-test.sh` y ejecutar en un entorno aislado:

```bash
DESK_TEST_PROJECT=catalogo-final-test DESK_TEST_WEB_PORT=18180 DESK_TEST_API_PORT=18181 bash scripts/service-desk-test.sh
DESK_TEST_PROJECT=catalogo-final-test DESK_TEST_WEB_PORT=18180 DESK_TEST_API_PORT=18181 bash scripts/service-desk-persistence.sh
```

Explicación: «El harness construye la aplicación, verifica originales, aplica migraciones, importa el catálogo y prepara datos aislados. Ejecuta pruebas de permisos, concurrencia y los siete flujos, con código de salida ante errores. Otro control reinicia la base y API y compara registros y un adjunto. Docker es el entorno de ejecución, no toda la validación».

No ejecutar `down -v`. No usar el proyecto de demostración como valor de `DESK_TEST_PROJECT`. Las pruebas PostgreSQL se omiten localmente sin su variable, pero eso no es un PASS de integración. Mostrar la [salida real](evidencias/service-desk-cierre-salida.txt), no solo una captura de un comando.

## Si selecciona una funcionalidad

| Funcionalidad | Recorrido breve | Regla que debes explicar |
| --- | --- | --- |
| Tickets | Consola, abrir un caso y ver historial | Nivel 1 derivado del servicio nivel 2 y acciones transaccionales |
| Colas y roles | Colas y permisos, seleccionar cola y sujeto | Grants por rol o usuario, consulta no puede escribir |
| Cambios | `SD-00000015` pendiente y `SD-00000016` en ejecución | No ejecutar sin aprobación y conservar riesgo, pruebas y reversión |
| Accesos | `SD-00000030` pendiente o `SD-00000032` rechazado | Solicitante y aprobador distintos, incluso con rol administrador |
| Incidentes | `SD-00000004` pausado y `SD-00000006` reabierto | I0–I6, duración desconocida no es cero y ciclos preservados |
| Eventos | `SD-00000020` o `SD-00000021` | E2/E3 tienen un incidente vinculado y notificación sin duplicados |
| Problemas | `SD-00000025` con causa y `SD-00000027` cerrado | Causa, alternativa temporal y solución definitiva, sin KEDB independiente |
| Nuevos servicios | `SD-00000010` pendiente | Tras aprobar se puede vincular un servicio existente sin duplicar código |
| Servicios existentes | `SD-00000035` asignado o `SD-00000037` cerrado | Cumplimiento y confirmación, distinto de incidentes o altas del catálogo |
| FAQ | Centro de ayuda, buscar `acceso` | Publicación explícita, sesión y edición administrativa con versión |
| CMDB | Abrir un CI y sus relaciones | Clasificación no es organización, dependencia no implica interrupción |
| Fuente CMDB | CMDB, Origen del Excel | Se conservan dos filas y 40 valores, no se inventan responsables o CI faltantes |
| Indicadores | Tablero y detalle de ciclo | Tiempo calendario, pausas sin doble resta y objetivos históricos |

No realices una acción de escritura durante la defensa si primero necesitas conservar el conjunto inicial intacto. Puedes explicar con casos ya preparados y ejecutar las pruebas negativas en el entorno aislado.

## Si pregunta por IA

Mostrar `docs/IA.md` para el índice. Elegir una mejora de prompt con su salida real, una actualización de contexto con su commit y el loop del trigger de colas que fallaba al resolver un campo de usuario. Explicar la corrección y abrir la salida de la nueva prueba.

Para grafos, distinguir el árbol de clasificación, donde no se aceptan ciclos, del grafo técnico. `depende_de` y `alojado_en` rechazan ciclos y permiten analizar impacto potencial. `conectado_a` es simétrica y admite ciclos. El recorrido tiene límites declarados y no demuestra una caída real del servicio.

## Qué no debes afirmar

No atribuir las seis subclases técnicas a las fuentes docentes. No afirmar que P04 está completamente resuelto mientras siga sin confirmarse la correspondencia exacta. No presentar una captura, un SKIP o una propuesta de ChatGPT como resultado de una prueba. No asegurar la entrega en Wedy sin haberla realizado y comprobado personalmente.

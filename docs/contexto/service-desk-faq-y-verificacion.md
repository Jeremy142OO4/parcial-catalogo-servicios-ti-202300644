# Contexto de FAQ, demostración y verificación

Fecha: 28 de octubre de 2026. Solicitud: preparar la FAQ, cuentas por rol, siete colas, 37 tickets y verificar P01–P18.

Nota posterior: la reserva de P04 descrita en este registro histórico se resolvió documentalmente en el alcance acordado con Jeremy. Ver [decisión y salvedad interpretativa](../decisiones/002-interpretacion-cmdb-p04.md). No se alteran los resultados obtenidos antes de la aclaración.

## Decisiones

- Separar la demostración en proyecto, base, volúmenes y puertos propios. Los tests pueden crear fixtures sin alterar los 37 tickets preparados.
- FAQ de texto con sesión, publicación explícita, baja lógica, versión optimista e historial inmutable.
- Preparar tickets mediante los mismos servicios y transiciones del servidor, con aprobador diferente del solicitante. No simular estados con actualizaciones SQL en la demostración.
- Identificar casos por dataset y clave estable, conservando los que ya terminaron su preparación. Los detalles son sintéticos declarados, no valores inventados de Excel.
- Mantener P04 parcial hasta confirmar la taxonomía docente exacta. Los ciclos sí tienen pruebas reales, incluso concurrentes.

## Resultados y límites

La [guía de demostración](../DEMOSTRACION-SERVICE-DESK.md) contiene cuentas, flujo y recibo 37/7/5000/8. La [verificación final](../evidencias/service-desk-final.md) recoge resultados de cada ID, mediciones y reinicio. Se conservaron dos observaciones de mejora: medición SQL prematura y fixture de relaciones no repetible. La salida fallida no se sustituyó por texto inventado.

Esta actualización existe en el worktree y aún no constituye un commit adicional. No se modifica el tag histórico ni se afirma haber hecho push en esta fase.

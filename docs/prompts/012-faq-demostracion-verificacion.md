# FAQ, datos de demostración y verificación final

## Solicitud recibida

Mensaje literal del usuario, 28 de octubre de 2026:

> esto ahora:
>
> 1. **FAQ y datos de demostración:** cuentas por rol, siete colas y 37 tickets preparados.
> 2. **Verificación final:** completar P01–P18 y documentar resultados reales.

## Instrucción operativa de esta fase

Esta formulación documenta cómo se estructuró el trabajo en Codex. No se presenta como un mensaje enviado a ChatGPT web.

> Actúa como un ingeniero de software responsable de una aplicación Service Desk en Go, React y PostgreSQL, con atención a la seguridad y a pruebas reproducibles.
>
> Implementa una FAQ y prepara cuentas por rol, siete colas y 37 tickets de demostración. Amplía la verificación P01–P18 y registra resultados obtenidos, no resultados supuestos.
>
> La aplicación extiende el catálogo del parcial anterior. Hay trabajo local sin commit y la laptop aloja las pruebas. Deben conservarse el tag histórico, los archivos originales y las bases existentes.
>
> Usa el enunciado conservado, el modelo existente, los siete flujos, los tests y la matriz de pruebas como entrada. La demostración debe estar separada de las pruebas que crean registros temporales.
>
> Entrega implementación, comandos repetibles, recibos de conteos, matriz de resultados, limitaciones y capturas reales de la FAQ y del conjunto preparado.
>
> Acepta el resultado solo si hay cinco tickets por proceso y siete incidentes I0–I6, las escrituras de consulta se rechazan, repetir la preparación no duplica tickets y la persistencia se demuestra con un reinicio real. Un requisito sin fuente o sin prueba suficiente debe quedar explícitamente parcial.

## Resultado y revisión

- Migración 008 con FAQ, historial protegido y mapeo estable del conjunto de demostración.
- Interfaz con búsqueda de ayuda y administración protegida por servidor.
- Entorno nuevo de demostración con 37 tickets, seis cuentas específicas, siete colas, ocho artículos y 5,000 dispositivos.
- Se detectó que el primer plan SQL se medía antes de generar los dispositivos. Se corrigió el orden y se repitió la medición con 5,000 dispositivos activos. La salida inicial se conserva para mostrar la diferencia.
- Una repetición de P14 falló porque el fixture reutilizaba relaciones. Se conservaron la salida y el fallo, se crearon CI exclusivos por ejecución y se comprobó otra repetición satisfactoria sin limpiar la base.
- Se añadieron pruebas de sesiones inactivas, importación repetida, FAQ, carrera con auditoría única, eventos concurrentes, ciclos de clasificación, impacto y rollback ante fallo de auditoría.
- Se comprobó P18 reiniciando realmente PostgreSQL y la API, comparando registros y descargando el mismo adjunto.
- P04 no se declara completo sin confirmar la taxonomía docente exacta.

Evidencias: [demostración](../DEMOSTRACION-SERVICE-DESK.md), [resultado final](../evidencias/service-desk-final.md), [primer intento](../evidencias/service-desk-final-primer-intento.txt) y [FAQ capturada](../Imagenes/service-desk-faq.png).

Herramienta utilizada: Codex desktop. No se atribuye a esta ejecución un modelo o nivel de razonamiento que no esté confirmado por la sesión. No se realizó una conversación nueva en ChatGPT web para esta fase.
